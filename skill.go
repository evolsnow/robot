package main

import (
	"fmt"
	"github.com/evolsnow/robot/conn"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

//zmz.tv needs to login before downloading
var zmzClient http.Client

func (rb *Robot) Start(update tgbotapi.Update) string {
	user := update.Message.Chat.UserName
	go conn.SetUserChatId(user, update.Message.Chat.ID)
	return "welcome: " + user
}

func (rb *Robot) Help(update tgbotapi.Update) string {
	helpMsg := `
/alarm - set a reminder
/allmemos - get all of your memos
/evolve	- self evolution of me
/memo  save a memo
/movie - find movie download links
/show - find American show download links
/trans - translate words between english and chinese
/help - show this help message
`
	return helpMsg
}

func (rb *Robot) Evolve(update tgbotapi.Update) {
	if update.Message.Chat.FirstName != "Evol" || update.Message.Chat.LastName != "Gan" {
		rb.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "sorry, unauthorized"))
		return
	}
	<-saidGoodBye
	close(saidGoodBye)
	cmd := exec.Command("bash", "/root/evolve_"+rb.nickName)
	cmd.Start()
	os.Exit(1)
}

func (rb *Robot) Translate(update tgbotapi.Update) string {
	var info string
	if string(update.Message.Text[0]) == "/" {
		raw := strings.SplitAfterN(update.Message.Text, " ", 2)
		if len(raw) < 2 {
			return "what do you want to translate, try '/trans cat'?"
		} else {
			info = "翻译" + raw[1]
		}
	} else {
		info = update.Message.Text
	}

	return qinAI(info)

}
func (rb *Robot) Talk(update tgbotapi.Update) string {
	info := update.Message.Text
	if strings.Contains(info, rb.name) {
		if strings.Contains(info, "闭嘴") || strings.Contains(info, "别说话") {
			rb.shutUp = true
		} else if rb.shutUp && strings.Contains(info, "说话") {
			rb.shutUp = false
			return fmt.Sprintf("%s终于可以说话啦", rb.nickName)
		}
		info = strings.Replace(info, fmt.Sprintf("@%s", rb.name), "", -1)
	}

	if rb.shutUp {
		return ""
	}
	log.Printf(info)

	if rb.nickName == "samaritan" {
		if chinese(info) {
			return tlAI(info)
		} else {
			return mitAI(info)
		}
	} else { //jarvis use another AI
		return qinAI(info)
	}
	//	return response
}

func (rb *Robot) SetReminder(update tgbotapi.Update, step int) string {
	user := update.Message.Chat.UserName
	switch step {
	case 0:
		//known issue of go, you can not just assign update.Message.Chat.ID to userTask[user].ChatId
		tmpTask := userTask[user]
		tmpTask.ChatId = update.Message.Chat.ID
		tmpTask.Owner = update.Message.Chat.UserName
		userTask[user] = tmpTask

		tmpAction := userAction[user]
		tmpAction.ActionStep++
		userAction[user] = tmpAction
		return "Ok, what should I remind you to do?"
	case 1:
		//save thing
		tmpTask := userTask[user]
		tmpTask.Desc = update.Message.Text
		userTask[user] = tmpTask

		tmpAction := userAction[user]
		tmpAction.ActionStep++
		userAction[user] = tmpAction
		return "When or how much time after?\n" +
			"You can type:\n" +
			"'*2/14 11:30*' means 11:30 at 2/14 \n" + //first format
			"'*11:30*' means  11:30 today\n" + //second format
			"'*5m10s*' means 5 minutes 10 seconds later" //third format
	case 2:
		defer delete(userAction, user)
		//save time duration
		text := update.Message.Text
		text = strings.Replace(text, "：", ":", -1)
		firstFormat := "1/02 15:04"
		secondFormat := "15:04"
		thirdFormat := "15:04:05"
		var showTime string
		var scheduledTime time.Time
		var nowTime = time.Now()
		var du time.Duration
		var err error
		if strings.Contains(text, ":") {
			scheduledTime, err = time.Parse(firstFormat, text)
			nowTime, _ = time.Parse(firstFormat, nowTime.Format(firstFormat))
			showTime = scheduledTime.Format(firstFormat)
			if err != nil { //try to parse with first format
				scheduledTime, err = time.Parse(secondFormat, text)
				nowTime, _ = time.Parse(secondFormat, nowTime.Format(secondFormat))
				showTime = scheduledTime.Format(secondFormat)
				if err != nil {
					return "wrong format, try '2/14 11:30' or '11:30'?"
				}
			}
			du = scheduledTime.Sub(nowTime)
		} else {

			du, err = time.ParseDuration(text)
			scheduledTime = nowTime.Add(du)
			showTime = scheduledTime.Format(thirdFormat)
			if err != nil {
				return "wrong format, try '1h2m3s'?"
			}
		}
		//		tmpTask := userTask[user]
		//		tmpTask.When = scheduledTime
		//		userTask[user] = tmpTask
		go func(rb *Robot, ts Task) {
			timer := time.NewTimer(du)
			rawMsg := fmt.Sprintf("Hi %s, maybe it's time to:\n*%s*", ts.Owner, ts.Desc)
			msg := tgbotapi.NewMessage(ts.ChatId, rawMsg)
			msg.ParseMode = "markdown"
			<-timer.C
			_, err := rb.bot.Send(msg)
			if err != nil {
				rb.bot.Send(tgbotapi.NewMessage(conn.GetUserChatId(ts.Owner), rawMsg))
			}
			delete(userTask, user)
		}(rb, userTask[user])

		//		delete(userAction, user)
		return fmt.Sprintf("Ok, I will remind you that\n*%s* - *%s*", showTime, userTask[user].Desc)
	}
	return ""
}

func (rb *Robot) DownloadMovie(update tgbotapi.Update, step int, results chan string) (ret string) {
	user := update.Message.Chat.UserName
	switch step {
	case 0:
		//known issue of go, you can not just assign update.Message.Chat.ID to userTask[user].ChatId
		tmpAction := userAction[user]
		tmpAction.ActionStep++
		userAction[user] = tmpAction
		ret = "Ok, which movie do you want to download?"
	case 1:
		defer func() {
			delete(userAction, user)
			results <- "done"
		}()
		results <- "Searching movie..."
		movie := update.Message.Text
		var wg sync.WaitGroup
		wg.Add(2)
		go getMovieFromZMZ(movie, results, &wg)
		go getMovieFromLBL(movie, results, &wg)
		wg.Wait()
	}
	return
}

func (rb *Robot) DownloadShow(update tgbotapi.Update, step int, results chan string) (ret string) {
	user := update.Message.Chat.UserName
	switch step {
	case 0:
		//known issue of go, you can not just assign update.Message.Chat.ID to userTask[user].ChatId
		tmpAction := userAction[user]
		tmpAction.ActionStep++
		userAction[user] = tmpAction
		ret = "Ok, which American show do you want to download?"
	case 1:
		results <- "Searching American show..."
		info := strings.Split(update.Message.Text, " ")
		if len(info) < 3 {
			results <- "Please specify the season and episode,like:\n*疑犯追踪 1 10*"
			return
		}
		GetShowFromZMZ(info[0], info[1], info[2], results)
		delete(userAction, user)
		results <- "done"
	}
	return
}

func (rb *Robot) SaveMemo(update tgbotapi.Update, step int) (ret string) {
	user := update.Message.Chat.UserName
	switch step {
	case 0:
		//known issue of go, you can not just assign update.Message.Chat.ID to userTask[user].ChatId
		tmpAction := userAction[user]
		tmpAction.ActionStep++
		userAction[user] = tmpAction
		ret = "Ok, what do you want to save?"
	case 1:
		defer delete(userAction, user)
		time := time.Now().Format("2006-1-02 15:04")
		memo := update.Message.Text
		go conn.HSetMemo(user, time, memo)
		ret = "Ok, type '/allmemos' to see all your memos"
	}
	return
}

func (rb *Robot) GetAllMemos(update tgbotapi.Update) (ret string) {
	user := update.Message.Chat.UserName
	//	beforeParse := conn.HGetAllMemos(user)
	//	memos := make([]map[string]string, len(beforeParse))
	memos := conn.HGetAllMemos(user)
	if len(*memos) == 0 {
		return "You have no memo now, type '/memo' to save one?"
	}
	//	for i, before := range beforeParse {
	//		memos[i] = before.(map[string]string)
	//	}
	for time, content := range *memos {
		ret += fmt.Sprintf("%s:%s", time, content)
	}
	return
}

func getMovieFromLBL(movie string, results chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	var id string
	resp, _ := http.Get("http://www.lbldy.com/search/" + movie)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	re, _ := regexp.Compile("<div class=\"postlist\" id=\"post-(.*?)\">")
	//find first match case
	firstId := re.FindSubmatch(body)
	if len(firstId) == 0 {
		results <- fmt.Sprintf("No results for *%s* from LBL", movie)
		return
	} else {
		id = string(firstId[1])
		resp, _ = http.Get("http://www.lbldy.com/movie/" + id + ".html")
		defer resp.Body.Close()
		re, _ = regexp.Compile(`<p><a href="(.*?)"( target="_blank">|>)(.*?)</a></p>`)
		body, _ := ioutil.ReadAll(resp.Body)
		//go does not support (?!) regex
		body = []byte(strings.Replace(string(body), `<a href="/xunlei/"`, "", -1))
		downloads := re.FindAllSubmatch(body, -1)
		if len(downloads) == 0 {
			results <- fmt.Sprintf("No results for *%s* from LBL", movie)
			return
		} else {
			ret := "Results from LBL:\n\n"
			for i := range downloads {
				ret += fmt.Sprintf("*%s*\n```%s```\n\n", string(downloads[i][3]), string(downloads[i][1]))
				if i%5 == 0 && i > 0 {
					results <- ret
					ret = fmt.Sprintf("*LBL Part %d*\n\n", i/5+1)
				}
			}
			results <- ret
		}
	}
}

func getMovieFromZMZ(movie string, results chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	loginZMZ()
	if downloads := getZMZResource(movie); downloads == nil {
		results <- fmt.Sprintf("No results for *%s* from ZMZ", movie)
		return
	} else {
		ret := "Results from ZMZ:\n\n"
		for i := range downloads {
			name := string(downloads[i][1])
			size := string(downloads[i][2])
			link := string(downloads[i][3])
			ret += fmt.Sprintf("*%s*(%s)\n```%s```\n\n", name, size, link)
			if i%5 == 0 && i > 0 {
				results <- ret
				ret = fmt.Sprintf("*ZMZ Part %d*\n\n", i/5+1)
			}
		}
		results <- ret
	}
	return
}

func GetShowFromZMZ(show, s, e string, results chan string) {
	loginZMZ()
	downloads := getZMZResource(show)
	if downloads == nil {
		results <- fmt.Sprintf("No results for *%s* from ZMZ", show)
		return
	}
	//second parse
	re, _ := regexp.Compile(fmt.Sprintf(".*?season=\"%s\" episode=\"%s\">.*?", s, e))
	results <- "Results from ZMZ:\n\n"
	count := 0
	for i := range downloads {
		if re.Find(downloads[i][0]) != nil {
			name := string(downloads[i][1])
			size := string(downloads[i][2])
			link := string(downloads[i][3])
			results <- fmt.Sprintf("*ZMZ %s*(%s)\n```%s```\n\n", name, size, link)
			count++
		}
	}
	if count == 0 {
		results <- fmt.Sprintf("No results found for *S%sE%s*", s, e)

	}
	return
}

func getZMZResource(name string) [][][]byte {
	id := getZMZResourceId(name)
	if id == "" {
		return nil
	}
	resourceURL := "http://www.zimuzu.tv/resource/list/" + id
	resp, _ := zmzClient.Get(resourceURL)
	defer resp.Body.Close()
	//1.name 2.size 3.link
	re, _ := regexp.Compile(`<li class="clearfix".*?<input type="checkbox"><a title="(.*?)".*?<font class="f3">(.*?)</font>.*?<a href="(.*?)" type="ed2k">`)
	body, _ := ioutil.ReadAll(resp.Body)
	body = []byte(strings.Replace(string(body), "\n", "", -1))
	//	tmp := (strings.Replace(string(body), "</div>\n", "", -1))
	//	body = []byte(strings.Replace(tmp, "<div class=\"fr\">\n", "", -1))
	downloads := re.FindAllSubmatch(body, -1)
	if len(downloads) == 0 {
		return nil
	}
	return downloads
}

func getZMZResourceId(name string) (id string) {
	queryURL := fmt.Sprintf("http://www.zimuzu.tv/search?keyword=%s&type=resource", name)
	re, _ := regexp.Compile(`<div class="t f14"><a href="/resource/(.*?)"><strong class="list_title">`)
	resp, _ := zmzClient.Get(queryURL)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	//find first match case
	firstId := re.FindSubmatch(body)
	if len(firstId) == 0 {
		return
	} else {
		log.Println(id)
		id = string(firstId[1])
		return
	}
}

func loginZMZ() {
	gCookieJar, _ := cookiejar.New(nil)
	zmzURL := "http://www.zimuzu.tv/User/Login/ajaxLogin"
	zmzClient = http.Client{
		Jar: gCookieJar,
	}
	zmzClient.PostForm(zmzURL, url.Values{"account": {"evol4snow"}, "password": {"104545"}, "remember": {"0"}})
}

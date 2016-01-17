package main

import (
	"fmt"
	"github.com/evolsnow/robot/conn"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

func (rb *Robot) Start(update tgbotapi.Update) string {
	user := update.Message.Chat.UserName
	go conn.SetUserChatId(user, update.Message.Chat.ID)
	return "welcome: " + user
}

func (rb *Robot) Help(update tgbotapi.Update) string {
	helpMsg := `
/alarm - set a reminder
/trans - translate words between english and chinese
/evolve	- self evolution of me
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
		results <- "Searching..."
		movie := update.Message.Text
		var wg sync.WaitGroup
		wg.Add(2)
		go getMovieFromZmz(movie, results, &wg)
		go getMovieFromLbl(movie, results, &wg)
		wg.Wait()
	}
	return
}

func getMovieFromZmz(movie string, results chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	return
}

func getMovieFromLbl(movie string, results chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	var id string
	resp, _ := http.Get("http://www.lbldy.com/search/" + movie)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	re, _ := regexp.Compile("<div class=\"postlist\" id=\"post-(.*?)\">")
	//find first match case
	firstId := re.FindSubmatch(body)
	if len(firstId) == 0 {
		results <- fmt.Sprintf("no answer for *%s* from lbl", movie)
		return
	} else {
		id = string(firstId[1])
		resp, _ = http.Get("http://www.lbldy.com/movie/" + id + ".html")
		defer resp.Body.Close()
		//		re, _ = regexp.Compile("<p><a href=\"(.*?)\" target=\"_blank\">(.*?)</a></p>")
		re, _ = regexp.Compile("<p><a href=\"(.*?)\"(| target=\"_blank)\">(.*?)</a></p>")

		body, _ = ioutil.ReadAll(resp.Body)
		downloads := re.FindAllSubmatch(body, -1)
		if len(downloads) == 0 {
			results <- fmt.Sprintf("no answer for *%s* from lbl", movie)
			return
		} else {
			results <- "Results from lbl:\n\n"
			ret := ""
			for i := range downloads {
				ret += string(downloads[i][2]) + "\n" + string(downloads[i][1]) + "\n\n"
			}
			results <- ret
		}
	}
}

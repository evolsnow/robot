package main

import (
	"encoding/json"
	"fmt"
	"github.com/evolsnow/robot/conn"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"strings"
	"time"
	"unicode"
)

var saidGoodBye = make(chan int, 1)
var userAction = make(map[string]Action) //map[user]Action
var userTask = make(map[string]Task)

type Robot struct {
	bot     *tgbotapi.BotAPI
	updates <-chan tgbotapi.Update
	shutUp  bool
	//	language []string
	name     string //name from telegram
	nickName string //user defined name
}

type Action struct {
	ActionName string
	ActionStep int
}

type Task struct {
	ChatId   int
	Owner    string
	Desc     string
	Duration time.Duration
}

func newRobot(token, nickName, webHook string) *Robot {
	var rb = new(Robot)
	var err error
	rb.bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}
	rb.name = rb.bot.Self.UserName
	rb.nickName = nickName
	log.Printf("%s: Authorized on account %s", rb.nickName, rb.name)
	_, err = rb.bot.SetWebhook(tgbotapi.NewWebhook(webHook + rb.bot.Token))
	if err != nil {
		log.Fatal(err)
	}
	rb.updates, _ = rb.bot.ListenForWebhook("/" + rb.bot.Token)
	return rb
}

func (rb *Robot) run() {
	if rb.nickName == "samaritan" {
		chatId := conn.GetMasterId()
		msg := tgbotapi.NewMessage(chatId, "samaritan is coming back!")
		if _, err := rb.bot.Send(msg); err != nil {
			log.Println("evolution failed")
		}
	}
	for update := range rb.updates {
		go handlerUpdate(rb, update)
	}
}

func handlerUpdate(rb *Robot, update tgbotapi.Update) {
	defer func() {
		if p := recover(); p != nil {
			err := fmt.Errorf("internal error: %v", p)
			log.Println(err)
		}
	}()

	user := update.Message.Chat.UserName + ":" + rb.nickName
	text := update.Message.Text
	chatId := update.Message.Chat.ID
	var endPoint, rawMsg string
	if action, ok := userAction[user]; ok { //detect if user is in interaction mode
		switch action.ActionName {
		case "setReminder":
			rawMsg = rb.SetReminder(update, action.ActionStep)
		}
	} else if string(text[0]) == "/" {
		received := strings.Split(text, " ")
		endPoint = received[0]
		log.Println(endPoint)
		switch endPoint {
		case "/start":
			rawMsg = rb.Start(update)
		case "/help":
			rawMsg = rb.Help(update)
		case "/trans":
			rawMsg = rb.Translate(update)
		case "/alarm":
			tmpAction := userAction[user]
			tmpAction.ActionName = "setReminder"
			userAction[user] = tmpAction
			rawMsg = rb.SetReminder(update, 0)
		case "/evolve":
			rawMsg = "upgrading..."
			go conn.SetMasterId(chatId)
			go rb.Evolve(update)
		default:
			rawMsg = "unknow command, type /help?"
		}
	} else {
		rawMsg = rb.Talk(update)
	}
	if rawMsg == "" {
		return
	}
	msg := tgbotapi.NewMessage(chatId, rawMsg)
	msg.ParseMode = "markdown"
	log.Println(rawMsg)
	_, err := rb.bot.Send(msg)
	if endPoint == "/evolve" {
		saidGoodBye <- 1
	}
	if err != nil {
		panic(err)
	}

}

////parse "/help text msg" to "text msg"
//func parseText(text string) string {
//	return strings.SplitAfterN(text, " ", 2)[1]
//}

func (rb *Robot) Start(update tgbotapi.Update) string {
	user := update.Message.Chat.UserName
	go conn.SetUserChatId(user+":"+rb.nickName, update.Message.Chat.ID)
	return "welcome: " + user
}

func (rb *Robot) Help(update tgbotapi.Update) string {
	helpMsg := `
/alarm - set a reminder
/trans - translate words between english and chinese
/evolve	- self evolution of samaritan
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
	cmd := exec.Command("bash", "/root/evolve")
	if err := cmd.Start(); err != nil {
		log.Println(err.Error())
	}
}

func (rb *Robot) Translate(update tgbotapi.Update) string {
	raw := strings.SplitAfterN(update.Message.Text, " ", 2)
	info := ""
	if len(raw) < 2 {
		return "what do you want to translate, try '/trans cat'?"
	} else {
		info = "翻译" + raw[1]
	}
	return qinAI(info)

}
func (rb *Robot) Talk(update tgbotapi.Update) string {
	info := update.Message.Text
	chinese := false
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
	log.Println(info)
	//	var response string
	for _, r := range info {
		if unicode.Is(unicode.Scripts["Han"], r) {
			chinese = true
			break
		}
	}
	if rb.nickName == "samaritan" {
		if chinese {
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
	user := update.Message.Chat.UserName + ":" + rb.nickName
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
			"You can type '*2/14 11:30*' means 11:30 at 2/14 \n" + //first format
			"Or type '*11:30*' means at 11:30 today\n" + //second format
			"Or type '*5m10s*' means 5 minutes 10 seconds later" //third format
	case 2:
		//save time duration
		text := update.Message.Text
		firstFormat := "1/02 15:04"
		secondFormat := "15:04"
		var scheduledTime time.Time
		var nowTime = time.Now()
		var du time.Duration
		var err error
		if strings.Contains(text, ":") {
			scheduledTime, err = time.Parse(firstFormat, text)
			nowTime, _ = time.Parse(firstFormat, nowTime.Format(firstFormat))
			if err != nil { //try to parse with first format
				scheduledTime, err = time.Parse(secondFormat, text)
				nowTime, _ = time.Parse(secondFormat, nowTime.Format(secondFormat))
				if err != nil {
					return "wrong format, try '2/14 11:30' or '11:30'?"
				}
			}
			du = scheduledTime.Sub(nowTime)
		} else {

			du, err = time.ParseDuration(text)
			if err != nil {
				return "wrong format, try '1h2m3s'?"
			}
		}
		tmpTask := userTask[user]
		tmpTask.Duration = du
		userTask[user] = tmpTask
		go func(rb *Robot, ts Task) {
			timer := time.NewTimer(du)
			<-timer.C
			rawMsg := fmt.Sprintf("Hi %s, maybe it's time to:\n*%s*", ts.Owner, ts.Desc)
			msg := tgbotapi.NewMessage(ts.ChatId, rawMsg)
			msg.ParseMode = "markdown"
			_, err := rb.bot.Send(msg)
			if err != nil {
				rb.bot.Send(tgbotapi.NewMessage(conn.GetUserChatId(ts.Owner), rawMsg))
			}

		}(rb, userTask[user])

		delete(userAction, user)
		delete(userTask, user)
		return fmt.Sprintf("Ok, I will remind you to %s later", userTask[user].Desc)
	}
	return ""
}

func tlAI(info string) string {
	info = strings.Replace(info, " ", "", -1)
	key := "a5052a22b8232be1e387ff153e823975"
	tuLingURL := fmt.Sprintf("http://www.tuling123.com/openapi/api?key=%s&info=%s", key, info)
	resp, err := http.Get(tuLingURL)
	if err != nil {
		log.Println(err.Error())
	}
	defer resp.Body.Close()
	reply := new(tlReply)
	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(reply)
	log.Printf("reply from tuling machine: %s", reply.Text+"\n"+reply.Url)
	return reply.Text + "\n" + reply.Url
}

type tlReply struct {
	code int    `json:"code"`
	Url  string `json:"url,omitempty"`
	Text string `json:"text"`
}

//func simAI(info, lc string) string {
//	info = strings.Replace(info, " ", "+", -1)
//	simURL := fmt.Sprintf("http://www.simsimi.com/requestChat?lc=%s&ft=1.0&req=%s&uid=58642449&did=0", lc, info)
//	resp, err := http.Get(simURL)
//	if err != nil {
//		log.Println(err.Error())
//	}
//	defer resp.Body.Close()
//	reply := new(simReply)
//	decoder := json.NewDecoder(resp.Body)
//	decoder.Decode(reply)
//	return strings.Replace(reply.Res.Msg, "<br>", "\n", -1)
//}
//
//type simReply struct {
//	result int `json:"code"`
//	Res    res
//}
//type res struct {
//	Msg string `json:"msg"`
//}

func qinAI(info string) string {
	info = strings.Replace(info, " ", "+", -1)
	qinURL := fmt.Sprintf("http://api.qingyunke.com/api.php?key=free&appid=0&msg=%s", info)
	resp, err := http.Get(qinURL)
	if err != nil {
		log.Println(err.Error())
	}
	defer resp.Body.Close()
	reply := new(qinReply)
	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(reply)
	log.Printf("reply from qingyunke machine: %s", reply.Content)
	ret := strings.Replace(reply.Content, "{br}", "\n", -1)
	return strings.Replace(ret, "菲菲", "Jarvis", -1)
}

type qinReply struct {
	result  int    `json:"resulte"`
	Content string `json:"content"`
}

func mitAI(info string) string {
	mitURL := "http://fiddle.pandorabots.com/pandora/talk?botid=9fa364f2fe345a10&skin=demochat"
	resp, err := http.PostForm(mitURL, url.Values{"message": {info}, "botcust2": {"d064e07d6e067535"}})
	if err != nil {
		log.Println(err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	re, _ := regexp.Compile("Mitsuku:</B>(.*?)<br> <br>")
	all := re.FindAll(body, -1)
	if len(all) == 0 {
		return "change another question?"
	}
	found := (string(all[0]))
	log.Printf("reply from mitsuku machine: %s", found)
	ret := strings.Replace(found, `<P ALIGN="CENTER"><img src="http://`, "", -1)
	ret = strings.Replace(ret, `"></img></P>`, "", -1)
	ret = strings.Replace(ret[13:], "<br>", "\n", -1)
	ret = strings.Replace(ret, "Mitsuku", "samaritan", -1)
	return ret
}

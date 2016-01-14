package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"strings"
	"unicode"
)

var saidGoodBye = make(chan int, 1)

type robot struct {
	bot     *tgbotapi.BotAPI
	updates <-chan tgbotapi.Update
	shutUp  bool
	//	language []string
	name     string //name from telegram
	nickName string //user defined name
}

func (rb *robot) run() {
	for update := range rb.updates {
		go handlerUpdate(rb, update)
	}
}

func newRobot(token, nickName string) *robot {
	var rb = new(robot)
	var err error
	rb.bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}
	rb.name = rb.bot.Self.UserName
	rb.nickName = nickName
	log.Printf("%s: Authorized on account %s", rb.nickName, rb.name)
	_, err = rb.bot.SetWebhook(tgbotapi.NewWebhook("https://www.samaritan.tech:8443/" + rb.bot.Token))
	if err != nil {
		log.Fatal(err)
	}
	rb.updates, _ = rb.bot.ListenForWebhook("/" + rb.bot.Token)
	return rb
}

func handlerUpdate(rb *robot, update tgbotapi.Update) {
	defer func() {
		if p := recover(); p != nil {
			err := fmt.Errorf("internal error: %v", p)
			log.Println(err)
		}
	}()
	text := update.Message.Text
	chatId := update.Message.Chat.ID
	var endPoint, rawMsg string
	if string(text[0]) == "/" {
		received := strings.Split(text, " ")
		endPoint = received[0]
		log.Println(endPoint)
		switch endPoint {
		case "/start":
			rawMsg = rb.Start(update)
		case "/talk":
			rawMsg = rb.Talk(update)
		case "/evolve":
			rawMsg = "self upgrading..."
			log.Println(rawMsg)
		//go rb.Evolve()
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
	_, err := rb.bot.Send(msg)
	if endPoint == "/evolve" {
		saidGoodBye <- 1
	}
	if err != nil {
		panic(err)
	}

}

func (rb *robot) Start(update tgbotapi.Update) string {
	return "welcome: " + update.Message.Chat.UserName
}

func (rb *robot) Evolve() {
	select {
	case <-saidGoodBye:
		close(saidGoodBye)
		cmd := exec.Command("/root/evolve")
		if err := cmd.Start(); err != nil {
			log.Println(err.Error())
		}
	}
}

func (rb *robot) Talk(update tgbotapi.Update) string {
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
			info = strings.Replace(info, " ", "", -1)
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

func tlAI(info string) string {
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
	return strings.Replace(reply.Text+reply.Url, "<br>", "\n", -1)
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

type simReply struct {
	result int `json:"code"`
	Res    res
}
type res struct {
	Msg string `json:"msg"`
}

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
	log.Println(found)
	ret := strings.Replace(found, `<P ALIGN="CENTER"><img src="http://`, "", -1)
	ret = strings.Replace(ret, `"></img></P>`, "", -1)
	ret = strings.Replace(ret[13:], "<br>", "\n", -1)
	ret = strings.Replace(ret, "Mitsuku", "samaritan", -1)
	return ret
}

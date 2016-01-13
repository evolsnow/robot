package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"unicode"
)

type robot struct {
	bot     *tgbotapi.BotAPI
	updates tgbotapi.Update
	shutUp  bool
	//	language []string
	name string
}

func (rb *robot) run() {
	for update := range rb.updates {
		go handlerUpdate(rb, update)
	}
}

func newRobot(token string) *robot {
	var rb = new(robot)
	var err error
	rb.bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}
	rb.bot.Debug = false
	rb.name = rb.bot.Self.UserName
	log.Printf("Authorized on account %s", rb.name)
	_, err = rb.bot.SetWebhook(tgbotapi.NewWebhook("https://www.samaritan.tech:8443/" + rb.bot.Token))
	if err != nil {
		log.Fatal(err)
	}
	rb.updates, _ = rb.bot.ListenForWebhook("/" + rb.bot.Token)
	return rb
}

//func main() {
//	//test()
//	//used for 104
//	//go http.ListenAndServeTLS("0.0.0.0:8443", "server.crt", "server.key", nil)
//
//	//	go func() {
//	//		http.Handle("/websocket", websocket.Handler(socketHandler))
//	//		log.Fatal(http.ListenAndServe("localhost:8000", nil))
//	//	}()
//	go http.ListenAndServe("0.0.0.0:8000", nil)
//	go jarvis()
//	var err error
//	bot, err = tgbotapi.NewBotAPI("164760320:AAEE0sKLgCwHGYJ0Iqz7o-GYH4jVTQZAZho")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	bot.Debug = false
//	log.Printf("Authorized on account %s", bot.Self.UserName)
//
//	//used for 104
//	//_, err = bot.SetWebhook(tgbotapi.NewWebhookWithCert("https://104.236.156.226:8443/"+bot.Token, "server.crt"))
//	_, err = bot.SetWebhook(tgbotapi.NewWebhook("https://www.samaritan.tech:8443/" + bot.Token))
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	updates, _ := bot.ListenForWebhook("/" + bot.Token)
//	for update := range updates {
//		go handlerUpdate(update)
//	}
//}

//func jarvis() {
//	jabot, err := tgbotapi.NewBotAPI("176820788:AAH26vgFIk7oWKibd7P8XHHZX2t2_2Jvke8")
//	if err != nil {
//		log.Fatal(err)
//	}
//	log.Printf("Authorized on account %s", jabot.Self.UserName)
//	_, err = jabot.SetWebhook(tgbotapi.NewWebhook("https://www.samaritan.tech:8443/" + jabot.Token))
//	if err != nil {
//		log.Fatal(err)
//	}
//	updates, _ := jabot.ListenForWebhook("/" + jabot.Token)
//	jabot.Debug = false
//	for update := range updates {
//		go func(update tgbotapi.Update) {
//			//			zh := false
//			info := update.Message.Text
//			rawMsg := ""
//			defer func() {
//				if p := recover(); p != nil {
//					err := fmt.Errorf("internal error: %v", p)
//					log.Println(err)
//				}
//			}()
//
//			if strings.Contains(info, "@SnowJarvisBot") {
//				if strings.Contains(info, "闭嘴") || strings.Contains(info, "别说话") {
//					shutUpJa = true
//				} else if strings.Contains(info, "说话") {
//					shutUpJa = false
//					rawMsg = "jarvis终于可以说话啦~"
//				}
//				info = strings.Replace(info, "@EvolsnowBot", "", -1)
//			}
//			if shutUpJa {
//				return
//			}
//			log.Println(info)
//			// SimBot
//			//			for _, r := range info {
//			//				if unicode.Is(unicode.Scripts["Han"], r) {
//			//					info = strings.Replace(info, " ", "", -1)
//			//					zh = true
//			//					break
//			//				}
//			//			}
//
//			//			if rawMsg != "" {
//			//			} else if zh {
//			//				rawMsg = simAI(info, "ch")
//			//			} else {
//			//				rawMsg = simAI(info, "en")
//			//			}
//			//			if strings.Contains(rawMsg, "I HAVE NO RESPONSE.") {
//			//				rawMsg = "ah...I don't know"
//			//			}
//			if rawMsg != "" {
//			} else {
//				rawMsg = qinAI(info)
//			}
//			msg := tgbotapi.NewMessage(update.Message.Chat.ID, rawMsg)
//			msg.ParseMode = "markdown"
//			_, err := jabot.Send(msg)
//			if err != nil {
//				panic(err)
//			}
//
//		}(update)
//	}
//}

func handlerUpdate(rb *robot, update tgbotapi.Update) {
	defer func() {
		if p := recover(); p != nil {
			err := fmt.Errorf("internal error: %v", p)
			log.Println(err)
		}
	}()
	text := update.Message.Text
	chatId := update.Message.Chat.ID
	rawMsg := ""
	funcMap := map[string]func(update tgbotapi.Update) string{
		"/start": start,
		"/talk":  talk,
	}
	if string(text[0]) == "/" {
		received := strings.Split(text, " ")
		endPoint := received[0]
		if _, ok := funcMap[endPoint]; ok {
			rawMsg = funcMap[endPoint](update)
		} else {
			rawMsg = "unknown command"
		}
	} else {
		rawMsg = talk(rb, update)
	}
	if rawMsg == "" {
		return
	}
	msg := tgbotapi.NewMessage(chatId, rawMsg)
	msg.ParseMode = "markdown"
	_, err := rb.bot.Send(msg)
	if err != nil {
		panic(err)
	}

}

func start(update tgbotapi.Update) string {
	return "welcome: " + update.Message.Chat.UserName
}

func talk(rb *robot, update tgbotapi.Update) string {
	info := update.Message.Text
	chinese := false
	if strings.Contains(info, "@EvolsnowBot") {
		if strings.Contains(info, "闭嘴") || strings.Contains(info, "别说话") {
			rb.shutUp = true
		} else if strings.Contains(info, "说话") {
			rb.shutUp = false
			return "samaritan终于可以说话啦~"
		}
		info = strings.Replace(info, "@EvolsnowBot", "", -1)
	}
	//	if text := strings.Split(info, " "); text[0] == "@EvolsnowBot" {
	//		info = strings.Join(text[1:], " ")
	//	}
	if rb.shutUp {
		return ""
	}
	log.Println(info)
	var response string
	for _, r := range info {
		if unicode.Is(unicode.Scripts["Han"], r) {
			info = strings.Replace(info, " ", "", -1)
			chinese = true
			break
		}
	}
	if chinese {
		response = tlAI(info)
	} else {
		response = mitAI(info)
	}
	return response
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

func simAI(info, lc string) string {
	info = strings.Replace(info, " ", "+", -1)
	simURL := fmt.Sprintf("http://www.simsimi.com/requestChat?lc=%s&ft=1.0&req=%s&uid=58642449&did=0", lc, info)
	resp, err := http.Get(simURL)
	if err != nil {
		log.Println(err.Error())
	}
	defer resp.Body.Close()
	reply := new(simReply)
	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(reply)
	return strings.Replace(reply.Res.Msg, "<br>", "\n", -1)
}

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

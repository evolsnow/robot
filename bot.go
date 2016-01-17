package main

import (
	"fmt"
	"github.com/evolsnow/robot/conn"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strings"
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
	ChatId int
	Owner  string
	Desc   string
	//	When   time.Time
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
	chatId := conn.GetMasterId()
	msg := tgbotapi.NewMessage(chatId, fmt.Sprintf("%s is coming back!", rb.nickName))
	rb.bot.Send(msg)
	go loginZMZ()
	for update := range rb.updates {
		go handlerUpdate(rb, update)
	}
}
func (rb *Robot) Reply(update tgbotapi.Update, rawMsg string) (err error) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, rawMsg)
	msg.ParseMode = "markdown"
	log.Printf(rawMsg)
	_, err = rb.bot.Send(msg)
	return
}
func handlerUpdate(rb *Robot, update tgbotapi.Update) {
	defer func() {
		if p := recover(); p != nil {
			err := fmt.Errorf("internal error: %v", p)
			log.Println(err)
		}
	}()
	user := update.Message.Chat.UserName
	text := update.Message.Text
	chatId := update.Message.Chat.ID
	var endPoint, rawMsg string
	if action, ok := userAction[user]; ok { //detect if user is in interaction mode
		switch action.ActionName {
		case "setReminder":
			rawMsg = rb.SetReminder(update, action.ActionStep)
		case "downloadMovie":
			var results = make(chan string)
			//			rb.Reply(update, "Searching...")
			go rb.DownloadMovie(update, action.ActionStep, results)
			for {
				select {
				case msg := <-results:
					if msg == "done" {
						return
					}
					rb.Reply(update, msg)
				}
			}
			//			rawMsg = rb.DownloadMovie(update, action.ActionStep)
		}
	} else if string([]rune(text)[:2]) == "翻译" {
		rawMsg = rb.Translate(update)
	} else if string(text[0]) == "/" {
		text = strings.Replace(text, "@"+rb.name, "", 1)
		received := strings.Split(text, " ")
		endPoint = received[0]
		log.Printf(endPoint)
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
		case "/movie":
			tmpAction := userAction[user]
			tmpAction.ActionName = "downloadMovie"
			userAction[user] = tmpAction
			rawMsg = rb.DownloadMovie(update, 0, nil)
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

	//	msg := tgbotapi.NewMessage(chatId, rawMsg)
	//	msg.ParseMode = "markdown"
	//	log.Printf(rawMsg)
	//	_, err := rb.bot.Send(msg)

	if err := rb.Reply(update, rawMsg); err != nil {
		panic(err)
	}
	if endPoint == "/evolve" {
		saidGoodBye <- 1
	}
}

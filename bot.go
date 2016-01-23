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
var userTask = make(map[string]conn.Task)

type Robot struct {
	bot      *tgbotapi.BotAPI
	updates  <-chan tgbotapi.Update
	shutUp   bool   //shut up the robot
	name     string //name from telegram
	nickName string //user defined name
}

type Action struct {
	ActionName string
	ActionStep int
}

func newRobot(token, nickName, webHook string) *Robot {
	var rb = new(Robot)
	var err error
	rb.bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}
	rb.name = rb.bot.Self.UserName //name from telegram
	rb.nickName = nickName         //name from yourself
	log.Printf("%s: Authorized on account %s", rb.nickName, rb.name)
	_, err = rb.bot.SetWebhook(tgbotapi.NewWebhook(webHook + rb.bot.Token))
	if err != nil {
		log.Fatal(err)
	}
	rb.updates, _ = rb.bot.ListenForWebhook("/" + rb.bot.Token)
	return rb
}

func (rb *Robot) run() {
	chatId := conn.ReadMasterId()
	rawMsg := fmt.Sprintf("%s is coming back!", rb.nickName)
	rb.Reply(chatId, rawMsg)
	//go loginZMZ()
	//reload tasks from redis
	go restoreTasks(rb)
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
	user := update.Message.Chat.UserName
	text := update.Message.Text
	chatId := update.Message.Chat.ID
	var endPoint, rawMsg string
	if action, ok := userAction[user]; ok { //detect if user is in interaction mode
		switch action.ActionName {
		case "setReminder":
			rawMsg = rb.SetReminder(update, action.ActionStep)
		case "saveMemo":
			rawMsg = rb.SaveMemo(update, action.ActionStep)
		case "removeMemo":
			rawMsg = rb.RemoveMemo(update, action.ActionStep)
		case "downloadMovie":
			results := make(chan string, 2)
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
		case "downloadShow":
			results := make(chan string, 5)
			go rb.DownloadShow(update, action.ActionStep, results)
			for {
				select {
				case msg := <-results:
					if msg == "done" {
						return
					}
					rb.Reply(update, msg)
				}

			}
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
		case "/alarms":
			rawMsg = rb.GetTasks(update)
		case "/memos":
			rawMsg = rb.GetAllMemos(update)
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
		case "/memo":
			tmpAction := userAction[user]
			tmpAction.ActionName = "saveMemo"
			userAction[user] = tmpAction
			rawMsg = rb.SaveMemo(update, 0)
		case "/removememo":
			tmpAction := userAction[user]
			tmpAction.ActionName = "removeMemo"
			userAction[user] = tmpAction
			rawMsg = rb.RemoveMemo(update, 0)
		case "/show":
			tmpAction := userAction[user]
			tmpAction.ActionName = "downloadShow"
			userAction[user] = tmpAction
			rawMsg = rb.DownloadShow(update, 0, nil)
		case "/evolve":
			rawMsg = "upgrading..."
			go conn.CreateMasterId(chatId)
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

	if err := rb.Reply(update, rawMsg); err != nil {
		panic(err)
	}
	if endPoint == "/evolve" {
		saidGoodBye <- 1
	}
}

func restoreTasks(rb *Robot) {
	tasks := conn.ReadAllTasks()
	log.Println("unfinished tasks:", len(tasks))
	for i := range tasks {
		go rb.DoTask(tasks[i])
	}
}

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/evolsnow/robot/conn"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	FirstFormat  = "1/2 15:04"
	SecondFormat = "15:04"
	ThirdFormat  = "15:04:05"
	RedisFormat  = "1/2 15:04:05" //save to redis format
)

var saidGoodBye = make(chan int, 1)

var abortTask = make(map[int]chan int)
var userAction = make(map[string]Action) //map[user]Action
var userTask = make(map[string]conn.Task)
var userTaskIds = make(map[string][]int)

// Robot is a bot carried with additional information
type Robot struct {
	bot      *tgbotapi.BotAPI
	updates  <-chan tgbotapi.Update //update msg
	shutUp   bool                   //shut up the robot
	name     string                 //name from telegram
	nickName string                 //user defined name
}

// Action used in interaction mode
type Action struct {
	ActionName string
	ActionStep int
}

//return a initialized robot
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
	rb.updates = rb.bot.ListenForWebhook("/" + rb.bot.Token)
	return rb
}

//robot run and handle update msg
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

// Reply is encapsulated robot message send action
func (rb *Robot) Reply(v interface{}, rawMsg string) (err error) {
	var chatId int64
	switch v.(type) {
	case tgbotapi.Update:
		chatId = v.(tgbotapi.Update).Message.Chat.ID
	case int64:
		chatId = v.(int64)
	}
	msg := tgbotapi.NewMessage(chatId, rawMsg)
	msg.ParseMode = "markdown"
	log.Printf(rawMsg)
	_, err = rb.bot.Send(msg)
	return
}

// Start is command '/start'
func (rb *Robot) Start(update tgbotapi.Update) string {
	user := update.Message.Chat.UserName
	go conn.CreateUserChatId(user, update.Message.Chat.ID)
	return fmt.Sprintf("welcome: %s.\nType '/help' see what can I do.", user)
}

// Help show help message
func (rb *Robot) Help(update tgbotapi.Update) string {
	helpMsg := `
/alarm - set a reminder
/alarms - show all of your alarms
/rmalarm - remove alarm
/memo - save a memo
/memos - show all of your memos
/rmmemo - remove memo
/movie - find movie download links
/show - find American show download links
/trans - translate words between english and chinese
/exit - exit any interaction mode
/help - show this help message
`
	return helpMsg
}

// Repeat is markdown test function
func (rb *Robot) Repeat(update tgbotapi.Update) {
	rb.Reply(update, update.Message.Text)
}

// Evolve remote executes self evolve script, exit the robot, only for test.
func (rb *Robot) Evolve(update tgbotapi.Update) {
	if update.Message.Chat.FirstName != "Evol" || update.Message.Chat.LastName != "Gan" {
		rb.Reply(update, "sorry, unauthorized")
		return
	}
	conn.CreateMasterId(update.Message.Chat.ID)
	<-saidGoodBye
	close(saidGoodBye)
	//git pull and restart
	cmd := exec.Command("bash", "/root/evolve_"+rb.nickName)
	cmd.Start()
	os.Exit(1)
}

// Translate Zh<-->En
// command '/trans'
func (rb *Robot) Translate(update tgbotapi.Update) string {
	var info string
	if string(update.Message.Text[0]) == "/" {
		//'trans cat'
		raw := strings.SplitAfterN(update.Message.Text, " ", 2) //at most 2 substring
		if len(raw) < 2 {
			return "what do you want to translate, try '/trans cat'?"
		}
		info = "翻译" + raw[1] //'翻译cat'
	} else {
		info = update.Message.Text
	}

	return qinAI(info)

}

// Talk with AI
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

	if rb.nickName != "jarvis" {
		if chinese(info) {
			return tlAI(info)
		}
		return mitAI(info)
	} else {
		//jarvis use another AI
		return qinAI(info)
	}
}

// SetReminder set an alarm
// command '/alarm'
func (rb *Robot) SetReminder(update tgbotapi.Update, step int) string {
	user := update.Message.Chat.UserName
	tmpTask := userTask[user]
	tmpAction := userAction[user]

	switch step {
	case 0:
		//known issue of go, you can not just assign update.Message.Chat.Id to userTask[user].ChatId
		tmpTask.ChatId = update.Message.Chat.ID
		tmpTask.Owner = user
		userTask[user] = tmpTask

		tmpAction.ActionStep++
		userAction[user] = tmpAction
		return "Ok, what should I remind you to do?"
	case 1:
		//save task content
		tmpTask.Desc = update.Message.Text
		userTask[user] = tmpTask

		tmpAction.ActionStep++
		userAction[user] = tmpAction
		return "When or how much time after?\n" +
			"You can type:\n" +
			"'*2/14 11:30*' means 11:30 at 2/14 \n" + //first format
			"'*11:30*' means  11:30 today\n" + //second format
			"'*5m10s*' means 5 minutes 10 seconds later" //third format
	case 2:
		//format time
		text := update.Message.Text
		text = strings.Replace(text, "：", ":", -1)
		var showTime string  //show to user
		var redisTime string //time string to save to redis
		var scheduledTime time.Time
		var nowTime = time.Now()
		var du time.Duration
		var err error
		if strings.Contains(text, ":") {
			//first and second case
			scheduledTime, err = time.Parse(FirstFormat, text)
			//			nowTime, _ = time.Parse(FirstFormat, nowTime.Format(FirstFormat))
			showTime = scheduledTime.Format(FirstFormat)
			redisTime = scheduledTime.Format(RedisFormat)
			//second case
			if err != nil {
				//try to parse with second format
				scheduledTime, err = time.Parse(SecondFormat, text)
				redisTime = nowTime.Format("1/02 ") + scheduledTime.Format(ThirdFormat)
				//				nowTime, _ = time.Parse(SecondFormat, nowTime.Format(SecondFormat))
				showTime = scheduledTime.Format(SecondFormat)

				if err != nil {
					return "wrong format, try '2/14 11:30' or '11:30'?"
				}
			}
		} else {
			//third case
			du, err = time.ParseDuration(text)
			scheduledTime = nowTime.Add(du)
			showTime = scheduledTime.Format(ThirdFormat)
			redisTime = scheduledTime.Format(RedisFormat)
			if err != nil {
				return "wrong format, try '1h2m3s'?"
			}
		}
		//save time
		tmpTask.When = redisTime
		tmpTask.Id = conn.UpdateTaskId()
		userTask[user] = tmpTask
		//arrange to do the task
		go rb.DoTask(userTask[user])
		//save task in redis
		go conn.CreateTask(userTask[user])
		delete(userAction, user) //delete userAction, prevent to be stuck
		return fmt.Sprintf("Ok, I will remind you that\n*%s* - *%s*", showTime, userTask[user].Desc)
	}
	return ""
}

// DoTask accomplish all tasks here
func (rb *Robot) DoTask(ts conn.Task) {
	nowString := time.Now().Format(RedisFormat)
	now, _ := time.Parse(RedisFormat, nowString)
	when, _ := time.Parse(RedisFormat, ts.When)
	if when.After(now) {
		//set timer
		du := when.Sub(now)
		timer := time.NewTimer(du)
		abortTask[ts.Id] = make(chan int)
		for {
			select {
			case <-abortTask[ts.Id]:
				//triggered by 'rm alarm' command
				log.Println("abort mission:", ts.Id)
				conn.DeleteTask(ts)
				return
			case <-timer.C:
				break
			}
			break
		}
	}
	//else if now is after when means we miss the time to do the task, so do it immediately
	rawMsg := fmt.Sprintf("Hi %s, maybe it's time to:\n*%s*", ts.Owner, ts.Desc)
	if rb.Reply(ts.ChatId, rawMsg) != nil {
		//if failed to send with the given chatId, load it from redis
		rb.Reply(conn.ReadUserChatId(ts.Owner), rawMsg)
	}
	//delete the task from redis, we won't save it
	conn.DeleteTask(ts)
}

// GetTasks get the given  user's all tasks
// command 'alarms'
func (rb *Robot) GetTasks(update tgbotapi.Update) (ret string) {
	user := update.Message.Chat.UserName
	tasks := conn.ReadUserTasks(user)
	if len(tasks) == 0 {
		return "You have no alarm now, type '/alarm' to set one?"
	}
	for i := range tasks {
		ret += fmt.Sprintf("%d. %s:  %s\n", i+1, tasks[i].When, tasks[i].Desc)
	}
	return
}

// RemoveReminder cancel a task
// command '/rmalarm'
func (rb *Robot) RemoveReminder(update tgbotapi.Update, step int) (ret string) {
	user := update.Message.Chat.UserName
	tmpAction := userAction[user]
	switch step {
	case 0:
		//init the struct
		tmpAction.ActionStep++
		userAction[user] = tmpAction
		tasks := conn.ReadUserTasks(user)
		if len(tasks) == 0 {
			delete(userAction, user)
			return "You have no alarm now, type '/alarm' to set one?"
		}
		ret = "Ok, which alarm do you want to remove?(type id)\n"
		userTaskIds[user] = make([]int, len(tasks))
		for i := range tasks {
			userTaskIds[user][i] = tasks[i].Id
			ret += fmt.Sprintf("%d. %s:  %s\n", i+1, tasks[i].When, tasks[i].Desc)
		}
	case 1:
		defer func() {
			delete(userAction, user)
			delete(userTaskIds, user)
		}()
		index, err := strconv.Atoi(update.Message.Text)
		if err != nil {
			return "please select the alarm id"
		}
		taskId := userTaskIds[user][index-1]
		//cancel the task
		abortTask[taskId] <- 1
		ret = "Ok, type '/alarms' to see your new alarms"
	}
	return
}

// DownloadMovie download movie from lbl and zmz
// command '/movie'
func (rb *Robot) DownloadMovie(update tgbotapi.Update, step int, results chan<- string) (ret string) {
	user := update.Message.Chat.UserName
	tmpAction := userAction[user]
	switch step {
	case 0:
		tmpAction.ActionStep++
		userAction[user] = tmpAction
		ret = "Ok, which movie do you want to download?"
	case 1:
		defer func() {
			delete(userAction, user)
			//close(results)
		}()
		results <- "Searching movie..."
		movie := update.Message.Text
		//var wg sync.WaitGroup
		//wg.Add(2)
		go getMovieFromZMZ(movie, results)
		go getMovieFromLBL(movie, results)
		//wg.Wait()
	}
	return
}

// DownloadShow download American show from zmz
// command '/show'
func (rb *Robot) DownloadShow(update tgbotapi.Update, step int, results chan string) (ret string) {
	user := update.Message.Chat.UserName
	tmpAction := userAction[user]
	switch step {
	case 0:
		tmpAction.ActionStep++
		userAction[user] = tmpAction
		ret = "Ok, which American show do you want to download?"
	case 1:
		results <- "Searching American show..."
		info := strings.Fields(update.Message.Text)
		if len(info) < 3 {
			if rct := conn.ReadDownloadRecord(user, info[0]); rct != "" {
				results <- fmt.Sprintf("Recent downloads of %s: %s", info[0], rct)
			} else {
				results <- "Please specify the season and episode, like:\n*疑犯追踪 1 10*"
			}
			return
		}
		if getShowFromZMZ(info[0], info[1], info[2], results) {
			//found resource
			conn.CreateDownloadRecord(user, info[0], fmt.Sprintf("S%sE%s", info[1], info[2]))
		}
		delete(userAction, user)
		close(results)
	}
	return
}

// SaveMemo create a memo for user, saved in redis
// command '/memo'
func (rb *Robot) SaveMemo(update tgbotapi.Update, step int) (ret string) {
	user := update.Message.Chat.UserName
	tmpAction := userAction[user]
	switch step {
	case 0:
		tmpAction.ActionStep++
		userAction[user] = tmpAction
		ret = "Ok, what do you want to save?"
	case 1:
		defer delete(userAction, user)
		when := time.Now().Format("2006-1-02 15:04")
		memo := update.Message.Text
		go conn.CreateMemo(user, when, memo)
		ret = "Ok, type '/memos' to see all your memos"
	}
	return
}

// GetAllMemos reads all user's memos in redis
// command '/memos'
func (rb *Robot) GetAllMemos(update tgbotapi.Update) (ret string) {
	user := update.Message.Chat.UserName
	memos := conn.ReadAllMemos(user)
	if len(memos) == 0 {
		return "You have no memo now, type '/memo' to save one?"
	}
	for i := range memos {
		ret += fmt.Sprintf("%d. %s:  *%s*\n", i+1, memos[i].Time, memos[i].Content)
	}
	return
}

// RemoveMemo deletes a memo
// command '/rmmemo'
func (rb *Robot) RemoveMemo(update tgbotapi.Update, step int) (ret string) {
	user := update.Message.Chat.UserName
	tmpAction := userAction[user]
	switch step {
	case 0:
		tmpAction.ActionStep++
		userAction[user] = tmpAction
		ret = "Ok, which memo do you want to remove?(type id)\n" + rb.GetAllMemos(update)
	case 1:
		defer delete(userAction, user)
		index, err := strconv.Atoi(update.Message.Text)
		if err != nil {
			return "please select the memo id"
		}
		go conn.DeleteMemo(user, index-1)
		ret = "Ok, type '/memos' to see your new memos"
	}
	return
}

//restore task when robot run
func restoreTasks(rb *Robot) {
	tasks := conn.ReadAllTasks()
	log.Println("unfinished tasks:", len(tasks))
	for i := range tasks {
		go rb.DoTask(tasks[i])
	}
}

//all telegram updates are handled here
func handlerUpdate(rb *Robot, update tgbotapi.Update) {
	defer func() {
		if p := recover(); p != nil {
			err := fmt.Errorf("internal error: %v", p)
			log.Println(err)
		}
	}()
	user := update.Message.Chat.UserName
	text := update.Message.Text
	var endPoint, rawMsg string
	text = strings.Replace(text, "@"+rb.name, "", 1)
	received := strings.Split(text, " ")
	endPoint = received[0]
	if endPoint == "/exit" {
		delete(userAction, user)
		return
	}
	if action, ok := userAction[user]; ok {
		//if user is in interaction mode
		rawMsg = inAction(rb, action, update)
	} else if string([]rune(text)[:2]) == "翻译" {
		rawMsg = rb.Translate(update)
	} else if string(text[0]) == "/" {
		//new command
		log.Printf(endPoint)
		rawMsg = inCommand(rb, endPoint, update)
	} else {
		//just talk
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

func inAction(rb *Robot, action Action, update tgbotapi.Update) (rawMsg string) {
	switch action.ActionName {
	case "setReminder":
		rawMsg = rb.SetReminder(update, action.ActionStep)
	case "saveMemo":
		rawMsg = rb.SaveMemo(update, action.ActionStep)
	case "removeMemo":
		rawMsg = rb.RemoveMemo(update, action.ActionStep)
	case "removeReminder":
		rawMsg = rb.RemoveReminder(update, action.ActionStep)
	case "downloadMovie":
		results := make(chan string, 2)
		go rb.DownloadMovie(update, action.ActionStep, results)
		for {
			msg, ok := <-results
			if !ok {
				return
			}
			rb.Reply(update, msg)

		}
	case "downloadShow":
		results := make(chan string, 5)
		go rb.DownloadShow(update, action.ActionStep, results)
		for {
			msg, ok := <-results
			if !ok {
				return
			}
			rb.Reply(update, msg)
		}
	}
	return
}

func inCommand(rb *Robot, endPoint string, update tgbotapi.Update) (rawMsg string) {
	user := update.Message.Chat.UserName
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
	case "/rmmemo":
		tmpAction := userAction[user]
		tmpAction.ActionName = "removeMemo"
		userAction[user] = tmpAction
		rawMsg = rb.RemoveMemo(update, 0)
	case "/rmalarm":
		tmpAction := userAction[user]
		tmpAction.ActionName = "removeReminder"
		userAction[user] = tmpAction
		rawMsg = rb.RemoveReminder(update, 0)
	case "/show":
		tmpAction := userAction[user]
		tmpAction.ActionName = "downloadShow"
		userAction[user] = tmpAction
		rawMsg = rb.DownloadShow(update, 0, nil)
	case "/evolve":
		rawMsg = "upgrading..."
		go rb.Evolve(update)
	case "/repeat":
		rb.Repeat(update)
	default:
		rawMsg = "unknow command, type /help?"
	}
	return
}

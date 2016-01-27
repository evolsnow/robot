package main

import (
	"flag"
	"fmt"
	"github.com/evolsnow/robot/conn"
	"golang.org/x/net/websocket"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func main() {
	var configFile string
	var debug bool

	flag.StringVar(&configFile, "c", "config.json", "specify config file")
	flag.BoolVar(&debug, "d", false, "debug mode")
	flag.Parse()
	config, err := ParseConfig(configFile)
	if err != nil {
		log.Fatal("a vailid json config file must exist")
	}

	//connect to redis
	redisPort := strconv.Itoa(config.RedisPort)
	redisServer := net.JoinHostPort(config.RedisAddress, redisPort)
	if !conn.Ping(redisServer, config.RedisPassword) {
		log.Fatal("connect to redis server failed")
	}
	conn.Pool = conn.NewPool(redisServer, config.RedisPassword, config.RedisDB)

	//create robot and run
	robot := newRobot(config.RobotToken, config.RobotName, config.WebHookUrl)
	robot.bot.Debug = debug
	go robot.run()

	go groupTalk()

	//run server and web samaritan
	srvPort := strconv.Itoa(config.Port)
	http.HandleFunc("/ajax", ajax)
	http.Handle("/websocket", websocket.Handler(socketHandler))
	log.Fatal(http.ListenAndServeTLS(net.JoinHostPort(config.Server, srvPort), config.Cert, config.CertKey, nil))

}

func groupTalk() {
	tlChan := make(chan string)
	qinChan := make(chan string)
	//iceChan := make(chan string, 5)
	initSentence := "你好"
	//iceChan <- tlAI(initSentence)
	go func() {
		qinChan <- tlAI(initSentence)
	}()
	for {
		select {
		//case msgToIce := <-iceChan:
		//	replyFromIce := iceAI(msgToIce)
		//	tlChan <- replyFromIce
		//	qinChan <- replyFromIce
		case msgToTl := <-tlChan:
			replyFromTl := tlAI(msgToTl)
			qinChan <- replyFromTl
		//iceChan <- replyFromTl
		case msgToQin := <-qinChan:
			replyFromQin := qinAI(msgToQin)
			//iceChan <- replyFromQin
			tlChan <- replyFromQin
		}
	}
}

//used for web samaritan robot
func socketHandler(ws *websocket.Conn) {
	for {
		var in string
		var ret []string
		if err := websocket.Message.Receive(ws, &in); err != nil {
			log.Println("socket closed")
			return
		}
		ret = receive(in)

		for i := range ret {
			websocket.Message.Send(ws, ret[i])
			time.Sleep(time.Second)
		}
		websocket.Message.Send(ws, "")
	}
}

//when webSocket unavailable, fallback to ajax long polling
func ajax(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	var messages = make(chan string)
	if r.Method == "GET" {
		w.Write([]byte(<-messages))
	} else {
		body := r.FormValue("text")
		if body != "" {
			go func(string) {
				ret := receive(body)
				for i := range ret {
					messages <- ret[i]
					time.Sleep(time.Second)
				}
				messages <- ""
			}(body)
		}
	}
}

//receive from client
func receive(in string) (ret []string) {
	defer func() {
		if p := recover(); p != nil {
			err := fmt.Errorf("client closed error: %v", p)
			log.Println(err)
		}
	}()
	fmt.Printf("Received: %s\n", in)
	var response string
	var answer = make(chan string)
	sf := func(c rune) bool {
		return c == ',' || c == '，' || c == ';' || c == '。' || c == '.' || c == '？' || c == '?'
	}
	if chinese(in) {
		go func() {
			answer <- iceAI(in)
		}()
		go func() {
			answer <- tlAI(in)
		}()
		go func() {
			ret := qinAI(in)
			if ret != "" {
				answer <- strings.Replace(ret, "Jarvis", "samaritan", -1)
			}
		}()
		response = <-answer
		// Separate into fields with func.
		ret = strings.FieldsFunc(response, sf)

	} else {
		response = mitAI(in)
		ret = strings.FieldsFunc(response, sf)
	}
	return
}

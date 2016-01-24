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

var messages = make(chan string)

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

	//run server and web samaritan
	srvPort := strconv.Itoa(config.Port)
	http.HandleFunc("/ajax", ajax)
	http.Handle("/websocket", websocket.Handler(socketHandler))
	log.Fatal(http.ListenAndServeTLS(net.JoinHostPort(config.Server, srvPort), config.Cert, config.CertKey, nil))

}

//used for web samaritan robot
func socketHandler(ws *websocket.Conn) {
	for {
		var in string
		var ret []string
		if err := websocket.Message.Receive(ws, &in); err != nil {
			log.Println(err)
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
	fmt.Printf("Received: %s\n", in)
	var response string
	sf := func(c rune) bool {
		return c == ',' || c == '，' || c == ';' || c == '。' || c == '.' || c == '？' || c == '?'
	}
	if chinese(in) {
		response = qinAI(in)
		// Separate into fields with func.
		ret = strings.FieldsFunc(response, sf)

	} else {
		response = mitAI(in)
		ret = strings.FieldsFunc(response, sf)
	}
	return
}

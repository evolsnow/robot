package main

import (
	"flag"
	"fmt"
	"github.com/evolsnow/robot/conn"
	"golang.org/x/net/websocket"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"
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

	redisPort := strconv.Itoa(config.RedisPort)
	redisServer := net.JoinHostPort(config.RedisAddress, redisPort)
	if !conn.Ping(redisServer, config.RedisPassword) {
		log.Fatal("connect to redis server failed")
	}
	conn.Pool = conn.NewPool(redisServer, config.RedisPassword, config.RedisDB)
	robot := newRobot(config.RobotToken, config.RobotName, config.WebHookUrl)
	robot.bot.Debug = debug
	go robot.run()
	srvPort := strconv.Itoa(config.Port)
	http.HandleFunc("/ajax", ajax)
	http.Handle("/websocket", websocket.Handler(socketHandler))
	//	log.Fatal(http.ListenAndServe(net.JoinHostPort(config.Server, srvPort), nil))
	log.Fatal(http.ListenAndServeTLS(net.JoinHostPort(config.Server, srvPort), config.Cert, config.CertKey, nil))

}

//used for web samaritan robot
func socketHandler(ws *websocket.Conn) {
	for {
		var in, response string
		var ret []string
		sf := func(c rune) bool {
			return c == ',' || c == '，' || c == ';' || c == '。' || c == '.' || c == '？' || c == '?'
		}
		if err := websocket.Message.Receive(ws, &in); err != nil {
			log.Println(err)
			return
		}
		fmt.Printf("Received: %s\n", in)
		zh := false
		for _, r := range in {
			if unicode.Is(unicode.Scripts["Han"], r) {
				log.Printf(in)
				zh = true
				break
			}
		}
		if zh {
			response = tlAI(in)
			// Separate into fields with func.
			ret = strings.FieldsFunc(response, sf)

		} else {
			response = mitAI(in)
			ret = strings.FieldsFunc(response, sf)
		}
		for i := range ret {
			websocket.Message.Send(ws, ret[i])
			time.Sleep(time.Second)
		}
		websocket.Message.Send(ws, "")
	}
}
func ajax(w http.ResponseWriter, r *http.Request) {
	var messages = make(chan string)
	w.Header().Add("Access-Control-Allow-Origin", "*")
	go func() {
		for {
			time.Sleep(time.Second * 2)
			log.Println("hi")
			messages <- "from ajax"
		}
	}()
	io.WriteString(w, <-messages)
}

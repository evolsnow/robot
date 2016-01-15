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
	"unicode"
)

func main() {
	var configFile string
	flag.StringVar(&configFile, "c", "config.json", "specify config file")
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

	robot := newRobot(config.RobotToken, config.RobotName, config.WebHookAddress)
	go robot.run()

	http.Handle("/websocket", websocket.Handler(socketHandler))
	srvPort := strconv.Itoa(config.Port)
	log.Fatal(http.ListenAndServe(net.JoinHostPort(config.Server, srvPort), nil))
	//used for 104
	//go http.ListenAndServeTLS("0.0.0.0:8443", "server.crt", "server.key", nil)
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
				log.Println(in)
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

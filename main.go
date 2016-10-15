/*
Licensed under the MIT License (MIT)

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/evolsnow/robot/conn"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
	return true
}} // use default options for webSocket
var visitor = 0

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
	robot := newRobot(config.RobotToken, config.RobotName, config.WebHookURL)
	robot.bot.Debug = debug
	go robot.run()

	//run server and web samaritan
	srvPort := strconv.Itoa(config.Port)
	http.HandleFunc("/ajax", ajax)
	http.HandleFunc("/websocket", socketHandler)
	http.HandleFunc("/groupTalk", groupTalk)
	log.Fatal(http.ListenAndServeTLS(net.JoinHostPort(config.Server, srvPort), config.Cert, config.CertKey, nil))

}

//3 robot's group talk
func groupTalk(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	visitor++
	tlChan := make(chan string, 5)
	qinChan := make(chan string, 5)
	iceChan := make(chan string, 5)
	result := make(chan string, 10)
	initSentence := "你好"
	tlChan <- tlAI(initSentence)
	go func() {
		for {
			if visitor > 0 {
				msgToTl := <-tlChan
				replyFromTl := tlAI(msgToTl)
				time.Sleep(time.Second * 2)
				go func() {
					if replyFromTl != "" {
						result <- "samaritan: " + replyFromTl
						qinChan <- replyFromTl
						iceChan <- replyFromTl
					}
				}()
				//c.WriteMessage(websocket.TextMessage, []byte("samaritan: " + replyFromTl))
			} else {
				break
			}
		}
	}()

	go func() {
		for {
			if visitor > 0 {
				msgToQin := <-qinChan
				replyFromQin := qinAI(msgToQin)
				time.Sleep(time.Second * 2)
				go func() {
					if replyFromQin != "" {
						result <- "菲菲: " + replyFromQin
						tlChan <- replyFromQin
						iceChan <- replyFromQin
					}
				}()
				//c.WriteMessage(websocket.TextMessage, []byte("菲菲: " + replyFromQin))
			} else {
				break
			}
		}
	}()

	go func() {
		for {
			if visitor > 0 {
				msgToIce := <-iceChan
				replyFromIce := iceAI(msgToIce)
				go func() {
					if replyFromIce != "" {
						result <- "小冰: " + replyFromIce
						tlChan <- replyFromIce
						qinChan <- replyFromIce
					}
				}()
				//c.WriteMessage(websocket.TextMessage, []byte("小冰: " + replyFromIce))
			} else {
				break
			}
		}
	}()

	go func() {
		for {
			if visitor > 0 {
				c.WriteMessage(websocket.TextMessage, []byte(<-result))
			} else {
				break
			}
		}
	}()

	for {

		_, _, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			visitor--
			break
		}
	}
}

//used by web samaritan robot
func socketHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, in, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		ret := receive(string(in))
		for i := range ret {
			c.WriteMessage(mt, []byte(ret[i]))
			time.Sleep(time.Second)
		}
		c.WriteMessage(mt, []byte(""))
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
	fmt.Printf("Received: %s\n", in)
	var response string
	var answer = make(chan string, 3)
	sf := func(c rune) bool {
		return c == ',' || c == '，' || c == ';' || c == '。' || c == '.' || c == '？' || c == '?'
	}
	if chinese(in) {
		go func() {
			if ret := iceAI(in); ret != "" {
				answer <- ret
			}
		}()
		go func() {
			if ret := tlAI(in); ret != "" {
				answer <- ret
			}
		}()
		go func() {
			if ret := qinAI(in); ret != "" {
				answer <- strings.Replace(ret, "Jarvis", "samaritan", -1)
			}
		}()
		response = <-answer // accept the first reply
		// Separate into fields with func.
		ret = strings.FieldsFunc(response, sf)

	} else {
		response = mitAI(in)
		ret = strings.FieldsFunc(response, sf)
	}
	return
}

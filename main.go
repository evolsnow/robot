package main

import (
	"fmt"
	"github.com/evolsnow/robot/conn"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode"
)

func main() {
	conn.Pool = conn.NewPool("127.0.0.1:6379", "", 1)
	if !conn.Ping("127.0.0.1:6379", "") {
		log.Fatal("connect to redis server failed")
	}

	samaritan := newRobot("164760320:AAEE0sKLgCwHGYJ0Iqz7o-GYH4jVTQZAZho", "samaritan")
	jarvis := newRobot("176820788:AAH26vgFIk7oWKibd7P8XHHZX2t2_2Jvke8", "jarvis")
	//	samaritan.bot.Debug = true
	go samaritan.run()
	go jarvis.run()

	http.HandleFunc("/ajax", ajax)
	http.Handle("/websocket", websocket.Handler(socketHandler))
	log.Fatal(http.ListenAndServe("localhost:8000", nil))
	//used for 104
	//go http.ListenAndServeTLS("0.0.0.0:8443", "server.crt", "server.key", nil)
}

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
				in = strings.TrimSpace(in)
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
			//			ret = strings.Split(response, ",")
			ret = strings.FieldsFunc(response, sf)
		}
		go func(ws *websocket.Conn, ret []string) {
			for i := range ret {
				websocket.Message.Send(ws, ret[i])
				time.Sleep(time.Second)
			}
			websocket.Message.Send(ws, "")
		}(ws, ret)
	}
}

func ajax(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "https://samaritan.tech")
	for i:=0;i<10;i++ {
		fmt.Fprint(w, "reply from ajax")
	}
}

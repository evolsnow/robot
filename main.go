package main

import (
	"fmt"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode"
)

func main() {
	//test()
	//used for 104
	//go http.ListenAndServeTLS("0.0.0.0:8443", "server.crt", "server.key", nil)
	samaritan := newRobot("164760320:AAEE0sKLgCwHGYJ0Iqz7o-GYH4jVTQZAZho")
	go samaritan.run()
	http.Handle("/websocket", websocket.Handler(socketHandler))
	log.Fatal(http.ListenAndServe("localhost:8000", nil))
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
				in = strings.Replace(in, " ", "", -1)
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
			websocket.Message.Send(ws, "      ")
		}(ws, ret)
	}
}

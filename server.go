package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode"
)

func main() {
	//test()
	//used for 104
	//go http.ListenAndServeTLS("0.0.0.0:8443", "server.crt", "server.key", nil)

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
			log.Println("汉语")
			response = tlAI2(in)
			// Separate into fields with func.
			ret = strings.FieldsFunc(response, sf)

		} else {
			log.Println("英语")
			response = mitAI2(in)
			//			ret = strings.Split(response, ",")
			ret = strings.FieldsFunc(response, sf)
		}
		go func(ws *websocket.Conn, ret []string) {
			for i := range ret {
				websocket.Message.Send(ws, ret[i])
				time.Sleep(time.Second)
			}
			websocket.Message.Send(ws, "again?")
		}(ws, ret)
	}
}

func tlAI2(info string) string {
	key := "a5052a22b8232be1e387ff153e823975"
	tuLingURL := fmt.Sprintf("http://www.tuling123.com/openapi/api?key=%s&info=%s", key, info)
	resp, err := http.Get(tuLingURL)
	if err != nil {
		log.Println(err.Error())
	}
	defer resp.Body.Close()
	reply := new(tlReply)
	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(reply)
	return strings.Replace(reply.Text+reply.Url, "<br>", "\n", -1)
}

type tlReply struct {
	code int    `json:"code"`
	Url  string `json:"url,omitempty"`
	Text string `json:"text"`
}

func mitAI2(info string) string {
	mitURL := "http://fiddle.pandorabots.com/pandora/talk?botid=9fa364f2fe345a10&skin=demochat"
	resp, err := http.PostForm(mitURL, url.Values{"message": {info}, "botcust2": {"d064e07d6e067535"}})
	if err != nil {
		log.Println(err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	re, _ := regexp.Compile("Mitsuku:</B>(.*?)<br> <br>")
	all := re.FindAll(body, -1)
	if len(all) == 0 {
		return "change another question?"
	}
	found := (string(all[0]))
	log.Println(found)
	ret := strings.Replace(found, `<P ALIGN="CENTER"><img src="http://`, "", -1)
	ret = strings.Replace(ret, `"></img></P>`, "", -1)
	ret = strings.Replace(ret[13:], "<br>", "\n", -1)
	ret = strings.Replace(ret, "Mitsuku", "samaritan", -1)
	return ret
}

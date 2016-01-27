package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

func tlAI(info string) string {
	key := "a5052a22b8232be1e387ff153e823975"
	tuLingURL := fmt.Sprintf("http://www.tuling123.com/openapi/api?key=%s&info=%s", key, url.QueryEscape(info))
	resp, err := http.Get(tuLingURL)
	if err != nil {
		log.Printf(err.Error())
	}
	defer resp.Body.Close()
	reply := new(tlReply)
	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(reply)
	log.Printf("reply from tuling machine: %s", reply.Text+"\n"+reply.Url)
	wl := []string{"<cd.url=互动百科@", "", "&prd=button_doc_jinru>", "", "<br>", "\n"}
	srp := strings.NewReplacer(wl...)
	ret := srp.Replace(reply.Text + "\n" + reply.Url)
	return ret
}

type tlReply struct {
	code int    `json:"code"`
	Url  string `json:"url,omitempty"`
	Text string `json:"text"`
}

func qinAI(info string) string {
	//info = strings.Replace(info, " ", "+", -1)
	qinURL := fmt.Sprintf("http://api.qingyunke.com/api.php?key=free&appid=0&msg=%s", url.QueryEscape(info))
	resp, err := http.Get(qinURL)
	if err != nil {
		log.Println(err)
		return ""
	}
	defer resp.Body.Close()
	reply := new(qinReply)
	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(reply)
	log.Printf("reply from qingyunke machine: %s", reply.Content)
	wl := []string{"{br}", "\n", "菲菲", "Jarvis"}
	srp := strings.NewReplacer(wl...)
	ret := srp.Replace(reply.Content)
	return ret
}

type qinReply struct {
	result  int    `json:"resulte"`
	Content string `json:"content"`
}

func mitAI(info string) string {
	mitURL := "http://fiddle.pandorabots.com/pandora/talk?botid=9fa364f2fe345a10&skin=demochat"
	resp, err := http.PostForm(mitURL, url.Values{"message": {info}, "botcust2": {"d064e07d6e067535"}})
	if err != nil {
		log.Printf(err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	re, _ := regexp.Compile("Mitsuku:</B>(.*?)<br> <br>")
	all := re.FindSubmatch(body)
	if len(all) == 0 {
		return "change another question?"
	}
	found := (string(all[1]))
	log.Printf("reply from mitsuku machine: %s", found)
	wl := []string{`<P ALIGN="CENTER"><img src="http://`, "", `"></img></P>`, " ", "<br>", "\n", "xloadswf2.", "", "Mitsuku", "samaritan"}
	srp := strings.NewReplacer(wl...)
	ret := srp.Replace(found)
	ret = strings.TrimLeft(ret, " ")
	return ret
}

func iceAI(info string) string {
	//Ice may failed sometimes
	defer func() {
		if p := recover(); p != nil {
			err := fmt.Errorf("xiaoice error: %v", p)
			log.Println(err)
		}
	}()
	iceURL := fmt.Sprintf("http://127.0.0.1:8008/openxiaoice/ask?q=%s", url.QueryEscape(info))
	resp, err := http.Get(iceURL)
	if err != nil {
		log.Printf(err.Error())
	}
	defer resp.Body.Close()
	reply := new(iceReply)
	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(reply)
	log.Printf("reply from xiaoice: %s", reply.Answer)
	return reply.Answer
}

type iceReply struct {
	Code   int    `json:"code"`
	Answer string `json:"answer"`
}

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	tlKey = "a5052a22b8232be1e387ff153e823975"
)

//get reply from tlAI
func tlAI(info string) string {

	tuLingURL := fmt.Sprintf("http://www.tuling123.com/openapi/api?key=%s&info=%s", tlKey, url.QueryEscape(info))
	resp, err := http.Get(tuLingURL)
	if err != nil {
		log.Println(err)
		return ""
	}
	defer resp.Body.Close()
	reply := new(tlReply)
	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(reply)
	log.Printf("reply from tuling machine: %s", reply.Text + "\n" + reply.URL)
	wl := []string{"<cd.url=互动百科@", "", "&prd=button_doc_jinru>", "", "<br>", "\n"}
	srp := strings.NewReplacer(wl...)
	ret := srp.Replace(reply.Text + "\n" + reply.URL)
	return ret
}

type tlReply struct {
	code int
	URL  string `json:"url,omitempty"`
	Text string `json:"text"`
}

//get reply from qinAI
func qinAI(info string) string {
	//info = strings.Replace(info, " ", "+", -1)
	qinURL := fmt.Sprintf("http://api.qingyunke.com/api.php?key=free&appid=0&msg=%s", url.QueryEscape(info))
	timeout := time.Duration(2 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(qinURL)
	//resp, err := http.Get(qinURL)
	if err != nil {
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
	result  int
	Content string `json:"content"`
}

//get reply from mitAI
func mitAI(info string) string {
	mitURL := "https://demo.pandorabots.com/atalk/mitsuku/mitsukudemo"
	resp, err := http.PostForm(mitURL, url.Values{"input": {info}, "user_key": {"pb3568993377180953528873199695415106305"}})
	if err != nil {
		log.Printf(err.Error())
		return ""
	}
	defer resp.Body.Close()
	reply := new(mitReply)
	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(reply)
	log.Printf("reply from mit machine: %s", reply.Responses[0])
	return reply.Responses[0]
}

type mitReply struct {
	Responses []string `json:"responses"`
}

//get reply from iceAI
func iceAI(info string) string {
	iceURL := fmt.Sprintf("http://127.0.0.1:8008/openxiaoice/ask?q=%s", url.QueryEscape(info))
	timeout := time.Duration(4 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(iceURL)
	if err != nil {
		return ""
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

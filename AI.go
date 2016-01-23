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
	info = strings.Replace(info, " ", "", -1)
	key := "a5052a22b8232be1e387ff153e823975"
	tuLingURL := fmt.Sprintf("http://www.tuling123.com/openapi/api?key=%s&info=%s", key, info)
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
	info = strings.Replace(info, " ", "+", -1)
	qinURL := fmt.Sprintf("http://api.qingyunke.com/api.php?key=free&appid=0&msg=%s", info)
	resp, err := http.Get(qinURL)
	if err != nil {
		log.Printf(err.Error())
	}
	defer resp.Body.Close()
	reply := new(qinReply)
	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(reply)
	log.Printf("reply from qingyunke machine: %s", reply.Content)
	ret := strings.Replace(reply.Content, "{br}", "\n", -1)
	return strings.Replace(ret, "菲菲", "Jarvis", -1)
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
	ret := strings.Replace(found, `<P ALIGN="CENTER"><img src="http://`, "", -1)
	ret = strings.Replace(ret, `"></img></P>`, "", -1)
	ret = strings.Replace(ret, "<br>", "\n", -1)
	ret = strings.Replace(ret, "xloadswf2.", "", -1)
	ret = strings.Replace(ret, "Mitsuku", "samaritan", -1)
	ret = strings.TrimLeft(ret, " ")
	return ret
}

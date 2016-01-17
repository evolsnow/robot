package main

//import (
//	"net/http"
//	"io/ioutil"
//	"regexp"
//	"net/url"
//	"log"
//	"strings"
//	"fmt"
//)
//
//func main() {
//	//	log.Println(mitAI2("hello"))
//	movie := "盗梦间"
//	var id string
//	resp, _ := http.Get("http://www.lbldy.com/search/" + movie)
//	defer resp.Body.Close()
//	body, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		fmt.Println("http read error")
//		return
//	}
//
//	re, _ := regexp.Compile("<div class=\"postlist\" id=\"post-(.*?)\">")
//
//	//查找符合正则的第一个
//	firstId := re.FindSubmatch(body)
//	if len(firstId) == 0 {
//		log.Printf("no answer for %s", movie)
//	}else {
//		id = string(firstId[1])
//		fmt.Println("Find:", id)
//		resp, _ = http.Get("http://www.lbldy.com/movie/" + id + ".html")
//		defer resp.Body.Close()
//		re, _ = regexp.Compile("<p><a href=\"(.*?)\" target=\"_blank\">(.*?)</a></p>")
//		body, _ = ioutil.ReadAll(resp.Body)
//		downloads := re.FindAllSubmatch(body, -1)
//		for i := range downloads {
//			log.Printf("found %s: %s\n", string(downloads[i][2]), string(downloads[i][1]))
//		}
//	}
//
//}
//
//func mitAI2(info string) string {
//	mitURL := "http://fiddle.pandorabots.com/pandora/talk?botid=9fa364f2fe345a10&skin=demochat"
//	resp, err := http.PostForm(mitURL, url.Values{"message": {info}, "botcust2": {"d064e07d6e067535"}})
//	if err != nil {
//		log.Printf(err.Error())
//	}
//	defer resp.Body.Close()
//	body, err := ioutil.ReadAll(resp.Body)
//	re, _ := regexp.Compile("Mitsuku:</B>(.*?)<br> <br>")
//	all := re.FindSubmatch(body)
//	if len(all) == 0 {
//		return "change another question?"
//	}
//	found := (string(all[1]))
//	log.Printf("reply from mitsuku machine: %s", found)
//	ret := strings.Replace(found, `<P ALIGN="CENTER"><img src="http://`, "", -1)
//	ret = strings.Replace(ret, `"></img></P>`, "", -1)
//	ret = strings.Replace(ret, "<br>", "\n", -1)
//	ret = strings.Replace(ret, "xloadswf2.", "", -1)
//	ret = strings.Replace(ret, "Mitsuku", "samaritan", -1)
//	ret = strings.TrimLeft(ret, " ")
//	return ret
//}

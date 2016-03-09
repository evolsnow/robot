// get show and movie source download links
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"sync"
)

//zmz.tv needs to login before downloading
var zmzClient http.Client

func getMovieFromLBL(movie string, results chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	var id string
	resp, _ := http.Get("http://www.lbldy.com/search/" + movie)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	re, _ := regexp.Compile("<div class=\"postlist\" id=\"post-(.*?)\">")
	//find first match case
	firstId := re.FindSubmatch(body)
	if len(firstId) == 0 {
		results <- fmt.Sprintf("No results for *%s* from LBL", movie)
		return
	} else {
		id = string(firstId[1])
		resp, _ = http.Get("http://www.lbldy.com/movie/" + id + ".html")
		defer resp.Body.Close()
		re, _ = regexp.Compile(`<p><a href="(.*?)"( target="_blank">|>)(.*?)</a></p>`)
		body, _ := ioutil.ReadAll(resp.Body)
		//go does not support (?!) regex
		body = []byte(strings.Replace(string(body), `<a href="/xunlei/"`, "", -1))
		downloads := re.FindAllSubmatch(body, -1)
		if len(downloads) == 0 {
			results <- fmt.Sprintf("No results for *%s* from LBL", movie)
			return
		} else {
			ret := "Results from LBL:\n\n"
			for i := range downloads {
				ret += fmt.Sprintf("*%s*\n```%s```\n\n", string(downloads[i][3]), string(downloads[i][1]))
				//when results are too large, we split it.
				if i%5 == 0 && i > 0 {
					results <- ret
					ret = fmt.Sprintf("*LBL Part %d*\n\n", i/5+1)
				}
			}
			results <- ret
		}
	}
}

func getMovieFromZMZ(movie string, results chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	loginZMZ()
	if downloads := getZMZResource(movie); downloads == nil {
		results <- fmt.Sprintf("No results for *%s* from ZMZ", movie)
		return
	} else {
		ret := "Results from ZMZ:\n\n"
		for i := range downloads {
			name := string(downloads[i][1])
			size := string(downloads[i][2])
			link := string(downloads[i][3])
			ret += fmt.Sprintf("*%s*(%s)\n```%s```\n\n", name, size, link)
			if i%5 == 0 && i > 0 {
				results <- ret
				ret = fmt.Sprintf("*ZMZ Part %d*\n\n", i/5+1)
			}
		}
		results <- ret
	}
	return
}

func getShowFromZMZ(show, s, e string, results chan string) {
	loginZMZ()
	downloads := getZMZResource(show)
	if downloads == nil {
		results <- fmt.Sprintf("No results for *%s* from ZMZ", show)
		return
	}
	//second parse
	re, _ := regexp.Compile(fmt.Sprintf(".*?season=\"%s\" episode=\"%s\">.*?", s, e))
	results <- "Results from ZMZ:\n\n"
	count := 0
	for i := range downloads {
		if re.Find(downloads[i][0]) != nil {
			name := string(downloads[i][1])
			size := string(downloads[i][2])
			link := string(downloads[i][3])
			results <- fmt.Sprintf("*ZMZ %s*(%s)\n```%s```\n\n", name, size, link)
			count++
		}
	}
	if count == 0 {
		results <- fmt.Sprintf("No results found for *S%sE%s*", s, e)

	}
	return
}

//get show and get movie from zmz both uses this function
func getZMZResource(name string) [][][]byte {
	id := getZMZResourceId(name)
	if id == "" {
		return nil
	}
	resourceURL := "http://www.zimuzu.tv/resource/list/" + id
	resp, _ := zmzClient.Get(resourceURL)
	defer resp.Body.Close()
	//1.name 2.size 3.link
	re, _ := regexp.Compile(`<li class="clearfix".*?<input type="checkbox"><a title="(.*?)".*?<font class="f3">(.*?)</font>.*?<a href="(.*?)" type="ed2k">`)
	body, _ := ioutil.ReadAll(resp.Body)
	body = []byte(strings.Replace(string(body), "\n", "", -1))
	downloads := re.FindAllSubmatch(body, -1)
	if len(downloads) == 0 {
		return nil
	}
	return downloads
}

func getZMZResourceId(name string) (id string) {
	queryURL := fmt.Sprintf("http://www.zimuzu.tv/search?keyword=%s&type=resource", name)
	re, _ := regexp.Compile(`<div class="t f14"><a href="/resource/(.*?)"><strong class="list_title">`)
	resp, _ := zmzClient.Get(queryURL)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	//find first match case
	firstId := re.FindSubmatch(body)
	if len(firstId) == 0 {
		return
	} else {
		log.Println(id)
		id = string(firstId[1])
		return
	}
}

//login zmz first because zmz don't allow login at different browsers, but I have two robots...
func loginZMZ() {
	gCookieJar, _ := cookiejar.New(nil)
	zmzURL := "http://www.zimuzu.tv/User/Login/ajaxLogin"
	zmzClient = http.Client{
		Jar: gCookieJar,
	}
	//post with my public account, you can use it also
	zmzClient.PostForm(zmzURL, url.Values{"account": {"evol4snow"}, "password": {"104545"}, "remember": {"0"}})
}

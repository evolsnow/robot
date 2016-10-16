// get show and movie source download links
package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
)

type Media struct {
	Name string
	Size string
	Link string
}

//zmz.tv needs to login before downloading
var zmzClient http.Client

//get movie resource from lbl
func getMovieFromLBL(movie string, results chan<- string) {
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
		var ms []Media
		resp, _ = http.Get("http://www.lbldy.com/movie/" + id + ".html")
		defer resp.Body.Close()
		doc, err := goquery.NewDocumentFromReader(io.Reader(resp.Body))
		if err != nil {
			return
		}
		doc.Find("p").Each(func(i int, selection *goquery.Selection) {
			name := selection.Find("a").Text()
			link, _ := selection.Find("a").Attr("href")
			if strings.HasPrefix(link, "ed2k") || strings.HasPrefix(link, "magnet") || strings.HasPrefix(link, "thunder") {
				m := Media{
					Name: name,
					Link: link,
				}
				ms = append(ms, m)
			}
		})

		if len(ms) == 0 {
			results <- fmt.Sprintf("No results for *%s* from LBL", movie)
			return
		} else {
			ret := "Results from LBL:\n\n"
			for i, m := range ms {
				ret += fmt.Sprintf("*%s*\n```%s```\n\n", m.Name, m.Link)
				//when results are too large, we split it.
				if i%4 == 0 && i < len(ms)-1 && i > 0 {
					results <- ret
					ret = fmt.Sprintf("*LBL Part %d*\n\n", i/4+1)
				}
			}
			results <- ret
		}
	}
}

//get movie resource from zmz
func getMovieFromZMZ(movie string, results chan<- string) {
	loginZMZ()
	if ms := getZMZResource(movie, "0", "0"); ms == nil {
		results <- fmt.Sprintf("No results for *%s* from ZMZ", movie)
		return
	} else {
		ret := "Results from ZMZ:\n\n"
		for i, m := range ms {
			name := m.Name
			size := m.Size
			link := m.Link
			ret += fmt.Sprintf("*%s* %s\n```%s```\n\n", name, size, link)
			if i%4 == 0 && i < len(ms)-1 && i > 0 {
				results <- ret
				ret = fmt.Sprintf("*ZMZ Part %d*\n\n", i/4+1)
			}
		}
		results <- ret
	}
	return
}

//get American show resource from zmz
func getShowFromZMZ(show, s, e string, results chan<- string) bool {
	loginZMZ()
	ms := getZMZResource(show, s, e)
	if ms == nil {
		results <- fmt.Sprintf("No results found for *S%sE%s*", s, e)
		return false
	}
	for _, m := range ms {
		name := m.Name
		size := m.Size
		link := m.Link
		results <- fmt.Sprintf("*ZMZ %s* %s\n```%s```\n\n", name, size, link)
	}
	return true
}

//get show and get movie from zmz both uses this function
func getZMZResource(name, season, episode string) []Media {
	id := getZMZResourceId(name)
	if id == "" {
		return nil
	}
	resourceURL := "http://www.zimuzu.tv/gresource/list/" + id
	resp, _ := zmzClient.Get(resourceURL)
	defer resp.Body.Close()
	//1.name 2.size 3.link
	var ms []Media
	doc, err := goquery.NewDocumentFromReader(io.Reader(resp.Body))
	if err != nil {
		return nil
	}
	doc.Find("li.clearfix").Each(func(i int, selection *goquery.Selection) {
		s, _ := selection.Attr("season")
		e, _ := selection.Attr("episode")
		if s != season || e != episode {
			return
		}
		name := selection.Find(".fl a.lk").Text()
		link, _ := selection.Find(".fr a").Attr("href")
		var size string
		if strings.HasPrefix(link, "ed2k") || strings.HasPrefix(link, "magnet") {
			size = selection.Find(".fl font.f3").Text()
			if size == "" || size == "0" {
				size = "(?)"
			}
			m := Media{
				Name: name,
				Link: link,
				Size: size,
			}
			ms = append(ms, m)
		}
	})
	return ms
}

//get source id before find zmz source
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
	}
	id = string(firstId[1])
	return
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

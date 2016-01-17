package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"sync"
)

func main() {
	//	log.Println(mitAI2("hello"))
	movie := "盗梦空间"
	results := make(chan string)
	log.Println("begin")
	go DownloadMovie2(movie, results)
	for {
		select {
		case msg := <-results:
			if msg == "done" {
				log.Println("ok")
				return
			}
			log.Println(msg)
		}
	}

}

func DownloadMovie2(movie string, results chan string) {
	var wg sync.WaitGroup
	go func() { results <- "Searching..." }()
	wg.Add(1)

	go getMovieFromLbl2(movie, results, &wg)
	wg.Wait()
	results <- "done"
	return
}

func getMovieFromLbl2(movie string, results chan string, wg *sync.WaitGroup) {
	log.Println("here")
	wg.Done()
	var id string
	resp, _ := http.Get("http://www.lbldy.com/search/" + movie)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//		fmt.Println("http read error")
		//		results <- "network error, try again later"
		return
	}
	re, _ := regexp.Compile("<div class=\"postlist\" id=\"post-(.*?)\">")
	//find first match case
	firstId := re.FindSubmatch(body)
	if len(firstId) == 0 {
		//		results <- fmt.Sprintf("no answer for %s from lbldy.com", movie)
		return
	} else {
		id = string(firstId[1])
		log.Println("Find:", id)
		resp, _ = http.Get("http://www.lbldy.com/movie/" + id + ".html")
		defer resp.Body.Close()
		re, _ = regexp.Compile("<p><a href=\"(.*?)\" target=\"_blank\">(.*?)</a></p>")
		body, _ = ioutil.ReadAll(resp.Body)
		downloads := re.FindAllSubmatch(body, -1)
		if len(downloads) == 0 {
			//			results <- fmt.Sprintf("no answer for *%s*", movie)
			return
		} else {
			results <- "Results from lbldy.com:\n\n"
			ret := ""
			for i := range downloads {
				ret += string(downloads[i][2]) + "\n" + string(downloads[i][1]) + "\n\n"
			}
			results <- ret
		}
	}
}

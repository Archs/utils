package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"github.com/opesun/goquery"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

var (
	url      = flag.String("url", "http://www.baidu.com", "url to start")
	word     = flag.String("k", "keyword", "keyword to search")
	ofname   = flag.String("o", "", "output file for collected links")
	nThreads = flag.Int("n", 1, "number of cocurrent scrawlers")
)

var (
	urls    chan string
	writer  *os.File
	bufSize = 1000
	visited map[string]bool
)

// store only md5 of the url to reduce memory usage
func addNewUrl(url string) {
	h := md5.New()
	io.WriteString(h, url)
	hstr := fmt.Sprintf("%x", h.Sum(nil))
	// only add unvisited urls
	if _, ok := visited[hstr]; !ok {
		visited[hstr] = true
		urls <- url
	}
}

func spider(word string) {
	for {
		url := <-urls
		log.Println(url)
		nodes, err := goquery.ParseUrl(url)
		if err != nil {
			log.Println("Err:", err.Error())
			continue
		}
		// put found url into channel
		nodes.Find("a").Each(func(idx int, el *goquery.Node) {
			go func() {
				newurl := ""
				for _, attr := range el.Attr {
					if attr.Key == "href" {
						if strings.HasPrefix(attr.Val, "http") {
							newurl = attr.Val
						} else {
							// handle relative path
							// TODO this needs care
							newurl = strings.Trim(url, "/") + "/" + strings.Trim(attr.Val, "/")
						}
						addNewUrl(newurl)
					}
				}
			}()
		})
		// match word
		if -1 != strings.Index(nodes.Html(), word) {
			log.Println("Found keyword in:", url)
			writer.WriteString(url + "\n")
		}
	}
}

func main() {
	flag.Parse()
	if *ofname == "" {
		*ofname = "output_" + time.Now().Format("20060102_030405") + ".txt"
	}
	var err error
	writer, err = os.OpenFile(*ofname, os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Panicln(err.Error())
	}
	defer writer.Close()
	urls = make(chan string, bufSize)
	visited = make(map[string]bool)
	// add the very first url to scrawl with
	addNewUrl(*url)
	for i := 0; i < *nThreads; i++ {
		go spider(*word)
	}
	// monitor
	for {
		log.Println(len(urls), "to search ...")
		if len(urls) == 0 {
			log.Println("Done Scrawling")
			os.Exit(1)
		}
		time.Sleep(time.Minute)
	}
}

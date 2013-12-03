package main

import (
	"code.google.com/p/mahonia"
	"crypto/md5"
	"errors"
	"flag"
	"fmt"
	"github.com/opesun/goquery"
	"io"
	"io/ioutil"
	"log"
	"net/http"
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

// for decodings
var (
	gbk  = mahonia.NewDecoder("gb18030")
	big5 = mahonia.NewDecoder("big5")
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

// open
func openUrl(url string) (data string, err error) {
	var resp *http.Response
	var raw []byte
	var dec mahonia.Decoder = nil
	resp, err = http.Get(url)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = errors.New("Bad Status:" + resp.Status)
		return
	}
	// only handle html files
	ctype := resp.Header.Get("Content-Type")
	if -1 == strings.Index(ctype, "text/html") {
		err = errors.New("Nota a html file")
		return
	}
	// try enconding: gbk\big5\utf8
	charset := ""
	if seps := strings.Split(ctype, "="); len(seps) >= 2 {
		charset = seps[1]
		charset = strings.ToLower(charset)
		if strings.HasPrefix(charset, "gb") {
			charset = "gb18030"
			dec = gbk
		} else if strings.HasPrefix(charset, "big") {
			charset = "big5"
			dec = big5
		} else if strings.HasPrefix(charset, "utf") || charset == "" {
			charset = "utf8"
			dec = nil
		} else {
			err = errors.New("Unsupported charset:" + charset)
			return
		}
	} else {
		dec = nil
	}
	// TODO gzip handle
	contentEncoding := resp.Header.Get("Content-Encoding")
	if contentEncoding == "gzip" {
		err = errors.New("Content-Encoding:" + contentEncoding + "temporally not supported")
		return
	}
	// read the response
	if dec != nil {
		raw, err = ioutil.ReadAll(dec.NewReader(resp.Body))
	} else {
		raw, err = ioutil.ReadAll(resp.Body)
	}
	if err != nil {
		return
	}
	defer resp.Body.Close()
	data = string(raw)
	return
}

func spider(word string) {
	for {
		url := <-urls
		log.Println(url)
		html, err := openUrl(url)
		nodes, err := goquery.ParseString(html)
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
		if -1 != strings.Index(html, word) {
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

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
	dbg      = flag.Bool("debug", false, "enable debug or not")
)

var (
	urls        chan string
	errChan     chan string
	foundChan   chan string
	writer      *os.File
	errWriter   *os.File
	bufSize     = 1000
	visited     map[string]bool
	errFileName = "spider.errors.txt"

	// debug handle
	debug Debug
)

// for decodings
var (
	gbk  = mahonia.NewDecoder("gb18030")
	big5 = mahonia.NewDecoder("big5")
)

// handle debug
type Debug bool

func (d Debug) Println(s ...string) {
	if d {
		log.Println(s)
	}
}

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
	debug.Println("Get:", url)
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
		err = errors.New("Not a html file")
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
	debug.Println("Using charset:", charset)
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
	debug.Println("Data:", data)
	return
}

func spider(word string) {
	for {
		url := <-urls
		html, err := openUrl(url)
		if err != nil {
			errChan <- fmt.Sprintf("%s [Open]\t[%s]\t%s\r\n",
				time.Now().Format("2006/01/02 03:04:05"), url, err.Error())
			continue
		}
		nodes, err := goquery.ParseString(html)
		if err != nil {
			errChan <- fmt.Sprintf("%s [Parsing]:\t[%s]\t%s\r\n",
				time.Now().Format("2006/01/02 03:04:05"), url, err.Error())
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
			log.Println("*Found*\t", url)
			foundChan <- fmt.Sprintf("%s\t%s\r\n",
				time.Now().Format("2006/01/02 03:04:05"), url)
		}
	}
}

func main() {
	// parse cmd line
	flag.Parse()
	debug = Debug(*dbg)
	if *ofname == "" {
		*ofname = "output_" + time.Now().Format("20060102_030405") + ".txt"
	}
	var err error
	writer, err = os.OpenFile(*ofname, os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Println(err.Error())
		os.Exit(-1)
	}
	defer writer.Close()
	errWriter, err = os.OpenFile(errFileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Println(err.Error())
		os.Exit(-1)
	}
	defer errWriter.Close()
	urls = make(chan string, bufSize)
	errChan = make(chan string, bufSize)
	foundChan = make(chan string, bufSize)
	visited = make(map[string]bool)
	// add the very first url to scrawl with
	addNewUrl(*url)
	for i := 0; i < *nThreads; i++ {
		go spider(*word)
	}
	// write to file
	go func() {
		for {
			errWriter.WriteString(<-errChan)
		}
	}()
	go func() {
		for {
			writer.WriteString(<-foundChan)
		}
	}()
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

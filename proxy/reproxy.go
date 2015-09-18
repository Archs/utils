package main

import (
	"crypto/aes"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	cipherKey    = []byte("forward.exe11111")
	CookieMaxAge = 60 // in seconds
)

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

type ProxyChecker struct {
	rproxy *httputil.ReverseProxy
}

func newProxyChecker(rproxy *httputil.ReverseProxy) *ProxyChecker {
	return &ProxyChecker{
		rproxy: rproxy,
	}
}

func createAuthCookie(strtime string) *http.Cookie {
	cipher, _ := aes.NewCipher(cipherKey)
	dat := []byte(strtime)
	cipher.Encrypt(dat, dat)
	cookie := http.Cookie{
		Name:   "pAuthentication",
		Value:  hex.EncodeToString(dat),
		MaxAge: CookieMaxAge,
	}
	return &cookie
}

func decodeCookie(cookieValue string) (time.Time, error) {
	cipher, _ := aes.NewCipher(cipherKey)
	dat, err := hex.DecodeString(cookieValue)
	if err != nil {
		return time.Time{}, err
	}
	cipher.Decrypt(dat, dat)
	return time.Parse(time.RFC3339, string(dat))
}

func (c *ProxyChecker) hasValidCookie(rw http.ResponseWriter, req *http.Request) (ok bool) {
	cookies := req.Cookies()
	// 正常访问
	for _, cookie := range cookies {
		if cookie.Name == "pAuthentication" {
			_, err := decodeCookie(cookie.Value)
			if err != nil {
				log.Println("invalid pAuthentication", err.Error())
				return false
			}
			return true
		}
	}
	// 没有正确cookie的时候测试是否为/authRequest
	// 是，则发回服务器时间
	// 登录器将 正确响应：aes(服务器时间)，作为首页url参数auth发送给本代理
	if req.URL.Path == "/authRequest" {
		now := time.Now().Format(time.RFC3339)
		io.WriteString(rw, now)
		fmt.Printf("/authResponse?auth=%s\n", createAuthCookie(now).Value)
		return false
	}
	// 是否为/authReponse
	// 是，则设置新的cookie,跳转到/
	if req.URL.Path == "/authResponse" {
		val := req.URL.Query().Get("auth")
		t, err := decodeCookie(val)
		if err != nil {
			log.Println("/authResponse", err.Error())
			return false
		}
		http.SetCookie(rw, createAuthCookie(t.Format(time.RFC3339)))
		http.Redirect(rw, req, "/", http.StatusMovedPermanently)
		return false
	}
	// 既没有正确的cookie也不是/authXXX请求
	io.WriteString(rw, "请使用登录器")
	return false
}

func (c *ProxyChecker) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if c.hasValidCookie(rw, req) {
		c.rproxy.ServeHTTP(rw, req)
	} else {
		rw.Write(nil)
	}
}

func main() {
	// args handlings
	flag.IntVar(&CookieMaxAge, "MaxAge", 3600*4, "proxy authentication cookie time out in seconds")
	flag.Parse()
	// begin proxy handings
	target, err := url.Parse("http://localhost")
	if err != nil {
		log.Fatal(err)
	}
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
	}
	rproxy := httputil.NewSingleHostReverseProxy(target)
	rproxy.Director = director
	rproxy.ErrorLog = log.New(os.Stdout, "", log.LstdFlags)
	// that's it! our reverse proxy is ready!
	s := &http.Server{
		Addr:    ":8088",
		Handler: newProxyChecker(rproxy),
	}

	s.ListenAndServe()
}

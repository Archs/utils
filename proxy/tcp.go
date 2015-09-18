// TcpProxy project main.go
package main

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("missing message!")
		return
	}
	ip := os.Args[1]
	port, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("error happened ,exit")
		return
	}
	addr := os.Args[3]
	host := "Host: " + addr

	Service(ip, port, addr, host)
}

func Service(ip string, port int, dstaddr string, dsthost string) {
	// listen and accept
	listen, err := net.ListenTCP("tcp", &net.TCPAddr{net.ParseIP(ip), port, ""})
	if err != nil {
		fmt.Println("listen error: ", err.Error())
		return
	}
	fmt.Println("init done...")

	for {
		client, err := listen.AcceptTCP()
		if err != nil {
			fmt.Println("accept error: ", err.Error())
			continue
		}
		go Channal(client, dstaddr, dsthost)
	}
}

func Channal(client *net.TCPConn, addr string, rhost string) {

	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		fmt.Println("ResolveTCPAddr error: ", err.Error())
		client.Close()
		return
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Println("connection error: ", err.Error())
		client.Close()
		return
	}

	go ReadRequest(client, conn, rhost)
	ReadResponse(conn, client)
}

func ReadRequest(lconn *net.TCPConn, rconn *net.TCPConn, dsthost string) {
	for {
		buf := make([]byte, 10240)
		n, err := lconn.Read(buf)
		if err != nil {
			break
		}

		mesg := changeHost(string(buf[:n]), dsthost)
		// print request
		fmt.Println(mesg)
		rconn.Write([]byte(mesg))
	}
	lconn.Close()
}

func ReadResponse(lconn *net.TCPConn, rconn *net.TCPConn) {
	for {
		buf := make([]byte, 10240)
		n, err := lconn.Read(buf)
		if err != nil {
			break
		}

		// fmt.Println(string(buf[:n]))
		// rconn.Write(buf[:n])
		rmsg := changeUrl(string(buf[:n]), "http://localhost/", "http://localhost:8088/")
		rconn.Write([]byte(rmsg))
	}
	lconn.Close()
}

func changeUrl(response, oldUrl, newUrl string) string {
	return strings.Replace(response, oldUrl, newUrl, -1)
}

// change Host
func changeHost(request string, newhost string) string {
	reg := regexp.MustCompile(`Host[^\r\n]+`)
	return reg.ReplaceAllString(request, newhost)
}

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
)

func proxy(conn net.Conn) {
	if conn == nil {
		return
	}
	defer conn.Close()

	connBuf := bufio.NewReader(conn)
	req, err := http.ReadRequest(connBuf)
	if err != nil {
		return
	}

	if req.Method == "CONNECT" {

	} else {
		if req.Method == "POST" {
			cl := req.Header.Get("Content-Length")
			req.Header.Add("Connection", "keep-alive")
			length, _ := strconv.Atoi(cl)

			if length == 0 {
				log.Panic("内容长度不一致")
			}
		}
		req.Header.Del("Proxy-Connection")

		client, _ := net.Dial("tcp", req.Host)
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)

			}

			defer func() {
				if err := recover(); err != nil {
					fmt.Println(err)
				}
			}()
			client.Close()
		}()
		reqBuf := bufio.NewWriter(client)
		//写入头部
		reqBuf.WriteString(fmt.Sprintf("%s %s %s\r\n", req.Method, req.URL.Path, req.Proto))
		reqBuf.WriteString(fmt.Sprintf("Host: %s\r\n", req.Host))
		for head := range req.Header {
			reqBuf.WriteString(fmt.Sprintf("%s: %s\r\n", head, strings.Join(req.Header[head], ",")))
		}
		reqBuf.WriteString("\r\n")
		//写入body
		len := connBuf.Buffered()
		body, _ := connBuf.Peek(len)
		reqBuf.Write(body)
		err = reqBuf.Flush()

		bs, _ := ioutil.ReadAll(client)
		conn.Write(bs)

	}

}

func main() {
	l, err := net.Listen("tcp", ":1081")
	if err != nil {
		log.Panic(err)
	}

	for {
		client, err := l.Accept()
		if err != nil {
			log.Panic(err)
		}
		go proxy(client)
	}
}

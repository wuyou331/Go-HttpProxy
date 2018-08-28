package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
)

const (
	NewLine = "\r\n"
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
		req.Header.Del("Proxy-Connection")
		req.Header.Add("Connection", "close")

		//发起http请求
		client, _ := net.Dial("tcp", req.Host)
		defer func() {
			if client != nil {
				client.Close()
			}
		}()

		reqWriter := bufio.NewWriter(client)
		//写入头部
		reqWriter.WriteString(fmt.Sprintf("%s %s %s%s", req.Method, req.URL.Path, req.Proto, NewLine))
		reqWriter.WriteString(fmt.Sprintf("Host: %s%s", req.Host, NewLine))
		for head := range req.Header {
			reqWriter.WriteString(fmt.Sprintf("%s: %s%s", head, strings.Join(req.Header[head], ","), NewLine))
		}
		reqWriter.WriteString(NewLine)
		cl := req.Header.Get("Content-Length")
		length, _ := strconv.Atoi(cl)
		if length > 0 {
			//写入body
			body, _ := connBuf.Peek(length)
			reqWriter.Write(body)
		}

		if client != nil {
			reqWriter.Flush()
			rspReader := bufio.NewReader(client)
			var buf = make([]byte, 1024*10)
			for {
				length, err = rspReader.Read(buf)
				if length == 0 || err != nil {
					break
				}
				conn.Write(buf[0:length])
			}

		}

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

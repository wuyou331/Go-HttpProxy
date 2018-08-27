package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
)

func proxy(conn net.Conn) {
	if conn == nil {
		return
	}
	defer conn.Close()
	req, err := http.ReadRequest(bufio.NewReader(conn))
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

		var buf bytes.Buffer
		buf.WriteString(fmt.Sprintf("%s %s %s\r\n", req.Method, req.URL.Path, req.Proto))
		buf.WriteString(fmt.Sprintf("Host: %s\r\n", req.Host))
		for head := range req.Header {
			buf.WriteString(fmt.Sprintf("%s: %s\r\n", head, req.Header[head]))
		}
		var str = string(buf.Bytes())
		str += ""
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
		proxy(client)
	}
}

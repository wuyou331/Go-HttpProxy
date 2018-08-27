package main

import (
	"bufio"
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
		req.RequestURI = ""
		client := &http.Client{}

		resp, respErr := client.Do(req)
		if respErr != nil {
			return
		}
		defer resp.Body.Close()

		resp.Write(conn)
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

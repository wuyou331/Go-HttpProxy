package main

import (
	"fmt"
	"log"
	"net"
	"request"
)

func proxy(conn net.Conn) {
	if conn == nil {
		return
	}
	defer conn.Close()

	request, err := req.CreateRequest(conn)
	if err != nil {
		fmt.Println(err)
		return
	}

	if request.Method == "CONNECT" {
		request.Connect()
	} else {
		request.Proxy()
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

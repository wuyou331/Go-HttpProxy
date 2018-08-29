package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"req"
)

func proxy(conn net.Conn) {
	if conn == nil {
		return
	}
	defer conn.Close()

	var connReader = bufio.NewReader(conn)
	request, err := req.CreateRequest(connReader)
	if err != nil {
		fmt.Println(err)
		return
	}

	if request.Method == "Connect" {
		request.Connect(conn)
	} else {
		request.Proxy(conn, connReader)
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

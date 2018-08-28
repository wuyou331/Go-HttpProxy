package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
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

	var header = make(map[string]string)
	var method, proto string
	var host *url.URL
	var connReader = bufio.NewReader(conn)

	//读取请求头第一行
	if line, err := readLine(connReader); err != nil {
		if err != nil {
			fmt.Print(err)
			return
		}
	} else {
		args := strings.Split(line, " ")
		if len(args) != 3 {
			//非标准请求头
			fmt.Printf("首行非标准请求头:%s%s", line, NewLine)
			return
		}
		method, proto = args[0], args[2]
		host, err = url.Parse(args[1])
		if err != nil {
			fmt.Printf("Host非标准URL:%s%s", args[1], NewLine)
			return
		}

	}
	//组装header
	for {
		line, err := readLine(connReader)
		if err != nil {
			fmt.Print(err)
			return
		} else if len(line) == 0 {
			break
		}
		spitIndex := strings.IndexAny(line, ":")
		if spitIndex == -1 {
			//非标准请求头
			fmt.Printf("非标准请求头:%s%s", line, NewLine)
			return
		}
		header[line[0:spitIndex]] = strings.Trim(line[spitIndex+1:], " ")
	}

	if method == "CONNECT" {
		//隧道请求
		host := fmt.Sprintf("%s:%s", host.Scheme, host.Opaque)
		socket, err := net.Dial("tcp", host)
		if err != nil {
			fmt.Printf("服务器无法连接:%s%s", host, NewLine)
			return
		}
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 Connection Established%s%s", NewLine, NewLine)))
		go io.Copy(socket, conn)
		io.Copy(conn, socket)
	} else {
		//普通请求
		if strings.IndexByte(host.Host, byte(':')) == -1 {
			host.Host += ":80"
		}

		//修改请求头，不保持tcp连接
		delete(header, "Proxy-Connection")
		delete(header, "Connection")
		header["Connection"] = "close"

		//发起http请求

		httpClient, err := net.Dial("tcp", host.Host)
		if err != nil {
			fmt.Printf("服务器无法连接:%s%s", host.Host, NewLine)
			return
		}
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)
			}
			if httpClient != nil {
				httpClient.Close()
			}
		}()

		//写入头部
		httpClient.Write([]byte(fmt.Sprintf("%s %s %s%s", method, host.Path, proto, NewLine)))
		for head := range header {
			httpClient.Write([]byte(fmt.Sprintf("%s: %s%s", head, header[head], NewLine)))
		}
		httpClient.Write([]byte(NewLine))
		if cl, ok := header["Content-Length"]; ok {
			//获取body长度，写入	body
			if length, _ := strconv.Atoi(cl); length > 0 {
				buf, _ := connReader.Peek(length)
				httpClient.Write(buf[0:length])
			}
		}

		io.Copy(conn, httpClient)
	}

}

func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if len(line) >= 2 {
		line = line[:len(line)-2]
	}
	return line, err
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

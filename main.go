package main

import (
	"bufio"
	"fmt"
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
			fmt.Printf("首行非标准请求头:%s", line)
			return
		}
		method, proto = args[0], args[2]
		host, err = url.Parse(args[1])
		if err != nil {
			fmt.Printf("Host非标准URL:%s", args[1])
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
			fmt.Printf("非标准请求头:%s", line)
			return
		}
		header[line[0:spitIndex]] = strings.Trim(line[spitIndex+1:], " ")
	}

	//发起请求
	if method == "CONNECT" {

	} else {
		//修改请求头，不保持tcp连接
		delete(header, "Proxy-Connection")
		delete(header, "Connection")
		header["Connection"] = "close"

		//发起http请求
		httpClient, err := net.Dial("tcp", host.Host)
		if err != nil {
			fmt.Printf("服务器无法连接:%s", host.Host)
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

		reqWriter := bufio.NewWriter(httpClient)
		//写入头部
		reqWriter.WriteString(fmt.Sprintf("%s %s %s%s", method, host.Path, proto, NewLine))
		for head := range header {
			reqWriter.WriteString(fmt.Sprintf("%s: %s%s", head, header[head], NewLine))
		}
		reqWriter.WriteString(NewLine)
		if cl, ok := header["Content-Length"]; ok {
			//获取body长度，写入	body
			if len, _ := strconv.Atoi(cl); len > 0 {
				buf, _ := connReader.Peek(len)
				reqWriter.Write(buf[0:len])
			}
		}

		//有时候数据写完，socket即断开连接
		if httpClient != nil {
			reqWriter.Flush()
			rspReader := bufio.NewReader(httpClient)
			var buf = make([]byte, 1024*100)
			for {
				len, err := rspReader.Read(buf)
				if len == 0 || err != nil {
					break
				}
				conn.Write(buf[0:len])
			}
		}

	}

}

func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	return line[:len(line)-2], err
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

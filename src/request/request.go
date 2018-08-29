package req

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
)

const (
	NewLine = "\r\n"
)

type Request struct {
	URL    *url.URL
	Method string
	Proto  string
	Header map[string]string
}

//CreateRequest 打包Request
func CreateRequest(reader *bufio.Reader) (*Request, error) {

	var request = Request{}
	//读取请求头第一行
	if line, err := readLine(reader); err != nil {
		if err != nil {
			return nil, err
		}
	} else {
		args := strings.Split(line, " ")
		if len(args) != 3 {
			return nil, errors.New(fmt.Sprintf("首行非标准请求头:%s%s", line, NewLine))
		}
		request.Method, request.Proto = args[0], args[2]
		request.URL, err = url.Parse(args[1])
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Host非标准URL:%s%s", args[1], NewLine))
		}
	}
	request.Header = make(map[string]string)
	//组装header
	for {
		line, err := readLine(reader)
		if err != nil {
			return nil, err
		} else if len(line) == 0 {
			break
		}
		spitIndex := strings.IndexAny(line, ":")
		if spitIndex == -1 {
			return nil, errors.New(fmt.Sprintf("非标准请求头:%s%s", line, NewLine))
		}
		request.Header[line[0:spitIndex]] = strings.Trim(line[spitIndex+1:], " ")
	}
	return &request, nil
}

func (request *Request) Connect(conn net.Conn) {
	//隧道请求
	host := fmt.Sprintf("%s:%s", request.URL.Scheme, request.URL.Opaque)
	socket, err := net.Dial("tcp", host)
	if err != nil {
		fmt.Printf("服务器无法连接:%s%s", host, NewLine)
		return
	}
	conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 Connection Established%s%s", NewLine, NewLine)))
	go io.Copy(socket, conn)
	io.Copy(conn, socket)
}

func (request *Request) Proxy(conn net.Conn, reader *bufio.Reader) {
	//普通请求
	if len(request.URL.Host) == 0 {
		request.URL.Host = request.URL.Scheme
	}

	if strings.IndexByte(request.URL.Host, byte(':')) == -1 {
		request.URL.Host += ":80"
	}

	//修改请求头，不保持tcp连接
	delete(request.Header, "Proxy-Connection")
	request.Header["Connection"] = "close"

	//发起http请求

	httpClient, err := net.Dial("tcp", request.URL.Host)
	if err != nil {
		fmt.Printf("服务器无法连接:%s%s", request.URL.Host, NewLine)
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
	httpClient.Write([]byte(fmt.Sprintf("%s %s %s%s", request.Method, request.URL.Path, request.Proto, NewLine)))
	for head := range request.Header {
		httpClient.Write([]byte(fmt.Sprintf("%s: %s%s", head, request.Header[head], NewLine)))
	}
	httpClient.Write([]byte(NewLine))
	if cl, ok := request.Header["Content-Length"]; ok {
		//获取body长度，写入	body
		if length, _ := strconv.Atoi(cl); length > 0 {
			buf, _ := reader.Peek(length)
			httpClient.Write(buf[0:length])
		}
	}

	io.Copy(conn, httpClient)
}

func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if len(line) >= 2 {
		line = line[:len(line)-2]
	}
	return line, err
}

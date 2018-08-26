package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
)

func proxy(client net.Conn) {
	if client == nil {
		return
	}
	defer client.Close()

	body, err := ioutil.ReadAll(client)
	if err != nil {
		log.Println(err)
	}

	var method, host string
	var header = getHeader(body)

	fmt.Sscanf(header[0][1], "%s%s", &method, &host)
	hostPortURL, err := url.Parse(host)
	if err != nil {
		log.Println(err)
		return
	}
	if hostPortURL.Opaque == "443" {

	}

	http.NewRequest(method, host, strings.NewReader(""))
	//if err != nil {
	//	// handle error
	//}
	//r.Header.Del("Proxy-Connection")
	//req.Header=r.Header
	//
	//resp, err := client.Do(req)
	//
	//defer resp.Body.Close()
	//rHead := w.Header()
	//for h:= range resp.Header{
	//
	//	rHead.Add(h,resp.Header.Get(h))
	//}
	//
	//body, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	// handle error
	//}
	//
	//w.Write(body)
}

func getHeader(body []byte) map[string][]string {
	var req *http.Request

	var index = 0
	var header = make(map[string][]string)
	for {
		end := bytes.IndexAny(body[index:], "\r\n")
		if end == -1 {
			break
		}
		txt := string(body[index : index+end])
		if len(txt) == 0 {
			break
		}
		var s []string
		if len(header) == 0 {
			s = strings.Split(txt, " ")
		} else {
			s = strings.Split(txt, ":")
		}
		header[s[0]] = s[1:]

		index = index + end + 2
	}
	return header
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

package main

import (
	"fmt"
	"log"
	"net"

	"github.com/hakkiir/httpfromtcp/internal/request"
)

func main() {

	tcpListener, err := net.Listen("tcp4", "127.0.0.1:42069")
	if err != nil {
		fmt.Println(err)
	}
	defer tcpListener.Close()

	for {
		conn, err := tcpListener.Accept()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("connection accepted!")

		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n", req.RequestLine.Method, req.RequestLine.RequestTarget, req.RequestLine.HttpVersion)
		fmt.Printf("Headers:\n")
		for key, val := range req.Headers {
			fmt.Printf("- %s: %s\n", key, val)
		}
		fmt.Println("Body:")
		fmt.Println(string(req.Body))
	}
}

/*
func getLinesChannel(c io.ReadCloser) <-chan string {
	strChan := make(chan string)
	go func() {
		b8 := make([]byte, 8)
		var curline string

		for i, err := c.Read(b8); err != io.EOF; {
			strings := strings.Split(string(b8[0:i]), "\n")
			if len(strings) > 1 {
				for _, string := range strings[0 : len(strings)-1] {
					strChan <- curline + string
					curline = ""
				}
			}
			curline += strings[len(strings)-1]
			i, err = c.Read(b8)
		}
		if curline != "" {
			strChan <- curline
		}

		close(strChan)
		c.Close()
	}()
	return strChan
}
*/

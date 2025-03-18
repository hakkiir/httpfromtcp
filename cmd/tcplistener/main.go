package main

import (
	"fmt"
	"io"
	"net"
	"strings"
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
		channel := getLinesChannel(conn)
		for line := range channel {
			fmt.Printf("%s\n", line)
		}
		fmt.Println("connection closed")
	}
}

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

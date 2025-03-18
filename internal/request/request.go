package request

import (
	"bytes"
	"fmt"
	"io"
	"slices"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const crlf = "\r\n"

func RequestFromReader(reader io.Reader) (*Request, error) {

	req, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	requestLine, err := parseRequestLine(req)
	if err != nil {
		return nil, err
	}
	return &Request{
		RequestLine: *requestLine,
	}, err

}

func parseRequestLine(data []byte) (*RequestLine, error) {
	index := bytes.Index(data, []byte(crlf))
	if index == -1 {
		return nil, fmt.Errorf("could not find CRLF in request line")
	}
	reqLineStr := string(data[:index])
	requestLine, err := requestLineFromString(reqLineStr)
	if err != nil {
		return nil, err
	}

	return requestLine, nil
}

func requestLineFromString(str string) (*RequestLine, error) {

	methods := []string{"GET", "POST", "PUT", "DELETE"}

	parts := strings.Split(str, " ")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid number of parts in request line: %s", str)
	}

	method := parts[0]
	if !slices.Contains(methods, method) {
		return nil, fmt.Errorf("invalid method: %s", method)
	}

	requestPath := parts[1]

	httpVersionParts := strings.Split(parts[2], "/")
	if len(httpVersionParts) != 2 {
		return nil, fmt.Errorf("malformed start-line: %s", str)
	}

	httpPart := httpVersionParts[0]
	if httpPart != "HTTP" {
		return nil, fmt.Errorf("unknown HTTP version: %s", httpPart)
	}

	httpVersionNumber := httpVersionParts[1]
	if httpVersionNumber != "1.1" {
		return nil, fmt.Errorf("unknown HTTP version: %s", httpPart)
	}

	return &RequestLine{
		HttpVersion:   httpVersionNumber,
		RequestTarget: requestPath,
		Method:        method,
	}, nil
}

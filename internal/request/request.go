package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/hakkiir/httpfromtcp/internal/headers"
)

type Request struct {
	RequestLine RequestLine
	State       ParserState
	Headers     headers.Headers
	Body        []byte
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const crlf = "\r\n"

type ParserState int

const (
	requestStateInitialized ParserState = iota
	requestStateDone
	requestStateParsingHeaders
	requestStateParsingBody
)
const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize, bufferSize)
	readToIndex := 0
	req := &Request{
		State:   requestStateInitialized,
		Headers: make(headers.Headers),
	}
	for req.State != requestStateDone {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		numBytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(io.EOF, err) {
				if req.State != requestStateDone {
					return nil, fmt.Errorf("incomplete request, in state: %d, read n bytes on EOF: %d", req.State, numBytesRead)
				}
				break
			}
			return nil, err
		}
		readToIndex += numBytesRead

		numBytesParsed, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[numBytesParsed:])
		readToIndex -= numBytesParsed
	}
	return req, nil
}

func parseRequestLine(data []byte) (numBytes int, rl *RequestLine, err error) {
	index := bytes.Index(data, []byte(crlf))
	if index == -1 {
		return 0, nil, nil
	}
	reqLineStr := string(data[:index])
	requestLine, err := requestLineFromString(reqLineStr)
	if err != nil {
		return 0, nil, err
	}

	return index + 2, requestLine, nil
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

func (r *Request) parse(data []byte) (int, error) {

	totalBytesParsed := 0
	for r.State != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		totalBytesParsed += n
		if n == 0 {
			break
		}
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.State {
	case requestStateInitialized:
		numBytes, requestLine, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if numBytes == 0 {
			return 0, nil
		}
		r.RequestLine = *requestLine
		r.State = requestStateParsingHeaders
		return numBytes, nil

	case requestStateParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.State = requestStateParsingBody
		}
		return n, nil

	case requestStateParsingBody:
		val := r.Headers.Get("Content-Length")
		if val == "" {
			r.State = requestStateDone
			return 0, nil
		}
		contentLen, err := strconv.Atoi(val)
		if err != nil {
			return 0, err
		}
		r.Body = append(r.Body, data...)
		if len(r.Body) > contentLen {
			return 0, fmt.Errorf("too long content")
		}
		if contentLen == len(r.Body) {
			r.State = requestStateDone
		}
		return len(data), nil

	case requestStateDone:
		return 0, fmt.Errorf("error: trying to read data in done state")

	default:
		return 0, fmt.Errorf("unknown state")
	}
}

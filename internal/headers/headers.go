package headers

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

const crlf = "\r\n"

type Headers map[string]string

func (h Headers) Get(key string) (len string) {

	key = strings.ToLower(key)

	if v, ok := h[key]; ok {
		return v
	}
	return ""
}

func (h Headers) Set(key, value string) {
	key = strings.ToLower(key)
	v, ok := h[key]
	if ok {
		value = strings.Join([]string{
			v,
			value,
		}, ", ")
	}
	h[key] = value
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {

	crlfIndex := bytes.Index(data, []byte(crlf))
	// if data starts with CRLF, headers are parsed
	if crlfIndex == 0 {
		return 2, true, nil
	}
	//not enought data
	if crlfIndex == -1 {
		return 0, false, nil
	}

	parts := bytes.SplitN(data[:crlfIndex], []byte(":"), 2)
	keyStr := string(parts[0])

	if keyStr[len(keyStr)-1] == ' ' {
		return 0, false, fmt.Errorf("malformed header name: %s", keyStr)
	}

	if len(strings.Split(keyStr, " ")) > 1 {
		return 0, false, fmt.Errorf("malformed header name: %s", keyStr)
	}

	val := strings.ToLower(string(bytes.TrimSpace(parts[1])))
	key := strings.ToLower(strings.TrimSpace(keyStr))

	pattern := "^[A-Za-z0-9!#$%&'*+\\-\\.\\^_`|~]+$"

	re := regexp.MustCompile(pattern)
	if !re.MatchString(key) {
		return 0, false, fmt.Errorf("invalid characters in key: %s", key)
	}

	if v, ok := h[key]; ok {
		h[key] = v + ", " + val
	} else {
		h[key] = val
	}
	return crlfIndex + 2, false, nil
}

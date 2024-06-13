package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
)

type Method int

const (
	MethodGet Method = iota + 1
	MethodPost
	MethodPut
	MethodDelete
)

type RequestInterface interface {
	Method() Method
	Path() string
	HeaderGet(key string) (string, bool)
	RouteTo(path string) bool
	IsPost() bool
	IsGet() bool
	ContentLength() int
	Data() []byte
}

func (m Method) String() string {
	switch m {
	case MethodGet:
		return "get"
	case MethodPost:
		return "post"
	case MethodPut:
		return "put"
	case MethodDelete:
		return "delete"
	}
	panic("Invalid method")
}

func MethodFromString(method string) (Method, error) {
	switch strings.ToLower(method) {
	case "get":
		return MethodGet, nil
	case "post":
		return MethodPost, nil
	case "put":
		return MethodPut, nil
	case "delete":
		return MethodDelete, nil
	}
	return 0, fmt.Errorf("Invalid method: %s", method)
}

func ParseRequest(reader io.Reader) (RequestInterface, error) {
	bufReader := bufio.NewReader(reader)
	requestLine, err := bufReader.ReadString('\n')
	if err != nil {
		log.Println("Error reading request line: ", err.Error())
		return &Request{}, err
	}
	var methodStr, path, protocol string
	fmt.Sscanf(requestLine, "%s %s %s", &methodStr, &path, &protocol)

	method, err := MethodFromString(methodStr)
	if err != nil {
		log.Println("Error parsing method: ", err.Error())
		return &Request{}, err
	}

	httpHeaders := make(map[string]string)
	for {
		headerLine, err := bufReader.ReadString('\n')
		if err != nil {
			log.Println("Error reading header line: ", err.Error())
			return &Request{}, err
		}
		headerLine = strings.TrimSuffix(headerLine, CRLF)
		if headerLine == "" {
			break
		}
		parts := strings.SplitN(headerLine, ":", 2)
		if len(parts) != 2 {
			log.Println("Invalid header line:", headerLine)
			continue
		}

		headerName := strings.TrimSpace(parts[0])
		headerValue := strings.TrimSpace(parts[1])
		httpHeaders[strings.ToLower(headerName)] = headerValue
	}

	var data []byte
	if contentLengthStr, ok := httpHeaders["content-length"]; ok {
		contentLength, err := strconv.Atoi(contentLengthStr)
		if err != nil {
			log.Println("Error parsing content-length: ", err.Error())
			return &Request{}, err
		}
		data = make([]byte, contentLength)
		_, err = io.ReadFull(bufReader, data)
		if err != nil {
			log.Println("Error reading request body: ", err.Error())
			return &Request{}, err
		}
	}

	return &Request{
		method:   method,
		path:     path,
		protocol: protocol,
		headers:  httpHeaders,
		data:     data,
	}, nil
}

type ResponseInterface interface{}

type Response struct{}

type ResponseBuilderInterface interface {
	SetProtocol(protocol string)
	SetStatus(status int)
	SetHeader(key, value string)
	SetBody(body []byte)
	Build() ResponseInterface
}

type Request struct {
	method   Method
	path     string
	protocol string
	headers  map[string]string
	data     []byte
}

func (r *Request) Method() Method {
	return r.method
}

func (r *Request) Path() string {
	return r.path
}

func (r *Request) HeaderGet(key string) (string, bool) {
	value, ok := r.headers[strings.ToLower(key)]
	return value, ok
}

func (r *Request) RouteTo(path string) bool {
	return strings.HasPrefix(r.path, path)
}

func (r *Request) IsPost() bool {
	return r.method == MethodPost
}

func (r *Request) IsGet() bool {
	return r.method == MethodGet
}

func (r *Request) ContentLength() int {
	if contentLengthStr, ok := r.headers["content-length"]; ok {
		contentLength, err := strconv.Atoi(contentLengthStr)
		if err != nil {
			log.Println("Error parsing content-length: ", err.Error())
			return 0
		}
		return contentLength
	}
	return 0
}

func (r *Request) Data() []byte {
	return r.data
}

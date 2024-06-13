package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
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
	CheckEncoding(encoding string) bool
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

type responseData struct {
	status   int
	headers  map[string]string
	body     []byte
	protocol string
}

type Response struct {
	data *responseData
}

var StatusMap = map[int]string{
	200: "OK",
	201: "Created",
	404: "Not Found",
	405: "Method Not Allowed",
	500: "Internal Server Error",
}

func StatusText(status int) string {
	if text, ok := StatusMap[status]; ok {
		return text
	}
	return "Unknown"
}

func (r Response) WriteTo(writer io.Writer) (int64, error) {
	// writer.Write([]byte(fmt.Sprintf("%s %d %s", r.data.protocol, r.data.status, StatusText(r.data.status)))
	writer.Write([]byte(fmt.Sprintf("%s %d %s", r.data.protocol, r.data.status, StatusText(r.data.status))))
	writer.Write([]byte(CRLF))

	var buf bytes.Buffer
	if r.data.body != nil {
		if r.data.headers["Content-Encoding"] == "gzip" {
			gzipWriter := gzip.NewWriter(&buf)
			gzipWriter.Write(r.data.body)
			gzipWriter.Close()
		} else {
			buf.Write(r.data.body)
		}
		r.data.headers["Content-Length"] = strconv.Itoa(len(buf.Bytes()))
	}
	for key, value := range r.data.headers {
		writer.Write([]byte(fmt.Sprintf("%s: %s", key, value)))
		writer.Write([]byte(CRLF))
	}
	writer.Write([]byte(CRLF))
	writer.Write(buf.Bytes())
	return 0, nil
}

type ResponseBuilderInterface interface {
	SetProtocol(protocol string) *ResponseBuilder
	SetStatus(status int) *ResponseBuilder
	SetHeader(key, value string) *ResponseBuilder
	SetBody(body []byte) *ResponseBuilder
	Build() io.WriterTo
}

type ResponseBuilder struct {
	data *responseData
}

func (r *ResponseBuilder) SetProtocol(protocol string) *ResponseBuilder {
	r.data.protocol = protocol
	return r
}

func (r *ResponseBuilder) SetStatus(status int) *ResponseBuilder {
	r.data.status = status
	return r
}

func (r *ResponseBuilder) SetHeader(key, value string) *ResponseBuilder {
	r.data.headers[key] = value
	return r
}

func (r *ResponseBuilder) SetBody(body []byte) *ResponseBuilder {
	r.data.body = body
	return r
}

func (r *ResponseBuilder) Build() io.WriterTo {
	return Response{}
}

type BuilderOption func(*ResponseBuilder)

func WithProtocol(protocol string) BuilderOption {
	return func(builder *ResponseBuilder) {
		builder.data.protocol = protocol
	}
}

func NewResponseBuilder(options ...BuilderOption) ResponseBuilderInterface {
	builder := ResponseBuilder{
		data: &responseData{
			headers:  make(map[string]string),
			protocol: "HTTP/1.1",
		},
	}

	for _, option := range options {
		option(&builder)
	}

	return &builder
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

func (r *Request) CheckEncoding(encoding string) bool {
	if value, ok := r.headers[ACCEPT_ENCODING]; ok {
		listEncodings := strings.Split(value, ",")
		for _, enc := range listEncodings {
			if strings.TrimSpace(strings.ToLower(enc)) == strings.ToLower(encoding) {
				return true
			}
		}
	}
	return false
}

func (r *Request) Data() []byte {
	return r.data
}

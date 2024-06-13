package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"

	// Uncomment this block to pass the first stage
	"net"
	"os"
)

const (
	CRLF            = "\r\n"
	CONTENT_LENGTH  = "content-length"
	ACCEPT_ENCODING = "accept-encoding"
	USER_AGENT      = "user-agent"
)

func main() {
	// read --directory /tmp/ in flag
	directory := flag.String("directory", "/tmp/", "the directory to serve files from")
	flag.Parse()
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go func(conn net.Conn) {
			defer conn.Close()
			request, err := ParseRequest(conn)
			if err != nil {
				log.Println("Error parsing request: ", err.Error())
				return
			}

			if request.Path() == "/" {
				conn.Write([]byte("HTTP/1.1 200 OK" + CRLF + CRLF))
			} else if request.RouteTo("/echo/") {
				var contentString string
				fmt.Sscanf(request.Path(), "/echo/%s", &contentString)
				contentLength := len(contentString)
				if value, ok := request.HeaderGet(ACCEPT_ENCODING); ok && (value == "gzip") {
					conn.Write([]byte("HTTP/1.1 200 OK" + CRLF + "Content-Type: text/plain" + CRLF + "Content-Encoding: gzip" + CRLF + "Content-Length: " + strconv.Itoa(contentLength) + CRLF + CRLF + contentString))
				} else {
					returnString := fmt.Sprintf("HTTP/1.1 200 OK%sContent-Type: text/plain%sContent-Length: %d%s%s", CRLF, CRLF, contentLength, CRLF+CRLF, contentString)
					conn.Write([]byte(returnString))
				}
			} else if request.Path() == "/user-agent" {
				userAgent, _ := request.HeaderGet(USER_AGENT)
				returnString := fmt.Sprintf("HTTP/1.1 200 OK%sContent-Type: text/plain%sContent-Length: %d%s%s", CRLF, CRLF, len(userAgent), CRLF+CRLF, userAgent)
				conn.Write([]byte(returnString))
			} else if strings.HasPrefix(request.Path(), "/files/") {
				fileName := strings.TrimPrefix(request.Path(), "/files/")
				filePath := *directory + fileName
				if request.IsGet() {
					if _, err := os.Stat(filePath); os.IsNotExist(err) {
						conn.Write([]byte("HTTP/1.1 404 Not Found" + CRLF + CRLF))
						return
					}
					file, err := os.Open(filePath)
					if err != nil {
						conn.Write([]byte("HTTP/1.1 500 Internal Server Error" + CRLF + CRLF))
						return
					}
					var contentString string
					scanner := bufio.NewScanner(file)
					for scanner.Scan() {
						contentString += scanner.Text()
					}
					contentLength := len(contentString)
					returnString := fmt.Sprintf("HTTP/1.1 200 OK%sContent-Type: application/octet-stream%sContent-Length: %d%s%s", CRLF, CRLF, contentLength, CRLF+CRLF, contentString)
					conn.Write([]byte(returnString))
				} else if request.IsPost() {
					// open or create file
					file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
					if err != nil {
						conn.Write([]byte("HTTP/1.1 500 Internal Server Error" + CRLF + CRLF))
						return
					}

					_, err = file.Write(request.Data())
					if err != nil {
						conn.Write([]byte("HTTP/1.1 500 Internal Server Error" + CRLF + CRLF))
						return
					}
					conn.Write([]byte("HTTP/1.1 201 Created" + CRLF + CRLF))

				} else {
					conn.Write([]byte("HTTP/1.1 405 Method Not Allowed" + CRLF + CRLF))
				}
			} else {
				conn.Write([]byte("HTTP/1.1 404 Not Found" + CRLF + CRLF))
			}
		}(conn)
	}
}

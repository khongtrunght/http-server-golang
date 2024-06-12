package main

import (
	"bufio"
	"fmt"
	"log"
	"strings"

	// Uncomment this block to pass the first stage
	"net"
	"os"
)

const CRLF = "\r\n"

func main() {
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
			reader := bufio.NewReader(conn)
			requestLine, err := reader.ReadString('\n')
			if err != nil {
				log.Println("Error reading request line: ", err.Error())
				return
			}
			fmt.Printf("Request: %s", requestLine)
			// Parse the request line
			var method, path, protocol string
			fmt.Sscanf(requestLine, "%s %s %s", &method, &path, &protocol)
			fmt.Printf("Method: %s, Path: %s, Protocol: %s\n", method, path, protocol)

			httpHeaders := make(map[string]string)
			for {
				headerLine, err := reader.ReadString('\n')
				if err != nil {
					log.Println("Error reading header line: ", err.Error())
					return
				}
				if headerLine == CRLF {
					break
				}
				fmt.Printf("Header: %s", headerLine)
				var headerName, headerValue string
				fmt.Sscanf(headerLine, "%s: %s", &headerName, &headerValue)
				httpHeaders[headerName] = headerValue
			}

			// Respond to the request
			if path == "/" {
				conn.Write([]byte("HTTP/1.1 200 OK" + CRLF + CRLF))
			} else if strings.HasPrefix(path, "/echo/") {
				var contentString string
				fmt.Sscanf(path, "/echo/%s", &contentString)
				contentLength := len(contentString)
				contentType := "text/plain"
				conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK%sContent-Length: %d%sContent-Type: %s%s%s", CRLF, contentLength, CRLF, contentType, CRLF, contentString)))
			} else {
				conn.Write([]byte("HTTP/1.1 404 Not Found" + CRLF + CRLF))
			}
		}(conn)
	}
}

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"strings"

	// Uncomment this block to pass the first stage
	"net"
	"os"
)

const CRLF = "\r\n"

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
				headerLine = strings.TrimSuffix(headerLine, CRLF)
				if headerLine == "" {
					break
				}
				fmt.Printf("Header: %q\n", headerLine)
				parts := strings.SplitN(headerLine, ":", 2)
				if len(parts) != 2 {
					log.Println("Invalid header line:", headerLine)
					continue
				}

				headerName := strings.TrimSpace(parts[0])
				headerValue := strings.TrimSpace(parts[1])
				httpHeaders[headerName] = headerValue
			}

			// print all headers
			for key, value := range httpHeaders {
				fmt.Printf("Header: %q: %q\n", key, value)
			}

			// Respond to the request
			if path == "/" {
				conn.Write([]byte("HTTP/1.1 200 OK" + CRLF + CRLF))
			} else if strings.HasPrefix(path, "/echo/") {
				var contentString string
				fmt.Sscanf(path, "/echo/%s", &contentString)
				contentLength := len(contentString)
				returnString := fmt.Sprintf("HTTP/1.1 200 OK%sContent-Type: text/plain%sContent-Length: %d%s%s", CRLF, CRLF, contentLength, CRLF+CRLF, contentString)
				conn.Write([]byte(returnString))
			} else if path == "/user-agent" {
				userAgent := httpHeaders["User-Agent"]
				returnString := fmt.Sprintf("HTTP/1.1 200 OK%sContent-Type: text/plain%sContent-Length: %d%s%s", CRLF, CRLF, len(userAgent), CRLF+CRLF, userAgent)
				conn.Write([]byte(returnString))
			} else if strings.HasPrefix(path, "/files/") {
				fileName := strings.TrimPrefix(path, "/files/")
				filePath := *directory + fileName
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
			} else {
				conn.Write([]byte("HTTP/1.1 404 Not Found" + CRLF + CRLF))
			}
		}(conn)
	}
}

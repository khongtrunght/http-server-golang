package main

import (
	"fmt"
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
			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil {
				fmt.Printf("Error reading from connection: %v", err)
				return
			}

			// get resourse path
			var path string
			fmt.Sscanf(string(buf[:n]), "GET %s HTTP/1.1", &path)
			if path == "/" {
				conn.Write([]byte("HTTP/1.1 200 OK" + CRLF + CRLF))
			} else {
				// 404
				conn.Write([]byte("HTTP/1.1 404 Not Found" + CRLF + CRLF))
			}
		}(conn)
	}
}

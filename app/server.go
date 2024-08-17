package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	tcp, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	fmt.Println("Server Running on Port 0.0.0.0:4221")
	defer tcp.Close()
	for {
		conn, err := tcp.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnections(conn)
	}
}

func handleConnections(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024)

	connBuffer, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Failed to read Request")
		log.Fatal(err)
	}
	req := string(buf[:connBuffer])
	lines := strings.Split(req, "\r\n")
	path := strings.Split(req, " ")[1]

	m := lines[0]
	method := strings.Split(m, " ")[0]
	if method == "GET" && path == "/" {
		response := "HTTP/1.1 200 OK\r\n\r\nHello world\r\n"
		conn.Write([]byte(response))
	} else if method == "GET" && strings.Split(path, "/")[1] == "greet" {
		message := strings.Split(path, "/")[2]
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(message), message)
		conn.Write([]byte(response))
	} else if method == "GET" && path == "/user-agent" {
		var userAgent string
		for _, line := range lines {
			if strings.HasPrefix(line, "User-Agent: ") {
				userAgent = strings.TrimPrefix(line, "User-Agent: ")
				break
			}
		}
		if userAgent != "" {
			response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent)*3, "Request User-Agent: "+userAgent)
			conn.Write([]byte(response))
		} else {
			conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\nUser-Agent not found"))
		}
	} else if method == "GET" && strings.Split(path, "/")[1] == "files" {
		file := strings.Split(path, "/")[2]
		str, err := os.ReadFile(file)
		if err != nil {
			fmt.Println(err)
			conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		} else {
			response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(str), string(str))
			conn.Write([]byte(response))
		}
	} else if method == "POST" && strings.Split(path, "/")[1] == "files" {
		content := strings.Split(req, "\r\n\r\n")
		if len(content) < 2 {
			conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\nMissing body content"))
			return
		}
		body := content[1]
		err := os.WriteFile("bar", []byte(body), 0644)
		if err != nil {
			conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\nFailed to save file"))
			return
		}
		conn.Write([]byte("HTTP/1.1 201 Created\r\n\r\nFile uploaded successfully"))
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}

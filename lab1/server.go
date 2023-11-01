package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
)

const maxChildProcesses = 10000

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: ./http_server <port>")
		os.Exit(1)
	}

	port := os.Args[1]
	listen, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	defer listen.Close()
	fmt.Printf("Server is listening on port %s\n", port)

	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err)
		return
	}

	request := string(buffer)
	requestLines := strings.Split(request, "\n")

	if len(requestLines) < 1 {
		return
	}

	requestLine := strings.Fields(requestLines[0])

	if len(requestLine) < 2 {
		return
	}

	method := requestLine[0]
	path := requestLine[1]

	if method == "GET" {
		handleGETRequest(conn, path)
	} else if method == "POST" {
		handlePOSTRequest(conn, path, requestLines)
	} else {
		sendResponse(conn, "HTTP/1.1 501 Not Implemented", "text/plain", []byte("501 Not Implemented"))
	}
}

func handleGETRequest(conn net.Conn, path string) {
	if path == "/" {
		path = "/index.html"
	}

	filePath := "." + path
	contentType := getContentType(path)

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			sendResponse(conn, "HTTP/1.1 404 Not Found", "text/plain", []byte("404 Not Found"))
		} else {
			sendResponse(conn, "HTTP/1.1 400 Bad Request", "text/plain", []byte("400 Bad Request"))
		}
		return
	}

	sendResponse(conn, "HTTP/1.1 200 OK", contentType, data)
}

func handlePOSTRequest(conn net.Conn, path string, requestLines []string) {
	if path == "/" {
		path = "/index.html"
	}

	filePath := "." + path
	contentType := getContentType(path)

	bodyStart := false
	body := []byte{}

	for _, line := range requestLines {
		if bodyStart {
			body = append(body, []byte(line)...)
		}
		if line == "\r\n" {
			bodyStart = true
		}
	}

	err := ioutil.WriteFile(filePath, body, 0644)
	if err != nil {
		sendResponse(conn, "HTTP/1.1 400 Bad Request", "text/plain", []byte("400 Bad Request"))
		return
	}

	sendResponse(conn, "HTTP/1.1 200 OK", contentType, []byte("File uploaded successfully"))
}

func getContentType(path string) string {
	if strings.HasSuffix(path, ".html") {
		return "text/html"
	} else if strings.HasSuffix(path, ".txt") {
		return "text/plain"
	} else if strings.HasSuffix(path, ".gif") {
		return "image/gif"
	} else if strings.HasSuffix(path, ".jpeg") || strings.HasSuffix(path, ".jpg") {
		return "image/jpeg"
	} else if strings.HasSuffix(path, ".css") {
		return "text/css"
	}
	return "text/plain"
}

func sendResponse(conn net.Conn, status, contentType string, content []byte) {
	header := fmt.Sprintf("%s\r\nContent-Type: %s\r\nContent-Length: %d\r\n\r\n", status, contentType, len(content))
	response := header + string(content)
	conn.Write([]byte(response))
}
package main

import (
	"bufio"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
)

const maxGoroutines = 10

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: ./http_server <port>")
		os.Exit(1)
	}
	port := os.Args[1]

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("Server is running on port", port)

	var wg sync.WaitGroup
	sem := make(chan struct{}, maxGoroutines)
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}
		sem <- struct{}{} // Acquire a token
		wg.Add(1)
		go func() {
			defer wg.Done()
			handleConnection(conn)
			<-sem // Release the token
			fmt.Println(sem)
		}()
	}
	wg.Wait()
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	request, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		fmt.Println("Error reading request:", err)
		sendResponse(conn, http.StatusBadRequest, "text/plain", "Bad Request\n")
		return
	}
	defer request.Body.Close()

	switch request.Method {
	case http.MethodGet:
		handleGet(conn, request)
	case http.MethodPost:
		handlePost(conn, request)
	default:
		sendResponse(conn, http.StatusNotImplemented, "text/plain", "Not Implemented\n")
	}
}

func handleGet(conn net.Conn, request *http.Request) {
	path := request.URL.Path
	if !isValidPath(path) {
		sendResponse(conn, http.StatusBadRequest, "text/plain", "Bad Request\n")
		return
	}

	data, err := os.ReadFile(path[1:]) // Removing the leading '/'
	if os.IsNotExist(err) {
		sendResponse(conn, http.StatusNotFound, "text/plain", "Not Found\n")
		return
	} else if err != nil {
		fmt.Println("Error reading file:", err)
		sendResponse(conn, http.StatusInternalServerError, "text/plain", "Internal Server Error\n")
		return
	}

	contentType := mime.TypeByExtension(path)
	sendResponse(conn, http.StatusOK, contentType, string(data))
}

func handlePost(conn net.Conn, request *http.Request) {
	path := request.URL.Path
	if !isValidPath(path) {
		sendResponse(conn, http.StatusBadRequest, "text/plain", "Bad Request\n")
		return
	}

	if isImage(path) {
		file, _, _ := request.FormFile("file")
		os.Mkdir("./uploaded/", 0777)
		saveFile, _ := os.OpenFile("./uploaded/"+path[1:], os.O_WRONLY|os.O_CREATE, 0666)
		io.Copy(saveFile, file)

		defer file.Close()
		defer saveFile.Close()
	} else {
		data, err := io.ReadAll(request.Body)
		if err != nil {
			fmt.Println("Error reading POST data:", err)
			sendResponse(conn, http.StatusInternalServerError, "text/plain", "Internal Server Error\n")
			return
		}

		err = os.WriteFile(path[1:], data, 0644) // Removing the leading '/'
		if err != nil {
			fmt.Println("Error writing file:", err)
			sendResponse(conn, http.StatusInternalServerError, "text/plain", "Internal Server Error\n")
			return
		}
	}

	sendResponse(conn, http.StatusOK, "text/plain", "OK\n")
}

func sendResponse(conn net.Conn, status int, contentType, body string) {
	response := http.Response{
		StatusCode: status,
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
	}
	response.Header.Set("Content-Type", contentType)
	response.Body = io.NopCloser(strings.NewReader(body))

	response.Write(conn)
}

func isValidPath(path string) bool {
	if !strings.HasPrefix(path, "/") {
		return false
	}
	ext := strings.ToLower(path)
	for _, validExt := range []string{".html", ".txt", ".gif", ".jpeg", ".jpg", ".css"} {
		if strings.HasSuffix(ext, validExt) {
			return true
		}
	}
	return false
}

func isImage(path string) bool {
	if !strings.HasPrefix(path, "/") {
		return false
	}
	ext := strings.ToLower(path)
	for _, validExt := range []string{".gif", ".jpeg", ".jpg"} {
		if strings.HasSuffix(ext, validExt) {
			return true
		}
	}
	return false
}

package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
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
		}()
	}
	wg.Wait()
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	request, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		log.Println(err)
		return
	}

	// Handle the request
	switch req.Method {
	case http.MethodGet:
		handleGETRequest(c, req)
	case http.MethodPost:
		handlePOSTRequest(c, req)
	default:
		// Return a "Not Implemented" error
		http.Error(c, "Not implemented", http.StatusNotImplemented)
	}

	// Close the connection
	c.Close()
}

// Handler for GET requests
func handleGETRequest(c net.Conn, req *http.Request) {
	// Get the file path from the request URL
	filePath := req.URL.Path

	// Check if the file is in the cache
	if contents, ok := fileCache.Load(filePath); ok {
		// If the file is in the cache, write it to the connection
		_, err := c.Write(contents.([]byte))
		if err != nil {
			log.Println(err)
			return
		}

		return
	}

	// If the file is not in the cache, read it from the local filesystem
	file, err := os.Open(filePath)
	if err != nil {
		// If the file does not exist, return a 404 error
		if os.IsNotExist(err) {
			http.NotFound(c, req)
			return
		}

		// If there is another error, log it and return a 500 error
		log.Println(err)
		http.Error(c, "Internal server error", http.StatusInternalServerError)
		return
	}

	defer file.Close()

	// Read the file contents into memory
	contents, err := io.Copy(os.Stdout, file)
	if err != nil {
		// If there is an error reading the file, log it and return a 500 error
		log.Println(err)
		http.Error(c, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Add the file contents to the cache
	fileCache.Store(filePath, contents)

	// Write the file contents to the connection
	_, err = c.Write(contents)
	if err != nil {
		log.Println(err)
		return
	}
}

// Handler for POST requests
func handlePOSTRequest(c net.Conn, req *http.Request) {
	// Read the request body
	body, err := io.Copy(os.Stdout, req.Body)
	if err != nil {
		log.Println(err)
		return
	}

	// Save the request body to a file
	filePath := "./post_request.txt"
	file, err := os.Create(filePath)
	if err != nil {
		log.Println(err)
		return
	}

	defer file.Close()

	_, err = file.Write(body)
	if err != nil {
		log.Println(err)
		return
	}

	// Write a response to the connection
	_, err = c.Write([]byte("POST request body saved to file"))
	if err != nil {
		log.Println(err)
		return
	}
}

func main() {
	// Create a new TCP listener
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	// Accept incoming connections and handle them concurrently
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handleGET(conn)
	}
}

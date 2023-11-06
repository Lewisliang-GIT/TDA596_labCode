package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

var customTransport = http.DefaultTransport

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// Create a new HTTP request with the same method, URL, and body as the original request
	targetURL := r.URL
	proxyReq, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
	if err != nil {
		http.Error(w, "Error creating proxy request", http.StatusInternalServerError)
		return
	}

	// Copy the headers from the original request to the proxy request
	for name, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}

	// Send the proxy request using the custom transport
	resp, err := customTransport.RoundTrip(proxyReq)
	if err != nil {
		http.Error(w, "Error sending proxy request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Copy the headers from the proxy response to the original response
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// Set the status code of the original response to the status code of the proxy response
	w.WriteHeader(resp.StatusCode)

	// Copy the body of the proxy response to the original response
	io.Copy(w, resp.Body)
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: ./http_server <port>")
		os.Exit(1)
	}
	port := os.Args[1]

	// Create a new HTTP server with the handleRequest function as the handler
	server := http.Server{
		Addr:    ":" + port,
		Handler: http.HandlerFunc(handleRequest),
	}

	// Start the server and log any errors
	log.Println("Starting proxy server on :" + port)
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("Error starting proxy server: ", err)
	}
}

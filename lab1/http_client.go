//NOT A USED FILE, JUST FOR TEST

package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

func main() {
	// Get the file and URL from the user
	fmt.Println("Enter the file path:")
	filePath := ""
	fmt.Scanln(&filePath)

	fmt.Println("Enter the URL to upload or download the file:")
	uploadURL := ""
	fmt.Scanln(&uploadURL)

	// Ask the user to choose POST or GET
	fmt.Println("Do you want to upload or download the file? (POST/GET)")
	method := ""
	fmt.Scanln(&method)

	// Create a new HTTP client
	client := &http.Client{}

	// Create a new request based on the user's choice
	switch method {
	case "POST":
		// Create a new POST request with a file upload
		postRequest, err := http.NewRequest(http.MethodPost, uploadURL, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		// Create a new buffer to write the multipart request to
		buffer := bytes.Buffer{}

		// Create a new multipart writer
		multipartWriter := multipart.NewWriter(&buffer)

		// Open the file to upload
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Create a new form field for the file upload
		filePart, err := multipartWriter.CreateFormFile("file", filePath)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Copy the file contents to the multipart writer
		defer file.Close()
		io.Copy(filePart, file)

		// Close the multipart writer
		defer multipartWriter.Close()

		// Set the request headers
		postRequest.Header.Set("Content-Type", multipartWriter.FormDataContentType())

		// Send the POST request and get the response
		response, err := client.Do(postRequest)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Read the response body
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Print the response body
		fmt.Println(string(body))

	case "GET":
		// Create a new GET request
		getRequest, err := http.NewRequest(http.MethodGet, uploadURL, nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Send the GET request and get the response
		response, err := client.Do(getRequest)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Read the response body
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Print the response body
		fmt.Println(string(body))

	default:
		fmt.Println("Invalid method")
		return
	}
}

// package main

// import (
// 	"bufio"
// 	"fmt"
// 	"net"
// 	"os"
// 	"strings"
// )

// func main() {
// 	arguments := os.Args
// 	if len(arguments) == 1 {
// 		fmt.Println("Please provide host:port.")
// 		return
// 	}

// 	CONNECT := arguments[1]
// 	c, err := net.Dial("tcp", CONNECT)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	for {
// 		reader := bufio.NewReader(os.Stdin)
// 		fmt.Print(">> ")
// 		text, _ := reader.ReadString('\n')
// 		fmt.Fprintf(c, text+"\n")

// 		message, _ := bufio.NewReader(c).ReadString('\n')
// 		fmt.Print("->: " + message)
// 		if strings.TrimSpace(string(text)) == "STOP" {
// 			fmt.Println("TCP client exiting...")
// 			return
// 		}
// 	}
// }

// package main

// import (
// 	"fmt"
// 	"io"
// 	"log"
// 	"net/http"
// )

// func main() {
// 	client := &http.Client{}
// 	req, err := http.NewRequest("GET", "http://example.com", nil)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer resp.Body.Close()
// 	bodyText, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	fmt.Printf("%s\n", bodyText)
// }

// func main() {
// 	form := new(bytes.Buffer)
// 	writer := multipart.NewWriter(form)
// 	fw, err := writer.CreateFormFile("file", filepath.Base("asd.css"))
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	fd, err := os.Open("asd.css")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer fd.Close()
// 	_, err = io.Copy(fw, fd)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	writer.Close()

// 	client := &http.Client{}
// 	req, err := http.NewRequest("POST", "http://localhost:1234/asd.css", form)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	req.Header.Set("Content-Type", writer.FormDataContentType())
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer resp.Body.Close()
// 	bodyText, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	fmt.Printf("%s\n", bodyText)
// }

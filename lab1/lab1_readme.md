# lab1

Our program is running in go version 1.21.4.

In this lab, we have two code file here.

`http_server.go` is the code of the http server. `proxy.go` is the code of the proxy.

## `http_server.go`

The lab requires us not to use *http.ListenAnd Serve* to deploying the http server. We use *net.Listen* to listen incomming TCP connecting and analyzing what the http request header is. Depending on the require of the incoming http we split two functions. One is handle *'GET'* method and one is used to handle *'POST'* method.

In the beginning we read in input from user as the port number which can the server running on it. Then we listen all the incomming conneciton. We also defind a waiting group `var wg sync.WaitGroup` to limited the incomming connection only can have 10 connections at same time. We use a lock to implementing this function.

```go
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
```

Inside the `func handleConnection(conn net.Conn)` we use a switch to figure out the incomming connecting is *POST* or *GET*.

We read the connction and write it into buffer. `request, err := http.ReadRequest(bufio.NewReader(conn))`

Then depending the request meth, we create tow functions `handleGet(conn, request)` and `handlePost(conn, request)`. If it is other type methods, we will write `http.StatusNotImplemented` into the http response.

### `handleGet(conn, request)`

We use `request.URL.Path` as the file name also as the file identifier. `isValidPath(path)` will check suffix of the file and `IsNotExist` checks the file exist in our system or not. `Bad Request` and `Not Found` will write into the http responde is not pass those check.

If the get file is correct. A data string `data, err := os.ReadFile(path[1:])` will open and write this string into http respond. So that the user can grt the correct file from the server.

#### Function Parameter

* `handleGet` takes two parameters: conn net.Conn and request *http.Request.
* `conn` is a network connection to the client. This is used to send responce back to the client.
* `request` is a pointer to an `http.Request` object, which represents the HTTP request received from the client.

#### Path Validation

The function first retrieves the requested path from the URL: path := request.URL.Path.
It then checks if this path is valid using the isValidPath function. This is important for security, to prevent path traversal attacks and ensure the requested path is something the server is willing to serve.
If the path is not valid, the function sends a "400 Bad Request" response to the client and returns early.

#### Reading the Requested File:

If the path is valid, the function attempts to read the file corresponding to the requested path using `os.ReadFile(path[1:])`
The [1:] slice operation removes the leading / from the path. This is because the file system paths typically don't start with a /, but URLs do.
It handles errors that might occur during file reading:
Note that if an error occurs, the function returns early after sending the appropriate response.

#### Determining the current type

The function uses `mime.TypeByExtension(path)` to determine the MIME type of the file based on its extension. This is important for the client (usually a web browser) to understand what type of file is being sent.
If the MIME type cannot be determined, `TypeByExtension` returns an empty string, which could lead the client to make its own determination about the file type.

#### Sending the responce

Finally, if everything is successful, the function sends an `"200 OK"` response along with the content of the file.
The content type (determined earlier) and the file data are included in the response.

#### Security Consideration

The function assumes that the path provided is safe after running through `isValidPath`. However, it's crucial to ensure that `isValidPath` robustly prevents directory traversal attacks. For instance, a path like `/../secret.txt` should not be allowed.

### `handlePost(conn, request)`

In the funcion of handle the *POST*. Same as `handleGet(conn, request)`, we will check the file type to *POST* is correct or not. After that we read the file from **file** from the http reqiure. Then we save the post file into the directory.

#### Function paramter

* `handlePost` takes two parameters: conn net.Conn and request *http.Request.
* `conn` is a network connection to the client, used for sending responses back.
* `request` is a pointer to an HTTP request, representing the client's request.

#### Path valiation

The function checks if the requested path corresponds to an image using isImage.
If the path is for an image, the function processes a file upload:  
It retrieves the file from the form-data `(file, _, _ := request.FormFile("file")). Error handling is omitted here (_, _)`, which is risky as it can lead to panics or undefined behavior if the form file isn't present.  
It creates (or ensures the existence of) a directory named uploaded where the images will be stored.  
It then opens (or creates) a file in the uploaded directory corresponding to the requested path and writes the uploaded file data to it using `io.Copy`.  
File handles for both the uploaded file and the new file are deferred to close at the end of the function execution.

#### Handling Image Uploads

If the path is not for an image, the function appears to intend a different handling for non-image files (as suggested by the commented-out code), but this part is not active.
Instead, it treats non-image files similarly to image files:
A file is retrieved from the form-data.
A file directory is created (or checked for existence).
The file is saved in the file directory.

#### Handling Non-Image File Uploads

If the path is not for an image, the function appears to intend a different handling for non-image files (as suggested by the commented-out code), but this part is not active.

Instead, it treats non-image files similarly to image files:
A file is retrieved from the form-data.
A file directory is created (or checked for existence).
The file is saved in the file directory.

#### Response

After processing the upload, the function sends an `"200 OK"` response with a plain text message `"OK\n"`.

## proxy.go

This code implements a simple proxy server that can forward HTTP requests to other servers. The server listens on a specified port and accepts incoming connections. When a connection is accepted, the server reads the request line from the client and parses it to determine the destination host and port. The server then establishes a connection to the destination host and forwards the request header and body to the destination server. The server then copies data between the client and the destination server until the connection is closed.

### Sever struct and functions

```go
addr     string
listener net.Listener
```

contains the address of the server and a listener.

`start` function:  
Call `net.Listen` to listen on the specified port and get the listener.  
Enter the `runLoop` loop.

`runLoop` funtion:  
Continuously accepts client connections.  
For each connection, the `handleConnection` method is called to handle the connection.

`handleConnection`function:  
Create a new connection instance `c`.  
Call `c.serve()` to handle the connection.

### Connection struct and functions

```go
reqConn net.Conn
```

contains a virable of type net.Conn representing the connection.

`serve` function:  
Call readRequestInfo to read the request information, get the destination address and whether it is a secure request.  
Send the request header to the target server.  
Establish a connection to the target server remoteConn.  
Start two goroutines, execute io.Copy(remoteConn, c.reqConn) and io.Copy(c.reqConn, remoteConn) respectively for bi-directional data transfer.

### Other functions

`newConnection` function:  
Creates and returns a new connection instance, passing in the connection `conn`.

`NewServer` function:  
Creates a new proxy server instance.

`readRequestInfo`:  
Call poachRequestLine to read the request line and get the remaining data and the request line string.  

Parses the request line to get the request method, path, HTTP version and whether it is a secure request.  

If it is a CONNECT request, call drainConnectRequestHeader to read the CONNECT request header.

`poachRequestLine`:  
Keep reading the data in the connection until the request line terminator "\r\n" is found.  

Returns the request line string and the remaining data.

`parseRequestLine`:  
Parses the request line string, returning the request method, path, HTTP version, and parsing result.

`drainConnectRequestHeader`:  
Reads the data in the *CONNECT* request header to ensure that there are no residuals from the TLS handshake.

### Main function of proxy.go

```go
func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: ./http_server <port>")
		os.Exit(1)
	}
	input, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Print(err)
	}
	port := flag.Int("port", input, "listening port number")
	flag.Parse()

	server := NewServer(*port)
	server.Start()
}
```

Parses command line arguments for port numbers.  
Create a proxy server instance.  
Call `server.Start()` to start the proxy server.

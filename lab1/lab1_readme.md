TODO:
- [ ] discription of proxcy

# lab1
In this lab, we have two code file here. 

*http_server.go* is the code of the http server. *proxy.go* is the code of the proxy.

## http_server.go
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

### `handlePost(conn, request)`

In the funcion of handle the *POST*. Same as `handleGet(conn, request)`, we will check the file type to *POST* is correct or not. After that we read the file from **file** from the http reqiure. Then we save the post file into the directory.  

## proxcy.go ~~以下内容仅为初步草稿~~

This code implements a simple proxy server that can forward HTTP requests to other servers. The server listens on a specified port and accepts incoming connections. When a connection is accepted, the server reads the request line from the client and parses it to determine the destination host and port. The server then establishes a connection to the destination host and forwards the request header and body to the destination server. The server then copies data between the client and the destination server until the connection is closed.

### Code Structure
The code is organized into several functions and structs:

`NewServer`: Creates a new proxy server instance.
`server.Start`: Starts the proxy server listening on a specified port.
`server.handleConnection`: Handles an incoming connection from a client.
`connection.serve`: Processes an incoming connection from a client.
`readRequestInfo`: Reads and parses the request information from a client connection.
`poachRequestLine`: Reads the request line from a client connection and extracts the request method, path, and version.
parseRequestLine: Parses the request line into its components.
drainConnectRequestHeader: Drains the CONNECT request header from a client connection.
`newConnection`: Creates a new connection instance.

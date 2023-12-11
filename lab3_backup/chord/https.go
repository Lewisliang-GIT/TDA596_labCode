package chord

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
)

/* HTTP Server Functions*/

// Gets rid of annoying favicon requests
func faviconHandler(w http.ResponseWriter, r *http.Request) {
	return
}

// serverHandler Main handler for http requests to server, with sync mechanism
func ServerHandler(w http.ResponseWriter, r *http.Request) {

	if !checkMethod(w, r) {
		return
	}

	res, err := httputil.DumpRequest(r, true)
	if err != nil {
		return
	}
	fmt.Print(string(res) + "\n")

	if r.Method == "POST" {
		postHandler(w, r)
	} else {
		getHandler(w, r, getFileType(path.Base(r.URL.Path)))
	}
}

// Handles GET requests
func getHandler(w http.ResponseWriter, r *http.Request, fType string) {
	if path.Base(r.URL.Path) == "/" {
		fmt.Fprintf(w, "Hello, Welcome to the Main Page")
		return
	} else if !validFileType(w, fType) {
		return
	}

	if fileExists(path.Base(r.URL.Path)) {
		http.ServeFile(w, r, "./"+r.URL.String())
		return
	}
	//proxyHandler(w, r)
}

// postHandler Handles POST requests
func postHandler(w http.ResponseWriter, r *http.Request) {
	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)
	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, handler, err := r.FormFile("myFile")
	if err != nil {
		http.Error(w, "Error on retrieving file <POST>: "+err.Error(), 500)
		return
	}
	defer file.Close()

	//check for valid file type
	if !validFileType(w, getFileType(handler.Filename)) {
		return
	}

	// Create Local file
	f, err := os.Create(handler.Filename)
	if err != nil {
		http.Error(w, "Error on Create file <POST>: "+err.Error(), 500)
		return
	}
	defer f.Close()

	// Write to local file within our directory that follows
	_, err = io.Copy(f, file)
	if err != nil {
		http.Error(w, "Error on writing to file <POST>: "+err.Error(), 500)
		return
	}

	// return that we have successfully uploaded client file!
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Successfully Uploaded File,\nFeel free to visit again!!"))

	/**/
	//fileKey := Hash(handler.Filename).String()
	belongsTo, _ := RPCFindSuccessor(handler.Filename, Hash(handler.Filename))

	//_, ok := cNode.Bucket[fileKey]
	//_, ok1 := cNode.Backups[fileKey]
	//if belongsTo.Id == cNode.LocalNode.Id || ok {
	//	fmt.Println("<Received file>")
	//	cNode.Bucket[fileKey] = handler.Filename
	//	return
	//} else if ok1 {
	//	fmt.Println("Received file backup")
	//	cNode.Backups[fileKey] = handler.Filename
	//	return
	//}

	fmt.Println("<Received then sent file>")
	postSender(belongsTo, handler.Filename)
	/**/
}

// Check for valid Http Requesting Method
func checkMethod(w http.ResponseWriter, r *http.Request) bool {
	switch r.Method {
	case "GET":
		return true
	case "POST":
		return true
	default:
		http.Error(w, "Request Method Is Currently Not Supported <"+r.Method+">", 501)
		return false
	}
}

// Serves get requests by forwarding them to the Main server and sends back response
//func proxyHandler(w http.ResponseWriter, r *http.Request) {
//	forwardtoNode := cNode.lookUp(path.Base(r.URL.Path)).Address
//	if forwardtoNode == cNode.LocalNode.Address {
//		http.ServeFile(w, r, "/"+r.URL.String())
//		return
//	}
//
//	forwardtoNodeAddress, _ := setHttpsPort(forwardtoNode)
//	forwardToServerURL := "https://" + forwardtoNodeAddress
//	client, err := HttpsClient("secure_chord.crt")
//	if err != nil {
//		http.Error(w, "Error in setting https client for main server: "+err.Error(), 500)
//		return
//	}
//	resp, err := client.Get(forwardToServerURL + "/" + r.URL.String())
//	if err != nil {
//		http.Error(w, "Failure in forwarding request: "+err.Error(), 400)
//		return
//	}
//
//	defer resp.Body.Close()
//	copyHeader(w.Header(), resp.Header)
//	w.WriteHeader(resp.StatusCode)
//	_, err = io.Copy(w, resp.Body)
//	if err != nil {
//		http.Error(w, "Error in forwarding main server response: "+err.Error(), 500)
//		return
//	}
//}

// Get file type from request file name
func getFileType(filename string) string {
	var fType string
	var extension = filepath.Ext(filename)
	if len(extension) > 0 {
		fType = extension[1:]
	} else {
		fType = extension
	}
	return fType
}

// Check for a valid file type
func validFileType(w http.ResponseWriter, fType string) bool {
	switch fType {
	case "html", "txt", "gif", "jpeg", "jpg", "css":
		return true
	default:
		http.Error(w, "File type not supported", 400)
		return false
	}
}

// Copy source headers to destination header (use for sending back main server response)
func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func fileExists(fileName string) bool {
	if _, err := os.Stat(fileName); err == nil {
		return true
	} else {
		fmt.Printf("File does not exist\n")
		return false
	}
}

// setHttpsPort increments address to the https port
func setHttpsPort(address string) (string, error) {
	u, err := url.Parse("http://" + address)
	if err != nil {
		fmt.Println("URL paring error:", err)
		return "", err
	}

	p, err := strconv.Atoi(u.Port())
	if err != nil {
		fmt.Println("URL paring error <port>:", err)
		return "", err
	}

	newAddress := u.Hostname() + ":" + strconv.Itoa(p+1) + u.Path
	return newAddress, nil
}

// HttpsClient sets the https client
func HttpsClient(crt string) (*http.Client, error) {
	f, err := os.Open(crt)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	caCert, err := os.ReadFile(crt)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	cert, err := tls.LoadX509KeyPair("client.crt", "client.key")
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:            caCertPool,
				Certificates:       []tls.Certificate{cert},
				InsecureSkipVerify: true,
			},
		},
	}
	return client, nil
}

// httpsServer sets the https server
func HttpsServer(crt string) (*http.Server, error) {
	f, err := os.Open(crt)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	caCert, err := os.ReadFile(crt)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	cfg := &tls.Config{
		//ClientAuth: tls.RequireAndVerifyClientCert,
		ClientAuth: tls.VerifyClientCertIfGiven,
		ClientCAs:  caCertPool,
	}
	srv := &http.Server{
		Addr:      "",
		Handler:   nil,
		TLSConfig: cfg,
	}
	return srv, nil
}

var maxAttempts int = 5

// postSender sends files to nodes through http
func postSender(address string, filePath string) {
	for i := 0; i < maxAttempts; i++ {
		err := postSender1(address, filePath)
		if err != nil {
			continue
		}
		break
	}
}

func postSender1(address string, filePath string) error {
	client, err := HttpsClient("secure_chord.crt")
	if err != nil {
		fmt.Println("Error on setting https client;", err)
		return err
	}
	newAddress, _ := setHttpsPort(address)
	uri := "https://" + newAddress
	fileName := path.Base(filePath)

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error on opening file:", err)
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fw, err := writer.CreateFormFile("myFile", fileName)
	if err != nil {
		fmt.Println("Error on writing to form:", err)
		return err
	}
	defer writer.Close()

	_, err = io.Copy(fw, file)
	if err != nil {
		fmt.Println("Failed to copy file for sending:", err)
		return err
	}
	writer.Close()

	//_, err = http.Post(uri, writer.FormDataContentType(), body) //
	response, err := client.Post(uri, writer.FormDataContentType(), body) //d
	if err != nil {
		fmt.Println("Failed to send Post:", err)
		return err
	}

	res, err := httputil.DumpResponse(response, true)
	if err != nil {
		return err
	}
	fmt.Print(string(res) + "\n")

	fmt.Println("<Sent file>")
	return nil
}

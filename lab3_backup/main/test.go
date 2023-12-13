package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"lab3_backup/chord"
	"log"
	"math/big"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	buffer bytes.Buffer
	node   chord.Node
	Cmd    chord.Command
	err    error
)

func init() {

}
func main() {
	fmt.Printf("--------------------- <%s> ---------------------\n", chord.GetAddress())
	c := &Cmd
	c = commandlineFlags(c)
	log.Println(c.Node)
	c.Server = chord.NewServer(c.Node)
	err := rpc.Register(c.Node)
	if err != nil {
		log.Fatal("error registering node", err)
	}

	c.Server.Listener, _ = net.Listen("tcp", ":"+c.Node.Port)
	c.Node.Listening = true
	log.Println(c.Node)
	newP, e1 := strconv.Atoi(c.Node.Port)
	l1, e := net.Listen("tcp", ":"+strconv.Itoa(newP+1))
	if e != nil || e1 != nil {
		log.Fatal("listen error:", e)
	}
	srv, err := chord.HttpsServer("client.crt")
	if err != nil {
		log.Fatal("Failed to set https Server:", err)
	}
	srv.Addr = ":" + strconv.Itoa(newP+1)
	go func() {
		rpc.HandleHTTP()
		if err := http.Serve(c.Server.Listener, nil); err != nil {
			fmt.Printf("Error serving HTTP requests: <%v>\n", err)
			os.Exit(1)
		}
	}()

	go func() {
		//http.HandleFunc("/favicon.ico", faviconHandler)
		http.HandleFunc("/", c.ServerHandler)
		if err := srv.ServeTLS(l1, "secure_chord.crt", "secure_chord.key"); err != nil {
			fmt.Printf("Error serving HTTP requests: <%v>\n", err)
			os.Exit(1)
		}
	}()
	for {
		line, _ := getline(os.Stdin)

		cnt := 0
		switch line[0] {
		case "create":
			err = c.Create(line[1:]...)
			cnt++
		case "port":
			port := rand.Int31()%50 + 8000
			if len(line) == 1 {
				err = c.Port(strconv.FormatInt(int64(port), 10))
			} else {
				err = c.Port(line[1:]...)
			}
		case "quit":
			err = c.Quit(line[1:]...)
			os.Exit(1)
		case "join":
			err = c.Join(line[1:]...)
			cnt++
		case "put":
			err = c.Put(line[1:]...)
		case "get":
			err = c.Get(line[1:]...)
		case "del":
			err = c.Del(line[1:]...)
		case "backup":
			err = c.Backup(line[1:]...)
		case "recover":
			err = c.Recover(line[1:]...)
		case "dump":
			err = c.Dump(line[1:]...)
		case "putrandom":
			c.Random(line[1:]...)
		case "remove": //|| line[0] == "clear" {
			c.Remove(line[1:]...)
		default:
			if line[0] == "help" {
				err = c.Help(line[1:]...)
			} else {
				err = c.Help(line[0:]...)
			}
		}

		if err != nil {
			log.Println(err)
		}
		if s, e := buffer.ReadString('\n'); e == nil {
			log.Print(s)
		}
	}
}

type errType struct {
	No   int    // order in the loop
	node string // join node
	k    string
	v    string
	// cnt  int //data-cnt
}

type sl []*big.Int

func getline(reader io.Reader) ([]string, error) {
	//reader := bufio.NewReader(os.Stdin)
	//返回结果包含'\n'？？
	buffer := make([]string, 0, 10)
	scanner := bufio.NewScanner(reader)

	if scanner.Err() != nil {
		fmt.Println("1")
		return []string{}, scanner.Err()
	}

	//_, buffer, err := s
	f := func(from string, to *[]string) {
		tmp := strings.Split(from, " ")
		for _, s := range tmp {
			if s != "" {
				*to = append(*to, s)
				// fmt.Println(*to)
			}
		}
	}
	split := func(data []byte, atEOF bool) (int, []byte, error) {
		return bufio.ScanLines(data, atEOF)
	}
	scanner.Split(split)
	if scanner.Scan() {
		f(scanner.Text(), &buffer)
		// fmt.Println(buffer)
		// for i, _ := range buffer {
		// 	fmt.Printf("Order buffers '%s'\n", buffer[i])
		// }
	}
	if len(buffer) == 0 {
		return buffer, errors.New("empty line")
	}
	return buffer, nil //delete all ' ' in buffer
}

func main3() {
	//green := color.New(color.FgGreen)
	//red := color.New(color.FgRed)
	var Cmd chord.Command
	cnt := 0
	// rand.Seed(time.Now().Unix())

	var err error
	for {
		c := &Cmd

		line, _ := getline(os.Stdin)

		if line[0] == "create" {
			err = c.Create(line[1:]...)
			cnt++
		} else if line[0] == "port" {
			port := rand.Int31()%50 + 8000
			if len(line) == 1 {
				err = c.Port(strconv.FormatInt(int64(port), 10))
			} else {
				err = c.Port(line[1:]...)
			}
		} else if line[0] == "quit" {
			err = c.Quit(line[1:]...)
			os.Exit(1)
		} else if line[0] == "join" {
			err = c.Join(line[1:]...)
			cnt++
		} else if line[0] == "put" {
			err = c.Put(line[1:]...)
		} else if line[0] == "get" {
			err = c.Get(line[1:]...)
		} else if line[0] == "del" {
			err = c.Del(line[1:]...)
		} else if line[0] == "backup" {
			err = c.Backup(line[1:]...)
		} else if line[0] == "recover" {
			err = c.Recover(line[1:]...)
		} else if line[0] == "dump" {
			err = c.Dump(line[1:]...)
		} else if line[0] == "putrandom" {
			c.Random(line[1:]...)
		} else if line[0] == "remove" { //|| line[0] == "clear" {
			c.Remove(line[1:]...)
		} else {
			if line[0] == "help" {
				err = c.Help(line[1:]...)
			} else {
				err = c.Help(line[0:]...)
			}
		}

		if err != nil {
			log.Println(err)
		}
		if s, e := buffer.ReadString('\n'); e == nil {
			log.Print(s)
		}
	}
}

//func(args...string) error is the cmd funcs return by error
//can't define as const

// type cmd_function interface {
// 	Quit(args ...string) error
// 	Help(args ...string) error

// 	CPort(args ...string) error
// 	Create(args ...string) error
// 	Join(args ...string) error
// 	Put(args ...string) error
// 	Get(args ...string) error
// 	Del(args ...string) error

// 	Ping(args ...string) error
// 	Dump(args ...string) error
// }

// var (
// 	Cquit cmd_function
// )

// struct cmd:::

//operator domains

func commandlineFlags(c *chord.Command) *chord.Command {
	ip1 := flag.String("a", "", "The IP address that the Chord client will bind to")
	port1 := flag.String("p", "", "The port that the Chord client will bind to and listen on")
	ip2 := flag.String("ja", "", "The IP address of the machine running a Chord node")
	port2 := flag.String("jp", "", "The port that an existing Chord node is bound to and listening on")
	delay1 := flag.Int("ts", 30000, "The time in milliseconds between invocations of ‘stabilize")
	delay2 := flag.Int("tff", 10000, "The time in milliseconds between invocations of ‘fix fingers’")
	delay3 := flag.Int("tcp", 40000, "The time in milliseconds between invocations of ‘check predecessor")
	delay4 := flag.Int("s", 1, "The time in minutes between invocations of ‘backupHandler")
	nbrSuccesors := flag.Int("r", 3, "The number of successors maintained by the Chord client")
	idOverwrite := flag.String("i", "", "The identifier (ID) assigned to the Chord client which will"+
		" override the ID computed by the SHA1 sum of the client’s IP address and port number")
	debuggingOn := flag.Bool("d", false, "The switch for debugging print")

	flag.Parse()

	if *ip1 == "" {
		fmt.Println("Local Node IP hasn't been set\n<Setting to default (localhost IP)>")
		*ip1 = chord.GetAddress()
	}
	if *port1 == "" {
		fmt.Println("Local Node CPort hasn't been set\n<Setting to default (8080)>")
		*port1 = "8080"
	}

	c.Debug = *debuggingOn
	if *delay1 > 60000 || *delay1 < 1 {
		*delay1 = 30000
	}
	if *delay2 > 60000 || *delay2 < 1 {
		*delay2 = 10000
	}
	if *delay3 > 60000 || *delay3 < 1 {
		*delay3 = 40000
	}
	if *delay4 > 10080 || *delay4 < 1 {
		*delay4 = 1
	}
	if *nbrSuccesors > 32 || *nbrSuccesors < 1 {
		*nbrSuccesors = 3
	}
	chord.RefreshTime = time.Duration(*delay1)
	chord.FixFingersDelay = time.Duration(*delay2)
	chord.PredeccesorCheckDelay = time.Duration(*delay3)
	chord.BackupTimeDelay = time.Duration(*delay4)
	chord.MSet = *nbrSuccesors

	ip := *ip1
	port := *port1
	address := ip + ":" + port
	id := chord.Hash(address).String()
	if (*idOverwrite != "") && (len(*idOverwrite) == 48) {
		id = *idOverwrite
	}
	c.Node = chord.NewNode(port, c.Debug, id)
	c.CPort = port

	fmt.Printf("<LocalNode>: %+v\n", *c.Node)

	if *ip2 != "" || *port2 != "" {
		hostIP := *ip2
		hostPort := *port2
		hostAddress := hostIP + ":" + hostPort
		hostID := chord.Hash(hostAddress).String()
		hostNode := chord.Node{Host: hostIP, Port: hostPort, Id: hostID, Address: hostAddress}
		fmt.Printf("<HostNode>: %+v\n", hostNode)
		return c
	} else {
		//c.Create()
		c.Node.Create()
		return c
	}
}

package chord

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"time"
)

var (
	Logger *log.Logger
	debug  bool
	logf   *os.File
)

func init() {
	debug = false
	var err error
	// logf, err := os.OpenFile("./logs.txt", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0664)
	logf, err := os.OpenFile("./logs.txt", os.O_CREATE|os.O_TRUNC|os.O_RDONLY, 0777)
	if err != nil {
		log.Panicln("Crash in Log file init")
	}
	Logger = log.New(logf, ">>> ", log.Ltime)
}

// GetAddress local address
func GetAddress() string {
	var address string
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Panicf("init: GetAddress of finding network: %v", err)
	}

	// find the first non-loopback interface with an IP address
	for _, elt := range interfaces {

		if elt.Flags&net.FlagLoopback == 0 && elt.Flags&net.FlagUp != 0 {
			addres, err := elt.Addrs()
			if err != nil {
				log.Panicf("init: Get addresses of local address: %v", err)
			}

			for _, addr := range addres {
				if ipnet, ok := addr.(*net.IPNet); ok {
					if ip4 := ipnet.IP.To4(); len(ip4) == net.IPv4len {
						address = ip4.String()
						break
					}
				}
			}
		}
	}
	if address == "" {
		log.Panicf("init: failed to find non-loopback interface with valid address on this node")
	}

	return address
}

type Server struct {
	node      *Node
	Listener  net.Listener
	Rpcserver *rpc.Server
	// listening bool

	// logfile *os.File
}

// type Nothing struct{}
// 通常情况下在程序日志里记录一些比较有意义的状态数据：
// 程序启动，退出的时间点；程序运行消耗时间；
// 耗时程序的执行进度；重要变量的状态变化。
// 初次之外，在公共的日志里规避打印程序的调试或者提示信息。

func NewServer(n *Node) *Server {
	//什么时候消亡？？
	p := &Server{
		node: n, // only one element is inited
		// listening: false,
	}
	Logger.Println("--------------------------------------------------------------- <<<")
	Logger.Println("Init a Server at ", time.Now())
	return p
}

// address is in s.node
// this is the function to run a Server
func (s *Server) Listen() error {
	// if s.listening {
	// return errors.New("Already listening")
	// }

	s.Rpcserver = rpc.NewServer()
	err := s.Rpcserver.Register(s.node)
	if err != nil {
		return err
	} //using s.node as a object to do things by rpc
	// rpc.HandleHTTP()

	ler, err := net.Listen("tcp", ":"+s.node.Port) // address
	if err != nil {
		fmt.Printf("listen error: %v", err)
		Logger.Panicf("listen error: %v", err)
		//panic(err)
	} else {
		Logger.Println("listen at ", ":"+s.node.Port)
	}

	s.node.Create()
	s.Listener = ler
	s.node.Listening = true

	go s.Rpcserver.Accept(s.Listener) // goroutine
	return nil
}

// avoid repete listening???
// the CPort is Server.node.CPort(not CPort outside)
func (s *Server) Join(address string) error {
	//if err := s.Listen(); err != nil {
	//	return err
	//}
	return s.node.Join(address)
}

// for a Server, it means unlisten
func (s *Server) Quit() error {
	s.node.merge()

	s.node.Listening = false
	if err := s.Listener.Close(); err != nil {
		fmt.Println(err)
	}

	// if err := logf.Close(); err != nil {
	// 	fmt.Printf("logs Close: %v", err)
	// }

	// fmt.Println("quit hard")

	if err := s.RemoveFile(); err != nil {
		fmt.Println(err)
	}
	return nil
}

func (s *Server) RemoveFile() error {
	err := os.Remove("./backup/" + Hash(s.node.Address).String() + ".txt")
	return err
}

func (s *Server) IsListening() bool {
	return s.Listener != nil
	// return s.listening
}

func (s *Server) Debug() string {

	return fmt.Sprintf(`
ID: %v
Listening: %v Address: %v
Data: %v
Successors: %v
Predecessor: %v
Fingers: %v
`, Hash(s.node.Address), s.IsListening(), s.node.Address, s.node.Data, s.node.SuccessorTable, s.node.Predecessor, s.node.FingerTable)
}

func (s *Server) Backup() {
	s.node.backup()
}

func (s *Server) Recover() {
	s.node.recover()
}

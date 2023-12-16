package chord

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"
)

const (
	m           = 160 // 0-base indexing
	suSize      = 32
	refreshTime = 100 * time.Millisecond // adjust to avoid error:connectex: Only one usage of each socket address (protocol/network address/CPort) is normally permitted. in Stabilize with Notify err:
)

var RefreshTime time.Duration
var FixFingersDelay time.Duration
var PredeccesorCheckDelay time.Duration
var BackupTimeDelay time.Duration
var MSet int

// Node of a virtual machine-- each resource is in Data
type Node struct {
	// HOst + ":" + Port = address of node(Server), using n.addr() to get
	Host           string
	Port           string
	Id             string //*big.Int // hash(addr)
	Address        string
	SuccessorTable [suSize + 1]string // do at 1, 2, 3; using 0 to do Successor, so that can simplify susize+1
	Successor      string
	Predecessor    string

	Data        map[string]string
	FingerTable [m + 1]string
	next        int

	Listening bool
	debug     bool // if true, is debugging
	// file      *os.File
}

type KVP struct {
	K, V string //key, value
}

func init() {
	os.MkdirAll("backup", 0777)
}

func NewNode(_port string, _debug bool, _id string) *Node {
	_host := GetAddress()

	p := &Node{
		Host: _host,
		Port: _port,
		Id:   _id,
		// Id:    Hash(fmt.Sprintf("%v:%v", _host, _port)),
		Address: fmt.Sprintf("%v:%v", _host, _port),
		Data:    make(map[string]string),
		debug:   _debug,
		// FingerTable: make([]string, 0, m),
		next: -1,
	}

	return p
}

// init predecessor and Successor
func (n *Node) Create() {
	n.Predecessor = "" //Addr(n) // or "", can be stablized later
	n.Successor = n.Address
	for i, _ := range n.SuccessorTable {
		n.SuccessorTable[i] = n.Address
	}
	n.recover()

	go n.checkSurvival()
	go n.checkPredecessorPeriod()
	go n.stabilizePeriod()
	go n.fixFingerTablePeriod()
	//go n.backupPeriod()
}

func (n *Node) Besplited(address string, response *bool) error {
	// if n.CPort == "8003" {
	// 	n.dump()
	// 	fmt.Println("WTF:::: ", address)
	// }

	id := Hash(address)
	ip := Hash(n.Address)
	for key, val := range n.Data {
		if InclusiveBetween(id, Hash(key), ip) {
			if err := Call(address, "Node.Put", KVP{K: key, V: val}, &response); err != nil {
				fmt.Printf("Splited err: %v\n", err)
			}
			delete(n.Data, key)
			// fmt.Printf("split %s to %s: %s, %s\n", n.Address, address, key, val)
		}
	}

	return nil
}

func (n *Node) merge() {
	var response bool
	for key, val := range n.Data {
		for err := Call(n.Successor, "Node.Put", KVP{K: key, V: val}, &response); err != nil; err = Call(n.Successor, "Node.Put", KVP{K: key, V: val}, &response) {
			// n.dump()
			// fmt.Printf("Merge err : %v\n", err)
		}
		// fmt.Printf("merge %s to %s: %s, %s\n", n.Address, n.Successor, key, val)
	}
	return
}

// init node's info
func (n *Node) Join(address string) error {
	n.Predecessor = "" // have to be stablized later
	addr, err := RPCFindSuccessor(address, Hash(n.Address))
	if err != nil {
		fmt.Printf("node Join-findsuccessor %v\n", err)
		fmt.Println("Address is ", n.Address, address, addr)
		return err
	}
	n.Successor = addr
	if err := n.fixSuccessorTable(); err != nil {
		return err
	}

	//notify the Successor[0]
	if err := RPCNotify(n.Successor, n.Address); err != nil {
		fmt.Println("Join but Notify err")
	}
	// if n.CPort == "8006" { //} || n.CPort == "8003" {
	// 	n.dump()
	// }

	var response bool
	if err := Call(n.Successor, "Node.Besplited", n.Address, &response); err != nil {
		fmt.Println("split err")
	}
	// if n.CPort == "8006" {
	// 	n.dump()
	// }

	go n.checkSurvival()
	go n.checkPredecessorPeriod()
	go n.stabilizePeriod()
	go n.fixFingerTablePeriod()
	//go n.backupPeriod()

	return nil
}

func (n *Node) CopySuccessor(a bool, table *[suSize + 1]string) error {
	// fmt.Println("From ", n.Address, n.SuccessorTable)
	for i := 0; i <= suSize; i++ {
		// fmt.Printf("change %s to %s\n", (*table)[i], n.SuccessorTable[i])
		table[i] = n.SuccessorTable[i]
	}
	// fmt.Println("Copy2 ", table)
	return nil
}

func (n *Node) fixSuccessorTable() error {
	var a bool
	// fmt.Println(">>>Before ", n.successor, n.SuccessorTable)
	if err := Call(n.Successor, "Node.CopySuccessor", a, &n.SuccessorTable); err != nil {
		return err
	}
	// fmt.Println("Mid ", n.successor, n.SuccessorTable)

	for i := suSize - 1; i >= 0; i-- {
		n.SuccessorTable[i+1] = n.SuccessorTable[i]
	}
	n.SuccessorTable[0] = n.Successor
	// fmt.Println("Copy ", n.successor, n.SuccessorTable)
	return nil
}

// search the local table for the highest predecessor of id
// n.closest_preceding_node(id)
// if empty, return "", left to let
// Find finger[i] ∈ (n, id)
func (n *Node) findFingerTable(id *big.Int) string {
	for i := m - 1; i >= 0; i-- {
		log.Printf("ping figertable %v", n.FingerTable[i])
		if _, err := RPCPing(n.FingerTable[i]); err != nil {
			continue
		}

		//倒着往前找，找到第一个finger使得finger在id和address之间时，此finger的下一个就是我们要找的pre——node
		//或者说，正着找，找到第一个finger使得finger是id的一个假后继，然后前一个finger就是pre——node
		if ExclusiveBetween(Hash(n.FingerTable[i]), Hash(n.Address), id) {
			// the return should be i - 1
			// what if i == 0?, impossible for Successor have to be included

			return n.FingerTable[i]
		}
	}
	// return n.FingerTable[m - 1]
	// n.dump()
	log.Printf("wrong")
	//i==0 error

	// all of them is included by n.Id and id, so as said should return n
	//however incase loop forever, return "", let the function to handle
	return n.Address ///??????

}

// If the id I'm looking for falls between me and mySuccessor
// Then the data for this id will be found on mySuccessor
// ------------>>>will do finger table later
// ask node n to find the Successor of id
// or a better node to continue the search with
// !!!!!maybe can repete with some node(in fixfinger)
func (n *Node) FindSuccessor(id *big.Int, successor *string) error {
	// fmt.Println(Hash(n.Address), id, n.Address)
	if _, err := RPCPing(n.Successor); err == nil && InclusiveBetween(id, Hash(n.Address), Hash(n.Successor)) {
		*successor = n.Successor
		// fmt.Println(id, " Successor find:", n.Address, n.Successor)
		return nil
	}

	for i := 1; i <= suSize; i++ {
		// fmt.Println("circle")
		if _, err := RPCPing(n.SuccessorTable[i]); err == nil && InclusiveBetween(id, Hash(n.SuccessorTable[i-1]), Hash(n.SuccessorTable[i])) { //id ∈ (n, Successor]
			*successor = n.SuccessorTable[i]
			// fmt.Println(id, " Successor find: ", n.SuccessorTable[i-1], n.SuccessorTable[i])
			return nil
		}
	}

	//wont loop if Successor is exists
	// fmt.Println(n)

	nextAddr := n.findFingerTable(id)
	// fmt.Printf("[%s]Finger :%s for %d\n", n.Address, nextAddr, id)
	if nextAddr == "" {
		return errors.New("findFingerTable Err")
	} else {
		var err error
		*successor, err = RPCFindSuccessor(nextAddr, id)
		return err
	}

}

func (n *Node) GetPredecessor(none bool, addr *string) error {
	*addr = n.Predecessor
	if n.Predecessor == "" {
		return errors.New("Predecessor is Empty")
	} else {
		return nil
	}
}

// address is the node that should be awared
// address maybe the n's Predecessor
func (n *Node) Notify(new_predecessor string, response *bool) error {
	if n.Predecessor == "" || ExclusiveBetween(Hash(new_predecessor), Hash(n.Predecessor), Hash(n.Address)) {
		// if n.CPort == "10.163.174.211:8004" {
		// 	fmt.Println("8004 CPort is notifying")
		// 	n.dump()
		// 	fmt.Printf("%s Notify %s\n", n.Address, new_predecessor)
		// }
		n.Predecessor = new_predecessor
		*response = true
		return nil
	}
	*response = false
	return nil
}

// check whther predecessor is failed
func (n *Node) checkPredecessor() {
	if _, err := RPCPing(n.Predecessor); err != nil {
		n.Predecessor = ""
	}
}

// call periodly, refresh.
// n.next is the index of finger to fix
func (n *Node) fixFingerTable() {
	var response string
	for true {
		n.next++
		if n.next >= m {
			n.next = -1
			return
		}
		id := fingerEntry(Hash(n.Address), n.next)

		if response == "" {
			if err := n.FindSuccessor(id, &response); err != nil || response == "" {
				log.Printf("fixFingertable err at: %v\n", err)
				// fmt.Println(n.next, n.Address, response)
				return
			}
			// color.Yellow("Successor")
		}

		if InclusiveBetween(id, Hash(n.Address), Hash(response)) {
			n.FingerTable[n.next] = response
		} else {
			n.next--
			return
		}
		// os.Exit(1)
	}
}

// Call periodly, verfiy succcessor, tell the Successor of n
// only check Successor 1
func (n *Node) stabilize() {
	for _, cur := range n.SuccessorTable {
		if _, err := RPCPing(cur); err != nil {
			continue
		}

		if pre_i, err := RPCGetPredecessor(cur); err == nil {
			if ExclusiveBetween(Hash(pre_i), Hash(n.Address), Hash(cur)) {
				n.Successor = pre_i
			}
		} else {
			// fmt.Printf("%s stabilize[%d] err %v\n", n.Address, i, err)
			// n.dump()
		}
		if err := n.fixSuccessorTable(); err != nil {
			n.Successor = n.SuccessorTable[0]
			// RPCNotify(n.Successor, Addr(n))
			log.Printf("[%s]Stabilize fix SuccessoTable %v\n", n.Address, err)
			continue
		}

		if err := RPCNotify(n.Successor, n.Address); err != nil {
			log.Printf("[%s]Stabilize Notify %v\n", n.Address, err)
		}

		return
	}
	// n.dump()
	log.Println("Successor List Failed")

}

func (n *Node) checkSurvival() {
	ticker := time.Tick(refreshTime)
	for {
		if !n.Listening {
			return
		}
		select { // if <-ticker == time.(0)
		case <-ticker:
			// fmt.Println("check stablize")
			response := 0
			if err := Call(n.Address, "Node.Ping", 51, &response); err != nil {
				n.Listening = false
				return
			}
		}
	}

}

// using 1 func-- 1 tick strategy, can not sync with(using frequency maybe different)
func (n *Node) stabilizePeriod() {
	ticker := time.Tick(refreshTime)
	for {
		if !n.Listening {
			return
		}
		select { // if <-ticker == time.(0)
		case <-ticker:
			// fmt.Println("check stablize")
			n.stabilize()
		}
	}

}
func (n *Node) checkPredecessorPeriod() {
	ticker := time.Tick(refreshTime)
	for {
		if !n.Listening {
			return
		}
		select { // if <-ticker == time.(0)
		case <-ticker:
			// fmt.Println("check Pre")
			n.checkPredecessor()
		}
	}
}
func (n *Node) fixFingerTablePeriod() {
	ticker := time.Tick(refreshTime)
	for {
		if !n.Listening {
			return
		}
		select { // if <-ticker == time.(0)
		case <-ticker:
			n.fixFingerTable()
		}
	}
}

func (n *Node) Put(args KVP, success *bool) error {
	// fmt.Println("Put ", n.Address, args.K, args.V)
	n.Data[args.K] = args.V
	// fmt.Println(n.Data)
	*success = true
	return nil
}
func (n *Node) Get(key string, response *string) error {
	// fmt.Println("Get ", n.Address, key, n.Data[key])
	*response = n.Data[key]
	return nil
}
func (n *Node) Del(key string, success *bool) error {
	if _, ok := n.Data[key]; ok {
		// fmt.Println("Del ", n.Address, key, n.Data[key])
		delete(n.Data, key)
		*success = true
		return nil
	} else {
		*success = false
		return nil //ok func, but can't find key
	}
}

func (n *Node) Ping(request int, response *int) error {
	//do something with request?? no need
	*response = len(n.Data)
	return nil
}

func (node *Node) dump() {
	fmt.Printf(`
ID: %v		Address: %v
S: %v		Successors: %v
Next: %v	Predecessor: %v
Data: %v
Finger: %v
`, Hash(node.Address), node.Address, node.Successor, node.SuccessorTable, node.next, node.Predecessor, node.Data, node.FingerTable[:60])
}

func (n *Node) backup() {
	file, err := os.OpenFile("./backup/"+Hash(n.Address).String()+".txt", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	b := bufio.NewWriter(file)

	if _, err := RPCPing(n.Predecessor); n.Predecessor != n.Address && err == nil {
		var response bool
		// fmt.Println("split backup")
		if err = n.Besplited(n.Predecessor, &response); err != nil {
			fmt.Printf("Backup split:%v\n", err)
		}
	}

	for k, v := range n.Data {
		// fmt.Println("backup", k, v)
		fmt.Fprintln(b, k, v)
		// fmt.Println(a, p)
		b.Flush()
	}
}

func (n *Node) backupPeriod() {
	ticker := time.Tick(refreshTime * 5)
	for {
		select {
		case <-ticker:
			n.backup()
		}
	}
}

func (n *Node) recover() {
	file, err := os.OpenFile("./backup/"+Hash(n.Address).String()+".txt", os.O_CREATE|os.O_RDONLY, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	// fmt.Println("safiwefsadjkvnsadsd")
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)
	for {
		if scanner.Scan() {
			var k, v string
			k = scanner.Text()
			if scanner.Scan() {
				v = scanner.Text()
			} else {
				break
			}

			n.Data[k] = v
		} else {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("reading input: %v\n", err)
	}
}

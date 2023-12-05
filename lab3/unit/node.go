package unit

import (
	"crypto/sha1"
	"log"
	"math/big"
	"net"
	"sync"
)

var m int// -r <Number> = The number of successors maintained by the Chord client. Represented as a base-10 integer. Must be specified, with a value in the range of [1,32].

type Key string

type NodeAddress string

type Node struct {
	Address     NodeAddress
	FingerTable []*fingerEntry
	Predecessor NodeAddress
	Successors  []*Node

	Bucket map[Key]string
	Mutex  sync.Mutex
}

func (node *Node)creatChord{
	log.Printf("Craeting chord")
	node.Predecessor = ""
	node.FingerTable=new([keySize + 2]*fingerEntry)[1:(keySize + 1)]
	for i := 0; i < keySize; i++ {
		node.FingerTable[i] = &fingerEntry{}
		node.FingerTable[i].Id = jump(string(node.Address), i+1).String()
		node.FingerTable[i].Successor = node
	}
	node.Successors= make([]*Node, keySize)
	for i := 0; i < keySize; i++ {
		node.Successors[i] = node.FingerTable[i].Successor
	}
}

func getLocalAddress() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

func hashString(elt string) *big.Int {
	hasher := sha1.New()
	hasher.Write([]byte(elt))
	return new(big.Int).SetBytes(hasher.Sum(nil))
}

const keySize = sha1.Size * 8

var two = big.NewInt(2)
var hashMod = new(big.Int).Exp(big.NewInt(2), big.NewInt(keySize), nil)

func jump(address string, fingerentry int) *big.Int {
	n := hashString(address)
	fingerentryminus1 := big.NewInt(int64(fingerentry) - 1)
	jump := new(big.Int).Exp(two, fingerentryminus1, nil)
	sum := new(big.Int).Add(n, jump)

	return new(big.Int).Mod(sum, hashMod)
}

func between(start, elt, end *big.Int, inclusive bool) bool {
	if end.Cmp(start) > 0 {
		return (start.Cmp(elt) < 0 && elt.Cmp(end) < 0) || (inclusive && elt.Cmp(end) == 0)
	} else {
		return start.Cmp(elt) < 0 || elt.Cmp(end) < 0 || (inclusive && elt.Cmp(end) == 0)
	}
}

n.find_successor(id)
if (id ∈ (n, successor])
return true, successor;
else
return false, closest_preceding_node(id);

// search the local table for the highest predecessor of id
n.closest_preceding_node(id)
// skip this loop if you do not have finger tables implemented yet
for i = m downto 1
if (finger[i] ∈ (n,id])
return finger[i];
return successor;

// find the successor of id
find(id, start)
found, nextNode = false, start;
i = 0
while not found and i < maxSteps
found, nextNode = nextNode.find_successor(id);
i += 1
if found
return nextNode;
else
report error;

func (n *Node) join(joinNode *Node) error {
	log.Printf("node join")
	n.Predecessor = ""
	n.Successors= make([]NodeAddress, m)
	for i := 0; i < m; i++ {
		n.Successors[i] = n.Address
	}
}

package chord

import (
	"crypto/sha1"
	"log"
	"math/big"
	"net"
)

func (node *Node) CreatChord() {
	log.Printf("Craeting chord node %v", node)
	node.Predecessor = nil
	node.FingerTable = newFingerTable(node, m)
	for i := 0; i < keySize; i++ {
		node.FingerTable[i] = newFingerEntry(big.NewInt(int64(i)), node)
		node.FingerTable[i].Id = jump(node.Id.String(), i+1)
		node.FingerTable[i].Successor = node
	}
	node.Successors = make([]*Node, keySize)
	for i := 0; i < keySize; i++ {
		node.Successors[i] = node.FingerTable[i].Successor
	}
	node.Bucket = make(map[*big.Int]string)
}

func getLocalAddress() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

func HashString(elt string) *big.Int {
	hasher := sha1.New()
	hasher.Write([]byte(elt))
	return new(big.Int).SetBytes(hasher.Sum(nil))
}

func between(start, elt, end *big.Int, inclusive bool) bool {
	if end.Cmp(start) > 0 {
		return (start.Cmp(elt) < 0 && elt.Cmp(end) < 0) || (inclusive && elt.Cmp(end) == 0)
	} else {
		return start.Cmp(elt) < 0 || elt.Cmp(end) < 0 || (inclusive && elt.Cmp(end) == 0)
	}
}

func jump(address string, fingerentry int) *big.Int {
	n := HashString(address)
	fingerentryminus1 := big.NewInt(int64(fingerentry) - 1)
	jump := new(big.Int).Exp(two, fingerentryminus1, nil)
	sum := new(big.Int).Add(n, jump)

	return new(big.Int).Mod(sum, hashMod)
}

/*
n.find_successor(id)
if (id ∈ (n, successor])
return true, successor;
else
return false, closest_preceding_node(id);
*/
func (node *Node) findSuccessor(id *big.Int) (bool, *Node) {
	if between(node.Id, id, node.Successors[0].Id, true) {
		return true, node.Successors[0]
	}
	return false, node.closestPrecedingNode(id)
}

/*
// search the local table for the highest predecessor of id
n.closest_preceding_node(id)
// skip this loop if you do not have finger tables implemented yet
for i = m downto 1
if (finger[i] ∈ (n,id])
return finger[i];
return successor;
*/
func (node *Node) closestPrecedingNode(id *big.Int) *Node {
	for i := m; i > 1; i-- {
		if between(node.Id, node.FingerTable[i].Id, id, true) {
			return node.FingerTable[i].Successor
		}
	}
	return node
}

/*
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
*/
func (node *Node) find(id *big.Int, startNode *Node) *Node {
	var nextNode = startNode
	var found bool
	var maxStep = keySize - 1
	for i := 0; i < maxStep; i++ {
		found, nextNode = nextNode.findSuccessor(id)
		if found == true {
			return nextNode
		}
	}
	return startNode
}

func (node *Node) join(joinNode *Node) error {
	node.Predecessor = nil
	node.Successors = make([]*Node, m)

	// depending on the set stabilize & fix_fingers delay time,
	//might get back myself when joining ring after just leaving
	found := node.find(node.Id, node)
	if found.Id == node.Id {
		node.Successors[0] = joinNode
	} else {
		node.Successors[0] = found
	}

	//node.Bucket = node.get_all()
	node.FingerTable = newFingerTable(node, m)
	for i := 0; i < keySize; i++ {
		node.FingerTable[i] = &fingerEntry{}
		node.FingerTable[i].Id = jump(string(node.Address), i+1)
		node.FingerTable[i].Successor = node.Successors[0]
	}

	/* Activate Background Processes*/
	node.stabilize()
	notify(joinNode)
	fixFingers(node, m)
	return nil
}

func (node *Node) stabilize() {

}

func notify(node *Node) {

}

//func (node *Node)find(startNode *Node,id *big.Int) *Node {
//	succ := node.Successors[0]
//
//	if between(startNode.Id,id,node.Id,true) {
//		return succ
//	}
//
//	cpn := this.closestPrecedingNode(keyId)
//	if cpn.Ip == this.Addr.Ip || cpn.Ip == "" { //all finger failed
//		cpn = succ
//	}
//
//	client, err := Diag(cpn.Ip)
//	if err != nil {
//		return AddrType{}
//	}
//	defer client.Close()
//	var ret AddrType
//	err = client.Call("ReceiverType.FindSuccessor", keyId, &ret)
//	return ret
//}

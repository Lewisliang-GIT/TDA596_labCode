package unit

import (
	"crypto/sha1"
	"math/big"
	"sync"
)

type Key string

type NodeAddress string

type Node struct {
	Address     NodeAddress
	FingerTable []NodeAddress
	Predecessor NodeAddress
	Successors  []NodeAddress

	Bucket map[Key]string
	Mutex  sync.Mutex
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

	// First check if node already present in the circle
	// Join this node to the same chord ring as parent
	var foo *Node

	// Ask if our id already exists on the ring.
	if joinNode != nil {
		remoteNode, err := n.findSuccessorRPC(joinNode, n.Address)
		if err != nil {
			return err
		}
		if isEqual(remoteNode.Address, n.Address) {
			return ERR_NODE_EXISTS
		}
		foo = joinNode
	} else {
		foo = n
	}

	succ, err := n.findSuccessorRPC(foo, n.Address)
	if err != nil {
		return err
	}
	n.Mutex.Lock()
	n.successor = succ
	n.Mutex.Unlock()
	return nil
}

package unit

import (
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

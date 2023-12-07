package unit

import (
	"crypto/sha1"
	"math/big"
)

var m int // -r <Number> = The number of successors maintained by the Chord client. Represented as a base-10 integer. Must be specified, with a value in the range of [1,32].

type Key string

type NodeAddress string

type Node struct {
	Id          *big.Int
	Address     NodeAddress
	FingerTable []*fingerEntry
	Predecessor *Node
	Successors  []*Node

	Bucket map[Key]string
	//Mutex  sync.Mutex
}

type ChordRing struct {
	Node []*Node
}

// fingerEntry represents a single finger table entry
type fingerEntry struct {
	Id        *big.Int // ID hash of (n + 2^i) mod (2^m)
	Successor *Node
}

//type fingerTable []*fingerEntry

const keySize = sha1.Size * 8

var two = big.NewInt(2)

var hashMod = new(big.Int).Exp(big.NewInt(2), big.NewInt(keySize), nil)

package unit

import "math/big"

// fingerEntry represents a single finger table entry
type fingerEntry struct {
	Id   []byte // ID hash of (n + 2^i) mod (2^m)
	Node *Node
}

type fingerTable []fingerEntry

func newFingerTable(node *Node, m int) fingerTable {
	ft := make([]*fingerEntry, m)
	for i := range ft {
		ft[i] = newFingerEntry(fingerID(node.Id, i, m), node)
	}
	return ft
}

// newFingerEntry returns an allocated new finger entry with the attributes set
func newFingerEntry(id []byte, node *Node) *fingerEntry {
	return &fingerEntry{
		Id:   id,
		Node: node,
	}
}

// Computes the offset by (n + 2^i) mod (2^m)
func fingerID(n []byte, i int, m int) []byte {
	// Convert the ID to a bigint
	idInt := (&big.Int{}).SetBytes(n)

	// Get the offset
	two := big.NewInt(2)
	offset := big.Int{}
	offset.Exp(two, big.NewInt(int64(i)), nil)

	// Sum
	sum := big.Int{}
	sum.Add(idInt, &offset)

	// Get the ceiling
	ceil := big.Int{}
	ceil.Exp(two, big.NewInt(int64(m)), nil)

	// Apply the mod
	idInt.Mod(&sum, &ceil)
	// Add together
	return idInt.Bytes()
}

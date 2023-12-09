package main

type Key string

type NodeAddress string

type Node struct {
	Address     NodeAddress //
	FingerTable []NodeAddress
	Predecessor NodeAddress
	Successors  []NodeAddress

	Bucket map[Key]string
}

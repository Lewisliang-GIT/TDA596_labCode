package main

import (
	"lab3/chord"
	"log"
)

func main() {
	ip := "0.0.0.0"
	port := "1234"
	address := ip + ":" + port
	var id = chord.HashString(address)
	//if (*idOverwrite != "") && (len(*idOverwrite) == 48) {
	//	id = *idOverwrite
	//}

	log.Printf("<LocalNode>: %+v\n")

	node := chord.Node{Id: id, Address: chord.NodeAddress(address)}
	node.CreatChord()
}

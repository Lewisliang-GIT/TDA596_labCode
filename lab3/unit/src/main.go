package main

import (
	"fmt"
	"os"
)

func main() {
	Arguments := getCmdArgs()
	fmt.Println(Arguments)
	valid := CheckArgsValid(Arguments)
	if valid == -1 {
		fmt.Println("Invalid command line arguments")
		os.Exit(1)
	} else {
		fmt.Println("Valid command line arguments")
		// Create new Node
		node := NewNode(Arguments)

		IPAddr := fmt.Sprintf("%s:%d", Arguments.Address, Arguments.Port)
	}

}

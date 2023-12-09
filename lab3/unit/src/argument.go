package main

import (
	"flag"
	"fmt"
	"net"
	"regexp"
)

type Arguments struct {
	// Read command line arguments
	Address     NodeAddress // Current node address
	Port        int         // Current node port
	JoinAddress NodeAddress // Joining node address
	JoinPort    int         // Joining node port
	Stabilize   int         // The time in milliseconds between invocations of stabilize.
	FixFingers  int         // The time in milliseconds between invocations of fix_fingers.
	CheckPred   int         // The time in milliseconds between invocations of check_predecessor.
	Successors  int
	ClientName  string
}

func getCmdArgs() Arguments {
	// Read command line arguments
	var a string  // Current node address
	var p int     // Current node port
	var ja string // Joining node address
	var jp int    // Joining node port
	var ts int    // The time in milliseconds between invocations of stabilize.
	var tff int   // The time in milliseconds between invocations of fix_fingers.
	var tcp int   // The time in milliseconds between invocations of check_predecessor.
	var r int     // The number of successors to maintain.
	var i string  // Client name

	// Parse command line arguments
	flag.StringVar(&a, "a", "localhost", "Current node address")
	flag.IntVar(&p, "p", 8000, "Current node port")
	flag.StringVar(&ja, "ja", "Unspecified", "Joining node address")
	flag.IntVar(&jp, "jp", 8000, "Joining node port")
	flag.IntVar(&ts, "ts", 3000, "The time in milliseconds between invocations of stabilize.")
	flag.IntVar(&tff, "tff", 1000, "The time in milliseconds between invocations of fix_fingers.")
	flag.IntVar(&tcp, "tcp", 3000, "The time in milliseconds between invocations of check_predecessor.")
	flag.IntVar(&r, "r", 3, "The number of successors to maintain.")
	flag.StringVar(&i, "i", "Default", "Client ID/Name")
	flag.Parse()

	// Return command line arguments
	return Arguments{
		Address:     NodeAddress(a),
		Port:        p,
		JoinAddress: NodeAddress(ja),
		JoinPort:    jp,
		Stabilize:   ts,
		FixFingers:  tff,
		CheckPred:   tcp,
		Successors:  r,
		ClientName:  i,
	}
}

func CheckArgsValid(args Arguments) int {
	// Check if Ip address is valid or not
	if net.ParseIP(string(args.Address)) == nil && args.Address != "localhost" {
		fmt.Println("IP address is invalid")
		return -1
	}
	// Check if port is valid
	if args.Port < 1024 || args.Port > 65535 {
		fmt.Println("Port number is invalid")
		return -1
	}

	// Check if durations are valid
	if args.Stabilize < 1 || args.Stabilize > 60000 {
		fmt.Println("Stabilize time is invalid")
		return -1
	}
	if args.FixFingers < 1 || args.FixFingers > 60000 {
		fmt.Println("FixFingers time is invalid")
		return -1
	}
	if args.CheckPred < 1 || args.CheckPred > 60000 {
		fmt.Println("CheckPred time is invalid")
		return -1
	}

	// Check if number of successors is valid
	if args.Successors < 1 || args.Successors > 32 {
		fmt.Println("Successors number is invalid")
		return -1
	}

	// Check if client name is s a valid string matching the regular expression [0-9a-fA-F]{40}
	if args.ClientName != "Default" {
		matched, err := regexp.MatchString("[0-9a-fA-F]*", args.ClientName)
		if err != nil || !matched {
			fmt.Println("Client Name is invalid")
			return -1
		}
	}

	// Check if joining address and port is valid or not
	if args.JoinAddress != "Unspecified" {
		// Addr is specified, check if addr & port are valid
		if net.ParseIP(string(args.JoinAddress)) != nil || args.JoinAddress == "localhost" {
			// Check if join port is valid
			if args.JoinPort < 1024 || args.JoinPort > 65535 {
				fmt.Println("Join port number is invalid")
				return -1
			}
			// Join the chord
			return 0
		} else {
			fmt.Println("Joining address is invalid")
			return -1
		}
	} else {
		// Join address is not specified, create a new chord ring
		// ignroe jp input
		return 1
	}
}

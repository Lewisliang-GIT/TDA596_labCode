package main

import (
	"errors"
	"fmt"
	"lab3_backup/chord"
	"math/rand"
	"strconv"
)

const (
	// layout  = "Jul 31, 2018 01:40:35 (CST)" // time format
	layout  = "2006-01-02 15:04:05.999999999 (MST)"
	MaxPara = 3
)

var (
	TOO_MANY_ARGUMENTS error
	FEW_ARGUMENTS      error
	ARGUMENTS_NUM      error
	NO_SERVICE         error
	IN_SERVICE         error
)

func init() {
	TOO_MANY_ARGUMENTS = errors.New("too many arguments")
	FEW_ARGUMENTS = errors.New("few arguments")
	NO_SERVICE = errors.New("No service")
	IN_SERVICE = errors.New("Already in service")
	ARGUMENTS_NUM = errors.New("arguments number error")
}

type command struct {
	node   *chord.Node
	server *chord.Server
	port   string // dht.DefaultPort
	host   string // dht.DefaultHost
	line   []string

	debug     bool
	listening bool // or begin maybe
}

// may change the port, once init(), cant change again, so dont use _init()
func (c *command) _init() {
	if c.port == "" {
		c.port = chord.DefaultPort
	}
	c.node = chord.NewNode(c.port, c.debug)
	c.server = chord.NewServer(c.node)
}

// port setting, before a server is init
func (c *command) Port(args ...string) error {
	if c.node != nil || c.listening {
		return errors.New("port can't set again after calling create or join")
	}

	if len(args) > 1 {
		return TOO_MANY_ARGUMENTS
	} else if len(args) == 0 {
		c.port = chord.DefaultPort
	} else {
		c.port = args[0]
	}

	fmt.Printf("Port set to %v\n", c.port)
	return nil
}

func (c *command) Create(args ...string) error {
	if len(args) > 0 {
		return TOO_MANY_ARGUMENTS
	}
	if c.listening {
		return IN_SERVICE
	} else {
		c.listening = true
	}

	c._init()
	c.server.Listen()
	fmt.Println("Node(created) listening at ", c.node.Address)
	return nil
}

// begin to listen server.node.address+port;
// node join at args[0](existing address)
func (c *command) Join(args ...string) error {
	if len(args) > 1 {
		return TOO_MANY_ARGUMENTS
	}
	if c.listening {
		return IN_SERVICE
	} else {
		c.listening = true
	}

	c._init()
	addres := chord.DefaultHost + ":" + chord.DefaultPort
	if len(args) == 1 {
		addres = args[0]
	}

	err := c.server.Join(addres)
	if err != nil {
		c.listening = false
		// log.Panicf("Join error %v", err)
		// return err
	}
	// fmt.Println("Joined at ", addres)
	return nil
}

func (c *command) Quit(args ...string) error {
	if len(args) > 1 {
		return TOO_MANY_ARGUMENTS
	}
	if !c.listening {
		return NO_SERVICE
	} else {
		c.listening = false
	}

	if c.server == nil {
		// fmt.Println("Pragram end")
		return nil
	}

	if err := c.server.Quit(); err != nil {
		fmt.Printf("Server Quit: %v\n", err)
	} else {
		// fmt.Println("Program end")
	}
	// os.Exit(1)
	return nil
}

func (c *command) Dump(args ...string) error {
	if len(args) != 0 {
		return TOO_MANY_ARGUMENTS
	}
	if !c.listening {
		return NO_SERVICE
	}

	fmt.Println(c.server.Debug())
	return nil
}

// Debug func----using dial
// fake ping
// test if args[0](address) is listening
func (c *command) Ping(args ...string) error {
	if len(args) == 0 {
		return FEW_ARGUMENTS
	} else if len(args) > 1 {
		return TOO_MANY_ARGUMENTS
	}
	if !c.listening {
		return NO_SERVICE
	}

	if response, err := chord.RPCPing(args[0]); err != nil {
		return err
	} else {
		fmt.Printf("Got response %d from Ping(3)\n", response)
		//fmt.Fprintln(&buffer, response) //???
		return nil
	}

}

// / put key value
func (c *command) Put(args ...string) error {
	if len(args) != 2 {
		return errors.New(TOO_MANY_ARGUMENTS.Error() + FEW_ARGUMENTS.Error())
	}

	if !c.listening {
		return NO_SERVICE
	}

	if err := chord.RPCPut(c.node.Address, args[0], args[1]); err != nil {
		//fmt.Fprintln(&buffer, false)
		return err
	} else {
		//fmt.Fprintln(&buffer, true)
		return nil
	}
}

func (c *command) Get(args ...string) error {
	if len(args) != 1 {
		return ARGUMENTS_NUM
	}
	if !c.listening {
		return NO_SERVICE
	}

	_, err := chord.RPCGet(c.node.Address, args[0])
	//fmt.Fprintln(&buffer, response)

	return err
}

func (c *command) Del(args ...string) error {
	if len(args) != 1 {
		return ARGUMENTS_NUM
	}
	if !c.listening {
		return NO_SERVICE
	}

	if _, err := chord.RPCDel(c.node.Address, args[0]); err != nil {
		//fmt.Fprintln(&buffer, resp)
		return err
	} else {
		//fmt.Fprintln(&buffer, resp)
		return nil
	}
}

func (c *command) Help(args ...string) error {
	var err error
	if len(args) > 1 {
		err = TOO_MANY_ARGUMENTS
	} else {
		err = nil
	}

	switch len(args) {
	case 0:

		fmt.Println(`Commands are:

Current command
	help		displays recognized commands<current command>

Commands related to DHT rings:
	port /<n>	set the listen-on port<n>. (default  3410)
	create		create a new ring.
	join <add>	join an existing ring.
	quit		shut down. This quits and ends the program. 

Commands related to finding and inserting keys and values
	put <k> <v>		insert the given key and value.
	putrandom <n>	randomly generate n <key, value> to insert.
	get <k>			find the given key in the currently active ring. 
	delete <k> 		the peer deletes it from the ring.

Commands that are useful mainly for debugging:
	dump			display information about the current node.
	dumpkey <k>		similar to dump, but this one finds the node resposible for <key>.
	dumpaddr <add>	similar to above, but query a specific host and dump its info.
	dumpall			walk around the ring, dumping all in clockwise order.

Get more details of each command, you can use order <help+command>
eg: help dump, then you will get details of 'dump'
`)
	case 1:

		switch args[0] {
		case "help":
			fmt.Println("the simplest command. This displays a list of recognized commands. Also, the current command")
		case "port":
			fmt.Println(`
port <n> or port
set the port that this node should listen on. 
By default, this should be port 3410, but users can set it to something else.
This command only works before a ring has been created or joined. After that point, trying to issue this command is an error.
`)
		case "create":
			fmt.Println(`
create
create a new ring.
This command only works before a ring has been created or joined. 
After that point, trying to issue this command is an error.
`)
		case "join":
			fmt.Println(`
join <address>
join an existing ring, one of whose nodes is at the address specified.
This command only works before a ring has been created or joined.
After that point, trying to issue this command is an error.
`)
		case "quit":
			fmt.Println(`
quit
shut down.This quits and ends the program. 
If this was the last instance in a ring, the ring is effectively shut down.
If this is not the last instance, it should send all of its data to its immediate successor before quitting. Other than that, it is not necessary to notify the rest of the ring when a node shuts down.
`)
		case "put":
			fmt.Println(`
there are those related to finding and inserting keys and values.
A <key> is any sequence of one or more non-space characters, as is a value.

put <key> <value> 
insert the given key and value into the currently active ring. 
The instance must find the peer that is responsible for the given key using a DHT lookup operation, 
then contact that host directly and send it the key and value to be stored.
`)
		case "putrandom":
			fmt.Println(`
Next, there are those related to finding and inserting keys and values.
A <key> is any sequence of one or more non-space characters, as is a value.

putrandom <n>
randomly generate n keys (and accompanying values) and put each pair into the ring. Useful for debugging.
`)
		case "get":
			fmt.Println(`
Next, there are those related to finding and inserting keys and values.
A <key> is any sequence of one or more non-space characters, as is a value.

get <key>
find the given key in the currently active ring.
The instance must find the peer that is responsible for the given key using a DHT lookup operation, 
then contact that host directly and retrieve the value and display it to the local user.
`)
		case "delete":
			fmt.Println(`
Next, there are those related to finding and inserting keys and values.
A <key> is any sequence of one or more non-space characters, as is a value.

delete <key>
similar to lookup, but instead of retrieving the value and displaying it, the peer deletes it from the ring.

`)
		case "dump":
			fmt.Println(`
For debugging

dump
display information about the current node, including the range of keys it is resposible for,
 its predecessor and successor links, its finger table, and the actual key/value pairs that it stores.
`)
		case "dumpkey":
			fmt.Println(`
For debugging

dumpkey <key>
similar to dump, but this one finds the node resposible for <key>, 
asks it for its dump info, and displays it to the local user. 
This allows a user at one terminal to query any part of the ring.
`)
		case "dumpaddr":
			fmt.Println(`
For debugging

dumpaddr <address>
similar to above, but query a specific host and dump its info.
`)
		case "dumpall":
			fmt.Println(`
For debugging

dumpall
walk around the ring, dumping all information about every peer in the ring in clockwise order 
(display the current host, then its successor, etc).
`)
		default:
			fmt.Println("Wrong command, get help from command help")
		}
	default:
		fmt.Println("Wrong command, get help from command help")
	}

	return err
}

// //can specially judge server and client.
// //switch server and client and some infos
// //order is like: "test server" or "test client msg"
// func Test(args ...string) error {

// 	}

// 	if args[0] == "server" {
// 		_init()
// 		fmt.Println("server is doing things")
// 		server.Listen()
// 	} else if args[0] == "client" {
// 		if len(args) != 2 {
// 			return errors.New("need command like : test client/server msg[only one]")
// 		}
// 		fmt.Println("client is dong things")
// 		dht.Testcli(host+":"+port, args[1])
// 	}

// 	return nil
// }

func (c *command) Backup(args ...string) error {
	if len(args) != 0 {
		return TOO_MANY_ARGUMENTS
	}
	if !c.listening {
		return NO_SERVICE
	}

	c.server.Backup()
	return nil
}
func (c *command) Recover(args ...string) error {
	if len(args) != 0 {
		return TOO_MANY_ARGUMENTS
	}
	if !c.listening {
		return NO_SERVICE
	}

	c.server.Recover()
	return nil
}

func (c *command) Random(args ...string) error {
	var x int
	var err error
	if len(args) == 0 {
		x = 1
	} else if len(args) == 1 {

		if x, err = strconv.Atoi(args[0]); err != nil {
			return err
		}

	} else {
		err = TOO_MANY_ARGUMENTS
	}
	for i := 1; i <= x; i++ {
		k := strconv.FormatInt(rand.Int63(), 10)
		v := strconv.FormatInt(rand.Int63(), 10)

		if err = c.Put(k, v); err != nil {
			return err
		}

	}

	return err
}
func (c *command) Remove(args ...string) error {
	if len(args) >= 1 {
		return TOO_MANY_ARGUMENTS
	}
	return c.server.RemoveFile()
}

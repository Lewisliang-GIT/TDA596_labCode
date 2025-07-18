package chord

import (
	"errors"
	"fmt"
	"log"
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

type Command struct {
	Node   *Node
	Server *Server
	CPort  string // dht.DefaultPort
	host   string // dht.DefaultHost
	id     string
	Line   []string

	Debug     bool
	listening bool // or begin maybe
}

// may change the CPort, once init(), cant change again, so dont use _init()
func (c *Command) _init() {
	if c.CPort == "" {
		c.CPort = DefaultPort
	}
	c.Node = NewNode(c.CPort, c.Debug, c.id)
	c.Server = NewServer(c.Node)
}

// CPort setting, before a Server is init
func (c *Command) Port(args ...string) error {
	if c.Node != nil || c.listening {
		return errors.New("CPort can't set again after calling create or join")
	}

	if len(args) > 1 {
		return TOO_MANY_ARGUMENTS
	} else if len(args) == 0 {
		c.CPort = DefaultPort
	} else {
		c.CPort = args[0]
	}

	fmt.Printf("CPort set to %v\n", c.CPort)
	return nil
}

func (c *Command) Create(args ...string) error {
	if len(args) > 0 {
		return TOO_MANY_ARGUMENTS
	}
	if c.listening {
		return IN_SERVICE
	} else {
		c.listening = true
	}

	//c._init()
	//err := c.Server.Listen()
	//if err != nil {
	//	return err
	//}
	c.Server.Snode.Create()
	fmt.Println("Node(created) listening at ", c.Node.Address)
	return nil
}

// begin to listen Server.node.address+CPort;
// node join at args[0](existing address)
func (c *Command) Join(args ...string) error {
	if len(args) > 1 {
		return TOO_MANY_ARGUMENTS
	}
	if c.listening {
		return IN_SERVICE
	} else {
		c.listening = true
	}

	//c._init()
	var addres string //addres := DefaultHost + ":" + DefaultPort
	if len(args) == 1 {
		addres = args[0]
	}

	err := c.Server.Join(addres)
	if err != nil {
		c.listening = false
		log.Panicf("Join error %v", err)
		return err
	}
	fmt.Println("Joined at ", addres)
	return nil
}

func (c *Command) Quit(args ...string) error {
	if len(args) > 1 {
		return TOO_MANY_ARGUMENTS
	}
	if !c.listening {
		return NO_SERVICE
	} else {
		c.listening = false
	}

	if c.Server == nil {
		// fmt.Println("Pragram end")
		return nil
	}

	if err := c.Server.Quit(); err != nil {
		fmt.Printf("Server Quit: %v\n", err)
	} else {
		// fmt.Println("Program end")
	}
	// os.Exit(1)
	return nil
}

func (c *Command) Dump(args ...string) error {
	if len(args) != 0 {
		return TOO_MANY_ARGUMENTS
	}
	if !c.listening {
		return NO_SERVICE
	}

	fmt.Println(c.Server.Debug())
	return nil
}

// Debug func----using dial
// fake ping
// test if args[0](address) is listening
func (c *Command) Ping(args ...string) error {
	if len(args) == 0 {
		return FEW_ARGUMENTS
	} else if len(args) > 1 {
		return TOO_MANY_ARGUMENTS
	}
	if !c.listening {
		return NO_SERVICE
	}

	if response, err := RPCPing(args[0]); err != nil {
		return err
	} else {
		fmt.Printf("Got response %d from Ping(3)\n", response)
		//fmt.Fprintln(&buffer, response) //???
		return nil
	}

}

// / put key value
func (c *Command) Put(args ...string) error {
	if len(args) != 2 {
		return errors.New(TOO_MANY_ARGUMENTS.Error() + FEW_ARGUMENTS.Error())
	}

	if !c.listening {
		return NO_SERVICE
	}
	k := Hash(args[0]).String()
	if err := RPCPut(c.Node.Address, k, args[0]); err != nil {
		//fmt.Fprintln(&buffer, false)
		return err
	} else {
		//fmt.Fprintln(&buffer, true)
		return nil
	}
}

func (c *Command) Get(args ...string) error {
	if len(args) != 1 {
		return ARGUMENTS_NUM
	}
	if !c.listening {
		return NO_SERVICE
	}

	_, err := RPCGet(c.Node.Address, args[0])
	//fmt.Fprintln(&buffer, response)

	return err
}

func (c *Command) Del(args ...string) error {
	if len(args) != 1 {
		return ARGUMENTS_NUM
	}
	if !c.listening {
		return NO_SERVICE
	}

	if _, err := RPCDel(c.Node.Address, args[0]); err != nil {
		//fmt.Fprintln(&buffer, resp)
		return err
	} else {
		//fmt.Fprintln(&buffer, resp)
		return nil
	}
}

func (c *Command) Help(args ...string) error {
	var err error
	if len(args) > 1 {
		err = TOO_MANY_ARGUMENTS
	} else {
		err = nil
	}

	switch len(args) {
	case 0:

		fmt.Println(`Commands are:

Current Command
	help		displays recognized commands<current Command>

Commands related to DHT rings:
	CPort /<n>	set the listen-on CPort<n>. (default  3410)
	create		create a new ring.
	join <add>	join an existing ring.
	quit		shut down. This quits and ends the program. 

Commands related to finding and inserting keys and values
	storefile <k> <v>		uploadfile of the given key and value.
	putrandom <n>	randomly generate n <key, value> to insert.
	lookup <k>			find the given key in the currently active ring. 
	delete <k> 		the peer deletes it from the ring.

Commands that are useful mainly for debugging:
	printstate		display information about the current node.
	showkey <k>		similar to printstate, but this one finds the node resposible for <key>.
	showaddr <add>	similar to above, but query a specific host and show its info.
	showall			walk around the ring, dumping all in clockwise order.

Get more details of each Command, you can use order <help+Command>
eg: help printstate, then you will get details of 'printstate'
`)
	case 1:

		switch args[0] {
		case "help":
			fmt.Println("the simplest Command. This displays a list of recognized commands. Also, the current Command")
		case "port":
			fmt.Println(`
port <n> or port
set the CPort that this node should listen on. 
By default, this should be CPort 3410, but users can set it to something else.
This Command only works before a ring has been created or joined. After that point, trying to issue this Command is an error.
`)
		case "quit":
			fmt.Println(`
quit
shut down.This quits and ends the program. 
If this was the last instance in a ring, the ring is effectively shut down.
If this is not the last instance, it should send all of its data to its immediate successor before quitting. Other than that, it is not necessary to notify the rest of the ring when a node shuts down.
`)
		case "storefile":
			fmt.Println(`
there are those related to finding and inserting keys and values.
A <key> is any sequence of one or more non-space characters, as is a value.

storefile <key> <value> 
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
		case "lookup":
			fmt.Println(`
Next, there are those related to finding and inserting keys and values.
A <key> is any sequence of one or more non-space characters, as is a value.

lookup <key>
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
		case "printstate":
			fmt.Println(`
For debugging

printstate
display information about the current node, including the range of keys it is resposible for,
 its predecessor and successor links, its finger table, and the actual key/value pairs that it stores.
`)
		case "showkey":
			fmt.Println(`
For debugging

showkey <key>
similar to dump, but this one finds the node resposible for <key>, 
asks it for its dump info, and displays it to the local user. 
This allows a user at one terminal to query any part of the ring.
`)
		case "showaddr":
			fmt.Println(`
For debugging

showaddr <address>
similar to above, but query a specific host and dump its info.
`)
		case "showall":
			fmt.Println(`
For debugging

showall
walk around the ring, dumping all information about every peer in the ring in clockwise order 
(display the current host, then its successor, etc).
`)
		default:
			fmt.Println("Wrong Command, get help from Command help")
		}
	default:
		fmt.Println("Wrong Command, get help from Command help")
	}

	return err
}

// //can specially judge Server and client.
// //switch Server and client and some infos
// //order is like: "test Server" or "test client msg"
// func Test(args ...string) error {

// 	}

// 	if args[0] == "Server" {
// 		_init()
// 		fmt.Println("Server is doing things")
// 		Server.Listen()
// 	} else if args[0] == "client" {
// 		if len(args) != 2 {
// 			return errors.New("need Command like : test client/Server msg[only one]")
// 		}
// 		fmt.Println("client is dong things")
// 		dht.Testcli(host+":"+CPort, args[1])
// 	}

// 	return nil
// }

func (c *Command) Backup(args ...string) error {
	if len(args) != 0 {
		return TOO_MANY_ARGUMENTS
	}
	if !c.listening {
		return NO_SERVICE
	}

	c.Server.Backup()
	return nil
}
func (c *Command) Recover(args ...string) error {
	if len(args) != 0 {
		return TOO_MANY_ARGUMENTS
	}
	if !c.listening {
		return NO_SERVICE
	}

	c.Server.Recover()
	return nil
}

func (c *Command) Random(args ...string) error {
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
func (c *Command) Remove(args ...string) error {
	if len(args) >= 1 {
		return TOO_MANY_ARGUMENTS
	}
	return c.Server.RemoveFile()
}

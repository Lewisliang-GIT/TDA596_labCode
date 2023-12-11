package chord

import (
	// "dht"
	// "dht"
	"errors"
	"fmt"
	"math/big"
	"net/rpc"
)

const (
	DefaultHost = ""
	DefaultPort = ""
)

//actually a struct is here:
// client, method, request, response

func Call(address string, method string, request interface{}, response interface{}) error {
	if address == "" {
		return errors.New("Call Err: No address")
	}

	client, err := rpc.Dial("tcp", address)
	if err != nil {
		// Logger.Printf("dial failed: %v", err)
		// fmt.Printf("Dial: %v\n", err)
		return err
	}
	defer client.Close()

	//get call
	err = client.Call(method, request, response)
	if err != nil {
		// Logger.Printf("call error: %v", err)
		// fmt.Printf("Call: %v\n", err)
		return err
	}

	return nil
}

// address.Predecessor --- be_noticed_node -- address
func RPCNotify(address string, new_predecessor string) error {
	if address == "" {
		return errors.New("Notify: rpc address is empty")
	}
	response := false
	// fmt.Println("RPC NOtify get")
	return Call(address, "Node.Notify", new_predecessor, &response)
}

func RPCGetPredecessor(addr string) (string, error) {
	if addr == "" {
		return "", errors.New("GetPredecessor: rpc address is empty")
	}

	response := ""
	if err := Call(addr, "Node.GetPredecessor", false, &response); err != nil {
		return "", err
	}
	if response == "" {
		return "", errors.New("GetPredecessor: rpc Empty predecessor")
	}

	return response, nil
}

func RPCFindSuccessor(addr string, id *big.Int) (string, error) {
	if addr == "" {
		return "", errors.New("RPCFindSuccessor: rpc address is empty")
	}

	response := ""
	if err := Call(addr, "Node.FindSuccessor", id, &response); err != nil {
		return response, err
	}

	return response, nil
}

func RPCPing(address string) (int, error) {
	response := -1
	if err := Call(address, "Node.Ping", 3, &response); err != nil {
		return response, err
	}

	// fmt.Printf("Got response %d from Ping(3)\n", response)
	return response, nil
}

// if return "", cant do the function anymore, return errors till main()
// if dial is panic, panic the whole program????????????
func find(address string, key string) string {
	response, err := RPCFindSuccessor(address, Hash(key))
	if err != nil {
		fmt.Printf("find address: %v\n", err) //maybe panic??
		return ""
	}

	return response
}

func RPCPut(address string, key string, val string) error {
	put_node := find(address, key) // key's successor
	if put_node == "" {
		return errors.New("can't get address")
	}

	var response bool
	if err := Call(put_node, "Node.Put", KVP{K: key, V: val}, &response); err != nil {
		return err
	}

	fmt.Printf("Put %s, %s in [%v]\n", key, val, put_node)
	if !response {
		return errors.New("No put")
	}
	return nil
}
func RPCGet(address string, key string) (string, error) {
	get_node := find(address, key) // key's successor
	if get_node == "" {
		return "", errors.New("can't get address")
	}

	var response string
	if err := Call(get_node, "Node.Get", key, &response); err != nil {
		return "", err
	}
	fmt.Printf("Get [%v] stored %v at %s\n", get_node, response, key)

	return response, nil
}

func RPCDel(address string, key string) (bool, error) {
	del_node := find(address, key) // key's successor
	if del_node == "" {
		return false, errors.New("can't get address")
	}

	var response bool
	if err := Call(del_node, "Node.Del", key, &response); err != nil {
		return false, err
	}

	fmt.Printf("Del [%v] KVPair(%v) is %t\n", del_node, key, response)
	return response, nil
}

func RPCGetSuccessors(address string, s []string) error {
	var a int
	if err := Call(address, "Node.GetSuccessors", a, &s); err != nil {
		return err
	}
	return nil
}

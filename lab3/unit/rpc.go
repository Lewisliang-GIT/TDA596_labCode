package unit

import (
	"errors"
	"fmt"
	"net/rpc"
	"net/rpc/jsonrpc"
	"strings"
)

func Call(targetNode NodeAddress, method string, request interface{}, reply interface{}) error {
	if len(strings.Split(string(targetNode), ":")) != 2 {
		fmt.Println("Error: targetNode address is not in the correct format: ", targetNode)
		return errors.New("Error: targetNode address is not in the correct format: " + string(targetNode))
	}
	ip := strings.Split(string(targetNode), ":")[0]
	port := strings.Split(string(targetNode), ":")[1]

	targetNodeAddr := ip + ":" + port
	// conn, err := tls.Dial("tcp", targetNodeAddr, &tls.Config{InsecureSkipVerify: true})
	// client := jsonrpc.NewClient(conn)
	client, err := jsonrpc.Dial("tcp", targetNodeAddr)
	if err != nil {
		fmt.Println("Method: ", method, "Dial Error: ", err)
		return err
	}
	defer func(client *rpc.Client) {
		err := client.Close()
		if err != nil {

		}
	}(client)
	err = client.Call(method, request, reply)
	if err != nil {
		fmt.Println("Call Error:", err)
		return err
	}
	return nil
}

package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"sort"
	"time"
)

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

// main/mrworker.go calls this function.
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	// Your worker implementation here.
	// Set up worker server
	for {
		args := RequestTaskReply{}
		reply := RequestTaskReply{}

		res := call("Coordinator.RequestTask", &args, &reply)
		if !res {
			break
		}

		switch reply.Task.Task {
		case Map:
			log.Printf("%v do map", reply.Task.Task)
			doMap(&reply, mapf)
		case Reduce:
			log.Printf("%v do reduce", reply.Task.Task)
			doReduce(&reply, reducef)
		case Wait:
			log.Printf("%v waiting", reply.Task.Task)
			time.Sleep(1 * time.Second)
		case Exit:
			log.Printf("%v exit", reply.Task.Task)
			os.Exit(0)
		default:
			time.Sleep(1 * time.Second)
		}
	}

	// uncomment to send the Example RPC to the coordinator.
	// CallExample()
}

func doReduce(reply *RequestTaskReply, reducef func(string, []string) string) {
	// Load intermediate files
	intermediate := []KeyValue{}
	for m := 0; m < len(reply.Task.InputFiles); m++ {
		file, err := os.Open(reply.Task.InputFiles[m])
		log.Printf("%v doing reduce %v", reply.TaskNo, reply.Task.InputFiles)
		if err != nil {
			log.Fatalf("cannot open %v", reply.Task.InputFiles[m])
		}
		dec := json.NewDecoder(file)
		for {
			var kv KeyValue
			if err := dec.Decode(&kv); err != nil {
				break
			}
			intermediate = append(intermediate, kv)
		}
		file.Close()
	}

	sort.Slice(intermediate, func(i, j int) bool {
		return intermediate[i].Key < intermediate[j].Key
	})

	oname := fmt.Sprintf("mr-out-%d", reply.Task.Index)
	ofile, _ := os.Create(oname)

	i := 0
	for i < len(intermediate) {
		j := i + 1
		for j < len(intermediate) && intermediate[j].Key == intermediate[i].Key {
			j++
		}
		values := []string{}
		for k := i; k < j; k++ {
			values = append(values, intermediate[k].Value)
		}
		output := reducef(intermediate[i].Key, values)
		fmt.Fprintf(ofile, "%v %v\n", intermediate[i].Key, output)
		i = j
	}

	ofile.Close()
	// Update task status
	reply.Task.Status = Finished
	replyEx := RequestTaskReply{}
	call("Coordinator.NotifyComplete", &reply, &replyEx)
}

func doMap(reply *RequestTaskReply, mapf func(string, string) []KeyValue) {

	file, err := os.Open(reply.Task.InputFiles[0])
	if err != nil {
		log.Fatalf("cannot open %v", reply.Task.InputFiles[0])
	}

	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("cannot read %v", reply.Task.InputFiles[0])
	}
	file.Close()
	log.Printf("%v doing map %v", reply.TaskNo, reply.Task.InputFiles)

	kva := mapf(reply.Task.InputFiles[0], string(content))
	intermediate := make([][]KeyValue, reply.NReduce)
	for _, kv := range kva {
		r := ihash(kv.Key) % reply.NReduce
		intermediate[r] = append(intermediate[r], kv)
	}

	for r, kva := range intermediate {
		oname := fmt.Sprintf("mr-%d-%d", reply.Task.Index, r)
		ofile, _ := os.Create(oname)
		enc := json.NewEncoder(ofile)
		for _, kv := range kva {
			enc.Encode(&kv)
		}
		ofile.Close()
	}
	log.Printf("%v done", reply.TaskNo)

	// Update server state of the task, by calling the RPC NotifyComplete
	reply.Task.Status = Finished
	replyEx := RequestTaskReply{}
	call("Coordinator.NotifyComplete", &reply, &replyEx)
}

// example function to show how to make an RPC call to the coordinator.
//
// the RPC argument and reply types are defined in fingertable.go.
func CallExample() {

	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	// the "Coordinator.Example" tells the
	// receiving server that we'd like to call
	// the Example() method of struct Coordinator.
	ok := call("Coordinator.Example", &args, &reply)
	if ok {
		// reply.Y should be 100.
		fmt.Printf("reply.Y %v\n", reply.Y)
	} else {
		fmt.Printf("call failed!\n")
	}
}

// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(rpcname string, args interface{}, reply interface{}) bool {
	sockname := coordinatorSock()
	// c, err := rpc.DialHTTP("tcp", "54.225.11.131"+":1234")
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}

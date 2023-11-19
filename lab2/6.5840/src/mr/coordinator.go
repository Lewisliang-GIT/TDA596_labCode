package mr

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

const (
	EMPTY       = 0
	IN_PROGRESS = 1
	COMPLETED   = 2
)

type Task struct {
	Status    string
	WorkerId  string
	StartedAt time.Time
}

type MapTask struct {
	FileName string
	NReduce  int
	Task
}

type ReduceTask struct {
	Region    int
	Locations []string
	Task
}

type Coordinator struct {
	// Your definitions here.
	MapTasks          []*MapTask
	RemindMapTasks    int
	ReduceTasks       []*ReduceTask
	RemindResuceTasks int
	Mu                sync.Mutex
}

// Your code here -- RPC handlers for the worker to call.

// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

// start a thread that listens for RPCs from worker.go
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	ret := false

	// Your code here.

	return ret
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}
	c.MapTasks = make([]*MapTask, len(files))
	c.RemindMapTasks = len(files)
	c.ReduceTasks = make([]*ReduceTask, nReduce)
	c.RemindResuceTasks = nReduce

	//Initialize map tasks
	for i, file := range files {
		c.MapTasks[i] = &MapTask{
			FileName: file,
			NReduce:  nReduce,
			Task:     Task{Status: EMPTY},
		}
	}

	//Initialize reduce tasks
	for i := 0; i < nReduce; i++ {
		c.ReduceTasks[i] = &ReduceTask{
			Region: nReduce + 1,
			Task:   Task{Status: EMPTY},
		}
	}

	fmt.Printf("Coordinator initialized with %v Map Tasks\n", len(files))
	fmt.Printf("Coordinator initialized with %v Reduce Tasks\n", nReduce)
	c.server()
	return &c
}

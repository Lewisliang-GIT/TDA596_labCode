package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
	"time"
)

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

// Add your RPC definitions here.
type Task int

const (
	Exit Task = iota
	Wait
	Map
	Reduce
)

type Status int

const (
	Unassigned Status = iota
	Assigned
	Finished
)

type MapReduceTask struct {
	Task      Task
	Status    Status
	TimeStamp time.Time
	Index     int

	InputFiles  []string
	OutputFiles []string
}

type RequestTaskReply struct {
	TaskNo  int
	Task    MapReduceTask
	NReduce int
}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	// s := "/var/tmp/5840-mr-"
	// s += strconv.Itoa(os.Getuid())
	s:=strconv.Itoa(os.Getuid())
	return s
}

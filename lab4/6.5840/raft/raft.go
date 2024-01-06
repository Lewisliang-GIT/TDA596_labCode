package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	"6.5840/labgob"
	"bytes"
	"log"

	//	"bytes"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	//	"6.5840/labgob"
	"6.5840/labrpc"
)

type VoteState int
type AppendState int

var HeartBeatTimeout = time.Duration(100) * time.Millisecond

// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in part 2D you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh, but set CommandValid to false for these
// other uses.
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int

	// For 2D:
	SnapshotValid bool
	Snapshot      []byte
	SnapshotTerm  int
	SnapshotIndex int
}

type logEntries struct {
	Command interface{}
	Term    int
}

type States int

const (
	Leader States = iota
	Candidate
	Follower
)

const (
	Normal VoteState = iota
	Killed
	Expired
	Voted
)

const (
	Appended AppendState = iota
	AppendTimeout
	AppendKilled   // killed by Raft
	AppendCommited //2b
	MissingLog     //2b
)

// A Go object implementing a single Raft peer.
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	// Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.
	currentTerm int
	votedFor    int
	states      States
	Logs        []logEntries
	overtime    time.Duration
	timer       *time.Ticker

	commitIndex int
	lastApplied int

	nextIndex  []int
	matchIndex []int

	applyCh chan ApplyMsg
}

type AppendEntriesArgs struct {
	Term         int          // leader的任期
	LeaderId     int          // leader自身的ID
	PrevLogIndex int          // 预计要从哪里追加的index，因此每次要比当前的len(logs)多1 args初始化为：rf.nextIndex[i] - 1
	PrevLogTerm  int          // 追加新的日志的任期号(这边传的应该都是当前leader的任期号 args初始化为：rf.currentTerm
	Entries      []logEntries // 预计存储的日志（为空时就是心跳连接）
	LeaderCommit int          // leader的commit index指的是最后一个被大多数机器都复制的日志Index
}

type AppendEntriesReply struct {
	Term        int         // leader的term可能是过时的，此时收到的Term用于更新他自己
	Success     bool        //	如果follower与Args中的PreLogIndex/PreLogTerm都匹配才会接过去新的日志（追加），不匹配直接返回false
	AppState    AppendState // 追加状态
	UpNextIndex int         //  用于更新请求节点的nextIndex[i]
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	var term int
	var isleader bool
	// Your code here (2A).
	term = rf.currentTerm
	log.Println("the peer[", rf.me, "] state is:", rf.states)
	if rf.states == Leader {
		isleader = true
	} else {
		isleader = false
	}
	return term, isleader
}

// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
// before you've implemented snapshots, you should pass nil as the
// second argument to persister.Save().
// after you've implemented snapshots, pass the current snapshot
// (or nil if there's not yet a snapshot).
func (rf *Raft) persist() {
	// Your code here (2C).
	w := new(bytes.Buffer)
	e := labgob.NewEncoder(w)
	e.Encode(rf.currentTerm)
	e.Encode(rf.votedFor)
	e.Encode(rf.Logs)
	data := w.Bytes()
	rf.persister.Save(data, nil)
}

// restore previously persisted state.
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (2C).
	r := bytes.NewBuffer(data)
	d := labgob.NewDecoder(r)
	var currentTerm int
	var votedFor int
	var logs []logEntries
	if d.Decode(&currentTerm) != nil ||
		d.Decode(&votedFor) != nil ||
		d.Decode(&logs) != nil {
		log.Println("decode error")
	} else {
		rf.currentTerm = currentTerm
		rf.votedFor = votedFor
		rf.Logs = logs
	}
}

// the service says it has created a snapshot that has
// all info up to and including index. this means the
// service no longer needs the log through (and including)
// that index. Raft should now trim its log as much as possible.
func (rf *Raft) Snapshot(index int, snapshot []byte) {
	// Your code here (2D).

}

// example RequestVote RPC arguments structure.
// field names must start with capital letters!
type RequestVoteArgs struct {
	// Your data here (2A, 2B).
	Term         int
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

// example RequestVote RPC reply structure.
// field names must start with capital letters!
type RequestVoteReply struct {
	// Your data here (2A).
	Term        int
	VoteGranted bool
	VoteState   VoteState
}

// example RequestVote RPC handler.
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (2A, 2B).

	defer log.Printf("[	    func-RequestVote-rf(%+v)		] : return %v\n", rf.me, reply)

	rf.mu.Lock()
	defer rf.mu.Unlock()

	// 当前节点crash
	if rf.killed() {
		reply.VoteState = Killed
		reply.Term = -1
		reply.VoteGranted = false
		return
	}

	//reason: 出现网络分区，该竞选者已经OutOfDate(过时）
	if args.Term < rf.currentTerm {
		reply.VoteState = Expired
		reply.Term = rf.currentTerm
		reply.VoteGranted = false
		return
	}

	if args.Term > rf.currentTerm {
		// 重置自身的状态
		rf.states = Follower
		rf.currentTerm = args.Term
		rf.votedFor = -1
	}

	log.Printf("[	    func-RequestVote-rf(%+v)		] : rf.voted: %v\n", rf.me, rf.votedFor)
	// 此时比自己任期小的都已经把票还原
	if rf.votedFor == -1 {

		lastLogTerm := 0
		if len(rf.Logs) > 0 {
			lastLogTerm = rf.Logs[len(rf.Logs)-1].Term
		}

		//  If votedFor is null or candidateId, and candidate’s log is at least as up-to-date as receiver’s log, grant vote (§5.2, §5.4)
		// 论文里的第二个匹配条件，当前peer要符合arg两个参数的预期
		if args.LastLogTerm < lastLogTerm || (len(rf.Logs) > 0 && args.LastLogTerm == rf.Logs[len(rf.Logs)-1].Term && args.LastLogIndex < len(rf.Logs)) {
			reply.VoteState = Expired
			reply.VoteGranted = false
			reply.Term = rf.currentTerm
			return
		}

		// 给票数，并且返回true
		rf.votedFor = args.CandidateId

		reply.VoteState = Normal
		reply.Term = rf.currentTerm
		reply.VoteGranted = true

		rf.timer.Reset(rf.overtime)

		log.Printf("[	    func-RequestVote-rf(%v)		] : voted rf[%v]\n", rf.me, rf.votedFor)

	} else { // 只剩下任期相同，但是票已经给了，此时存在两种情况

		reply.VoteState = Voted
		reply.VoteGranted = false

		// 1、当前的节点是来自同一轮，不同竞选者的，但是票数已经给了(又或者它本身自己就是竞选者）
		if rf.votedFor != args.CandidateId {
			// 告诉reply票已经没了返回false
			return
		} else { // 2. 当前的节点票已经给了同一个人了，但是由于sleep等网络原因，又发送了一次请求
			// 重置自身状态
			rf.states = Follower
		}

		rf.timer.Reset(rf.overtime)

	}

	return
}

// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply, voteNums *int) bool {

	if rf.killed() {
		return false
	}
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	for !ok {
		// 失败重传
		if rf.killed() {
			return false
		}
		ok = rf.peers[server].Call("Raft.RequestVote", args, reply)

	}

	rf.mu.Lock()
	defer rf.mu.Unlock()
	log.Printf("[	sendRequestVote(%v) ] : send a election to %v\n", rf.me, server)
	// 由于网络分区，请求投票的人的term的比自己的还小，不给予投票
	if args.Term < rf.currentTerm {
		return false
	}

	// 对reply的返回情况进行分支处理
	switch reply.VoteState {
	// 消息过期有两种情况:
	// 1.是本身的term过期了比节点的还小
	// 2.是节点日志的条目落后于节点了
	case Expired:
		{
			rf.states = Follower
			rf.timer.Reset(rf.overtime)
			if reply.Term > rf.currentTerm {
				rf.currentTerm = reply.Term
				rf.votedFor = -1
			}
		}
	case Normal, Voted:
		//根据是否同意投票，收集选票数量
		if reply.VoteGranted && reply.Term == rf.currentTerm && *voteNums <= (len(rf.peers)/2) {
			*voteNums++
		}

		// 票数超过一半
		if *voteNums >= (len(rf.peers)/2)+1 {

			*voteNums = 0
			// 本身就是leader在网络分区中更具有话语权的leader
			if rf.states == Leader {
				return ok
			}

			// 本身不是leader，那么需要初始化nextIndex数组
			rf.states = Leader
			rf.nextIndex = make([]int, len(rf.peers))
			for i := 0; i < len(rf.peers); i++ {
				rf.nextIndex[i] = len(rf.Logs) + 1
			}
			rf.timer.Reset(HeartBeatTimeout)
			log.Printf("[	sendRequestVote-func-rf(%v)		] be a leader\n", rf.me)
		}
	case Killed:
		return false
	}
	return ok
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply, appendNums *int) {

	if rf.killed() {
		return
	}

	// paper中5.3节第一段末尾提到，如果append失败应该不断的retries ,直到这个log成功的被store
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	for !ok {

		if rf.killed() {
			return
		}
		ok = rf.peers[server].Call("Raft.AppendEntries", args, reply)

	}

	// 必须在加在这里否则加载前面retry时进入时，RPC也需要一个锁，但是又获取不到，因为锁已经被加上了
	rf.mu.Lock()
	defer rf.mu.Unlock()

	// 对reply的返回状态进行分支
	switch reply.AppState {

	case AppendKilled:
		{
			return
		}

	// 目标节点正常返回
	case Appended:
		{

			// 2A的test目的是让Leader能不能连续任期，所以2A只需要对节点初始化然后返回就好
			// 2B需要判断返回的节点是否超过半数commit，才能将自身commit
			if reply.Success && reply.Term == rf.currentTerm && *appendNums <= len(rf.peers)/2 {
				*appendNums++
			}

			// 说明返回的值已经大过了自身数组
			if rf.nextIndex[server] > len(rf.Logs)+1 {
				return
			}
			rf.nextIndex[server] += len(args.Entries)
			if *appendNums > len(rf.peers)/2 {
				// 保证幂等性，不会提交第二次
				*appendNums = 0

				if len(rf.Logs) == 0 || rf.Logs[len(rf.Logs)-1].Term != rf.currentTerm {
					return
				}

				for rf.lastApplied < len(rf.Logs) {
					rf.lastApplied++
					applyMsg := ApplyMsg{
						CommandValid: true,
						Command:      rf.Logs[rf.lastApplied-1].Command,
						CommandIndex: rf.lastApplied,
					}
					rf.applyCh <- applyMsg
					rf.commitIndex = rf.lastApplied
					log.Printf("[	sendAppendEntries func-rf(%v)	] commitLog  \n", rf.me)
				}

			}
			log.Printf("[	sendAppendEntries func-rf(%v)	] rf.log :%+v  ; rf.lastApplied:%v\n",
				rf.me, rf.Logs, rf.lastApplied)

			return
		}

	case MissingLog, AppendCommited:
		if reply.Term > rf.currentTerm {
			rf.states = Follower
			rf.votedFor = -1
			rf.timer.Reset(rf.overtime)
			rf.currentTerm = reply.Term
			rf.persist()
		}
		rf.nextIndex[server] = reply.UpNextIndex
	//If AppendEntries RPC received from new leader: convert to follower(paper - 5.2)
	//reason: 出现网络分区，该Leader已经OutOfDate(过时）,term小于发送者
	case AppendTimeout:

		// 该节点变成追随者,并重置rf状态
		rf.states = Follower
		rf.votedFor = -1
		rf.timer.Reset(rf.overtime)
		rf.currentTerm = reply.Term
		rf.persist()

	}

	return
}

// AppendEntries 建立心跳、同步日志RPC
func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	// 节点crash
	if rf.killed() {
		reply.AppState = AppendKilled
		reply.Term = -1
		reply.Success = false
		return
	}

	// 出现网络分区，args的任期，比当前raft的任期还小，说明args之前所在的分区已经OutOfDate
	if args.Term < rf.currentTerm {
		reply.AppState = AppendTimeout
		reply.Term = rf.currentTerm
		reply.Success = false
		return
	}

	// 出现conflict的情况
	// paper:Reply false if log doesn’t contain an entry at prevLogIndex,whose term matches prevLogTerm (§5.3)
	// 首先要保证自身len(rf)大于0否则数组越界
	// 1、 如果preLogIndex的大于当前日志的最大的下标说明跟随者缺失日志，拒绝附加日志
	// 2、 如果preLog出`的任期和preLogIndex处的任期和preLogTerm不相等，那么说明日志存在conflict,拒绝附加日志
	if args.PrevLogIndex > 0 && (len(rf.Logs) < args.PrevLogIndex || rf.Logs[args.PrevLogIndex-1].Term != args.PrevLogTerm) {

		reply.AppState = MissingLog
		reply.Term = rf.currentTerm
		reply.Success = false
		reply.UpNextIndex = rf.lastApplied + 1
		return
	}

	// 如果当前节点提交的Index比传过来的还高，说明当前节点的日志已经超前,需返回过去
	if args.PrevLogIndex != -1 && rf.lastApplied > args.PrevLogIndex {
		reply.AppState = AppendCommited
		reply.Term = rf.currentTerm
		reply.Success = false
		reply.UpNextIndex = rf.lastApplied + 1
		return
	}

	// 对当前的rf进行ticker重置
	rf.currentTerm = args.Term
	rf.votedFor = args.LeaderId
	rf.states = Follower
	rf.timer.Reset(rf.overtime)

	// 对返回的reply进行赋值
	reply.AppState = Appended
	reply.Term = rf.currentTerm
	reply.Success = true

	// 如果存在日志包那么进行追加
	if args.Entries != nil {
		rf.Logs = rf.Logs[:args.PrevLogIndex]
		rf.Logs = append(rf.Logs, args.Entries...)

	}

	// 将日志提交至与Leader相同
	for rf.lastApplied < args.LeaderCommit {
		rf.lastApplied++
		applyMsg := ApplyMsg{
			CommandValid: true,
			CommandIndex: rf.lastApplied,
			Command:      rf.Logs[rf.lastApplied-1].Command,
		}
		rf.applyCh <- applyMsg
		rf.commitIndex = rf.lastApplied
		log.Printf("[	AppendEntries func-rf(%v)	] commitLog  \n", rf.me)
	}

	return
}

// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	// Your code here (2B).
	if rf.killed() {
		return index, term, false
	}

	rf.mu.Lock()
	defer rf.mu.Unlock()

	// 如果不是leader，直接返回
	if rf.states != Leader {
		return index, term, false
	}

	isLeader = true

	// 初始化日志条目。并进行追加
	appendLog := logEntries{Term: rf.currentTerm, Command: command}
	rf.Logs = append(rf.Logs, appendLog)
	index = len(rf.Logs)
	term = rf.currentTerm

	return index, term, isLeader
}

// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
	rf.mu.Lock()
	rf.timer.Stop()
	rf.mu.Unlock()
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

func (rf *Raft) ticker() {
	for rf.killed() == false {
		// Your code here (2A)
		// Check if a leader election should be started.
		select {

		case <-rf.timer.C:
			if rf.killed() {
				return
			}
			rf.mu.Lock()
			switch rf.states {
			case Follower:
				rf.states = Candidate
				fallthrough
			case Leader:
				// 进行心跳/日志同步
				appendNums := 1 // 对于正确返回的节点数量
				rf.timer.Reset(HeartBeatTimeout)
				// 构造msg
				for i := 0; i < len(rf.peers); i++ {
					if i == rf.me {
						continue
					}
					args := AppendEntriesArgs{
						Term:         rf.currentTerm,
						LeaderId:     rf.me,
						PrevLogIndex: 0,
						PrevLogTerm:  0,
						Entries:      nil,
						LeaderCommit: rf.commitIndex, // commitIndex为大多数log所认可的commitIndex
					}

					reply := AppendEntriesReply{}

					// 如果nextIndex[i]长度不等于rf.logs,代表与leader的log entries不一致，需要附带过去

					args.Entries = rf.Logs[rf.nextIndex[i]-1:]

					// 代表已经不是初始值0
					if rf.nextIndex[i] > 0 {
						args.PrevLogIndex = rf.nextIndex[i] - 1
					}

					if args.PrevLogIndex > 0 {
						log.Println("len(rf.log):", len(rf.Logs), "PrevLogIndex):", args.PrevLogIndex, "rf.nextIndex[i]", rf.nextIndex[i])
						args.PrevLogTerm = rf.Logs[args.PrevLogIndex-1].Term
					}

					log.Printf("[	ticker(%v) ] : send a election to %v\n", rf.me, i)
					go rf.sendAppendEntries(i, &args, &reply, &appendNums)
				}

			case Candidate:
				rf.currentTerm += 1
				rf.votedFor = rf.me
				votedNums := 1 //vote numbers for itself

				rf.overtime = time.Duration(150+rand.Intn(200)) * time.Millisecond
				rf.timer.Reset(rf.overtime)

				for i := 0; i < len(rf.peers); i++ {
					if i == rf.me {
						continue
					}

					voteArgs := RequestVoteArgs{
						Term:         rf.currentTerm,
						CandidateId:  rf.me,
						LastLogIndex: len(rf.Logs),
						LastLogTerm:  0,
					}
					if len(rf.Logs) > 0 {
						voteArgs.LastLogTerm = rf.Logs[len(rf.Logs)-1].Term
					}
					voteReply := RequestVoteReply{}
					go rf.sendRequestVote(i, &voteArgs, &voteReply, &votedNums)
				}
			}

			// pause for a random amount of time between 50 and 350
			// milliseconds.
			ms := 50 + (rand.Int63() % 300)
			time.Sleep(time.Duration(ms) * time.Millisecond)
			rf.mu.Unlock()
		}
	}
}

// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here (2A, 2B, 2C).

	// 对应论文中的初始化
	rf.applyCh = applyCh //2B

	rf.currentTerm = 0
	rf.votedFor = -1
	rf.Logs = make([]logEntries, 0)

	rf.commitIndex = 0
	rf.lastApplied = 0

	rf.nextIndex = make([]int, len(peers))
	rf.matchIndex = make([]int, len(peers))

	rf.states = Follower
	rf.overtime = time.Duration(150+rand.Intn(200)) * time.Millisecond // 随机产生150-350ms
	rf.timer = time.NewTicker(rf.overtime)

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	// start ticker goroutine to start elections
	go rf.ticker()

	return rf
}

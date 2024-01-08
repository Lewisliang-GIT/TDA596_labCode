package raft

import "log"

// Debugging
const Debug = false

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug {
		log.Printf(format, a...)
	}
	return
}

func (rf *Raft) restoreLog(curIndex int) LogEntry {
	return rf.logs[curIndex-rf.lastIncludeIndex]
}

// Restore true log tenure with snapshot offsets
func (rf *Raft) restoreLogTerm(curIndex int) int {
	if curIndex-rf.lastIncludeIndex == 0 {
		return rf.lastIncludeTerm
	}
	//fmt.Printf("[GET] curIndex:%v,rf.lastIncludeIndex:%v\n", curIndex, rf.lastIncludeIndex)
	return rf.logs[curIndex-rf.lastIncludeIndex].Term
}

// Get the last snapshot log subscript (for stored)
func (rf *Raft) getLastIndex() int {
	return len(rf.logs) - 1 + rf.lastIncludeIndex
}

// Get the last term (snapshot version)
func (rf *Raft) getLastTerm() int {
	if len(rf.logs)-1 == 0 {
		return rf.lastIncludeTerm
	} else {
		return rf.logs[len(rf.logs)-1].Term
	}
}

// Restore the real PrevLogInfo with snapshot offsets
func (rf *Raft) getPrevLogInfo(server int) (int, int) {
	newEntryBeginIndex := rf.nextIndex[server] - 1
	lastIndex := rf.getLastIndex()
	if newEntryBeginIndex == lastIndex+1 {
		newEntryBeginIndex = lastIndex
	}
	return newEntryBeginIndex, rf.restoreLogTerm(newEntryBeginIndex)
}

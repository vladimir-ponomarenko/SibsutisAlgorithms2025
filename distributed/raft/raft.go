package raft

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	StateFollower  = "Follower"
	StateCandidate = "Candidate"
	StateLeader    = "Leader"
)

type LogEntry struct {
	Term    int
	Command string
}

type Message struct {
	Kind string

	Term        int
	From        int
	To          int
	CandidateID int

	LeaderID     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int

	Success     bool
	MatchIndex  int
	VoteGranted bool

	LastLogIndex int
	LastLogTerm  int
}

type RaftNode struct {
	mu sync.Mutex

	id          int
	state       string
	currentTerm int
	votedFor    int
	leaderID    int

	log         []LogEntry
	commitIndex int
	lastApplied int

	nextIndex  map[int]int
	matchIndex map[int]int

	votesReceived int

	inbox chan Message
	Alive bool

	cluster *Cluster

	electionTimeout  time.Duration
	heartbeatTimeout time.Duration
}

func (rn *RaftNode) GetState() string {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	return rn.state
}

func (rn *RaftNode) GetCurrentTerm() int {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	return rn.currentTerm
}

func (rn *RaftNode) GetLog() []LogEntry {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	logCopy := make([]LogEntry, len(rn.log))
	copy(logCopy, rn.log)
	return logCopy
}

func (rn *RaftNode) GetID() int {
	return rn.id
}

func (rn *RaftNode) IsAlive() bool {
	return rn.Alive
}

func (rn *RaftNode) GetMatchIndex() map[int]int {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	matchIndexCopy := make(map[int]int)
	for k, v := range rn.matchIndex {
		matchIndexCopy[k] = v
	}
	return matchIndexCopy
}

func (rn *RaftNode) GetNextIndex() map[int]int {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	nextIndexCopy := make(map[int]int)
	for k, v := range rn.nextIndex {
		nextIndexCopy[k] = v
	}
	return nextIndexCopy
}

func (rn *RaftNode) AppendLogEntry(entry LogEntry) {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	rn.log = append(rn.log, entry)
}

func (rn *RaftNode) SetMatchIndex(nodeID int, index int) {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	rn.matchIndex[nodeID] = index
}

func (rn *RaftNode) SetNextIndex(nodeID int, index int) {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	rn.nextIndex[nodeID] = index
}

type Cluster struct {
	Nodes    map[int]*RaftNode
	dropRate float64
	delay    time.Duration
	mu       sync.Mutex
}

func NewCluster(nodeIDs []int, dropRate float64, delay time.Duration) *Cluster {
	c := &Cluster{
		Nodes:    make(map[int]*RaftNode),
		dropRate: dropRate,
		delay:    delay,
	}
	for _, id := range nodeIDs {
		rn := &RaftNode{
			id:       id,
			state:    StateFollower,
			votedFor: -1,
			leaderID: -1,
			log:      []LogEntry{},
			inbox:    make(chan Message, 1000),
			Alive:    true,

			nextIndex:  make(map[int]int),
			matchIndex: make(map[int]int),

			electionTimeout:  time.Duration(rand.Intn(1000)+1500) * time.Millisecond,
			heartbeatTimeout: 700 * time.Millisecond,
			cluster:          c,
		}
		c.Nodes[id] = rn
	}
	return c
}

func (c *Cluster) StartAll(wg *sync.WaitGroup) {
	for _, node := range c.Nodes {
		wg.Add(1)
		go node.runLoop(wg)
	}
}

func (c *Cluster) StopAll() {
	for _, node := range c.Nodes {
		node.Alive = false
	}
	for _, node := range c.Nodes {
		close(node.inbox)
	}
}

func (c *Cluster) SendMessage(msg Message) {
	c.mu.Lock()
	defer c.mu.Unlock()

	target, ok := c.Nodes[msg.To]
	if !ok {
		return
	}
	if rand.Float64() < c.dropRate {

		return
	}
	go func(m Message) {
		time.Sleep(c.delay)
		if target.Alive {
			target.inbox <- m
		}
	}(msg)
}

func (rn *RaftNode) runLoop(wg *sync.WaitGroup) {
	defer wg.Done()

	electionTimer := time.NewTimer(rn.electionTimeout)

	var heartbeatTicker *time.Ticker
	resetHeartbeat := func() {
		if heartbeatTicker != nil {
			heartbeatTicker.Stop()
		}
		if rn.state == StateLeader {
			heartbeatTicker = time.NewTicker(rn.heartbeatTimeout)
		} else {
			heartbeatTicker = nil
		}
	}
	resetHeartbeat()

	for rn.Alive {
		select {
		case <-electionTimer.C:
			rn.mu.Lock()
			if rn.state != StateLeader {
				rn.startElection()
			}
			electionTimer.Reset(rn.electionTimeout)
			rn.mu.Unlock()

		case <-func() <-chan time.Time {
			if heartbeatTicker != nil {
				return heartbeatTicker.C
			}
			return make(chan time.Time)
		}():
			rn.mu.Lock()
			if rn.state == StateLeader {
				rn.SendHeartbeats()
			}
			rn.mu.Unlock()

		case msg, ok := <-rn.inbox:
			if !ok {
				return
			}
			rn.mu.Lock()
			rn.handleMessage(msg)
			rn.mu.Unlock()
			electionTimer.Reset(rn.electionTimeout)
		}
		rn.mu.Lock()
		resetHeartbeat()
		rn.mu.Unlock()
	}
}

func (rn *RaftNode) startElection() {
	rn.currentTerm++
	rn.state = StateCandidate
	rn.votedFor = rn.id
	rn.leaderID = -1
	rn.votesReceived = 1

	fmt.Printf("[Node %d] Становлюсь кандидатом, term=%d\n", rn.id, rn.currentTerm)

	lastLogIndex := len(rn.log) - 1
	lastLogTerm := 0
	if lastLogIndex >= 0 {
		lastLogTerm = rn.log[lastLogIndex].Term
	}
	for _, node := range rn.cluster.Nodes {
		if node.id == rn.id {
			continue
		}
		msg := Message{
			Kind:         "RequestVote",
			Term:         rn.currentTerm,
			From:         rn.id,
			To:           node.id,
			CandidateID:  rn.id,
			LastLogIndex: lastLogIndex,
			LastLogTerm:  lastLogTerm,
		}
		rn.cluster.SendMessage(msg)
	}

	go func(termAtStart int) {
		time.Sleep(350 * time.Millisecond)
		rn.mu.Lock()
		defer rn.mu.Unlock()

		if rn.state == StateCandidate && rn.currentTerm == termAtStart {
			majority := len(rn.cluster.Nodes)/2 + 1
			if rn.votesReceived < majority {
				rn.state = StateFollower
				rn.votedFor = -1
			}
		}
	}(rn.currentTerm)
}

func (rn *RaftNode) handleMessage(msg Message) {
	switch msg.Kind {
	case "RequestVote":
		rn.handleRequestVote(msg)
	case "RequestVoteReply":
		rn.handleRequestVoteReply(msg)
	case "AppendEntries":
		rn.handleAppendEntries(msg)
	case "AppendEntriesReply":
		rn.handleAppendEntriesReply(msg)
	}
}

func (rn *RaftNode) handleRequestVote(msg Message) {
	if msg.Term < rn.currentTerm {
		rn.cluster.SendMessage(Message{
			Kind:        "RequestVoteReply",
			Term:        rn.currentTerm,
			From:        rn.id,
			To:          msg.From,
			VoteGranted: false,
		})
		return
	}
	if msg.Term > rn.currentTerm {
		rn.currentTerm = msg.Term
		rn.state = StateFollower
		rn.votedFor = -1
		rn.leaderID = -1
	}
	voteGranted := false
	if (rn.votedFor == -1 || rn.votedFor == msg.CandidateID) &&
		rn.isLogUpToDate(msg.LastLogIndex, msg.LastLogTerm) {
		voteGranted = true
		rn.votedFor = msg.CandidateID
		rn.state = StateFollower
	}
	rn.cluster.SendMessage(Message{
		Kind:        "RequestVoteReply",
		Term:        rn.currentTerm,
		From:        rn.id,
		To:          msg.From,
		VoteGranted: voteGranted,
	})
}

func (rn *RaftNode) isLogUpToDate(candidateLastLogIndex, candidateLastLogTerm int) bool {
	localLastIndex := len(rn.log) - 1
	localLastTerm := 0
	if localLastIndex >= 0 {
		localLastTerm = rn.log[localLastIndex].Term
	}
	if candidateLastLogTerm < localLastTerm {
		return false
	}
	if candidateLastLogTerm == localLastTerm && candidateLastLogIndex < localLastIndex {
		return false
	}
	return true
}

func (rn *RaftNode) handleRequestVoteReply(msg Message) {
	if msg.Term > rn.currentTerm {
		rn.currentTerm = msg.Term
		rn.state = StateFollower
		rn.votedFor = -1
		rn.leaderID = -1
		return
	}
	if rn.state != StateCandidate || msg.Term < rn.currentTerm {
		return
	}
	if msg.VoteGranted {
		rn.votesReceived++
		majority := len(rn.cluster.Nodes)/2 + 1
		if rn.votesReceived >= majority && rn.state == StateCandidate {
			rn.becomeLeader()
		}
	}
}

func (rn *RaftNode) becomeLeader() {
	rn.state = StateLeader
	rn.leaderID = rn.id
	fmt.Printf("[Node %d] Становлюсь лидером, term=%d\n", rn.id, rn.currentTerm)

	for _, node := range rn.cluster.Nodes {
		rn.nextIndex[node.id] = len(rn.log)
		rn.matchIndex[node.id] = -1
	}
	rn.SendHeartbeats()
}

func (rn *RaftNode) handleAppendEntries(msg Message) {
	if msg.Term < rn.currentTerm {
		rn.cluster.SendMessage(Message{
			Kind:       "AppendEntriesReply",
			Term:       rn.currentTerm,
			From:       rn.id,
			To:         msg.From,
			Success:    false,
			MatchIndex: -1,
		})
		return
	}
	if msg.Term > rn.currentTerm {
		rn.currentTerm = msg.Term
	}
	rn.state = StateFollower
	rn.leaderID = msg.LeaderID
	rn.votedFor = msg.LeaderID

	if msg.PrevLogIndex >= 0 {
		if msg.PrevLogIndex >= len(rn.log) {
			rn.cluster.SendMessage(Message{
				Kind:       "AppendEntriesReply",
				Term:       rn.currentTerm,
				From:       rn.id,
				To:         msg.From,
				Success:    false,
				MatchIndex: len(rn.log) - 1,
			})
			return
		} else {
			if rn.log[msg.PrevLogIndex].Term != msg.PrevLogTerm {

				rn.log = rn.log[:msg.PrevLogIndex]
				rn.cluster.SendMessage(Message{
					Kind:       "AppendEntriesReply",
					Term:       rn.currentTerm,
					From:       rn.id,
					To:         msg.From,
					Success:    false,
					MatchIndex: len(rn.log) - 1,
				})
				return
			}
		}
	}

	for i, entry := range msg.Entries {
		index := msg.PrevLogIndex + 1 + i
		if index < len(rn.log) {
			rn.log[index] = entry
		} else {
			rn.log = append(rn.log, entry)
		}
	}

	if msg.LeaderCommit > rn.commitIndex {
		lastNewEntry := msg.PrevLogIndex + len(msg.Entries)
		commitIndex := msg.LeaderCommit
		if lastNewEntry < commitIndex {
			commitIndex = lastNewEntry
		}
		if commitIndex > rn.commitIndex {
			rn.commitIndex = commitIndex
			rn.applyLogEntries()
		}
	}
	rn.cluster.SendMessage(Message{
		Kind:       "AppendEntriesReply",
		Term:       rn.currentTerm,
		From:       rn.id,
		To:         msg.From,
		Success:    true,
		MatchIndex: len(rn.log) - 1,
	})
}

func (rn *RaftNode) handleAppendEntriesReply(msg Message) {
	if msg.Term > rn.currentTerm {
		rn.state = StateFollower
		rn.currentTerm = msg.Term
		rn.votedFor = -1
		rn.leaderID = -1
		return
	}
	if rn.state != StateLeader {
		return
	}
	if msg.Success {
		rn.matchIndex[msg.From] = msg.MatchIndex
		rn.nextIndex[msg.From] = msg.MatchIndex + 1
		rn.updateCommitIndex()
	} else {
		rn.nextIndex[msg.From] = msg.MatchIndex + 1
		if rn.nextIndex[msg.From] < 0 {
			rn.nextIndex[msg.From] = 0
		}
		rn.SendAppendEntries(msg.From)
	}
}

func (rn *RaftNode) SendHeartbeats() {
	for _, node := range rn.cluster.Nodes {
		if node.id == rn.id {
			continue
		}
		rn.SendAppendEntries(node.id)
	}
}

func (rn *RaftNode) SendAppendEntries(followerID int) {
	prevIndex := rn.nextIndex[followerID] - 1
	prevTerm := 0
	if prevIndex >= 0 && prevIndex < len(rn.log) {
		prevTerm = rn.log[prevIndex].Term
	}
	var entries []LogEntry
	if rn.nextIndex[followerID] < len(rn.log) {
		entries = rn.log[rn.nextIndex[followerID]:]
	}
	rn.cluster.SendMessage(Message{
		Kind:         "AppendEntries",
		Term:         rn.currentTerm,
		From:         rn.id,
		To:           followerID,
		LeaderID:     rn.id,
		PrevLogIndex: prevIndex,
		PrevLogTerm:  prevTerm,
		Entries:      entries,
		LeaderCommit: rn.commitIndex,
	})
}

func (rn *RaftNode) updateCommitIndex() {
	for i := rn.commitIndex + 1; i < len(rn.log); i++ {
		count := 0
		for _, node := range rn.cluster.Nodes {
			if rn.matchIndex[node.id] >= i {
				count++
			}
		}
		if count >= (len(rn.cluster.Nodes)/2+1) && rn.log[i].Term == rn.currentTerm {
			rn.commitIndex = i
			rn.applyLogEntries()
		}
	}
}

func (rn *RaftNode) applyLogEntries() {
	for rn.lastApplied < rn.commitIndex {
		rn.lastApplied++
		entry := rn.log[rn.lastApplied]
		fmt.Printf("[Node %d] Применяю команду (index=%d, term=%d, cmd=%s)\n",
			rn.id, rn.lastApplied, entry.Term, entry.Command)
	}
}

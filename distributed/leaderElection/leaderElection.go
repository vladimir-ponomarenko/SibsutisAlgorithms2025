package leaderElection

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type MessageType string

// Сообщения в алгоритме
const (
	ElectionMsg        MessageType = "ELECTION"
	OkMsg              MessageType = "OK"
	CoordinatorMsg     MessageType = "COORDINATOR"
	RingCoordinatorMsg MessageType = "RING_COORDINATOR"
	RingElectionMsg    MessageType = "RING_ELECTION"
	CollectMsg         MessageType = "COLLECT"
	CollectReplyMsg    MessageType = "COLLECT_REPLY"
)

// Сообщения узлов
type Message struct {
	Kind   MessageType `json:"kind"`
	FromID int         `json:"from"`
	Data   interface{} `json:"data"`
	To     int         `json:"to"`
	Ids    []int       `json:"ids"`
	MaxID  int         `json:"max_id"`
}

// Узел
type Node struct {
	ID         int
	IsLeader   bool
	Alive      bool
	LeaderID   int
	NextID     int
	LocalCount int
	Inbox      chan Message
	nodes      map[int]chan Message
	Mu         sync.Mutex
	Wg         sync.WaitGroup
}

var Dispatcher = struct {
	sync.RWMutex
	Channels map[int]chan Message
}{Channels: make(map[int]chan Message)}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Новый узел
func NewNode(id, totalNodes int) *Node {
	n := &Node{
		ID:         id,
		Alive:      true,
		LeaderID:   -1,
		LocalCount: rand.Intn(50) + 50,
		Inbox:      make(chan Message, 10),
	}

	Dispatcher.Lock()
	Dispatcher.Channels[id] = n.Inbox
	Dispatcher.Unlock()

	if totalNodes > 0 {
		n.NextID = (id + 1) % totalNodes
	}
	return n
}

// Запускает обработку входящих сообщений
func (n *Node) Start() {
	n.Wg.Add(1)
	go n.Listen()
}

// Обрабатывает сообщения
func (n *Node) Listen() {
	defer n.Wg.Done()
	for msg := range n.Inbox {
		if !n.Alive {
			continue
		}

		switch msg.Kind {
		case ElectionMsg:
			n.HandleElection(msg)
		case OkMsg:
			n.HandleOK(msg)
		case CoordinatorMsg:
			n.HandleCoordinator(msg)
		case RingCoordinatorMsg: // Добавляем обработку нового типа сообщения
			n.HandleRingCoordinator(msg)
		case RingElectionMsg:
			n.HandleRingElection(msg)
		case CollectMsg:
			n.HandleCollect(msg)
		case CollectReplyMsg:
			n.HandleCollectReply(msg)
		}
	}
}

// Запуск выборов лидера
func (n *Node) StartElection() {
	n.Mu.Lock()
	defer n.Mu.Unlock()

	fmt.Printf("Node %d starts election\n", n.ID)
	higherNodes := false
	responseChan := make(chan bool, len(Dispatcher.Channels))

	for id := range Dispatcher.Channels {
		if id > n.ID {
			higherNodes = true
			go func(targetID int) {
				msg := Message{Kind: ElectionMsg, FromID: n.ID}
				Dispatcher.RLock()
				ch, ok := Dispatcher.Channels[targetID]
				Dispatcher.RUnlock()
				if ok {
					ch <- msg
					select {
					case <-time.After(1 * time.Second):
						responseChan <- false
					case <-time.After(100 * time.Millisecond):
						responseChan <- true
					}
				}
			}(id)
		}
	}

	if !higherNodes {
		n.DeclareLeader()
		return
	}

	timeout := time.After(2 * time.Second)
	receivedOK := false

WaitForResponses:
	for i := 0; i < cap(responseChan); i++ {
		select {
		case ok := <-responseChan:
			if ok {
				receivedOK = true
				break WaitForResponses
			}
		case <-timeout:
			break WaitForResponses
		}
	}

	if !receivedOK {
		n.DeclareLeader()
	}
}

// Объявляет узел лидером
func (n *Node) DeclareLeader() {
	fmt.Printf("Node %d declares itself leader\n", n.ID)
	n.IsLeader = true
	n.LeaderID = n.ID
	msg := Message{Kind: CoordinatorMsg, FromID: n.ID, Data: n.ID}
	for id, ch := range Dispatcher.Channels {
		if id != n.ID {
			ch <- msg
		}
	}
}

// Обрабатывает ELECTION
func (n *Node) HandleElection(msg Message) {
	if msg.FromID < n.ID && n.Alive {
		fmt.Printf("Node %d responds OK to %d\n", n.ID, msg.FromID)
		Dispatcher.RLock()
		ch, ok := Dispatcher.Channels[msg.FromID]
		Dispatcher.RUnlock()
		if ok {
			ch <- Message{Kind: OkMsg, FromID: n.ID}
		}
		n.StartElection()
	}
}

// Обрабатывает ОК
func (n *Node) HandleOK(msg Message) {
	fmt.Printf("Node %d received OK from %d\n", n.ID, msg.FromID)
}

// Обрабатывает COORDINATOR
func (n *Node) HandleCoordinator(msg Message) {
	n.LeaderID = msg.Data.(int)
	fmt.Printf("Node %d acknowledges leader %d\n", n.ID, n.LeaderID)
}

// Выборы на основе кольца
func (n *Node) StartRingElection() {
	n.Mu.Lock()
	defer n.Mu.Unlock()

	fmt.Printf("Node %d starts ring election\n", n.ID)
	msg := Message{
		Kind:   RingElectionMsg,
		FromID: n.ID,
		Data:   n.ID,
		To:     n.NextID,
	}

	Dispatcher.RLock()
	nextCh := Dispatcher.Channels[n.NextID]
	Dispatcher.RUnlock()
	fmt.Printf("Node %d sends ELECTION message with ID=%d to Node %d\n",
		n.ID, n.ID, n.NextID)
	nextCh <- msg
}

// Обрабатывает RINGELECTION
func (n *Node) HandleRingElection(msg Message) {
	if !n.Alive {
		return
	}

	currentMaxID := msg.Data.(int)
	initiator := msg.FromID

	fmt.Printf("Node %d received ELECTION message (currentMax=%d) from previous node\n",
		n.ID, currentMaxID)

	if initiator == n.ID {
		fmt.Printf("Election message returned to initiator Node %d. Maximum ID is %d\n",
			n.ID, currentMaxID)

		if currentMaxID > n.ID {
			msg := Message{
				Kind:   RingCoordinatorMsg,
				FromID: n.ID,
				Data:   currentMaxID,
				To:     n.NextID,
			}
			Dispatcher.RLock()
			nextCh := Dispatcher.Channels[n.NextID]
			Dispatcher.RUnlock()
			nextCh <- msg
		} else {
			n.IsLeader = true
			n.LeaderID = n.ID
			fmt.Printf("Node %d becomes leader and sends RingCoordinator message\n", n.ID)

			msg := Message{
				Kind:   RingCoordinatorMsg,
				FromID: n.ID,
				Data:   n.ID,
				To:     n.NextID,
			}
			Dispatcher.RLock()
			nextCh := Dispatcher.Channels[n.NextID]
			Dispatcher.RUnlock()
			nextCh <- msg
		}
		return
	}

	if n.ID > currentMaxID {
		fmt.Printf("Node %d replaces election ID %d with own ID %d\n",
			n.ID, currentMaxID, n.ID)
		currentMaxID = n.ID
	}

	msg.Data = currentMaxID
	msg.To = n.NextID

	Dispatcher.RLock()
	nextCh := Dispatcher.Channels[n.NextID]
	Dispatcher.RUnlock()
	fmt.Printf("Node %d forwards ELECTION message with ID=%d to Node %d\n",
		n.ID, currentMaxID, n.NextID)
	nextCh <- msg
}

// Обработчик для RingCoordinator сообщений
func (n *Node) HandleRingCoordinator(msg Message) {
	leaderID := msg.Data.(int)
	n.LeaderID = leaderID

	if leaderID == n.ID {
		n.IsLeader = true
	}

	fmt.Printf("Node %d acknowledges ring leader %d\n", n.ID, leaderID)

	if msg.FromID != n.ID {
		msg.To = n.NextID
		Dispatcher.RLock()
		nextCh := Dispatcher.Channels[n.NextID]
		Dispatcher.RUnlock()
		nextCh <- msg
	}
}

// Сбор данных от всех узлов
func (n *Node) StartGlobalCollection() {
	if !n.IsLeader {
		return
	}

	fmt.Printf("Leader %d initiating data collection\n", n.ID)
	n.Wg.Add(len(Dispatcher.Channels) - 1)
	total := n.LocalCount

	go func() {
		timeout := time.After(5 * time.Second)
		remaining := len(Dispatcher.Channels) - 1
		for remaining > 0 {
			select {
			case msg := <-n.Inbox:
				if msg.Kind == CollectReplyMsg {
					total += msg.Data.(int)
					remaining--
				}
			case <-timeout:
				fmt.Printf("Leader %d timed out waiting for responses\n", n.ID)
				return
			}
		}
		fmt.Printf("Leader %d collected total: %d\n", n.ID, total)
	}()

	for id := range Dispatcher.Channels {
		if id != n.ID {
			Dispatcher.RLock()
			ch := Dispatcher.Channels[id]
			Dispatcher.RUnlock()
			ch <- Message{Kind: CollectMsg, FromID: n.ID}
		}
	}
}

// Обрабатывает COLLECT
func (n *Node) HandleCollect(msg Message) {
	if n.Alive {
		reply := Message{
			Kind:   CollectReplyMsg,
			FromID: n.ID,
			Data:   n.LocalCount,
		}
		Dispatcher.RLock()
		ch := Dispatcher.Channels[msg.FromID]
		Dispatcher.RUnlock()
		ch <- reply
	}
}

// Обрабатывает COLLECTREPLY
func (n *Node) HandleCollectReply(msg Message) {
	n.Wg.Done()
}

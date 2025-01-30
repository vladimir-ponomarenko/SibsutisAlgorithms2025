package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"SibsutisAlgorithms2025/distributed/raft"
)

func sendCommandToLeader(c *raft.Cluster, leaderID int, command string) {
	leader := c.Nodes[leaderID]
	if leader.GetState() != raft.StateLeader {
		fmt.Printf("[CLIENT] Узел %d уже не лидер, команда '%s' не отправлена\n", leaderID, command)
		return
	}

	leader.AppendLogEntry(raft.LogEntry{
		Term:    leader.GetCurrentTerm(),
		Command: command,
	})
	logLen := len(leader.GetLog())
	index := logLen - 1
	matchIndexMap := leader.GetMatchIndex()
	nextIndexMap := leader.GetNextIndex()
	matchIndexMap[leader.GetID()] = index
	nextIndexMap[leader.GetID()] = index + 1
	leader.SetMatchIndex(leader.GetID(), index)
	leader.SetNextIndex(leader.GetID(), index+1)

	fmt.Printf("[CLIENT] Отправлена команда '%s' лидеру %d (локальный индекс=%d)\n",
		command, leaderID, index)
	leader.SendHeartbeats()
}

func demonstrateLogConflict(c *raft.Cluster) {
	fmt.Println("\n=== Конфликтная ситуация ===")

	var leaderID int
	for _, n := range c.Nodes {
		if n.IsAlive() && n.GetState() == raft.StateLeader {
			leaderID = n.GetID()
			break
		}
	}
	if leaderID == 0 {
		fmt.Println("[CONFLICT-DEMO] Нет лидера, пропускаем конфликтную демонстрацию")
		return
	}

	sendCommandToLeader(c, leaderID, "A")
	time.Sleep(200 * time.Millisecond)

	fmt.Printf("[CONFLICT-DEMO] Убиваем текущего лидера %d\n", leaderID)
	c.Nodes[leaderID].Alive = false

	time.Sleep(2 * time.Second)

	var newLeaderID int
	for _, n := range c.Nodes {
		if n.IsAlive() && n.GetState() == raft.StateLeader {
			newLeaderID = n.GetID()
			break
		}
	}
	if newLeaderID == 0 {
		// fmt.Println("[CONFLICT-DEMO] Не удалось выбрать нового лидера ")
		return
	}
	fmt.Printf("[CONFLICT-DEMO] Новый лидер: %d\n", newLeaderID)

	sendCommandToLeader(c, newLeaderID, "B")
	time.Sleep(2 * time.Second)

	fmt.Println("[CONFLICT-DEMO] Содержимое логов после конфликта:")
	for _, n := range c.Nodes {
		if !n.IsAlive() {
			fmt.Printf("  Node %d: (dead)\n", n.GetID())
			continue
		}
		fmt.Printf("  Node %d log: ", n.GetID())
		log := n.GetLog()
		for i, e := range log {
			fmt.Printf("[%d:%s:%d] ", i, e.Command, e.Term)
		}
		fmt.Println()
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	nodeIDs := []int{1, 2, 3, 4, 5}
	c := raft.NewCluster(nodeIDs, 0.15, 100*time.Millisecond)

	var wg sync.WaitGroup
	c.StartAll(&wg)

	time.Sleep(3 * time.Second)

	var leaderID int
	for _, n := range c.Nodes {
		if n.GetState() == raft.StateLeader {
			leaderID = n.GetID()
			break
		}
	}
	if leaderID == 0 {
		fmt.Println("[MAIN] Не удалось определить лидера :(")
	} else {
		fmt.Printf("[MAIN] Определён лидер: %d\n", leaderID)

		sendCommandToLeader(c, leaderID, "X")
		sendCommandToLeader(c, leaderID, "Y")

		time.Sleep(2 * time.Second)

		fmt.Println("=== Логи узлов перед конфликтной ситуацией ===")
		for _, n := range c.Nodes {
			fmt.Printf("  Node %d (state=%s, term=%d) log: ", n.GetID(), n.GetState(), n.GetCurrentTerm())
			log := n.GetLog()
			for i, e := range log {
				fmt.Printf("[%d:%s:%d] ", i, e.Command, e.Term)
			}
			fmt.Println()
		}
	}

	demonstrateLogConflict(c)

	time.Sleep(3 * time.Second)
	c.StopAll()
	wg.Wait()

	fmt.Println("[MAIN] Завершение программы")
}

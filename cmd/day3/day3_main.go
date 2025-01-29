package main

import (
	"SibsutisAlgorithms2025/distributed/leaderElection"
	"fmt"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	const numNodes = 5
	nodes := make([]*leaderElection.Node, numNodes)

	for i := 0; i < numNodes; i++ {
		nodes[i] = leaderElection.NewNode(i, numNodes)
		nodes[i].Start()
	}

	// Узел 0 замечает отсутствие лидера и запускает Bully ELECTION
	fmt.Println("=== Simulation 1: Bully Election ===")
	time.Sleep(1 * time.Second)
	fmt.Println("Node 0 notices the absence of a leader and initiates Bully Election")
	nodes[0].StartElection()

	time.Sleep(3 * time.Second)

	for _, node := range nodes {
		if node.IsLeader && node.Alive {
			fmt.Println("Leader starts collecting data...")
			node.StartGlobalCollection()
			break
		}
	}

	time.Sleep(6 * time.Second)

	// Узел 3 не отвечает "падение"
	fmt.Println("=== Simulation 2: Node Failure ===")
	fmt.Println("Simulating failure of Node 3...")
	nodes[3].Alive = false

	// Узел 1 замечает отсутствие лидера и запускает Bully ELECTION
	time.Sleep(3 * time.Second)
	fmt.Println("Node 1 notices the absence of a leader and initiates Bully Election")
	nodes[1].StartElection()

	time.Sleep(3 * time.Second)

	for _, node := range nodes {
		if node.IsLeader && node.Alive {
			fmt.Println("New leader starts collecting data...")
			node.StartGlobalCollection()
			break
		}
	}

	time.Sleep(6 * time.Second)

	// Узел 0 замечает отсутствие лидера и запускает Ring ELECTION
	fmt.Println("=== Simulation 3: Ring Election ===")
	leaderElection.Dispatcher.Lock()
	for i := 0; i < numNodes; i++ {
		nodes[i].IsLeader = false
		nodes[i].LeaderID = -1
	}
	leaderElection.Dispatcher.Unlock()
	nodes[3].Alive = true

	time.Sleep(1 * time.Second)
	fmt.Println("Node 0 notices the absence of a leader and initiates Ring Election")
	nodes[0].StartRingElection()

	time.Sleep(3 * time.Second)

	for _, node := range nodes {
		if node.IsLeader && node.Alive {
			fmt.Println("Leader starts collecting data...")
			node.StartGlobalCollection()
			break
		}
	}

	time.Sleep(6 * time.Second)
	fmt.Println("Simulation complete.")
}

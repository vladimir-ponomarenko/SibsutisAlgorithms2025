package main

import (
	bellmanford "SibsutisAlgorithms2025/algorithms/bellman-ford"
	"SibsutisAlgorithms2025/algorithms/dijkstra"
	Graph "SibsutisAlgorithms2025/graph"
	"SibsutisAlgorithms2025/graph/Mst"
	"fmt"
)

func main() {
	fmt.Println("--- DisjointSet ---")
	ds := Graph.NewDisjointSet(6)
	ds.Union(0, 1)
	ds.Union(1, 2)
	ds.Union(3, 4)

	fmt.Println("Find(0):", ds.Find(0))
	fmt.Println("Find(1):", ds.Find(1))
	fmt.Println("Find(2):", ds.Find(2))
	fmt.Println("Find(3):", ds.Find(3))
	fmt.Println("Find(4):", ds.Find(4))
	fmt.Println("Find(5):", ds.Find(5))

	ds.Union(2, 3)
	fmt.Println("After Union(2,3):")
	fmt.Println("Find(0):", ds.Find(0))
	fmt.Println("Find(1):", ds.Find(1))
	fmt.Println("Find(2):", ds.Find(2))
	fmt.Println("Find(3):", ds.Find(3))
	fmt.Println("Find(4):", ds.Find(4))
	fmt.Println("Find(5):", ds.Find(5))

	fmt.Println("Parent array:", ds.Parent)

	fmt.Println("\n--- MST ---")
	edges := []Graph.Edge{
		{U: 0, V: 1, W: 4},
		{U: 0, V: 2, W: 8},
		{U: 1, V: 2, W: 11},
		{U: 1, V: 3, W: 9},
		{U: 2, V: 3, W: 7},
		{U: 2, V: 4, W: 2},
		{U: 3, V: 4, W: 6},
		{U: 3, V: 5, W: 10},
		{U: 4, V: 5, W: 5},
	}

	mstEdges, totalWeight := Mst.MST(6, edges)
	fmt.Println("MST Edges:")
	for _, edge := range mstEdges {
		fmt.Printf("Edge{U:%d, V:%d, W:%d}\n", edge.U, edge.V, edge.W)
	}
	fmt.Println("Total MST Weight:", totalWeight)

	fmt.Println("\n--- Dijkstra ---")
	gDijkstra := Graph.NewGraph()
	gDijkstra.AddEdge(0, 1, 4, false)
	gDijkstra.AddEdge(0, 2, 8, false)
	gDijkstra.AddEdge(1, 2, 11, false)
	gDijkstra.AddEdge(1, 3, 9, false)
	gDijkstra.AddEdge(2, 3, 7, false)
	gDijkstra.AddEdge(2, 4, 2, false)
	gDijkstra.AddEdge(3, 4, 6, false)
	gDijkstra.AddEdge(3, 5, 10, false)
	gDijkstra.AddEdge(4, 5, 5, false)

	distDijkstra, parentDijkstra := dijkstra.Dijkstra(gDijkstra, 0)
	fmt.Println("Dijkstra Distances from vertex 0:", distDijkstra)
	fmt.Println("Dijkstra Parents:", parentDijkstra)

	fmt.Println("\n--- BellmanFord ---")
	gBellmanFord := Graph.NewGraph()
	gBellmanFord.AddEdge(0, 1, -1, false)
	gBellmanFord.AddEdge(0, 2, 4, false)
	gBellmanFord.AddEdge(1, 2, 3, false)
	gBellmanFord.AddEdge(1, 3, 2, false)
	gBellmanFord.AddEdge(1, 4, 2, false)
	gBellmanFord.AddEdge(3, 2, 5, false)
	gBellmanFord.AddEdge(3, 1, 1, false)
	gBellmanFord.AddEdge(4, 3, -3, false)

	distBellmanFord, parentBellmanFord, hasNegativeCycle := bellmanford.BellmanFord(gBellmanFord, 0)
	fmt.Println("BellmanFord Distances from vertex 0:", distBellmanFord)
	fmt.Println("BellmanFord Parents:", parentBellmanFord)
	fmt.Println("Negative Cycle Detected:", hasNegativeCycle)

	gNegativeCycle := Graph.NewGraph()
	gNegativeCycle.AddEdge(0, 1, 1, false)
	gNegativeCycle.AddEdge(1, 2, -3, false)
	gNegativeCycle.AddEdge(2, 0, 1, false)

}

package main

import (
	Graph "SibsutisAlgorithms2025/graph"
	"fmt"
)

func main() {

	gCC := Graph.NewGraph()

	gCC.AddEdge(0, 1, 1, true)
	gCC.AddEdge(1, 2, 1, true)

	gCC.AddEdge(4, 5, 1, true)
	gCC.AddEdge(5, 6, 1, true)

	g := Graph.NewGraph()

	g.AddEdge(0, 1, 1, false)
	g.AddEdge(0, 2, 1, false)
	g.AddEdge(1, 3, 1, false)
	g.AddEdge(1, 4, 1, false)
	g.AddEdge(2, 5, 1, false)

	fmt.Println("\n--- Проверка HasEdge ---")
	fmt.Println("HasEdge(0, 1):", g.HasEdge(0, 1))
	fmt.Println("HasEdge(0, 2):", g.HasEdge(0, 2))
	fmt.Println("HasEdge(1, 2):", g.HasEdge(1, 2))
	fmt.Println("HasEdge(2, 3):", g.HasEdge(2, 3))

	fmt.Println("HasEdge(1, 0):", g.HasEdge(1, 0))
	fmt.Println("HasEdge(2, 0):", g.HasEdge(2, 0))
	fmt.Println("HasEdge(2, 1):", g.HasEdge(2, 1))
	fmt.Println("HasEdge(3, 2):", g.HasEdge(3, 2))
	fmt.Println("HasEdge(0, 3):", g.HasEdge(0, 3))
	fmt.Println("HasEdge(3, 0):", g.HasEdge(3, 0))
	fmt.Println("HasEdge(1, 3):", g.HasEdge(1, 3))

	fmt.Println("\n--- Вывод BFS и DFS ---")
	bfsOrder := Graph.BFS(g, 0)
	dfsOrder := Graph.DFS(g, 0)

	fmt.Print("Порядок обхода BFS: ")
	for _, v := range bfsOrder {
		fmt.Printf("%d ", v)
	}
	fmt.Println()

	fmt.Print("Порядок обхода DFS: ")
	for _, v := range dfsOrder {
		fmt.Printf("%d ", v)
	}
	fmt.Println()

	fmt.Println("\n--- Проверка ConnectedComponents ---")
	countCC2, compMap2 := Graph.ConnectedComponents(gCC)
	fmt.Println("Количество компонент связности:", countCC2)
	fmt.Println("Компоненты по вершинам:", compMap2)
}

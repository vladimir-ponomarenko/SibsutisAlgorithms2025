package bellmanford

import (
	Graph "SibsutisAlgorithms2025/graph"
	"math"
)

func BellmanFord(g *Graph.Graph, start int) ([]int, []int, bool) {
	maxVertex := 0
	for u := range g.Adj {
		if u > maxVertex {
			maxVertex = u
		}
		for _, neighbor := range g.Adj[u] {
			if neighbor.To > maxVertex {
				maxVertex = neighbor.To
			}
		}
	}
	n := maxVertex + 1

	dist := make([]int, n)
	parent := make([]int, n)
	negativeCycle := false

	for i := 0; i < n; i++ {
		dist[i] = math.MaxInt
		parent[i] = -1
	}
	dist[start] = 0

	for i := 1; i < n; i++ {

		for _, edge := range g.GetAllEdges() {
			u, v, weight := edge.U, edge.V, edge.W

			if u < n && v < n {
				if dist[u] != math.MaxInt && dist[u]+weight < dist[v] {
					dist[v] = dist[u] + weight
					parent[v] = u
				}
			}
		}
	}

	for _, edge := range g.GetAllEdges() {
		u, v, weight := edge.U, edge.V, edge.W
		if u < n && v < n {
			if dist[u] != math.MaxInt && dist[u]+weight < dist[v] {
				negativeCycle = true
				break
			}
		}
	}

	return dist, parent, negativeCycle
}

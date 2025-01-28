package dijkstra

import (
	Graph "SibsutisAlgorithms2025/graph"
	"math"
)

func Dijkstra(g *Graph.Graph, start int) ([]int, []int) {
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
	visited := make([]bool, n)

	for i := 0; i < n; i++ {
		dist[i] = math.MaxInt
		parent[i] = -1
	}

	dist[start] = 0

	for i := 0; i < n; i++ {
		u := -1
		minDist := math.MaxInt
		for v := 0; v < n; v++ {
			if !visited[v] && dist[v] < minDist {
				minDist = dist[v]
				u = v
			}
		}

		if u == -1 {
			break
		}
		visited[u] = true

		for _, neighbor := range g.GetNeighbors(u) {
			v := neighbor.V
			weight := neighbor.W

			if dist[u]+weight < dist[v] {
				dist[v] = dist[u] + weight
				parent[v] = u
			}
		}
	}
	return dist, parent
}

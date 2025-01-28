package Mst

import (
	algorithms "SibsutisAlgorithms2025/algorithms/sort"
	Graph "SibsutisAlgorithms2025/graph"
)

func MST(n int, edges []Graph.Edge) (mst []Graph.Edge, totalWeight int) {
	sortedEdges := algorithms.MergeSort(edges)
	ds := Graph.NewDisjointSet(n)
	mst = make([]Graph.Edge, 0)
	totalWeight = 0

	for _, edge := range sortedEdges {
		u, v, w := edge.U, edge.V, edge.W

		rootU := ds.Find(u)
		rootV := ds.Find(v)

		if rootU != rootV {
			ds.Union(u, v)
			mst = append(mst, edge)
			totalWeight += w
		}

		if len(mst) == n-1 {
			break
		}
	}
	return mst, totalWeight
}

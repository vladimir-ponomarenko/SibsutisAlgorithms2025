package algorithms

import Graph "SibsutisAlgorithms2025/graph"

func MergeSort(edges []Graph.Edge) []Graph.Edge {
	if len(edges) <= 1 {
		return edges
	}

	mid := len(edges) / 2
	left := MergeSort(edges[:mid])
	right := MergeSort(edges[mid:])

	return merge(left, right)
}

func merge(left []Graph.Edge, right []Graph.Edge) []Graph.Edge {
	merged := make([]Graph.Edge, 0, len(left)+len(right))
	leftIndex, rightIndex := 0, 0
	for leftIndex < len(left) && rightIndex < len(right) {
		if left[leftIndex].W < right[rightIndex].W {
			merged = append(merged, left[leftIndex])
			leftIndex++
		} else {
			merged = append(merged, right[rightIndex])
			rightIndex++
		}
	}

	for leftIndex < len(left) {
		merged = append(merged, left[leftIndex])
		leftIndex++
	}

	for rightIndex < len(right) {
		merged = append(merged, right[rightIndex])
		rightIndex++
	}

	return merged
}

package Graph

type DisjointSet struct {
	Parent []int
	Size   []int
}

// Структура Disjoint Set Union
func NewDisjointSet(n int) *DisjointSet {
	ds := &DisjointSet{
		Parent: make([]int, n),
		Size:   make([]int, n),
	}
	for i := 0; i < n; i++ {
		ds.Parent[i] = i
		ds.Size[i] = 1
	}

	return ds
}

// Находит корень множества
func (ds *DisjointSet) Find(x int) int {
	if ds.Parent[x] != x {
		ds.Parent[x] = ds.Find(ds.Parent[x])
	}
	return ds.Parent[x]
}

// true - множества объединены, иначе false
func (ds *DisjointSet) Union(x, y int) bool {
	rx := ds.Find(x)
	ry := ds.Find(y)

	if rx == ry {
		return false
	}

	if ds.Size[rx] < ds.Size[ry] {
		ds.Parent[rx] = ry
		ds.Size[ry] += ds.Parent[rx]
	} else {
		ds.Parent[ry] = rx
		ds.Size[rx] += ds.Parent[ry]
	}

	return true
}

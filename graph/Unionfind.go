package Graph

type DisjointSet struct {
	parent []int // родитель i
	size   []int // размер множества, корень - i
}

// Структура Disjoint Set Union
func NewDisjointSet(n int) *DisjointSet {
	ds := &DisjointSet{
		parent: make([]int, n),
		size:   make([]int, n),
	}
	for i := 0; i < n; i++ {
		ds.parent[i] = i
		ds.size[i] = 1
	}

	return ds
}

// Находит корень множества
func (ds *DisjointSet) Find(x int) int {
	if ds.parent[x] != x {
		ds.parent[x] = ds.Find(ds.parent[x])
	}
	return ds.parent[x]
}

// true - множества объединены, иначе false
func (ds *DisjointSet) Union(x, y int) bool {
	rx := ds.Find(x)
	ry := ds.Find(y)

	if rx == ry {
		return false
	}

	if ds.size[rx] < ds.size[ry] {
		ds.parent[rx] = ry
		ds.size[ry] += ds.parent[rx]
	} else {
		ds.parent[ry] = rx
		ds.size[rx] += ds.parent[ry]
	}

	return true
}

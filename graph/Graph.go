package Graph

// Рёбра с весом
type Edge struct {
	U int
	V int
	W int
}

// Граф
// To: Соседняя вершина
// Weight: Вес ребра, ведущего к соседней вершине
type Graph struct {
	Adj map[int][]struct {
		To     int
		Weight int
	}
	edges []Edge
}

func NewGraph() *Graph {
	return &Graph{
		Adj: make(map[int][]struct {
			To     int
			Weight int
		}),
		edges: make([]Edge, 0),
	}
}

// Добавляет ребро в граф
func (g *Graph) AddEdge(u int, v int, weight int, undirected bool) {
	g.Adj[u] = append(g.Adj[u], struct {
		To     int
		Weight int
	}{To: v, Weight: weight})
	if undirected {
		g.Adj[v] = append(g.Adj[v], struct {
			To     int
			Weight int
		}{To: u, Weight: weight})
	}

	newEdge := Edge{U: u, V: v, W: weight}
	g.edges = append(g.edges, newEdge)

	if undirected {
		reverseEdge := Edge{U: v, V: u, W: weight}
		g.edges = append(g.edges, reverseEdge)
	}
}

// Проверка на ребро в графе
func (g *Graph) HasEdge(u int, v int) bool {
	if g.Adj[u] == nil {
		return false
	}
	for _, neighbor := range g.Adj[u] {
		if neighbor.To == v {
			return true
		}
	}
	return false
}

// Возвращает список ребёр
func (g *Graph) GetAllEdges() []Edge {
	return g.edges
}

func (g *Graph) GetNeighbors(u int) []struct {
	V int
	W int
} {
	neighborsWithWeights := []struct {
		V int
		W int
	}{}

	if g.Adj[u] == nil {
		return neighborsWithWeights
	}

	for _, neighbor := range g.Adj[u] {
		neighborsWithWeights = append(neighborsWithWeights, struct {
			V int
			W int
		}{V: neighbor.To, W: neighbor.Weight})
	}

	return neighborsWithWeights
}

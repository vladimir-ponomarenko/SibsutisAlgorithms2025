package Graph

type Graph struct {
	adj map[int][]int
}

func NewGraph() *Graph {
	return &Graph{adj: make(map[int][]int)}
}

// Добавляет ребро в граф
func (g *Graph) AddEdge(u int, v int, undirected bool) {
	g.adj[u] = append(g.adj[u], v)
	if undirected {
		g.adj[v] = append(g.adj[v], u)
	}
}

// Проверка на ребро в графе
func (g *Graph) HasEdge(u int, v int) bool {
	if g.adj[u] == nil {
		return false
	}
	for i := 0; i < len(g.adj[u]); i++ {
		if g.adj[u][i] == v {
			return true
		}
	}
	return false
}

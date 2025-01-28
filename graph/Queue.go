package Graph

type Queue []int

// Добавляет в очередь
func (q *Queue) Enqueue(x int) {
	*q = append(*q, x)
}

// Убирает из очереди
func (q *Queue) Dequeue() (int, bool) {
	if len(*q) == 0 {
		return 0, false
	}
	value := (*q)[0]
	*q = (*q)[1:]
	return value, true
}

// Обход в ширину
func BFS(g *Graph, start int) []int {
	visited := make(map[int]bool)
	order := []int{}
	queue := Queue{}
	queue.Enqueue(start)
	visited[start] = true

	for len(queue) > 0 {
		u, _ := queue.Dequeue()
		order = append(order, u)

		for _, neighbor := range g.Adj[u] {
			v := neighbor.To
			if !visited[v] {
				visited[v] = true
				queue.Enqueue(v)
			}
		}
	}
	return order
}

// Обход в глубину
func DFS(g *Graph, start int) []int {
	visited := make(map[int]bool)
	order := []int{}
	stack := NewStack()
	stack.Push(start)
	visited[start] = true

	for !stack.isEmpty() {
		u, _ := stack.Pop()
		order = append(order, u)

		for _, neighbor := range g.Adj[u] {
			v := neighbor.To
			if !visited[v] {
				visited[v] = true
				stack.Push(v)
			}
		}
	}
	return order
}

// Нахождение островов
func ConnectedComponents(g *Graph) (count int, comp map[int]int) {
	visited := make(map[int]bool)
	comp = make(map[int]int)
	count = 0

	for v := range g.Adj {
		if !visited[v] {
			count++
			componentVertices := DFSComponent(g, v, visited)
			for _, u := range componentVertices {
				comp[u] = count
			}
		}
	}
	return count, comp
}

func DFSComponent(g *Graph, start int, visited map[int]bool) []int {
	order := []int{}
	stack := NewStack()
	stack.Push(start)
	visited[start] = true

	for !stack.isEmpty() {
		u, _ := stack.Pop()
		order = append(order, u)

		for _, neighbor := range g.Adj[u] {
			v := neighbor.To
			if !visited[v] {
				visited[v] = true
				stack.Push(v)
			}
		}
	}
	return order
}

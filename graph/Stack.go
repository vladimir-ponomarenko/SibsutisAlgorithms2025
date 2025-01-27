package Graph

type Stack struct {
	data []int
}

func NewStack() *Stack {
	return &Stack{data: make([]int, 0)}
}

// Добавляет элемент в стек
func (s *Stack) Push(x int) {
	s.data = append(s.data, x)
}

// Получает последний элемент из стека и удаляет
func (s *Stack) Pop() (int, bool) {
	if len(s.data) == 0 {
		return 0, false
	}
	topIndex := len(s.data) - 1
	topValue := s.data[topIndex]
	s.data = s.data[:topIndex]
	return topValue, true
}

func (s *Stack) isEmpty() bool {
	return len(s.data) == 0
}

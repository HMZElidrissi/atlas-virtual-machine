package vm

type Stack struct {
	data []byte
}

func NewStack() *Stack {
	return &Stack{
		data: make([]byte, 0),
	}
}

func (s *Stack) Push(value byte) {
	s.data = append(s.data, value)
}

func (s *Stack) Pop() byte {
	if len(s.data) == 0 {
		return 0
	}

	value := s.data[len(s.data)-1]
	s.data = s.data[:len(s.data)-1]
	return value
}

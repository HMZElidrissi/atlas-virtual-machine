package vm

const StackSize = 256

type Stack struct {
	data   [StackSize]byte
	top    int
	memory *Memory
}

func NewStack(memory *Memory) *Stack {
	return &Stack{
		top:    -1,
		memory: memory,
	}
}

func (s *Stack) Push(value byte) {
	if s.top < StackSize-1 {
		s.top++
		s.memory.Write(uint16(s.top), value)
	} else {
		panic("Stack overflow")
	}
}

func (s *Stack) Pop() byte {
	if s.top >= 0 {
		value := s.memory.Read(uint16(s.top))
		s.top--
		return value
	}
	panic("Stack underflow")
}

package vm

// stackBase is the highest address the stack uses (top of the data segment).
// It grows downward to avoid colliding with variables at low addresses.
const stackBase = DataSegmentSize - 1 // 511

// Stack implements a LIFO stack backed by the VM's data segment,
// occupying the top portion of the data segment (growing downward).
type Stack struct {
	sp     int // stack pointer; stackBase+1 means empty
	memory *Memory
}

func NewStack(memory *Memory) *Stack {
	return &Stack{
		sp:     stackBase + 1, // empty: sp is one past the base
		memory: memory,
	}
}

// Push decrements the stack pointer then stores the value.
func (s *Stack) Push(value byte) {
	if s.sp <= 0 {
		panic("Stack overflow")
	}
	s.sp--
	s.memory.Write(uint16(s.sp), value)
}

// Pop reads the top value then increments the stack pointer.
func (s *Stack) Pop() byte {
	if s.sp > stackBase {
		panic("Stack underflow")
	}
	value := s.memory.Read(uint16(s.sp))
	s.sp++
	return value
}

// IsEmpty reports whether the stack contains no values.
func (s *Stack) IsEmpty() bool {
	return s.sp > stackBase
}

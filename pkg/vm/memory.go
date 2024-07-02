package vm

const (
	MemorySize      = 1024
	DataSegmentSize = 512
	CodeSegmentSize = 512
)

type Memory struct {
	data [MemorySize]byte
}

func NewMemory() *Memory {
	return &Memory{}
}

func (m *Memory) Read(address uint16) byte {
	if address < MemorySize {
		return m.data[address]
	}
	panic("Memory read out of bounds")
}

func (m *Memory) Write(address uint16, value byte) {
	if address < MemorySize {
		m.data[address] = value
	} else {
		panic("Memory write out of bounds")
	}
}

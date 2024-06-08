package vm

type Memory struct {
	Memory [MemorySize]byte
}

func (m *Memory) Read(address byte) byte {
	return m.Memory[address]
}

func (m *Memory) Write(address byte, value byte) {
	m.Memory[address] = value
}

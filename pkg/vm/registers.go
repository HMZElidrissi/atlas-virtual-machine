package vm

type Registers struct {
	PC  uint8 // Program Counter
	ACC int8  // Accumulator
}

func NewRegisters() *Registers {
	return &Registers{
		PC:  0,
		ACC: 0,
	}
}

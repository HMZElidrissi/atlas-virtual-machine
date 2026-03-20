package vm

type Opcode byte

const (
	ADD   Opcode = 0x00
	SUB   Opcode = 0x01
	MUL   Opcode = 0x02
	DIV   Opcode = 0x03
	AND   Opcode = 0x04
	OR    Opcode = 0x05
	XOR   Opcode = 0x06
	LOAD  Opcode = 0x07
	STORE Opcode = 0x08
	JUMP  Opcode = 0x09
	JZ    Opcode = 0x0A
	JNZ   Opcode = 0x0B
	IN    Opcode = 0x0C
	OUT   Opcode = 0x0D
	HALT  Opcode = 0x0E
)

type Instruction struct {
	Opcode  Opcode
	Operand byte
}

func DecodeInstruction(value byte) Instruction {
	return Instruction{
		Opcode:  Opcode(value >> 4),
		Operand: value & 0x0F,
	}
}

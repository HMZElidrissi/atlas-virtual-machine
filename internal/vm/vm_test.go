package vm_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/HMZElidrissi/atlas-virtual-machine/internal/vm"
)

// makeVM returns a VM wired to a byte buffer for output capture.
func makeVM(out *bytes.Buffer) *vm.VM {
	return vm.NewVM(os.Stdin, out)
}

// loadAndRun loads raw bytecode + optional initial data, then runs the VM.
func loadAndRun(t *testing.T, v *vm.VM, bytecode []byte, data map[uint8]byte) {
	t.Helper()
	if err := v.LoadProgram(bytecode); err != nil {
		t.Fatalf("LoadProgram: %v", err)
	}
	v.LoadData(data)
	v.Run()
}

// ---------------------------------------------------------------------------
// Memory
// ---------------------------------------------------------------------------

func TestMemory_ReadWrite(t *testing.T) {
	m := vm.NewMemory()
	m.Write(0, 42)
	if got := m.Read(0); got != 42 {
		t.Errorf("Read(0): want 42, got %d", got)
	}
	m.Write(511, 99)
	if got := m.Read(511); got != 99 {
		t.Errorf("Read(511): want 99, got %d", got)
	}
}

func TestMemory_OutOfBoundsRead(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on out-of-bounds read")
		}
	}()
	m := vm.NewMemory()
	m.Read(1024) // one beyond valid range
}

func TestMemory_OutOfBoundsWrite(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on out-of-bounds write")
		}
	}()
	m := vm.NewMemory()
	m.Write(1024, 0)
}

// ---------------------------------------------------------------------------
// Stack
// ---------------------------------------------------------------------------

func TestStack_PushPop(t *testing.T) {
	m := vm.NewMemory()
	s := vm.NewStack(m)

	s.Push(10)
	s.Push(20)

	if got := s.Pop(); got != 20 {
		t.Errorf("Pop: want 20, got %d", got)
	}
	if got := s.Pop(); got != 10 {
		t.Errorf("Pop: want 10, got %d", got)
	}
	if !s.IsEmpty() {
		t.Error("stack should be empty after popping all values")
	}
}

func TestStack_Underflow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on stack underflow")
		}
	}()
	m := vm.NewMemory()
	s := vm.NewStack(m)
	s.Pop()
}

// ---------------------------------------------------------------------------
// VM instructions
// ---------------------------------------------------------------------------

// encode packs opcode (upper nibble) and operand (lower nibble) into one byte.
func encode(opcode, operand byte) byte { return (opcode << 4) | (operand & 0x0F) }

const (
	opADD   = byte(0x0)
	opSUB   = byte(0x1)
	opMUL   = byte(0x2)
	opLOAD  = byte(0x7)
	opSTORE = byte(0x8)
	opJUMP  = byte(0x9)
	opJZ    = byte(0xA)
	opJNZ   = byte(0xB)
	opOUT   = byte(0xD)
	opHALT  = byte(0xE)
)

func TestVM_LoadStore(t *testing.T) {
	var out bytes.Buffer
	v := makeVM(&out)

	// data[0x00] = 77; program: LOAD 0x00, OUT, HALT
	bytecode := []byte{
		encode(opLOAD, 0x00),
		encode(opOUT, 0),
		encode(opHALT, 0),
	}
	loadAndRun(t, v, bytecode, map[uint8]byte{0x00: 77})

	if !strings.Contains(out.String(), "77") {
		t.Errorf("expected output '77', got %q", out.String())
	}
}

func TestVM_AddInstruction(t *testing.T) {
	var out bytes.Buffer
	v := makeVM(&out)

	// data[0x00]=5, data[0x01]=3
	// LOAD 0x00 → ACC=5; ADD 0x01 → ACC=8; OUT; HALT → prints 8
	bytecode := []byte{
		encode(opLOAD, 0x00),
		encode(opADD, 0x01),
		encode(opOUT, 0),
		encode(opHALT, 0),
	}
	loadAndRun(t, v, bytecode, map[uint8]byte{0x00: 5, 0x01: 3})

	if !strings.Contains(out.String(), "8") {
		t.Errorf("expected output '8', got %q", out.String())
	}
}

func TestVM_SubInstruction(t *testing.T) {
	var out bytes.Buffer
	v := makeVM(&out)

	// data[0x00]=10, data[0x01]=3 → ACC = 10 - 3 = 7
	bytecode := []byte{
		encode(opLOAD, 0x00),
		encode(opSUB, 0x01),
		encode(opOUT, 0),
		encode(opHALT, 0),
	}
	loadAndRun(t, v, bytecode, map[uint8]byte{0x00: 10, 0x01: 3})

	if !strings.Contains(out.String(), "7") {
		t.Errorf("expected output '7', got %q", out.String())
	}
}

func TestVM_JZ_TakenWhenZero(t *testing.T) {
	var out bytes.Buffer
	v := makeVM(&out)

	// data[0x00]=0
	// LOAD 0x00 → ACC=0
	// JZ  3     → PC=3 (branch taken, skip instruction 2)
	// OUT       ← skipped (instruction 2)
	// HALT      (instruction 3)
	bytecode := []byte{
		encode(opLOAD, 0x00), // 0
		encode(opJZ, 0x03),   // 1 → jump to 3
		encode(opOUT, 0),     // 2 (skipped)
		encode(opHALT, 0),    // 3
	}
	loadAndRun(t, v, bytecode, map[uint8]byte{0x00: 0})

	if strings.Contains(out.String(), "0") {
		t.Error("expected OUT to be skipped when ACC==0 and JZ taken, but output appeared")
	}
}

func TestVM_JNZ_NotTakenWhenZero(t *testing.T) {
	var out bytes.Buffer
	v := makeVM(&out)

	// data[0x00]=0; JNZ should NOT branch
	// LOAD 0x00 → ACC=0; JNZ 3 → not taken; OUT (prints 0); HALT
	bytecode := []byte{
		encode(opLOAD, 0x00), // 0
		encode(opJNZ, 0x03),  // 1 → branch NOT taken (ACC==0)
		encode(opOUT, 0),     // 2 ← executed
		encode(opHALT, 0),    // 3
	}
	loadAndRun(t, v, bytecode, map[uint8]byte{0x00: 0})

	if !strings.Contains(out.String(), "0") {
		t.Errorf("expected OUT to execute when JNZ not taken, got %q", out.String())
	}
}

func TestVM_LoadProgram_PCReset(t *testing.T) {
	v := vm.NewVM(os.Stdin, os.Stdout)
	bytecode := []byte{encode(opHALT, 0)}
	if err := v.LoadProgram(bytecode); err != nil {
		t.Fatalf("LoadProgram: %v", err)
	}
	// PC must be 0 (relative to code segment start) after loading.
	if v.Registers.PC != 0 {
		t.Errorf("expected PC=0 after LoadProgram, got %d", v.Registers.PC)
	}
}

func TestVM_DemoProgram_Even(t *testing.T) {
	// Full end-to-end: check that 10 is identified as even (output = 0).
	var out bytes.Buffer
	v := makeVM(&out)

	// Bytecode generated from the compiler for:
	//   var number: int; number = 10;
	//   if ((number & 1) == 0) { return (0); } else { return (1); }
	//
	// Memory layout: 0x00=number, 0x08=10, 0x09=0, 0x0A=1
	bytecode := []byte{
		0x78, // LOAD  0x08  → ACC = 10
		0x80, // STORE 0x00  → number = 10
		0x70, // LOAD  0x00  → ACC = number (10)
		0x4A, // AND   0x0A  → ACC = 10 & 1 = 0
		0x19, // SUB   0x09  → ACC = 0 - 0 = 0
		0xB8, // JNZ   0x08  → ACC==0 → not taken
		0x7A, // LOAD  0x0A  → ACC = 1 (equal → true)
		0x99, // JUMP  0x09  → PC = 9
		0x79, // LOAD  0x09  → ACC = 0 (not-equal)
		0xAE, // JZ    0x0E  → ACC==1 → not taken (condition true)
		0x79, // LOAD  0x09  → return value 0
		0xD0, // OUT
		0xE0, // HALT
	}
	data := map[uint8]byte{0x08: 10, 0x09: 0, 0x0A: 1}
	loadAndRun(t, v, bytecode, data)

	got := strings.TrimSpace(out.String())
	if got != "0" {
		t.Errorf("expected output '0' (even), got %q", got)
	}
}

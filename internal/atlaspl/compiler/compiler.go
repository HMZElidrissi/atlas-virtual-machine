package compiler

import (
	"fmt"

	"github.com/HMZElidrissi/atlas-virtual-machine/internal/atlaspl/ast"
)

// ---------------------------------------------------------------------------
// Data-segment memory layout (4-bit operand → addresses 0x00–0x0F only)
// ---------------------------------------------------------------------------
//   0x00 – 0x05  User variables   (up to 6)
//   0x06          tempReg1         scratch register 1
//   0x07          tempReg2         scratch register 2
//   0x08 – 0x0F  Constant pool    (up to 8 distinct integer literals)
// ---------------------------------------------------------------------------

const (
	maxUserVars   = 6
	varAreaBase   = byte(0x00)
	tempReg1      = byte(0x06)
	tempReg2      = byte(0x07)
	constAreaBase = byte(0x08)
	maxConsts     = 8
)

// Opcode bytes — upper nibble pre-shifted, OR'd with a 4-bit operand to
// produce one complete instruction byte.
const (
	bADD   = byte(0x00)
	bSUB   = byte(0x10)
	bMUL   = byte(0x20)
	bDIV   = byte(0x30)
	bAND   = byte(0x40)
	bOR    = byte(0x50)
	bXOR   = byte(0x60)
	bLOAD  = byte(0x70)
	bSTORE = byte(0x80)
	bJUMP  = byte(0x90)
	bJZ    = byte(0xA0)
	bJNZ   = byte(0xB0)
	bIN    = byte(0xC0)
	bOUT   = byte(0xD0)
	bHALT  = byte(0xE0)
)

// CompiledProgram is the output of a successful compilation.
type CompiledProgram struct {
	// InitialData maps data-segment addresses to their initial values.
	// The VM must call LoadData with this map before Run() to seed the
	// constant pool and any pre-initialised variables.
	InitialData map[uint8]byte

	// Bytecode is the raw code-segment bytes loaded by vm.LoadProgram.
	Bytecode []byte
}

// Compiler walks an AtlasPL AST and emits AtlasVM bytecode.
type Compiler struct {
	varTable    map[string]byte // variable name → data-segment address
	constTable  map[int64]byte  // constant value → data-segment address
	initialData map[uint8]byte  // initial memory values passed to vm.LoadData
	code        []byte          // emitted bytecode (code-segment bytes)
	nextVar     byte            // next free variable address
	nextConst   byte            // next free constant-pool address
}

// NewCompiler returns a ready-to-use Compiler.
func NewCompiler() *Compiler {
	return &Compiler{
		varTable:    make(map[string]byte),
		constTable:  make(map[int64]byte),
		initialData: make(map[uint8]byte),
		nextVar:     varAreaBase,
		nextConst:   constAreaBase,
	}
}

// Compile translates program into bytecode and initial data.
func (c *Compiler) Compile(program *ast.Program) (*CompiledProgram, error) {
	for _, stmt := range program.Statements {
		if err := c.compileStatement(stmt); err != nil {
			return nil, err
		}
	}
	c.emit(bHALT, 0) // guarantee termination

	return &CompiledProgram{
		InitialData: c.initialData,
		Bytecode:    c.code,
	}, nil
}

// ---------------------------------------------------------------------------
// Emission helpers
// ---------------------------------------------------------------------------

func (c *Compiler) emit(opcode, operand byte) {
	c.code = append(c.code, opcode|operand)
}

// currentPC returns the index of the next instruction to be emitted —
// i.e. the code-segment offset the runtime PC will hold at that point.
func (c *Compiler) currentPC() byte { return byte(len(c.code)) }

// emitJump appends a jump instruction with a placeholder operand (0) and
// returns the index so it can be back-patched via patch().
func (c *Compiler) emitJump(opcode byte) int {
	idx := len(c.code)
	c.code = append(c.code, opcode|0x00)
	return idx
}

// patch writes the correct 4-bit target into a previously emitted jump.
func (c *Compiler) patch(idx int, target byte) {
	c.code[idx] = (c.code[idx] & 0xF0) | (target & 0x0F)
}

// ---------------------------------------------------------------------------
// Address allocation
// ---------------------------------------------------------------------------

// allocVar reserves a data-segment address for name, reusing it on
// re-declaration to support simple variable shadowing.
func (c *Compiler) allocVar(name string) (byte, error) {
	if addr, ok := c.varTable[name]; ok {
		return addr, nil
	}
	if c.nextVar >= constAreaBase {
		return 0, fmt.Errorf("too many variables: maximum is %d", maxUserVars)
	}
	addr := c.nextVar
	c.varTable[name] = addr
	c.nextVar++
	return addr, nil
}

// allocConst returns the constant-pool address for val, allocating a new
// slot and seeding InitialData if this value has not been seen before.
func (c *Compiler) allocConst(val int64) (byte, error) {
	if addr, ok := c.constTable[val]; ok {
		return addr, nil
	}
	if c.nextConst > constAreaBase+maxConsts-1 {
		return 0, fmt.Errorf("constant pool full (max %d distinct constants)", maxConsts)
	}
	addr := c.nextConst
	c.constTable[val] = addr
	c.initialData[addr] = byte(val)
	c.nextConst++
	return addr, nil
}

package compiler_test

import (
	"strings"
	"testing"

	"github.com/HMZElidrissi/atlas-virtual-machine/internal/atlaspl/compiler"
	"github.com/HMZElidrissi/atlas-virtual-machine/internal/atlaspl/lexer"
	"github.com/HMZElidrissi/atlas-virtual-machine/internal/atlaspl/parser"
)

func compile(t *testing.T, src string) *compiler.CompiledProgram {
	t.Helper()
	l := lexer.NewLexer(strings.NewReader(src))
	p := parser.NewParser(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	c := compiler.NewCompiler()
	out, err := c.Compile(prog)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}
	return out
}

func TestCompile_DemoProgram(t *testing.T) {
	src := `
	var number: int;
	number = 10;
	if ((number & 1) == 0) {
	  return (0);
	} else {
	  return (1);
	}`
	out := compile(t, src)

	if len(out.Bytecode) == 0 {
		t.Fatal("expected non-empty bytecode")
	}
	// Constant pool must contain 10, 0, 1
	found10, found0, found1 := false, false, false
	for _, v := range out.InitialData {
		switch v {
		case 10:
			found10 = true
		case 0:
			found0 = true
		case 1:
			found1 = true
		}
	}
	if !found10 {
		t.Error("constant pool missing value 10")
	}
	if !found0 {
		t.Error("constant pool missing value 0")
	}
	if !found1 {
		t.Error("constant pool missing value 1")
	}
}

func TestCompile_ConstantDeduplication(t *testing.T) {
	// The same literal '1' appears twice; it should occupy only one pool slot.
	src := `
	var a: int;
	var b: int;
	a = 1;
	b = 1;`
	out := compile(t, src)

	count := 0
	for _, v := range out.InitialData {
		if v == 1 {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 1 pool slot for literal 1, got %d", count)
	}
}

func TestCompile_AlwaysEndsWithHalt(t *testing.T) {
	out := compile(t, "var x: int;")
	last := out.Bytecode[len(out.Bytecode)-1]
	// HALT opcode is 0xE, occupying the upper nibble → byte is 0xE0.
	if last != 0xE0 {
		t.Errorf("expected final byte 0xE0 (HALT), got 0x%02X", last)
	}
}

func TestCompile_SimpleAssignment(t *testing.T) {
	// var x: int; x = 5;
	// Expected: LOAD <const5>, STORE <x_addr>, HALT
	out := compile(t, "var x: int; x = 5;")

	if len(out.Bytecode) < 3 {
		t.Fatalf("expected at least 3 bytes, got %d", len(out.Bytecode))
	}

	// First instruction must be LOAD (opcode nibble 0x7)
	if out.Bytecode[0]>>4 != 0x7 {
		t.Errorf("expected LOAD (0x7x) as first instruction, got 0x%02X", out.Bytecode[0])
	}
	// Second instruction must be STORE (opcode nibble 0x8)
	if out.Bytecode[1]>>4 != 0x8 {
		t.Errorf("expected STORE (0x8x) as second instruction, got 0x%02X", out.Bytecode[1])
	}
}

func TestCompile_TooManyVariables(t *testing.T) {
	// maxUserVars=6 (addresses 0x00–0x05). Nine variables must fail.
	src := `var a: int;
var b: int;
var c: int;
var d: int;
var e: int;
var f: int;
var g: int;
var h: int;
var i: int;`
	l := lexer.NewLexer(strings.NewReader(src))
	p := parser.NewParser(l)
	prog := p.ParseProgram()
	c := compiler.NewCompiler()
	_, err := c.Compile(prog)
	if err == nil {
		t.Error("expected compilation error for too many variables, got nil")
	}
}

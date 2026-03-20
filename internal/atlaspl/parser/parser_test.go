package parser_test

import (
	"strings"
	"testing"

	"github.com/HMZElidrissi/atlas-virtual-machine/internal/atlaspl/ast"
	"github.com/HMZElidrissi/atlas-virtual-machine/internal/atlaspl/lexer"
	"github.com/HMZElidrissi/atlas-virtual-machine/internal/atlaspl/parser"
)

func parse(t *testing.T, src string) *ast.Program {
	t.Helper()
	l := lexer.NewLexer(strings.NewReader(src))
	p := parser.NewParser(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) != 0 {
		for _, e := range errs {
			t.Errorf("parse error: %s", e)
		}
		t.FailNow()
	}
	return prog
}

func TestParseVarStatement(t *testing.T) {
	prog := parse(t, "var x: int;")
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	vs, ok := prog.Statements[0].(*ast.VarStatement)
	if !ok {
		t.Fatalf("expected *ast.VarStatement, got %T", prog.Statements[0])
	}
	if vs.Name.Value != "x" {
		t.Errorf("expected name 'x', got %q", vs.Name.Value)
	}
	if vs.Type != "int" {
		t.Errorf("expected type 'int', got %q", vs.Type)
	}
}

func TestParseAssignmentStatement(t *testing.T) {
	prog := parse(t, "x = 42;")
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	as, ok := prog.Statements[0].(*ast.AssignmentStatement)
	if !ok {
		t.Fatalf("expected *ast.AssignmentStatement, got %T", prog.Statements[0])
	}
	if as.Name.Value != "x" {
		t.Errorf("expected name 'x', got %q", as.Name.Value)
	}
	lit, ok := as.Value.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expected *ast.IntegerLiteral, got %T", as.Value)
	}
	if lit.Value != 42 {
		t.Errorf("expected 42, got %d", lit.Value)
	}
}

func TestParseIfElseStatement(t *testing.T) {
	src := `if (x == 0) { return (1); } else { return (0); }`
	prog := parse(t, src)
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	is, ok := prog.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("expected *ast.IfStatement, got %T", prog.Statements[0])
	}
	if is.Consequence == nil {
		t.Error("expected non-nil consequence")
	}
	if is.Alternative == nil {
		t.Error("expected non-nil alternative")
	}
	// Condition should be an infix == expression
	cond, ok := is.Condition.(*ast.InfixExpression)
	if !ok {
		t.Fatalf("expected *ast.InfixExpression condition, got %T", is.Condition)
	}
	if cond.Operator != "==" {
		t.Errorf("expected operator ==, got %q", cond.Operator)
	}
}

func TestParseReturnStatement(t *testing.T) {
	prog := parse(t, "return (7);")
	rs, ok := prog.Statements[0].(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("expected *ast.ReturnStatement, got %T", prog.Statements[0])
	}
	lit, ok := rs.ReturnValue.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expected *ast.IntegerLiteral, got %T", rs.ReturnValue)
	}
	if lit.Value != 7 {
		t.Errorf("expected 7, got %d", lit.Value)
	}
}

func TestParseInfixPrecedence(t *testing.T) {
	// 2 + 3 * 4 should parse as 2 + (3 * 4)
	prog := parse(t, "2 + 3 * 4;")
	es := prog.Statements[0].(*ast.ExpressionStatement)
	outer, ok := es.Expression.(*ast.InfixExpression)
	if !ok {
		t.Fatalf("expected *ast.InfixExpression, got %T", es.Expression)
	}
	if outer.Operator != "+" {
		t.Errorf("expected outer operator '+', got %q", outer.Operator)
	}
	rhs, ok := outer.Right.(*ast.InfixExpression)
	if !ok {
		t.Fatalf("expected right to be *ast.InfixExpression, got %T", outer.Right)
	}
	if rhs.Operator != "*" {
		t.Errorf("expected inner operator '*', got %q", rhs.Operator)
	}
}

func TestParseCommentSkipped(t *testing.T) {
	prog := parse(t, "@ this entire line is a comment\nvar x: int;")
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement (comment skipped), got %d", len(prog.Statements))
	}
}

func TestParseDemoProgram(t *testing.T) {
	src := `
	var number: int;
	number = 10;
	if ((number & 1) == 0) {
	  return (0);
	} else {
	  return (1);
	}`
	prog := parse(t, src)
	if len(prog.Statements) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(prog.Statements))
	}
}

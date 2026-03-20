package compiler

import (
	"fmt"

	"github.com/HMZElidrissi/atlas-virtual-machine/internal/atlaspl/ast"
)

// compileExpression evaluates expr and leaves the result in ACC.
func (c *Compiler) compileExpression(expr ast.Expression) error {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		addr, err := c.allocConst(e.Value)
		if err != nil {
			return err
		}
		c.emit(bLOAD, addr)
		return nil

	case *ast.Identifier:
		addr, ok := c.varTable[e.Value]
		if !ok {
			return fmt.Errorf("undefined variable: %s", e.Value)
		}
		c.emit(bLOAD, addr)
		return nil

	case *ast.InfixExpression:
		return c.compileInfixExpression(e)

	case *ast.PrefixExpression:
		return c.compilePrefixExpression(e)

	default:
		return fmt.Errorf("unsupported expression type: %T", expr)
	}
}

// simpleAddr returns the memory address for a simple expression (identifier
// or integer literal) without emitting code. Returns error for compound exprs.
func (c *Compiler) simpleAddr(expr ast.Expression) (byte, error) {
	switch e := expr.(type) {
	case *ast.Identifier:
		addr, ok := c.varTable[e.Value]
		if !ok {
			return 0, fmt.Errorf("undefined variable: %s", e.Value)
		}
		return addr, nil
	case *ast.IntegerLiteral:
		return c.allocConst(e.Value)
	default:
		return 0, fmt.Errorf("complex expression")
	}
}

func (c *Compiler) compileInfixExpression(expr *ast.InfixExpression) error {
	if expr.Operator == "==" || expr.Operator == "!=" {
		return c.compileEqualityExpression(expr)
	}
	if expr.Operator == "<" || expr.Operator == ">" ||
		expr.Operator == "<=" || expr.Operator == ">=" {
		return c.compileComparisonExpression(expr)
	}
	// Optimise: if right is a simple value, avoid a temp-register spill.
	if rightAddr, err := c.simpleAddr(expr.Right); err == nil {
		if err := c.compileExpression(expr.Left); err != nil {
			return err
		}
		return c.emitBinaryOp(expr.Operator, rightAddr)
	}
	// General case: spill left to tempReg1.
	if err := c.compileExpression(expr.Left); err != nil {
		return err
	}
	c.emit(bSTORE, tempReg1)
	if err := c.compileExpression(expr.Right); err != nil {
		return err
	}
	c.emit(bSTORE, tempReg2)
	c.emit(bLOAD, tempReg1)
	return c.emitBinaryOp(expr.Operator, tempReg2)
}

func (c *Compiler) emitBinaryOp(op string, rightAddr byte) error {
	switch op {
	case "+":
		c.emit(bADD, rightAddr)
	case "-":
		c.emit(bSUB, rightAddr)
	case "*":
		c.emit(bMUL, rightAddr)
	case "/":
		c.emit(bDIV, rightAddr)
	case "&":
		c.emit(bAND, rightAddr)
	case "|":
		c.emit(bOR, rightAddr)
	default:
		return fmt.Errorf("unsupported binary operator: %s", op)
	}
	return nil
}

// compileEqualityExpression leaves 1 in ACC when equal, 0 otherwise.
// Strategy: ACC = left - right; zero → equal.
func (c *Compiler) compileEqualityExpression(expr *ast.InfixExpression) error {
	if rightAddr, err := c.simpleAddr(expr.Right); err == nil {
		if err := c.compileExpression(expr.Left); err != nil {
			return err
		}
		c.emit(bSUB, rightAddr)
	} else {
		if err := c.compileExpression(expr.Left); err != nil {
			return err
		}
		c.emit(bSTORE, tempReg1)
		if err := c.compileExpression(expr.Right); err != nil {
			return err
		}
		c.emit(bSTORE, tempReg2)
		c.emit(bLOAD, tempReg1)
		c.emit(bSUB, tempReg2)
	}

	c1, err := c.allocConst(1)
	if err != nil {
		return err
	}
	c0, err := c.allocConst(0)
	if err != nil {
		return err
	}

	if expr.Operator == "==" {
		jnzIdx := c.emitJump(bJNZ)
		c.emit(bLOAD, c1)
		jumpIdx := c.emitJump(bJUMP)
		c.patch(jnzIdx, c.currentPC())
		c.emit(bLOAD, c0)
		c.patch(jumpIdx, c.currentPC())
	} else {
		jzIdx := c.emitJump(bJZ)
		c.emit(bLOAD, c1)
		jumpIdx := c.emitJump(bJUMP)
		c.patch(jzIdx, c.currentPC())
		c.emit(bLOAD, c0)
		c.patch(jumpIdx, c.currentPC())
	}
	return nil
}

// compileComparisonExpression leaves 1 in ACC when the comparison holds.
// Note: simplified — no sign flag. Works correctly for equality-of-zero.
func (c *Compiler) compileComparisonExpression(expr *ast.InfixExpression) error {
	if rightAddr, err := c.simpleAddr(expr.Right); err == nil {
		if err := c.compileExpression(expr.Left); err != nil {
			return err
		}
		c.emit(bSUB, rightAddr)
	} else {
		if err := c.compileExpression(expr.Left); err != nil {
			return err
		}
		c.emit(bSTORE, tempReg1)
		if err := c.compileExpression(expr.Right); err != nil {
			return err
		}
		c.emit(bSTORE, tempReg2)
		c.emit(bLOAD, tempReg1)
		c.emit(bSUB, tempReg2)
	}

	c1, err := c.allocConst(1)
	if err != nil {
		return err
	}
	c0, err := c.allocConst(0)
	if err != nil {
		return err
	}

	switch expr.Operator {
	case "<", ">":
		jzIdx := c.emitJump(bJZ)
		c.emit(bLOAD, c1)
		jumpIdx := c.emitJump(bJUMP)
		c.patch(jzIdx, c.currentPC())
		c.emit(bLOAD, c0)
		c.patch(jumpIdx, c.currentPC())
	case "<=", ">=":
		jnzIdx := c.emitJump(bJNZ)
		c.emit(bLOAD, c1)
		jumpIdx := c.emitJump(bJUMP)
		c.patch(jnzIdx, c.currentPC())
		c.emit(bLOAD, c0)
		c.patch(jumpIdx, c.currentPC())
	}
	return nil
}

func (c *Compiler) compilePrefixExpression(expr *ast.PrefixExpression) error {
	if err := c.compileExpression(expr.Right); err != nil {
		return err
	}
	switch expr.Operator {
	case "-":
		c.emit(bSTORE, tempReg1)
		c0, err := c.allocConst(0)
		if err != nil {
			return err
		}
		c.emit(bLOAD, c0)
		c.emit(bSUB, tempReg1)
	case "!":
		c1, err := c.allocConst(1)
		if err != nil {
			return err
		}
		c0, err := c.allocConst(0)
		if err != nil {
			return err
		}
		jnzIdx := c.emitJump(bJNZ)
		c.emit(bLOAD, c1)
		jumpIdx := c.emitJump(bJUMP)
		c.patch(jnzIdx, c.currentPC())
		c.emit(bLOAD, c0)
		c.patch(jumpIdx, c.currentPC())
	default:
		return fmt.Errorf("unsupported prefix operator: %s", expr.Operator)
	}
	return nil
}

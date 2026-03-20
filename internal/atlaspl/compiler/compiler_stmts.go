package compiler

import (
	"fmt"

	"github.com/HMZElidrissi/atlas-virtual-machine/internal/atlaspl/ast"
)

func (c *Compiler) compileStatement(stmt ast.Statement) error {
	switch s := stmt.(type) {
	case *ast.VarStatement:
		return c.compileVarStatement(s)
	case *ast.AssignmentStatement:
		return c.compileAssignmentStatement(s)
	case *ast.IfStatement:
		return c.compileIfStatement(s)
	case *ast.ReturnStatement:
		return c.compileReturnStatement(s)
	case *ast.ExpressionStatement:
		return c.compileExpressionStatement(s)
	default:
		return fmt.Errorf("unsupported statement type: %T", stmt)
	}
}

func (c *Compiler) compileVarStatement(stmt *ast.VarStatement) error {
	_, err := c.allocVar(stmt.Name.Value)
	return err
}

func (c *Compiler) compileAssignmentStatement(stmt *ast.AssignmentStatement) error {
	if err := c.compileExpression(stmt.Value); err != nil {
		return err
	}
	addr, ok := c.varTable[stmt.Name.Value]
	if !ok {
		return fmt.Errorf("undefined variable: %s", stmt.Name.Value)
	}
	c.emit(bSTORE, addr)
	return nil
}

// compileIfStatement emits:
//
//	<condition>
//	JZ  [else_or_end]
//	<consequence>
//	JUMP [end]         ← only when alternative exists
//	[else]:
//	<alternative>
//	[end]:
func (c *Compiler) compileIfStatement(stmt *ast.IfStatement) error {
	if err := c.compileExpression(stmt.Condition); err != nil {
		return err
	}
	jzIdx := c.emitJump(bJZ)

	if err := c.compileBlockStatement(stmt.Consequence); err != nil {
		return err
	}

	if stmt.Alternative != nil {
		jumpIdx := c.emitJump(bJUMP)
		c.patch(jzIdx, c.currentPC())
		if err := c.compileBlockStatement(stmt.Alternative); err != nil {
			return err
		}
		c.patch(jumpIdx, c.currentPC())
	} else {
		c.patch(jzIdx, c.currentPC())
	}
	return nil
}

func (c *Compiler) compileReturnStatement(stmt *ast.ReturnStatement) error {
	if err := c.compileExpression(stmt.ReturnValue); err != nil {
		return err
	}
	c.emit(bOUT, 0)
	c.emit(bHALT, 0)
	return nil
}

func (c *Compiler) compileExpressionStatement(stmt *ast.ExpressionStatement) error {
	return c.compileExpression(stmt.Expression)
}

func (c *Compiler) compileBlockStatement(block *ast.BlockStatement) error {
	for _, stmt := range block.Statements {
		if err := c.compileStatement(stmt); err != nil {
			return err
		}
	}
	return nil
}

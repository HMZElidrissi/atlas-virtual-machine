package atlaspl

import (
	"fmt"
)

type Interpreter struct {
	environment map[string]interface{}
}

func NewInterpreter() *Interpreter {
	return &Interpreter{
		environment: make(map[string]interface{}),
	}
}

func (i *Interpreter) Interpret(program *Program) error {
	for _, statement := range program.Statements {
		if err := i.executeStatement(statement); err != nil {
			return err
		}
	}
	return nil
}

func (i *Interpreter) executeStatement(statement Statement) error {
	switch stmt := statement.(type) {
	case *VarStatement:
		return i.executeVarStatement(stmt)
	case *ReturnStatement:
		return i.executeReturnStatement(stmt)
	case *ExpressionStatement:
		return i.executeExpressionStatement(stmt)
	case *IfStatement:
		return i.executeIfStatement(stmt)
	case *AssignmentStatement:
		return i.executeAssignmentStatement(stmt)
	default:
		return fmt.Errorf("unknown statement type: %T", statement)
	}
}

func (i *Interpreter) executeVarStatement(stmt *VarStatement) error {
	var value interface{}
	var err error

	if stmt.Value != nil {
		value, err = i.evaluateExpression(stmt.Value)
		if err != nil {
			return err
		}
	}

	i.environment[stmt.Name.Value] = value
	return nil
}

func (i *Interpreter) executeReturnStatement(stmt *ReturnStatement) error {
	value, err := i.evaluateExpression(stmt.ReturnValue)
	if err != nil {
		return err
	}
	fmt.Printf("Return value: %v\n", value)
	return nil
}

func (i *Interpreter) executeExpressionStatement(stmt *ExpressionStatement) error {
	_, err := i.evaluateExpression(stmt.Expression)
	return err
}

func (i *Interpreter) executeIfStatement(stmt *IfStatement) error {
	condition, err := i.evaluateExpression(stmt.Condition)
	if err != nil {
		return err
	}

	if condition.(bool) {
		return i.executeBlockStatement(stmt.Consequence)
	} else if stmt.Alternative != nil {
		return i.executeBlockStatement(stmt.Alternative)
	}

	return nil
}

func (i *Interpreter) executeAssignmentStatement(stmt *AssignmentStatement) error {
	value, err := i.evaluateExpression(stmt.Value)
	if err != nil {
		return err
	}
	i.environment[stmt.Name.Value] = value
	return nil
}

func (i *Interpreter) executeBlockStatement(block *BlockStatement) error {
	for _, statement := range block.Statements {
		if err := i.executeStatement(statement); err != nil {
			return err
		}
	}
	return nil
}

func (i *Interpreter) evaluateExpression(expr Expression) (interface{}, error) {
	switch e := expr.(type) {
	case *Identifier:
		return i.evaluateIdentifier(e)
	case *IntegerLiteral:
		return e.Value, nil
	case *BooleanLiteral:
		return e.Value, nil
	case *PrefixExpression:
		return i.evaluatePrefixExpression(e)
	case *InfixExpression:
		return i.evaluateInfixExpression(e)
	default:
		return nil, fmt.Errorf("unknown expression type: %T", expr)
	}
}

func (i *Interpreter) evaluateIdentifier(ident *Identifier) (interface{}, error) {
	if val, ok := i.environment[ident.Value]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("identifier not found: %s", ident.Value)
}

func (i *Interpreter) evaluatePrefixExpression(expr *PrefixExpression) (interface{}, error) {
	right, err := i.evaluateExpression(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator {
	case "!":
		return !right.(bool), nil
	case "-":
		return -right.(int64), nil
	default:
		return nil, fmt.Errorf("unknown prefix operator: %s", expr.Operator)
	}
}

func (i *Interpreter) evaluateInfixExpression(expr *InfixExpression) (interface{}, error) {
	left, err := i.evaluateExpression(expr.Left)
	if err != nil {
		return nil, err
	}

	right, err := i.evaluateExpression(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator {
	case "+":
		return left.(int64) + right.(int64), nil
	case "-":
		return left.(int64) - right.(int64), nil
	case "*":
		return left.(int64) * right.(int64), nil
	case "/":
		return left.(int64) / right.(int64), nil
	case "==":
		return left == right, nil
	case "!=":
		return left != right, nil
	case "<":
		return left.(int64) < right.(int64), nil
	case ">":
		return left.(int64) > right.(int64), nil
	case "<=":
		return left.(int64) <= right.(int64), nil
	case ">=":
		return left.(int64) >= right.(int64), nil
	case "&":
		return left.(int64) & right.(int64), nil
	case "|":
		return left.(int64) | right.(int64), nil
	default:
		return nil, fmt.Errorf("unknown infix operator: %s", expr.Operator)
	}
}

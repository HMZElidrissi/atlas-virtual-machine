package ast

import "github.com/HMZElidrissi/atlas-virtual-machine/internal/atlaspl/lexer"

// Node is the base interface for every element of the AST.
type Node interface {
	TokenLiteral() string
}

// Statement represents an executable statement.
type Statement interface {
	Node
	statementNode()
}

// Expression represents a value-producing expression.
type Expression interface {
	Node
	expressionNode()
}

// Program is the root node — the entire parsed source file.
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

// ---------------------------------------------------------------------------
// Statements
// ---------------------------------------------------------------------------

type VarStatement struct {
	Token lexer.Token // the 'var' token
	Name  *Identifier
	Type  string
	Value Expression
}

func (vs *VarStatement) statementNode()       {}
func (vs *VarStatement) TokenLiteral() string { return vs.Token.Literal }

type ReturnStatement struct {
	Token       lexer.Token // the 'return' token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }

type ExpressionStatement struct {
	Token      lexer.Token // the first token of the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }

type IfStatement struct {
	Token       lexer.Token // the 'if' token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (is *IfStatement) statementNode()       {}
func (is *IfStatement) TokenLiteral() string { return is.Token.Literal }

type BlockStatement struct {
	Token      lexer.Token // the '{' token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }

type AssignmentStatement struct {
	Token lexer.Token // the '=' token
	Name  *Identifier
	Value Expression
}

func (as *AssignmentStatement) statementNode()       {}
func (as *AssignmentStatement) TokenLiteral() string { return as.Token.Literal }

// ---------------------------------------------------------------------------
// Expressions
// ---------------------------------------------------------------------------

type Identifier struct {
	Token lexer.Token // the IDENT token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

type IntegerLiteral struct {
	Token lexer.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }

type BooleanLiteral struct {
	Token lexer.Token
	Value bool
}

func (bl *BooleanLiteral) expressionNode()      {}
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Literal }

type PrefixExpression struct {
	Token    lexer.Token // the prefix token, e.g. !
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }

type InfixExpression struct {
	Token    lexer.Token // the operator token, e.g. +
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }

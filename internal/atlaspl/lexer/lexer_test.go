package lexer_test

import (
	"strings"
	"testing"

	"github.com/HMZElidrissi/atlas-virtual-machine/internal/atlaspl/lexer"
)

func TestNextToken_Keywords(t *testing.T) {
	input := "var if else return true false"
	want := []lexer.Token{
		{Type: lexer.VAR, Literal: "var"},
		{Type: lexer.IF, Literal: "if"},
		{Type: lexer.ELSE, Literal: "else"},
		{Type: lexer.RETURN, Literal: "return"},
		{Type: lexer.TRUE, Literal: "true"},
		{Type: lexer.FALSE, Literal: "false"},
		{Type: lexer.EOF},
	}
	assertTokens(t, input, want)
}

func TestNextToken_Operators(t *testing.T) {
	input := "= == != < <= > >= & | + - * /"
	want := []lexer.Token{
		{Type: lexer.EQUAL, Literal: "="},
		{Type: lexer.EQ, Literal: "=="},
		{Type: lexer.NEQ, Literal: "!="},
		{Type: lexer.LT, Literal: "<"},
		{Type: lexer.LTE, Literal: "<="},
		{Type: lexer.GT, Literal: ">"},
		{Type: lexer.GTE, Literal: ">="},
		{Type: lexer.AND, Literal: "&"},
		{Type: lexer.OR, Literal: "|"},
		{Type: lexer.PLUS, Literal: "+"},
		{Type: lexer.MINUS, Literal: "-"},
		{Type: lexer.ASTERISK, Literal: "*"},
		{Type: lexer.SLASH, Literal: "/"},
		{Type: lexer.EOF},
	}
	assertTokens(t, input, want)
}

func TestNextToken_Comment(t *testing.T) {
	// Comments start with @ and run to end of line — they should be emitted
	// as a COMMENT token (the parser skips them, but the lexer produces them).
	input := "var x: int; @ this is a comment\nx = 1;"
	l := lexer.NewLexer(strings.NewReader(input))

	tokens := collectTokens(l)
	for _, tok := range tokens {
		if tok.Type == lexer.ILLEGAL {
			t.Errorf("unexpected ILLEGAL token: %q", tok.Literal)
		}
	}
	// Verify the comment token is present
	found := false
	for _, tok := range tokens {
		if tok.Type == lexer.COMMENT {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected a COMMENT token, got none")
	}
}

func TestNextToken_FullProgram(t *testing.T) {
	input := `var number: int;
number = 10;
if ((number & 1) == 0) {
  return (0);
} else {
  return (1);
}`
	l := lexer.NewLexer(strings.NewReader(input))
	for {
		tok := l.NextToken()
		if tok.Type == lexer.ILLEGAL {
			t.Errorf("unexpected ILLEGAL token: %q", tok.Literal)
		}
		if tok.Type == lexer.EOF {
			break
		}
	}
}

func TestBangNotEqual(t *testing.T) {
	// '!' alone → BANG; '!=' → NEQ
	input := "! !="
	want := []lexer.Token{
		{Type: lexer.BANG, Literal: "!"},
		{Type: lexer.NEQ, Literal: "!="},
		{Type: lexer.EOF},
	}
	assertTokens(t, input, want)
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func assertTokens(t *testing.T, input string, want []lexer.Token) {
	t.Helper()
	l := lexer.NewLexer(strings.NewReader(input))
	for i, exp := range want {
		got := l.NextToken()
		if got.Type != exp.Type {
			t.Errorf("token[%d]: want type %s, got %s (literal=%q)", i, exp.Type, got.Type, got.Literal)
		}
		if exp.Literal != "" && got.Literal != exp.Literal {
			t.Errorf("token[%d]: want literal %q, got %q", i, exp.Literal, got.Literal)
		}
	}
}

func collectTokens(l *lexer.Lexer) []lexer.Token {
	var toks []lexer.Token
	for {
		tok := l.NextToken()
		toks = append(toks, tok)
		if tok.Type == lexer.EOF {
			break
		}
	}
	return toks
}

package atlaspl

import (
	"bufio"
	"io"
	"unicode"
)

type TokenType int

const (
	EOF TokenType = iota
	ILLEGAL
	IDENT
	INT
	BOOL
	VAR
	IF
	ELSE
	RETURN
	SEMICOLON
	EQUAL
	LPAREN
	RPAREN
	LBRACE
	RBRACE
	COLON
	PLUS
	MINUS
	ASTERISK
	SLASH
	AND
	OR
	NOT
	EQ
	NEQ
	LT
	GT
	LTE
	GTE
)

type Token struct {
	Type    TokenType
	Literal string
}

type Lexer struct {
	reader *bufio.Reader
}

func NewLexer(r io.Reader) *Lexer {
	return &Lexer{reader: bufio.NewReader(r)}
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	ch := l.readChar()
	switch {
	case ch == 0:
		return Token{Type: EOF}
	case isLetter(ch):
		return l.readIdentifier(ch)
	case isDigit(ch):
		return l.readNumber(ch)
	case ch == ';':
		return Token{Type: SEMICOLON, Literal: ";"}
	case ch == '=':
		if l.peekChar() == '=' {
			l.readChar()
			return Token{Type: EQ, Literal: "=="}
		}
		return Token{Type: EQUAL, Literal: "="}
	case ch == '(':
		return Token{Type: LPAREN, Literal: "("}
	case ch == ')':
		return Token{Type: RPAREN, Literal: ")"}
	case ch == '{':
		return Token{Type: LBRACE, Literal: "{"}
	case ch == '}':
		return Token{Type: RBRACE, Literal: "}"}
	case ch == ':':
		return Token{Type: COLON, Literal: ":"}
	case ch == '+':
		return Token{Type: PLUS, Literal: "+"}
	case ch == '-':
		return Token{Type: MINUS, Literal: "-"}
	case ch == '*':
		return Token{Type: ASTERISK, Literal: "*"}
	case ch == '/':
		return Token{Type: SLASH, Literal: "/"}
	case ch == '&':
		return Token{Type: AND, Literal: "&"}
	case ch == '|':
		return Token{Type: OR, Literal: "|"}
	case ch == '!':
		if l.peekChar() == '=' {
			l.readChar()
			return Token{Type: NEQ, Literal: "!="}
		}
		return Token{Type: NOT, Literal: "!"}
	case ch == '<':
		if l.peekChar() == '=' {
			l.readChar()
			return Token{Type: LTE, Literal: "<="}
		}
		return Token{Type: LT, Literal: "<"}
	case ch == '>':
		if l.peekChar() == '=' {
			l.readChar()
			return Token{Type: GTE, Literal: ">="}
		}
		return Token{Type: GT, Literal: ">"}
	default:
		return Token{Type: ILLEGAL, Literal: string(ch)}
	}
}

func (l *Lexer) readChar() byte {
	ch, _ := l.reader.ReadByte()
	return ch
}

func (l *Lexer) peekChar() byte {
	ch, _ := l.reader.Peek(1)
	if len(ch) == 0 {
		return 0
	}
	return ch[0]
}

func (l *Lexer) skipWhitespace() {
	for {
		ch := l.peekChar()
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			l.readChar()
		} else {
			break
		}
	}
}

func (l *Lexer) readIdentifier(first byte) Token {
	var literal string
	literal += string(first)

	for {
		ch := l.peekChar()
		if !isLetter(ch) && !isDigit(ch) {
			break
		}
		literal += string(l.readChar())
	}

	switch literal {
	case "var":
		return Token{Type: VAR, Literal: literal}
	case "if":
		return Token{Type: IF, Literal: literal}
	case "else":
		return Token{Type: ELSE, Literal: literal}
	case "return":
		return Token{Type: RETURN, Literal: literal}
	case "true", "false":
		return Token{Type: BOOL, Literal: literal}
	default:
		return Token{Type: IDENT, Literal: literal}
	}
}

func (l *Lexer) readNumber(first byte) Token {
	var literal string
	literal += string(first)

	for {
		ch := l.peekChar()
		if !isDigit(ch) {
			break
		}
		literal += string(l.readChar())
	}

	return Token{Type: INT, Literal: literal}
}

func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || ch == '_'
}

func isDigit(ch byte) bool {
	return unicode.IsDigit(rune(ch))
}

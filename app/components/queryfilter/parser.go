package queryfilter

import (
	"fmt"
	"strings"
	"unicode"
)

// tokenType represents lexer token categories.
type tokenType int

const (
	tokEOF tokenType = iota
	tokIDENT
	tokSTRING  // quoted string
	tokVALUE   // unquoted non-ident value (numbers, mixed chars)
	tokCOMMA   // ,  (OR)
	tokSEMI    // ;  (AND)
	tokLPAREN  // (
	tokRPAREN  // )
	tokEqEq    // ==
	tokNeq     // !=
	tokLT      // <
	tokLTE     // <=
	tokGT      // >
	tokGTE     // >=
	tokNamedOp // =name=
)

type tok struct {
	typ tokenType
	val string
}

// lexer tokenises an RSQL input string.
type lexer struct {
	input []rune
	pos   int
}

func newLexer(input string) *lexer {
	return &lexer{input: []rune(input)}
}

func (l *lexer) peek() rune {
	if l.pos >= len(l.input) {
		return 0
	}

	return l.input[l.pos]
}

func (l *lexer) advance() rune {
	r := l.input[l.pos]
	l.pos++

	return r
}

func (l *lexer) skipWS() {
	for l.pos < len(l.input) && unicode.IsSpace(l.input[l.pos]) {
		l.pos++
	}
}

func (l *lexer) next() (tok, error) {
	l.skipWS()

	if l.pos >= len(l.input) {
		return tok{typ: tokEOF}, nil
	}

	ch := l.peek()

	switch {
	case ch == ',':
		l.advance()
		return tok{typ: tokCOMMA}, nil

	case ch == ';':
		l.advance()
		return tok{typ: tokSEMI}, nil

	case ch == '(':
		l.advance()
		return tok{typ: tokLPAREN}, nil

	case ch == ')':
		l.advance()
		return tok{typ: tokRPAREN}, nil

	case ch == '=':
		l.advance()
		if l.pos < len(l.input) && l.input[l.pos] == '=' {
			l.advance()
			return tok{typ: tokEqEq, val: "=="}, nil
		}
		// Named op: =name=
		start := l.pos
		for l.pos < len(l.input) && (unicode.IsLetter(l.input[l.pos]) || unicode.IsDigit(l.input[l.pos])) {
			l.pos++
		}

		if l.pos < len(l.input) && l.input[l.pos] == '=' && l.pos > start {
			name := string(l.input[start:l.pos])
			l.advance()
			return tok{typ: tokNamedOp, val: name}, nil
		}

		return tok{}, fmt.Errorf("queryfilter: invalid operator at position %d", l.pos)

	case ch == '!':
		l.advance()
		if l.pos < len(l.input) && l.input[l.pos] == '=' {
			l.advance()
			return tok{typ: tokNeq, val: "!="}, nil
		}

		return tok{}, fmt.Errorf("queryfilter: unexpected '!' at position %d", l.pos)

	case ch == '<':
		l.advance()
		if l.pos < len(l.input) && l.input[l.pos] == '=' {
			l.advance()
			return tok{typ: tokLTE, val: "<="}, nil
		}

		return tok{typ: tokLT, val: "<"}, nil

	case ch == '>':
		l.advance()
		if l.pos < len(l.input) && l.input[l.pos] == '=' {
			l.advance()
			return tok{typ: tokGTE, val: ">="}, nil
		}

		return tok{typ: tokGT, val: ">"}, nil

	case ch == '\'' || ch == '"':
		quote := l.advance()
		start := l.pos
		for l.pos < len(l.input) && l.input[l.pos] != quote {
			if l.input[l.pos] == '\\' {
				l.pos++
			}

			l.pos++
		}

		if l.pos >= len(l.input) {
			return tok{}, fmt.Errorf("queryfilter: unterminated string")
		}

		val := string(l.input[start:l.pos])
		l.advance()

		return tok{typ: tokSTRING, val: val}, nil

	case unicode.IsLetter(ch) || ch == '_':
		start := l.pos
		for l.pos < len(l.input) && (unicode.IsLetter(l.input[l.pos]) || unicode.IsDigit(l.input[l.pos]) || l.input[l.pos] == '_') {
			l.pos++
		}

		return tok{typ: tokIDENT, val: string(l.input[start:l.pos])}, nil

	default:
		// Unquoted value: read until RSQL separator or whitespace
		start := l.pos
		for l.pos < len(l.input) && !strings.ContainsRune(",;()", l.input[l.pos]) && !unicode.IsSpace(l.input[l.pos]) {
			l.pos++
		}

		if l.pos == start {
			return tok{}, fmt.Errorf("queryfilter: unexpected character '%c' at position %d", ch, l.pos)
		}

		return tok{typ: tokVALUE, val: string(l.input[start:l.pos])}, nil
	}
}

// parser is a single-pass recursive descent RSQL parser.
type parser struct {
	lex     *lexer
	current tok
	peeked  bool
	lexErr  error
}

func newParser(input string) *parser {
	return &parser{lex: newLexer(input)}
}

func (p *parser) peek() tok {
	if !p.peeked {
		p.current, p.lexErr = p.lex.next()
		p.peeked = true
	}

	return p.current
}

func (p *parser) consume() tok {
	t := p.peek()
	p.peeked = false

	return t
}

func (p *parser) parseOr() (Node, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for p.peek().typ == tokCOMMA {
		p.consume()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}

		left = &LogicalNode{Op: LogicalOR, Left: left, Right: right}
	}

	return left, nil
}

func (p *parser) parseAnd() (Node, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	for p.peek().typ == tokSEMI {
		p.consume()
		right, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}

		left = &LogicalNode{Op: LogicalAND, Left: left, Right: right}
	}

	return left, nil
}

func (p *parser) parsePrimary() (Node, error) {
	if p.peek().typ == tokLPAREN {
		p.consume()
		node, err := p.parseOr()
		if err != nil {
			return nil, err
		}

		if p.peek().typ != tokRPAREN {
			return nil, fmt.Errorf("queryfilter: expected ')'")
		}

		p.consume()

		return node, nil
	}

	return p.parseComparison()
}

func (p *parser) parseComparison() (Node, error) {
	fieldTok := p.consume()
	if fieldTok.typ != tokIDENT {
		return nil, fmt.Errorf("queryfilter: expected field name, got %q", fieldTok.val)
	}

	opTok := p.consume()
	op, err := tokToOp(opTok)
	if err != nil {
		return nil, err
	}

	valTok := p.consume()
	if valTok.typ != tokIDENT && valTok.typ != tokSTRING && valTok.typ != tokVALUE {
		return nil, fmt.Errorf("queryfilter: expected value after operator, got %q", valTok.val)
	}

	return &ComparisonNode{Field: fieldTok.val, Op: op, Value: valTok.val}, nil
}

func tokToOp(t tok) (Op, error) {
	switch t.typ {
	case tokEqEq:
		return OpEq, nil
	case tokNeq:
		return OpNe, nil
	case tokLT:
		return OpLt, nil
	case tokLTE:
		return OpLte, nil
	case tokGT:
		return OpGt, nil
	case tokGTE:
		return OpGte, nil
	case tokNamedOp:
		switch strings.ToLower(t.val) {
		case "eq":
			return OpEq, nil
		case "ne":
			return OpNe, nil
		case "gt":
			return OpGt, nil
		case "lt":
			return OpLt, nil
		case "gte":
			return OpGte, nil
		case "lte":
			return OpLte, nil
		case "prefix":
			return OpPrefix, nil
		case "suffix":
			return OpSuffix, nil
		case "like":
			return OpLike, nil
		default:
			return "", fmt.Errorf("queryfilter: unknown operator %q", t.val)
		}
	default:
		return "", fmt.Errorf("queryfilter: expected comparison operator, got %q", t.val)
	}
}

// ParseFilter parses an RSQL filter string into a Node AST.
// Returns (nil, nil) for empty input.
func ParseFilter(q string) (Node, error) {
	if strings.TrimSpace(q) == "" {
		return nil, nil
	}

	p := newParser(q)
	node, err := p.parseOr()
	if err != nil {
		return nil, err
	}

	if p.lexErr != nil {
		return nil, p.lexErr
	}

	if p.peek().typ != tokEOF {
		return nil, fmt.Errorf("queryfilter: unexpected token %q", p.peek().val)
	}

	return node, nil
}

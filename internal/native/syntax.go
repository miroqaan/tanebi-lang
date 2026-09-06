// Package native compiles the TANEBI systems language to freestanding Rust.
// The Rust backend is an implementation detail, not a source-language escape.
package native

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type Source struct{ Name, Text string }
type position struct {
	file         string
	line, column int
}

func (p position) fail(format string, args ...any) error {
	return fmt.Errorf("%s:%d:%d: %s", p.file, p.line, p.column, fmt.Sprintf(format, args...))
}

type token struct {
	text, kind string
	pos        position
}

func lex(s Source) ([]token, error) {
	r := []rune(s.Text)
	out := []token{}
	line, col := 1, 1
	advance := func() rune {
		c := r[0]
		r = r[1:]
		if c == '\n' {
			line++
			col = 1
		} else {
			col++
		}
		return c
	}
	for len(r) > 0 {
		if unicode.IsSpace(r[0]) {
			advance()
			continue
		}
		if r[0] == '#' {
			for len(r) > 0 && r[0] != '\n' {
				advance()
			}
			continue
		}
		p := position{s.Name, line, col}
		c := r[0]
		if unicode.IsLetter(c) || c == '_' {
			var b strings.Builder
			for len(r) > 0 && (unicode.IsLetter(r[0]) || unicode.IsDigit(r[0]) || r[0] == '_') {
				b.WriteRune(advance())
			}
			out = append(out, token{b.String(), "name", p})
			continue
		}
		if c >= '0' && c <= '9' {
			var b strings.Builder
			for len(r) > 0 && ((r[0] >= '0' && r[0] <= '9') || (r[0] >= 'a' && r[0] <= 'f') || (r[0] >= 'A' && r[0] <= 'F') || r[0] == 'x' || r[0] == 'X' || r[0] == 'b' || r[0] == 'B' || r[0] == '_') {
				b.WriteRune(advance())
			}
			out = append(out, token{b.String(), "number", p})
			continue
		}
		if c == '"' {
			advance()
			var b strings.Builder
			closed := false
			for len(r) > 0 {
				c = advance()
				if c == '"' {
					closed = true
					break
				}
				if c == '\n' {
					return nil, p.fail("unterminated string")
				}
				if c == '\\' {
					if len(r) == 0 {
						break
					}
					c = advance()
					switch c {
					case 'n':
						c = '\n'
					case 'r':
						c = '\r'
					case 't':
						c = '\t'
					case '0':
						c = 0
					case '\\', '"':
					default:
						return nil, p.fail("unknown string escape \\%c", c)
					}
				}
				b.WriteRune(c)
			}
			if !closed {
				return nil, p.fail("unterminated string")
			}
			out = append(out, token{b.String(), "string", p})
			continue
		}
		if len(r) >= 2 {
			pair := string(r[:2])
			switch pair {
			case "->", "==", "!=", "<=", ">=", "&&", "||", "<<", ">>", "+=", "-=", "|=", "&=":
				advance()
				advance()
				out = append(out, token{pair, pair, p})
				continue
			}
		}
		if strings.ContainsRune("(){}[],:;=+-*/%<>&|^!~", c) {
			advance()
			out = append(out, token{string(c), string(c), p})
			continue
		}
		return nil, p.fail("unexpected character %q", c)
	}
	out = append(out, token{"", "eof", position{s.Name, line, col}})
	return out, nil
}

type expr struct {
	kind, name string
	pos        position
	number     uint64
	text       string
	args       []*expr
	typ        string
	local      bool
}
type stmt struct {
	kind, name, typ, op string
	pos                 position
	value, condition    *expr
	body, other         []*stmt
	local               bool
}
type parameter struct {
	name, typ string
	pos       position
}
type declaration struct {
	kind, name, typ string
	pos             position
	value           *expr
	params          []parameter
	body            []*stmt
}
type parser struct {
	tokens []token
	n      int
}

func (p *parser) peek() token { return p.tokens[p.n] }
func (p *parser) take() token {
	t := p.peek()
	if t.kind != "eof" {
		p.n++
	}
	return t
}
func (p *parser) match(s string) bool {
	if p.is(s) {
		p.take()
		return true
	}
	return false
}
func (p *parser) is(s string) bool {
	t := p.peek()
	return t.text == s && (t.kind == s || t.kind == "name")
}
func (p *parser) expect(s string) error {
	if !p.match(s) {
		return p.peek().pos.fail("expected %q, got %q", s, p.peek().text)
	}
	return nil
}
func (p *parser) name() (token, error) {
	t := p.take()
	if t.kind != "name" {
		return t, t.pos.fail("expected a name")
	}
	return t, nil
}
func (p *parser) typeName(allowVoid bool) (string, error) {
	t, e := p.name()
	if e != nil {
		return "", e
	}
	if t.text != "u64" && t.text != "bool" && t.text != "str" && !(allowVoid && t.text == "void") {
		return "", t.pos.fail("unknown type %q", t.text)
	}
	return t.text, nil
}

func parseSource(s Source) ([]*declaration, error) {
	t, e := lex(s)
	if e != nil {
		return nil, e
	}
	p := parser{tokens: t}
	var ds []*declaration
	for p.peek().kind != "eof" {
		if p.match(";") {
			continue
		}
		k := p.take()
		if k.kind != "name" {
			return nil, k.pos.fail("expected fn, const, global, buffer, or embed")
		}
		switch k.text {
		case "fn", "const", "global", "buffer", "embed":
		default:
			return nil, k.pos.fail("expected fn, const, global, buffer, or embed")
		}
		name, e := p.name()
		if e != nil {
			return nil, e
		}
		d := &declaration{kind: k.text, name: name.text, pos: name.pos}
		switch d.kind {
		case "fn":
			if e = p.expect("("); e != nil {
				return nil, e
			}
			if !p.match(")") {
				for {
					n, e := p.name()
					if e != nil {
						return nil, e
					}
					if e = p.expect(":"); e != nil {
						return nil, e
					}
					ty, e := p.typeName(false)
					if e != nil {
						return nil, e
					}
					d.params = append(d.params, parameter{n.text, ty, n.pos})
					if p.match(")") {
						break
					}
					if e = p.expect(","); e != nil {
						return nil, e
					}
				}
			}
			d.typ = "void"
			if p.match("->") {
				d.typ, e = p.typeName(true)
				if e != nil {
					return nil, e
				}
			}
			d.body, e = p.block()
			if e != nil {
				return nil, e
			}
		case "const", "global":
			if e = p.expect(":"); e != nil {
				return nil, e
			}
			d.typ, e = p.typeName(false)
			if e != nil {
				return nil, e
			}
			if e = p.expect("="); e != nil {
				return nil, e
			}
			d.value, e = p.expression(0)
			if e != nil {
				return nil, e
			}
			p.match(";")
		case "buffer":
			d.typ = "memory"
			if e = p.expect("["); e != nil {
				return nil, e
			}
			d.value, e = p.expression(0)
			if e != nil {
				return nil, e
			}
			if e = p.expect("]"); e != nil {
				return nil, e
			}
			p.match(";")
		case "embed":
			d.typ = "memory"
			p.match(";")
		}
		ds = append(ds, d)
	}
	return ds, nil
}
func (p *parser) block() ([]*stmt, error) {
	if e := p.expect("{"); e != nil {
		return nil, e
	}
	var out []*stmt
	for !p.match("}") {
		if p.peek().kind == "eof" {
			return nil, p.peek().pos.fail("expected '}'")
		}
		if p.match(";") {
			continue
		}
		s, e := p.statement()
		if e != nil {
			return nil, e
		}
		out = append(out, s)
	}
	return out, nil
}
func (p *parser) statement() (*stmt, error) {
	t := p.peek()
	s := &stmt{kind: t.text, pos: t.pos}
	var e error
	keyword := ""
	if t.kind == "name" {
		keyword = t.text
	}
	switch keyword {
	case "let":
		p.take()
		n, err := p.name()
		if err != nil {
			return nil, err
		}
		s.name = n.text
		if e = p.expect(":"); e != nil {
			return nil, e
		}
		s.typ, e = p.typeName(false)
		if e != nil {
			return nil, e
		}
		if e = p.expect("="); e != nil {
			return nil, e
		}
		s.value, e = p.expression(0)
	case "if":
		p.take()
		s.condition, e = p.expression(0)
		if e != nil {
			return nil, e
		}
		s.body, e = p.block()
		if e != nil {
			return nil, e
		}
		if p.match("else") {
			if p.is("if") {
				nested, err := p.statement()
				e = err
				s.other = []*stmt{nested}
			} else {
				s.other, e = p.block()
			}
		}
	case "while":
		p.take()
		s.condition, e = p.expression(0)
		if e != nil {
			return nil, e
		}
		s.body, e = p.block()
	case "return":
		p.take()
		if !p.is(";") && !p.is("}") {
			s.value, e = p.expression(0)
		}
	case "break", "continue":
		p.take()
	default:
		if t.kind == "name" && p.n+1 < len(p.tokens) {
			op := p.tokens[p.n+1].kind
			if op == "=" || op == "+=" || op == "-=" || op == "|=" || op == "&=" {
				p.take()
				p.take()
				s.kind = "assign"
				s.name = t.text
				s.op = op
				s.value, e = p.expression(0)
				break
			}
		}
		s.kind = "expr"
		s.value, e = p.expression(0)
	}
	if e != nil {
		return nil, e
	}
	p.match(";")
	return s, nil
}

var precedence = map[string]int{"||": 1, "&&": 2, "|": 3, "^": 4, "&": 5, "==": 6, "!=": 6, "<": 7, "<=": 7, ">": 7, ">=": 7, "<<": 8, ">>": 8, "+": 9, "-": 9, "*": 10, "/": 10, "%": 10}

func (p *parser) expression(min int) (*expr, error) {
	t := p.take()
	x := &expr{pos: t.pos}
	var e error
	switch {
	case t.kind == "!" || t.kind == "~" || t.kind == "-" || t.kind == "+":
		x.kind = "unary"
		x.name = t.text
		child, err := p.expression(11)
		e = err
		x.args = []*expr{child}
	case t.kind == "(":
		x, e = p.expression(0)
		if e == nil {
			e = p.expect(")")
		}
	case t.kind == "number":
		x.kind = "number"
		raw := strings.ReplaceAll(t.text, "_", "")
		base := 10
		if strings.HasPrefix(raw, "0x") || strings.HasPrefix(raw, "0X") {
			raw = raw[2:]
			base = 16
		} else if strings.HasPrefix(raw, "0b") || strings.HasPrefix(raw, "0B") {
			raw = raw[2:]
			base = 2
		}
		x.number, e = strconv.ParseUint(raw, base, 64)
		if e != nil {
			return nil, t.pos.fail("invalid u64 literal %q", t.text)
		}
	case t.kind == "string":
		x.kind = "string"
		x.text = t.text
	case t.kind == "name" && (t.text == "true" || t.text == "false"):
		x.kind = "bool"
		x.name = t.text
	case t.kind == "name":
		x.kind = "name"
		x.name = t.text
		if p.match("(") {
			x.kind = "call"
			if !p.match(")") {
				for {
					a, err := p.expression(0)
					if err != nil {
						return nil, err
					}
					x.args = append(x.args, a)
					if p.match(")") {
						break
					}
					if e = p.expect(","); e != nil {
						return nil, e
					}
				}
			}
		}
	default:
		return nil, t.pos.fail("expected expression, got %q", t.text)
	}
	if e != nil {
		return nil, e
	}
	for {
		op := p.peek()
		level, ok := precedence[op.kind]
		if !ok || level < min {
			break
		}
		p.take()
		right, e := p.expression(level + 1)
		if e != nil {
			return nil, e
		}
		x = &expr{kind: "binary", name: op.text, pos: op.pos, args: []*expr{x, right}}
	}
	return x, nil
}

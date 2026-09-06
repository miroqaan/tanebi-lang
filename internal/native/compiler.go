package native

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Options controls only generic compilation and explicit asset/config inputs.
type Options struct {
	// Target is "host", "uefi", or "library". Empty preserves the UEFI default.
	Target string
	Entry  string
	// Deprecated: use Target: "library". Kept for existing embedding callers.
	Library bool
	Assets  map[string]string
	Defines map[string]uint64
}
type signature struct {
	args   []string
	result string
}

var builtins = map[string]signature{
	"in8": {[]string{"u64"}, "u64"}, "out8": {[]string{"u64", "u64"}, "void"}, "out32": {[]string{"u64", "u64"}, "void"},
	"load8": {[]string{"u64"}, "u64"}, "load16": {[]string{"u64"}, "u64"}, "load32": {[]string{"u64"}, "u64"}, "load64": {[]string{"u64"}, "u64"},
	"store8": {[]string{"u64", "u64"}, "void"}, "store16": {[]string{"u64", "u64"}, "void"}, "store32": {[]string{"u64", "u64"}, "void"}, "store64": {[]string{"u64", "u64"}, "void"},
	"copy": {[]string{"u64", "u64", "u64"}, "void"}, "zero": {[]string{"u64", "u64"}, "void"}, "ticks": {nil, "u64"}, "pause": {nil, "void"}, "irq_disable": {nil, "void"},
	"byte": {[]string{"str", "u64"}, "u64"}, "config": {[]string{"str"}, "u64"},
	"print": {[]string{"printable"}, "void"}, "println": {[]string{"printable"}, "void"},
}

func builtin(name string) (signature, bool) {
	if s, ok := builtins[name]; ok {
		return s, true
	}
	if len(name) == 5 && strings.HasPrefix(name, "call") && name[4] >= '0' && name[4] <= '6' {
		args := make([]string, int(name[4]-'0')+1)
		for n := range args {
			args[n] = "u64"
		}
		return signature{args, "u64"}, true
	}
	if name == "addr" || name == "len" {
		return signature{[]string{"addressable"}, "u64"}, true
	}
	return signature{}, false
}

type constant struct {
	typ string
	n   uint64
	b   bool
	s   string
}
type compiler struct {
	options   Options
	decls     []*declaration
	symbols   map[string]*declaration
	constants map[string]constant
	visiting  map[string]bool
	sizes     map[string]uint64
	assets    map[string]string
}

// Compile validates all sources as one program before emitting Rust. It never
// accepts foreign code blocks, textual macros, imports, or arbitrary Rust.
func Compile(sources []Source, options Options) (string, error) {
	if options.Target == "" {
		options.Target = "uefi"
		if options.Library {
			options.Target = "library"
		}
	}
	if options.Library && options.Target != "library" {
		return "", fmt.Errorf("Library option conflicts with target %q; use Target: library", options.Target)
	}
	switch options.Target {
	case "host", "uefi", "library":
	default:
		return "", fmt.Errorf("unknown target %q; expected host, uefi, or library", options.Target)
	}
	if options.Entry == "" {
		if options.Target == "host" {
			options.Entry = "main"
		} else if options.Target == "uefi" {
			options.Entry = "entry"
		}
	}
	c := &compiler{options: options, symbols: map[string]*declaration{}, constants: map[string]constant{}, visiting: map[string]bool{}, sizes: map[string]uint64{}, assets: map[string]string{}}
	if len(sources) == 0 {
		return "", fmt.Errorf("no source files")
	}
	for _, source := range sources {
		ds, e := parseSource(source)
		if e != nil {
			return "", e
		}
		for _, d := range ds {
			if _, ok := c.symbols[d.name]; ok {
				return "", d.pos.fail("duplicate declaration %q", d.name)
			}
			if _, ok := builtin(d.name); ok {
				return "", d.pos.fail("declaration uses builtin name %q", d.name)
			}
			c.symbols[d.name] = d
			c.decls = append(c.decls, d)
		}
	}
	for _, d := range c.decls {
		if d.kind == "embed" {
			asset, ok := options.Assets[d.name]
			if !ok {
				for k, v := range options.Assets {
					if strings.EqualFold(k, d.name) {
						if ok {
							return "", d.pos.fail("ambiguous asset %q", d.name)
						}
						asset = v
						ok = true
					}
				}
			}
			if !ok {
				return "", d.pos.fail("missing -asset %s=PATH", d.name)
			}
			absolute, e := filepath.Abs(asset)
			if e != nil {
				return "", e
			}
			info, e := os.Stat(absolute)
			if e != nil {
				return "", d.pos.fail("asset: %v", e)
			}
			if !info.Mode().IsRegular() {
				return "", d.pos.fail("asset must be a regular file")
			}
			c.assets[d.name] = absolute
			c.sizes[d.name] = uint64(info.Size())
		}
	}
	for _, d := range c.decls {
		if d.kind == "const" || d.kind == "global" || d.kind == "buffer" {
			v, e := c.declarationConstant(d.name)
			if e != nil {
				return "", e
			}
			if d.kind == "buffer" {
				if v.n == 0 || v.n > 1<<31 {
					return "", d.pos.fail("buffer size must be 1..2147483648 bytes")
				}
				c.sizes[d.name] = v.n
			}
		}
	}
	for _, d := range c.decls {
		if d.kind == "fn" {
			v := validator{c: c, fn: d, scopes: []map[string]string{{}}}
			for _, p := range d.params {
				if _, ok := v.scopes[0][p.name]; ok {
					return "", p.pos.fail("duplicate parameter %q", p.name)
				}
				v.scopes[0][p.name] = p.typ
			}
			if e := v.block(d.body, 0, false); e != nil {
				return "", e
			}
			if d.typ != "void" && !returns(d.body) {
				return "", d.pos.fail("function %q can reach its end without returning %s", d.name, d.typ)
			}
		}
	}
	if options.Target != "library" {
		entry := options.Entry
		d, ok := c.symbols[entry]
		if !ok || d.kind != "fn" {
			return "", fmt.Errorf("entry function %q not found", entry)
		}
		if options.Target == "host" {
			if d.typ != "void" || len(d.params) != 0 {
				return "", d.pos.fail("host entry must have signature fn %s()->void with no arguments", entry)
			}
		} else if d.typ != "u64" || len(d.params) != 2 || d.params[0].typ != "u64" || d.params[1].typ != "u64" {
			return "", d.pos.fail("entry must have signature fn %s(image:u64, st:u64)->u64", entry)
		}
	}
	return c.emit(), nil
}

func (c *compiler) declarationConstant(name string) (constant, error) {
	if v, ok := c.constants[name]; ok {
		return v, nil
	}
	d, ok := c.symbols[name]
	if !ok {
		return constant{}, fmt.Errorf("unknown constant %q", name)
	}
	if c.visiting[name] {
		return constant{}, d.pos.fail("cyclic constant %q", name)
	}
	c.visiting[name] = true
	defer delete(c.visiting, name)
	checker := validator{c: c, scopes: []map[string]string{{}}}
	if _, e := checker.expression(d.value); e != nil {
		return constant{}, e
	}
	v, e := c.constExpr(d.value)
	if e != nil {
		return v, e
	}
	want := d.typ
	if d.kind == "buffer" {
		want = "u64"
	}
	if v.typ != want {
		return v, d.pos.fail("initializer has type %s, expected %s", v.typ, want)
	}
	c.constants[name] = v
	return v, nil
}
func (c *compiler) constExpr(x *expr) (constant, error) {
	if x == nil {
		return constant{}, fmt.Errorf("missing constant initializer")
	}
	err := func() (constant, error) { return constant{}, x.pos.fail("expression is not a compile-time constant") }
	switch x.kind {
	case "number":
		return constant{typ: "u64", n: x.number}, nil
	case "string":
		return constant{typ: "str", s: x.text}, nil
	case "bool":
		return constant{typ: "bool", b: x.name == "true"}, nil
	case "name":
		d, ok := c.symbols[x.name]
		if !ok {
			return constant{}, x.pos.fail("unknown symbol %q", x.name)
		}
		if d.kind != "const" {
			return err()
		}
		return c.declarationConstant(x.name)
	case "call":
		if x.name == "config" && len(x.args) == 1 && x.args[0].kind == "string" {
			return constant{typ: "u64", n: c.options.Defines[x.args[0].text]}, nil
		}
		if x.name == "len" && len(x.args) == 1 {
			a := x.args[0]
			if a.kind == "name" {
				d := c.symbols[a.name]
				if d.kind == "embed" {
					return constant{typ: "u64", n: c.sizes[a.name]}, nil
				}
				if d.kind == "buffer" {
					return c.declarationConstant(a.name)
				}
			}
			value, e := c.constExpr(a)
			if e != nil {
				return constant{}, e
			}
			if value.typ == "str" {
				return constant{typ: "u64", n: uint64(len(value.s))}, nil
			}
		}
		return err()
	case "unary":
		a, e := c.constExpr(x.args[0])
		if e != nil {
			return a, e
		}
		if x.name == "!" && a.typ == "bool" {
			a.b = !a.b
			return a, nil
		}
		if a.typ == "u64" {
			switch x.name {
			case "-":
				a.n = 0 - a.n
			case "~":
				a.n = ^a.n
			case "+":
			default:
				return constant{}, x.pos.fail("invalid unary operator for u64")
			}
			return a, nil
		}
		return constant{}, x.pos.fail("invalid unary operand type %s", a.typ)
	case "binary":
		a, e := c.constExpr(x.args[0])
		if e != nil {
			return a, e
		}
		// All operands have already been type-checked, but unreachable operands
		// are not evaluated, matching the generated Rust's short-circuit rules.
		if x.name == "&&" && a.typ == "bool" && !a.b {
			return a, nil
		}
		if x.name == "||" && a.typ == "bool" && a.b {
			return a, nil
		}
		b, e := c.constExpr(x.args[1])
		if e != nil {
			return b, e
		}
		ty, e := binaryType(x, a.typ, b.typ)
		if e != nil {
			return constant{}, e
		}
		v := constant{typ: ty}
		switch x.name {
		case "+":
			v.n = a.n + b.n
		case "-":
			v.n = a.n - b.n
		case "*":
			v.n = a.n * b.n
		case "/", "%":
			if b.n == 0 {
				return v, x.pos.fail("division by zero in constant")
			}
			if x.name == "/" {
				v.n = a.n / b.n
			} else {
				v.n = a.n % b.n
			}
		case "<<":
			v.n = a.n << (b.n & 63)
		case ">>":
			v.n = a.n >> (b.n & 63)
		case "&":
			v.n = a.n & b.n
		case "|":
			v.n = a.n | b.n
		case "^":
			v.n = a.n ^ b.n
		case "&&":
			v.b = a.b && b.b
		case "||":
			v.b = a.b || b.b
		case "==", "!=":
			v.b = a == b
			if x.name == "!=" {
				v.b = !v.b
			}
		case "<":
			v.b = a.n < b.n
		case "<=":
			v.b = a.n <= b.n
		case ">":
			v.b = a.n > b.n
		case ">=":
			v.b = a.n >= b.n
		}
		return v, nil
	}
	return err()
}
func binaryType(x *expr, a, b string) (string, error) {
	if a != b {
		return "", x.pos.fail("operator %s has mismatched operands %s and %s", x.name, a, b)
	}
	switch x.name {
	case "==", "!=":
		if a == "u64" || a == "str" || a == "bool" {
			return "bool", nil
		}
	case "&&", "||":
		if a == "bool" {
			return "bool", nil
		}
	case "<", "<=", ">", ">=":
		if a == "u64" {
			return "bool", nil
		}
	default:
		if a == "u64" {
			return "u64", nil
		}
	}
	return "", x.pos.fail("operator %s does not accept %s", x.name, a)
}

type validator struct {
	c      *compiler
	fn     *declaration
	scopes []map[string]string
}

func (v *validator) isLocal(name string) bool {
	for n := len(v.scopes) - 1; n >= 0; n-- {
		if _, ok := v.scopes[n][name]; ok {
			return true
		}
	}
	return false
}

func (v *validator) lookup(name string) (string, bool) {
	for n := len(v.scopes) - 1; n >= 0; n-- {
		if ty, ok := v.scopes[n][name]; ok {
			return ty, true
		}
	}
	if d, ok := v.c.symbols[name]; ok && d.kind != "fn" {
		return d.typ, true
	}
	return "", false
}
func (v *validator) expression(x *expr) (string, error) {
	var ty string
	switch x.kind {
	case "number":
		ty = "u64"
	case "string":
		ty = "str"
	case "bool":
		ty = "bool"
	case "name":
		var ok bool
		ty, ok = v.lookup(x.name)
		if !ok {
			return "", x.pos.fail("unknown symbol %q", x.name)
		}
		x.local = v.isLocal(x.name)
	case "unary":
		a, e := v.expression(x.args[0])
		if e != nil {
			return "", e
		}
		ty = "u64"
		if x.name == "!" {
			ty = "bool"
		}
		if a != ty {
			return "", x.pos.fail("operator %s requires %s, got %s", x.name, ty, a)
		}
	case "binary":
		a, e := v.expression(x.args[0])
		if e != nil {
			return "", e
		}
		b, e := v.expression(x.args[1])
		if e != nil {
			return "", e
		}
		ty, e = binaryType(x, a, b)
		if e != nil {
			return "", e
		}
	case "call":
		if (x.name == "print" || x.name == "println") && v.c.options.Target != "host" {
			return "", x.pos.fail("%s is available only for target host, not %s", x.name, v.c.options.Target)
		}
		sig, ok := builtin(x.name)
		if !ok {
			d, found := v.c.symbols[x.name]
			if !found || d.kind != "fn" {
				return "", x.pos.fail("unknown function %q", x.name)
			}
			sig.result = d.typ
			for _, p := range d.params {
				sig.args = append(sig.args, p.typ)
			}
		}
		if len(sig.args) != len(x.args) {
			return "", x.pos.fail("%s expects %d arguments, got %d", x.name, len(sig.args), len(x.args))
		}
		for n, a := range x.args {
			actual, e := v.expression(a)
			if e != nil {
				return "", e
			}
			want := sig.args[n]
			if want == "addressable" {
				if actual != "memory" && actual != "str" {
					return "", a.pos.fail("%s requires a buffer, embed, or str", x.name)
				}
			} else if want == "printable" {
				if actual != "u64" && actual != "bool" && actual != "str" {
					return "", a.pos.fail("%s requires u64, bool, or str, got %s", x.name, actual)
				}
			} else if actual != want {
				return "", a.pos.fail("argument %d of %s has type %s, expected %s", n+1, x.name, actual, want)
			}
		}
		if x.name == "config" && x.args[0].kind != "string" {
			return "", x.pos.fail("config requires a string literal")
		}
		ty = sig.result
	}
	x.typ = ty
	return ty, nil
}
func (v *validator) block(ss []*stmt, depth int, scope bool) error {
	if scope {
		v.scopes = append(v.scopes, map[string]string{})
		defer func() { v.scopes = v.scopes[:len(v.scopes)-1] }()
	}
	for _, s := range ss {
		switch s.kind {
		case "let":
			ty, e := v.expression(s.value)
			if e != nil {
				return e
			}
			if ty != s.typ {
				return s.pos.fail("local %s has initializer type %s, expected %s", s.name, ty, s.typ)
			}
			current := v.scopes[len(v.scopes)-1]
			if _, ok := current[s.name]; ok {
				return s.pos.fail("duplicate local %q", s.name)
			}
			current[s.name] = s.typ
		case "assign":
			want, ok := v.lookup(s.name)
			if !ok {
				return s.pos.fail("unknown assignment target %q", s.name)
			}
			local := false
			for _, scope := range v.scopes {
				if _, ok := scope[s.name]; ok {
					local = true
				}
			}
			if !local {
				d := v.c.symbols[s.name]
				if d.kind != "global" {
					return s.pos.fail("cannot assign to %s %q", d.kind, s.name)
				}
			}
			s.local = local
			got, e := v.expression(s.value)
			if e != nil {
				return e
			}
			if got != want {
				return s.pos.fail("assignment has type %s, expected %s", got, want)
			}
			if s.op != "=" && want != "u64" {
				return s.pos.fail("compound assignment requires u64")
			}
		case "if", "while":
			ty, e := v.expression(s.condition)
			if e != nil {
				return e
			}
			if ty != "bool" {
				return s.pos.fail("%s condition must be bool", s.kind)
			}
			childDepth := depth
			if s.kind == "while" {
				childDepth++
			}
			if e = v.block(s.body, childDepth, true); e != nil {
				return e
			}
			if e = v.block(s.other, depth, true); e != nil {
				return e
			}
		case "return":
			got := "void"
			if s.value != nil {
				var e error
				got, e = v.expression(s.value)
				if e != nil {
					return e
				}
			}
			if got != v.fn.typ {
				return s.pos.fail("return has type %s, expected %s", got, v.fn.typ)
			}
		case "break", "continue":
			if depth == 0 {
				return s.pos.fail("%s outside a loop", s.kind)
			}
		case "expr":
			if s.value.kind != "call" {
				return s.pos.fail("only a function call may be used as an expression statement")
			}
			if _, e := v.expression(s.value); e != nil {
				return e
			}
		}
	}
	return nil
}
func containsBreak(ss []*stmt) bool {
	for _, s := range ss {
		if s.kind == "break" {
			return true
		}
		if s.kind == "if" && (containsBreak(s.body) || containsBreak(s.other)) {
			return true
		}
	}
	return false
}
func returns(ss []*stmt) bool {
	for _, s := range ss {
		if s.kind == "return" {
			return true
		}
		if s.kind == "if" && len(s.other) > 0 && returns(s.body) && returns(s.other) {
			return true
		}
		if s.kind == "while" && s.condition.kind == "bool" && s.condition.name == "true" && !containsBreak(s.body) {
			return true
		}
	}
	return false
}

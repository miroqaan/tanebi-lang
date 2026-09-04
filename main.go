package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

type tokenKind int

const (
	tokenEOF tokenKind = iota
	tokenNewline
	tokenNumber
	tokenName
	tokenLet
	tokenPrint
	tokenPlus
	tokenMinus
	tokenStar
	tokenSlash
	tokenLeftParen
	tokenRightParen
	tokenEqual
)

type token struct {
	kind   tokenKind
	text   string
	line   int
	column int
}

type lexer struct {
	source []rune
	index  int
	line   int
	column int
}

func tokenize(source string) ([]token, error) {
	l := lexer{source: []rune(source), line: 1, column: 1}
	var tokens []token

	for !l.atEnd() {
		ch := l.peek()

		if ch == ' ' || ch == '\t' || ch == '\r' {
			l.advance()
			continue
		}

		if ch == '\n' {
			tokens = append(tokens, token{tokenNewline, "\n", l.line, l.column})
			l.advance()
			continue
		}

		if ch == '#' {
			for !l.atEnd() && l.peek() != '\n' {
				l.advance()
			}
			continue
		}

		if unicode.IsDigit(ch) {
			tok, err := l.number()
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, tok)
			continue
		}

		if unicode.IsLetter(ch) || ch == '_' {
			tokens = append(tokens, l.identifier())
			continue
		}

		startLine, startColumn := l.line, l.column
		l.advance()
		kind, ok := map[rune]tokenKind{
			';': tokenNewline,
			'+': tokenPlus,
			'-': tokenMinus,
			'*': tokenStar,
			'/': tokenSlash,
			'(': tokenLeftParen,
			')': tokenRightParen,
			'=': tokenEqual,
		}[ch]
		if !ok {
			return nil, miniError(startLine, startColumn, "알 수 없는 문자 %q", ch)
		}
		tokens = append(tokens, token{kind, string(ch), startLine, startColumn})
	}

	tokens = append(tokens, token{tokenEOF, "", l.line, l.column})
	return tokens, nil
}

func (l *lexer) number() (token, error) {
	start := l.index
	startLine, startColumn := l.line, l.column
	seenDot := false

	for !l.atEnd() {
		ch := l.peek()
		if ch == '.' && !seenDot {
			seenDot = true
		} else if !unicode.IsDigit(ch) {
			break
		}
		l.advance()
	}

	text := string(l.source[start:l.index])
	if strings.HasSuffix(text, ".") {
		return token{}, miniError(startLine, startColumn, "숫자 뒤에 소수점만 올 수 없습니다")
	}
	return token{tokenNumber, text, startLine, startColumn}, nil
}

func (l *lexer) identifier() token {
	start := l.index
	startLine, startColumn := l.line, l.column
	for !l.atEnd() && (unicode.IsLetter(l.peek()) || unicode.IsDigit(l.peek()) || l.peek() == '_') {
		l.advance()
	}

	text := string(l.source[start:l.index])
	kind := tokenName
	switch text {
	case "let":
		kind = tokenLet
	case "print":
		kind = tokenPrint
	}
	return token{kind, text, startLine, startColumn}
}

func (l *lexer) atEnd() bool {
	return l.index >= len(l.source)
}

func (l *lexer) peek() rune {
	return l.source[l.index]
}

func (l *lexer) advance() rune {
	ch := l.source[l.index]
	l.index++
	if ch == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
	return ch
}

type expression interface {
	expressionNode()
}

type numberExpression struct {
	value float64
}

func (numberExpression) expressionNode() {}

type variableExpression struct {
	name  string
	token token
}

func (variableExpression) expressionNode() {}

type unaryExpression struct {
	operator token
	value    expression
}

func (unaryExpression) expressionNode() {}

type binaryExpression struct {
	left     expression
	operator token
	right    expression
}

func (binaryExpression) expressionNode() {}

type statement interface {
	statementNode()
}

type letStatement struct {
	name  string
	value expression
}

func (letStatement) statementNode() {}

type printStatement struct {
	value expression
}

func (printStatement) statementNode() {}

type parser struct {
	tokens  []token
	current int
}

func parse(tokens []token) ([]statement, error) {
	p := parser{tokens: tokens}
	var statements []statement
	p.skipNewlines()

	for !p.check(tokenEOF) {
		stmt, err := p.statement()
		if err != nil {
			return nil, err
		}
		statements = append(statements, stmt)
		if !p.check(tokenNewline, tokenEOF) {
			tok := p.peek()
			return nil, miniError(tok.line, tok.column, "문장 끝에는 줄바꿈이나 ';'가 필요합니다")
		}
		p.skipNewlines()
	}
	return statements, nil
}

func (p *parser) statement() (statement, error) {
	if p.match(tokenLet) {
		name, err := p.consume(tokenName, "let 뒤에는 변수 이름이 필요합니다")
		if err != nil {
			return nil, err
		}
		if _, err := p.consume(tokenEqual, "변수 이름 뒤에는 '='가 필요합니다"); err != nil {
			return nil, err
		}
		value, err := p.expression()
		if err != nil {
			return nil, err
		}
		return letStatement{name: name.text, value: value}, nil
	}

	if p.match(tokenPrint) {
		value, err := p.expression()
		if err != nil {
			return nil, err
		}
		return printStatement{value: value}, nil
	}

	tok := p.peek()
	return nil, miniError(tok.line, tok.column, "문장은 let 또는 print로 시작해야 합니다")
}

func (p *parser) expression() (expression, error) {
	expr, err := p.term()
	if err != nil {
		return nil, err
	}
	for p.match(tokenPlus, tokenMinus) {
		operator := p.previous()
		right, err := p.term()
		if err != nil {
			return nil, err
		}
		expr = binaryExpression{left: expr, operator: operator, right: right}
	}
	return expr, nil
}

func (p *parser) term() (expression, error) {
	expr, err := p.unary()
	if err != nil {
		return nil, err
	}
	for p.match(tokenStar, tokenSlash) {
		operator := p.previous()
		right, err := p.unary()
		if err != nil {
			return nil, err
		}
		expr = binaryExpression{left: expr, operator: operator, right: right}
	}
	return expr, nil
}

func (p *parser) unary() (expression, error) {
	if p.match(tokenPlus, tokenMinus) {
		operator := p.previous()
		value, err := p.unary()
		if err != nil {
			return nil, err
		}
		return unaryExpression{operator: operator, value: value}, nil
	}
	return p.primary()
}

func (p *parser) primary() (expression, error) {
	if p.match(tokenNumber) {
		tok := p.previous()
		value, err := strconv.ParseFloat(tok.text, 64)
		if err != nil {
			return nil, miniError(tok.line, tok.column, "잘못된 숫자 %q", tok.text)
		}
		return numberExpression{value: value}, nil
	}

	if p.match(tokenName) {
		tok := p.previous()
		return variableExpression{name: tok.text, token: tok}, nil
	}

	if p.match(tokenLeftParen) {
		expr, err := p.expression()
		if err != nil {
			return nil, err
		}
		if _, err := p.consume(tokenRightParen, "괄호를 ')'로 닫아야 합니다"); err != nil {
			return nil, err
		}
		return expr, nil
	}

	tok := p.peek()
	return nil, miniError(tok.line, tok.column, "숫자, 변수 또는 괄호가 필요합니다")
}

func (p *parser) skipNewlines() {
	for p.match(tokenNewline) {
	}
}

func (p *parser) consume(kind tokenKind, message string) (token, error) {
	if p.check(kind) {
		return p.advance(), nil
	}
	tok := p.peek()
	return token{}, miniError(tok.line, tok.column, message)
}

func (p *parser) match(kinds ...tokenKind) bool {
	if p.check(kinds...) {
		p.advance()
		return true
	}
	return false
}

func (p *parser) check(kinds ...tokenKind) bool {
	for _, kind := range kinds {
		if p.peek().kind == kind {
			return true
		}
	}
	return false
}

func (p *parser) advance() token {
	tok := p.peek()
	if tok.kind != tokenEOF {
		p.current++
	}
	return tok
}

func (p *parser) peek() token {
	return p.tokens[p.current]
}

func (p *parser) previous() token {
	return p.tokens[p.current-1]
}

type interpreter struct {
	variables map[string]float64
	output    func(string)
}

func newInterpreter(output func(string)) *interpreter {
	return &interpreter{
		variables: make(map[string]float64),
		output:    output,
	}
}

func (i *interpreter) run(statements []statement) error {
	for _, stmt := range statements {
		switch stmt := stmt.(type) {
		case letStatement:
			value, err := i.evaluate(stmt.value)
			if err != nil {
				return err
			}
			i.variables[stmt.name] = value
		case printStatement:
			value, err := i.evaluate(stmt.value)
			if err != nil {
				return err
			}
			i.output(formatNumber(value))
		default:
			return errors.New("알 수 없는 문장입니다")
		}
	}
	return nil
}

func (i *interpreter) evaluate(expr expression) (float64, error) {
	switch expr := expr.(type) {
	case numberExpression:
		return expr.value, nil
	case variableExpression:
		value, ok := i.variables[expr.name]
		if !ok {
			return 0, miniError(expr.token.line, expr.token.column, "정의되지 않은 변수 %q", expr.name)
		}
		return value, nil
	case unaryExpression:
		value, err := i.evaluate(expr.value)
		if err != nil {
			return 0, err
		}
		if expr.operator.kind == tokenMinus {
			return -value, nil
		}
		return value, nil
	case binaryExpression:
		left, err := i.evaluate(expr.left)
		if err != nil {
			return 0, err
		}
		right, err := i.evaluate(expr.right)
		if err != nil {
			return 0, err
		}
		switch expr.operator.kind {
		case tokenPlus:
			return left + right, nil
		case tokenMinus:
			return left - right, nil
		case tokenStar:
			return left * right, nil
		case tokenSlash:
			if right == 0 {
				return 0, miniError(expr.operator.line, expr.operator.column, "0으로 나눌 수 없습니다")
			}
			return left / right, nil
		}
	}
	return 0, errors.New("알 수 없는 표현식입니다")
}

func runSource(source string, output func(string)) error {
	tokens, err := tokenize(source)
	if err != nil {
		return err
	}
	program, err := parse(tokens)
	if err != nil {
		return err
	}
	return newInterpreter(output).run(program)
}

func formatNumber(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func miniError(line, column int, format string, args ...any) error {
	return fmt.Errorf("%d:%d: %s", line, column, fmt.Sprintf(format, args...))
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "사용법: go run . <파일.mini>")
		os.Exit(2)
	}

	source, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "파일 오류: %v\n", err)
		os.Exit(1)
	}

	err = runSource(string(source), func(value string) {
		fmt.Println(value)
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Mini 오류: %v\n", err)
		os.Exit(1)
	}
}

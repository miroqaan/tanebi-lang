package tanebi

import "strconv"

type parser struct {
	tokens  []token
	current int
}

func parse(tokens []token) ([]statement, error) {
	p := parser{tokens: tokens}
	return p.statements(tokenEOF)
}

func (p *parser) statements(stop tokenKind) ([]statement, error) {
	var result []statement
	p.skipNewlines()
	for !p.check(stop) && !p.check(tokenEOF) {
		stmt, err := p.statement()
		if err != nil {
			return nil, err
		}
		result = append(result, stmt)
		if !p.check(tokenNewline, stop, tokenEOF) {
			return nil, sourceError(p.peek(), "expected a newline or ';' after statement")
		}
		p.skipNewlines()
	}
	if stop != tokenEOF && p.check(tokenEOF) {
		return nil, sourceError(p.peek(), "expected '}' before end of file")
	}
	return result, nil
}

func (p *parser) statement() (statement, error) {
	if p.match(tokenLet) {
		name, err := p.consume(tokenIdentifier, "expected a variable name after 'let'")
		if err != nil {
			return nil, err
		}
		if _, err := p.consume(tokenEqual, "expected '=' after variable name"); err != nil {
			return nil, err
		}
		value, err := p.expression()
		if err != nil {
			return nil, err
		}
		return letStatement{name: name, value: value}, nil
	}
	if p.match(tokenPrint) {
		value, err := p.expression()
		if err != nil {
			return nil, err
		}
		return printStatement{value: value}, nil
	}
	if p.match(tokenRepeat) {
		count, err := p.expression()
		if err != nil {
			return nil, err
		}
		body, err := p.block()
		if err != nil {
			return nil, err
		}
		return repeatStatement{count: count, body: body}, nil
	}
	if p.match(tokenIf) {
		condition, err := p.expression()
		if err != nil {
			return nil, err
		}
		thenBody, err := p.block()
		if err != nil {
			return nil, err
		}
		separatorStart := p.current
		p.skipNewlines()
		var elseBody []statement
		if p.match(tokenElse) {
			elseBody, err = p.block()
			if err != nil {
				return nil, err
			}
		} else {
			p.current = separatorStart
		}
		return ifStatement{condition: condition, thenBody: thenBody, elseBody: elseBody}, nil
	}
	if p.check(tokenIdentifier) && p.peekNext().kind == tokenEqual {
		name := p.advance()
		p.advance()
		value, err := p.expression()
		if err != nil {
			return nil, err
		}
		return assignmentStatement{name: name, value: value}, nil
	}
	return nil, sourceError(p.peek(), "expected 'let', assignment, 'print', 'repeat', or 'if'")
}

func (p *parser) block() ([]statement, error) {
	if _, err := p.consume(tokenLeftBrace, "expected '{'"); err != nil {
		return nil, err
	}
	body, err := p.statements(tokenRightBrace)
	if err != nil {
		return nil, err
	}
	if _, err := p.consume(tokenRightBrace, "expected '}'"); err != nil {
		return nil, err
	}
	return body, nil
}

func (p *parser) expression() (expression, error) { return p.or() }

func (p *parser) or() (expression, error) {
	expr, err := p.and()
	if err != nil {
		return nil, err
	}
	for p.match(tokenOr) {
		op := p.previous()
		right, err := p.and()
		if err != nil {
			return nil, err
		}
		expr = binaryExpression{left: expr, operator: op, right: right}
	}
	return expr, nil
}

func (p *parser) and() (expression, error) {
	expr, err := p.equality()
	if err != nil {
		return nil, err
	}
	for p.match(tokenAnd) {
		op := p.previous()
		right, err := p.equality()
		if err != nil {
			return nil, err
		}
		expr = binaryExpression{left: expr, operator: op, right: right}
	}
	return expr, nil
}

func (p *parser) equality() (expression, error) {
	expr, err := p.comparison()
	if err != nil {
		return nil, err
	}
	for p.match(tokenEqualEqual, tokenBangEqual) {
		op := p.previous()
		right, err := p.comparison()
		if err != nil {
			return nil, err
		}
		expr = binaryExpression{left: expr, operator: op, right: right}
	}
	return expr, nil
}

func (p *parser) comparison() (expression, error) {
	expr, err := p.term()
	if err != nil {
		return nil, err
	}
	for p.match(tokenLess, tokenLessEqual, tokenGreater, tokenGreaterEqual) {
		op := p.previous()
		right, err := p.term()
		if err != nil {
			return nil, err
		}
		expr = binaryExpression{left: expr, operator: op, right: right}
	}
	return expr, nil
}

func (p *parser) term() (expression, error) {
	expr, err := p.factor()
	if err != nil {
		return nil, err
	}
	for p.match(tokenPlus, tokenMinus) {
		op := p.previous()
		right, err := p.factor()
		if err != nil {
			return nil, err
		}
		expr = binaryExpression{left: expr, operator: op, right: right}
	}
	return expr, nil
}

func (p *parser) factor() (expression, error) {
	expr, err := p.unary()
	if err != nil {
		return nil, err
	}
	for p.match(tokenStar, tokenSlash, tokenPercent) {
		op := p.previous()
		right, err := p.unary()
		if err != nil {
			return nil, err
		}
		expr = binaryExpression{left: expr, operator: op, right: right}
	}
	return expr, nil
}

func (p *parser) unary() (expression, error) {
	if p.match(tokenBang, tokenMinus, tokenPlus) {
		op := p.previous()
		right, err := p.unary()
		if err != nil {
			return nil, err
		}
		return unaryExpression{operator: op, right: right}, nil
	}
	return p.primary()
}

func (p *parser) primary() (expression, error) {
	if p.match(tokenNumber) {
		tok := p.previous()
		number, err := strconv.ParseFloat(tok.text, 64)
		if err != nil {
			return nil, sourceError(tok, "invalid number %q", tok.text)
		}
		return literalExpression{value: numberValue(number)}, nil
	}
	if p.match(tokenString) {
		return literalExpression{value: stringValue(p.previous().text)}, nil
	}
	if p.match(tokenTrue) {
		return literalExpression{value: boolValue(true)}, nil
	}
	if p.match(tokenFalse) {
		return literalExpression{value: boolValue(false)}, nil
	}
	if p.match(tokenIdentifier) {
		return variableExpression{name: p.previous()}, nil
	}
	if p.match(tokenLeftParen) {
		expr, err := p.expression()
		if err != nil {
			return nil, err
		}
		if _, err := p.consume(tokenRightParen, "expected ')' after expression"); err != nil {
			return nil, err
		}
		return expr, nil
	}
	return nil, sourceError(p.peek(), "expected a number, string, boolean, variable, or '('")
}

func (p *parser) skipNewlines() {
	for p.match(tokenNewline) {
	}
}

func (p *parser) consume(kind tokenKind, message string) (token, error) {
	if p.check(kind) {
		return p.advance(), nil
	}
	return token{}, sourceError(p.peek(), message)
}

func (p *parser) match(kinds ...tokenKind) bool {
	if !p.check(kinds...) {
		return false
	}
	p.advance()
	return true
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

func (p *parser) peek() token     { return p.tokens[p.current] }
func (p *parser) previous() token { return p.tokens[p.current-1] }

func (p *parser) peekNext() token {
	if p.current+1 >= len(p.tokens) {
		return p.tokens[len(p.tokens)-1]
	}
	return p.tokens[p.current+1]
}

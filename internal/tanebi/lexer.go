package tanebi

import (
	"strings"
	"unicode"
)

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
		if ch == '\n' || ch == ';' {
			tokens = append(tokens, token{kind: tokenNewline, text: string(ch), line: l.line, column: l.column})
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
		if ch == '"' {
			tok, err := l.stringLiteral()
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, tok)
			continue
		}

		start := token{text: string(ch), line: l.line, column: l.column}
		l.advance()
		switch ch {
		case '+':
			start.kind = tokenPlus
		case '-':
			start.kind = tokenMinus
		case '*':
			start.kind = tokenStar
		case '/':
			start.kind = tokenSlash
		case '%':
			start.kind = tokenPercent
		case '(':
			start.kind = tokenLeftParen
		case ')':
			start.kind = tokenRightParen
		case '{':
			start.kind = tokenLeftBrace
		case '}':
			start.kind = tokenRightBrace
		case '=':
			start.kind, start.text = l.withOptionalEqual(tokenEqual, tokenEqualEqual, "=")
		case '!':
			start.kind, start.text = l.withOptionalEqual(tokenBang, tokenBangEqual, "!")
		case '<':
			start.kind, start.text = l.withOptionalEqual(tokenLess, tokenLessEqual, "<")
		case '>':
			start.kind, start.text = l.withOptionalEqual(tokenGreater, tokenGreaterEqual, ">")
		case '&':
			if l.atEnd() || l.peek() != '&' {
				return nil, sourceError(start, "expected '&' after '&'")
			}
			l.advance()
			start.kind, start.text = tokenAnd, "&&"
		case '|':
			if l.atEnd() || l.peek() != '|' {
				return nil, sourceError(start, "expected '|' after '|'")
			}
			l.advance()
			start.kind, start.text = tokenOr, "||"
		default:
			return nil, sourceError(start, "unknown character %q", ch)
		}
		tokens = append(tokens, start)
	}

	tokens = append(tokens, token{kind: tokenEOF, line: l.line, column: l.column})
	return tokens, nil
}

func (l *lexer) number() (token, error) {
	startIndex, line, column := l.index, l.line, l.column
	seenDot := false
	for !l.atEnd() {
		ch := l.peek()
		if unicode.IsDigit(ch) {
			l.advance()
			continue
		}
		if ch == '.' && !seenDot {
			seenDot = true
			l.advance()
			continue
		}
		break
	}
	text := string(l.source[startIndex:l.index])
	if strings.HasSuffix(text, ".") {
		return token{}, sourceError(token{line: line, column: column}, "a number cannot end with '.'")
	}
	return token{kind: tokenNumber, text: text, line: line, column: column}, nil
}

func (l *lexer) identifier() token {
	startIndex, line, column := l.index, l.line, l.column
	for !l.atEnd() {
		ch := l.peek()
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '_' {
			break
		}
		l.advance()
	}
	text := string(l.source[startIndex:l.index])
	kind := tokenIdentifier
	if keyword, ok := keywords[text]; ok {
		kind = keyword
	}
	return token{kind: kind, text: text, line: line, column: column}
}

func (l *lexer) stringLiteral() (token, error) {
	line, column := l.line, l.column
	l.advance()
	var value strings.Builder
	for !l.atEnd() && l.peek() != '"' {
		ch := l.advance()
		if ch == '\n' {
			return token{}, sourceError(token{line: line, column: column}, "unterminated string")
		}
		if ch != '\\' {
			value.WriteRune(ch)
			continue
		}
		if l.atEnd() {
			return token{}, sourceError(token{line: line, column: column}, "unterminated string")
		}
		escaped := l.advance()
		switch escaped {
		case 'n':
			value.WriteRune('\n')
		case 'r':
			value.WriteRune('\r')
		case 't':
			value.WriteRune('\t')
		case '"', '\\':
			value.WriteRune(escaped)
		default:
			return token{}, sourceError(token{line: l.line, column: l.column - 1}, "unknown escape sequence \\%c", escaped)
		}
	}
	if l.atEnd() {
		return token{}, sourceError(token{line: line, column: column}, "unterminated string")
	}
	l.advance()
	return token{kind: tokenString, text: value.String(), line: line, column: column}, nil
}

func (l *lexer) withOptionalEqual(single, doubled tokenKind, prefix string) (tokenKind, string) {
	if !l.atEnd() && l.peek() == '=' {
		l.advance()
		return doubled, prefix + "="
	}
	return single, prefix
}

func (l *lexer) atEnd() bool { return l.index >= len(l.source) }
func (l *lexer) peek() rune  { return l.source[l.index] }

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

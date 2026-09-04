package tanebi

type tokenKind int

const (
	tokenEOF tokenKind = iota
	tokenNewline
	tokenNumber
	tokenString
	tokenIdentifier
	tokenLet
	tokenPrint
	tokenRepeat
	tokenIf
	tokenElse
	tokenTrue
	tokenFalse
	tokenPlus
	tokenMinus
	tokenStar
	tokenSlash
	tokenPercent
	tokenLeftParen
	tokenRightParen
	tokenLeftBrace
	tokenRightBrace
	tokenEqual
	tokenEqualEqual
	tokenBang
	tokenBangEqual
	tokenLess
	tokenLessEqual
	tokenGreater
	tokenGreaterEqual
	tokenAnd
	tokenOr
)

type token struct {
	kind   tokenKind
	text   string
	line   int
	column int
}

var keywords = map[string]tokenKind{
	"let":    tokenLet,
	"print":  tokenPrint,
	"repeat": tokenRepeat,
	"if":     tokenIf,
	"else":   tokenElse,
	"true":   tokenTrue,
	"false":  tokenFalse,
}

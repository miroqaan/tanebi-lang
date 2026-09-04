package tanebi

import "fmt"

func sourceError(tok token, format string, args ...any) error {
	return fmt.Errorf("%d:%d: %s", tok.line, tok.column, fmt.Sprintf(format, args...))
}

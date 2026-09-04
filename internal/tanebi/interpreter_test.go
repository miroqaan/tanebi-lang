package tanebi

import (
	"bytes"
	"strings"
	"testing"
)

func outputOf(t *testing.T, source string) string {
	t.Helper()
	var output bytes.Buffer
	if err := Run(source, &output); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	return output.String()
}

func TestArithmeticAndUnicodeVariables(t *testing.T) {
	source := "let 가격 = 1200\nlet 수량 = 3\nprint 가격 * 수량 - 500\n"
	if got, want := outputOf(t, source), "3100\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestStringsRepeatAndJapaneseIdentifier(t *testing.T) {
	source := `let 名前 = "アカギ"
let 불꽃 = 3
repeat 불꽃 {
    print 名前 + "の世界が目を覚ます"
}
`
	want := strings.Repeat("アカギの世界が目を覚ます\n", 3)
	if got := outputOf(t, source); got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestAssignmentIfElseAndLogic(t *testing.T) {
	source := `let count = 0
repeat 4 {
    count = count + 1
}
if count == 4 && !(count < 0) {
    print "lit"
} else {
    print "dark"
}
`
	if got, want := outputOf(t, source), "lit\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestShortCircuitIsDeterministic(t *testing.T) {
	source := "if false && missing == 1 { print \"bad\" } else { print \"safe\" }\n"
	first := outputOf(t, source)
	second := outputOf(t, source)
	if first != "safe\n" || first != second {
		t.Fatalf("outputs = %q and %q", first, second)
	}
}

func TestIfWithoutElsePreservesFollowingStatement(t *testing.T) {
	source := "if true { print \"first\" }\nprint \"second\"\n"
	if got, want := outputOf(t, source), "first\nsecond\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestStringEscapes(t *testing.T) {
	if got, want := outputOf(t, `print "불씨\n種火"`), "불씨\n種火\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestSemicolonAndComments(t *testing.T) {
	source := "let x = 5; # spark\nprint x % 2;"
	if got, want := outputOf(t, source), "1\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestUsefulErrors(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{"undefined variable", "print missing\n", `1:7: undefined variable "missing"`},
		{"division by zero", "print 10 / 0\n", "1:10: division by zero"},
		{"repeat integer", "repeat 1.5 { print 1 }\n", "repeat count must be a non-negative integer"},
		{"if boolean", "if 1 { print 1 }\n", "if condition must be a boolean"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			err := Run(tt.source, &output)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want substring %q", err, tt.want)
			}
		})
	}
}

package main

import (
	"reflect"
	"strings"
	"testing"
)

func outputOf(t *testing.T, source string) []string {
	t.Helper()
	var output []string
	err := runSource(source, func(value string) {
		output = append(output, value)
	})
	if err != nil {
		t.Fatalf("runSource() error = %v", err)
	}
	return output
}

func TestVariablesAndPrecedence(t *testing.T) {
	got := outputOf(t, "let x = 2 + 3 * 4\nprint x\n")
	want := []string{"14"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("output = %v, want %v", got, want)
	}
}

func TestParenthesesAndDecimal(t *testing.T) {
	got := outputOf(t, "print (1 + 2) / 2\n")
	want := []string{"1.5"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("output = %v, want %v", got, want)
	}
}

func TestSemicolonsAndUnary(t *testing.T) {
	got := outputOf(t, "let x = -4; print x + 1;")
	want := []string{"-3"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("output = %v, want %v", got, want)
	}
}

func TestUndefinedVariable(t *testing.T) {
	err := runSource("print missing\n", func(string) {})
	if err == nil || !strings.Contains(err.Error(), "정의되지 않은 변수") {
		t.Fatalf("error = %v, want undefined-variable error", err)
	}
}

func TestDivisionByZero(t *testing.T) {
	err := runSource("print 10 / 0\n", func(string) {})
	if err == nil || !strings.Contains(err.Error(), "0으로 나눌 수 없습니다") {
		t.Fatalf("error = %v, want division-by-zero error", err)
	}
}

func TestUnicodeVariableName(t *testing.T) {
	got := outputOf(t, "let 가격 = 1200\nlet 수량 = 3\nprint 가격 * 수량\n")
	want := []string{"3600"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("output = %v, want %v", got, want)
	}
}

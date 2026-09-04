package tanebi

import (
	"fmt"
	"io"
	"math"
	"strconv"
)

type valueKind int

const (
	valueNumber valueKind = iota
	valueString
	valueBool
)

type value struct {
	kind    valueKind
	number  float64
	text    string
	boolean bool
}

func numberValue(v float64) value { return value{kind: valueNumber, number: v} }
func stringValue(v string) value  { return value{kind: valueString, text: v} }
func boolValue(v bool) value      { return value{kind: valueBool, boolean: v} }

type interpreter struct {
	variables map[string]value
	output    io.Writer
}

// Run tokenizes, parses, and executes a TANEBI program.
func Run(source string, output io.Writer) error {
	tokens, err := tokenize(source)
	if err != nil {
		return err
	}
	program, err := parse(tokens)
	if err != nil {
		return err
	}
	i := interpreter{variables: make(map[string]value), output: output}
	return i.execute(program)
}

func (i *interpreter) execute(statements []statement) error {
	for _, stmt := range statements {
		switch stmt := stmt.(type) {
		case letStatement:
			result, err := i.evaluate(stmt.value)
			if err != nil {
				return err
			}
			i.variables[stmt.name.text] = result
		case assignmentStatement:
			if _, exists := i.variables[stmt.name.text]; !exists {
				return sourceError(stmt.name, "undefined variable %q", stmt.name.text)
			}
			result, err := i.evaluate(stmt.value)
			if err != nil {
				return err
			}
			i.variables[stmt.name.text] = result
		case printStatement:
			result, err := i.evaluate(stmt.value)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintln(i.output, result.String()); err != nil {
				return fmt.Errorf("write output: %w", err)
			}
		case repeatStatement:
			count, err := i.evaluate(stmt.count)
			if err != nil {
				return err
			}
			if count.kind != valueNumber || count.number < 0 || math.Trunc(count.number) != count.number {
				return fmt.Errorf("repeat count must be a non-negative integer")
			}
			if count.number > 1_000_000 {
				return fmt.Errorf("repeat count exceeds 1000000")
			}
			for n := 0; n < int(count.number); n++ {
				if err := i.execute(stmt.body); err != nil {
					return err
				}
			}
		case ifStatement:
			condition, err := i.evaluate(stmt.condition)
			if err != nil {
				return err
			}
			if condition.kind != valueBool {
				return fmt.Errorf("if condition must be a boolean")
			}
			body := stmt.elseBody
			if condition.boolean {
				body = stmt.thenBody
			}
			if err := i.execute(body); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown statement")
		}
	}
	return nil
}

func (i *interpreter) evaluate(expr expression) (value, error) {
	switch expr := expr.(type) {
	case literalExpression:
		return expr.value, nil
	case variableExpression:
		result, ok := i.variables[expr.name.text]
		if !ok {
			return value{}, sourceError(expr.name, "undefined variable %q", expr.name.text)
		}
		return result, nil
	case unaryExpression:
		right, err := i.evaluate(expr.right)
		if err != nil {
			return value{}, err
		}
		switch expr.operator.kind {
		case tokenPlus:
			if right.kind != valueNumber {
				return value{}, sourceError(expr.operator, "unary '+' requires a number")
			}
			return right, nil
		case tokenMinus:
			if right.kind != valueNumber {
				return value{}, sourceError(expr.operator, "unary '-' requires a number")
			}
			return numberValue(-right.number), nil
		case tokenBang:
			if right.kind != valueBool {
				return value{}, sourceError(expr.operator, "'!' requires a boolean")
			}
			return boolValue(!right.boolean), nil
		}
	case binaryExpression:
		return i.evaluateBinary(expr)
	}
	return value{}, fmt.Errorf("unknown expression")
}

func (i *interpreter) evaluateBinary(expr binaryExpression) (value, error) {
	left, err := i.evaluate(expr.left)
	if err != nil {
		return value{}, err
	}

	if expr.operator.kind == tokenAnd || expr.operator.kind == tokenOr {
		if left.kind != valueBool {
			return value{}, sourceError(expr.operator, "logical operators require booleans")
		}
		if expr.operator.kind == tokenAnd && !left.boolean {
			return boolValue(false), nil
		}
		if expr.operator.kind == tokenOr && left.boolean {
			return boolValue(true), nil
		}
	}

	right, err := i.evaluate(expr.right)
	if err != nil {
		return value{}, err
	}
	switch expr.operator.kind {
	case tokenAnd, tokenOr:
		if right.kind != valueBool {
			return value{}, sourceError(expr.operator, "logical operators require booleans")
		}
		if expr.operator.kind == tokenAnd {
			return boolValue(left.boolean && right.boolean), nil
		}
		return boolValue(left.boolean || right.boolean), nil
	case tokenEqualEqual, tokenBangEqual:
		equal := valuesEqual(left, right)
		if expr.operator.kind == tokenBangEqual {
			equal = !equal
		}
		return boolValue(equal), nil
	case tokenPlus:
		if left.kind == valueNumber && right.kind == valueNumber {
			return numberValue(left.number + right.number), nil
		}
		if left.kind == valueString && right.kind == valueString {
			return stringValue(left.text + right.text), nil
		}
		return value{}, sourceError(expr.operator, "'+' requires two numbers or two strings")
	case tokenMinus, tokenStar, tokenSlash, tokenPercent:
		if left.kind != valueNumber || right.kind != valueNumber {
			return value{}, sourceError(expr.operator, "arithmetic operators require numbers")
		}
		switch expr.operator.kind {
		case tokenMinus:
			return numberValue(left.number - right.number), nil
		case tokenStar:
			return numberValue(left.number * right.number), nil
		case tokenSlash:
			if right.number == 0 {
				return value{}, sourceError(expr.operator, "division by zero")
			}
			return numberValue(left.number / right.number), nil
		default:
			if right.number == 0 {
				return value{}, sourceError(expr.operator, "division by zero")
			}
			return numberValue(math.Mod(left.number, right.number)), nil
		}
	case tokenLess, tokenLessEqual, tokenGreater, tokenGreaterEqual:
		return compareValues(left, right, expr.operator)
	}
	return value{}, sourceError(expr.operator, "unknown operator %q", expr.operator.text)
}

func valuesEqual(left, right value) bool {
	if left.kind != right.kind {
		return false
	}
	switch left.kind {
	case valueNumber:
		return left.number == right.number
	case valueString:
		return left.text == right.text
	case valueBool:
		return left.boolean == right.boolean
	default:
		return false
	}
}

func compareValues(left, right value, operator token) (value, error) {
	if left.kind != right.kind || (left.kind != valueNumber && left.kind != valueString) {
		return value{}, sourceError(operator, "comparison requires two numbers or two strings")
	}
	comparison := 0
	if left.kind == valueNumber {
		if left.number < right.number {
			comparison = -1
		} else if left.number > right.number {
			comparison = 1
		}
	} else {
		if left.text < right.text {
			comparison = -1
		} else if left.text > right.text {
			comparison = 1
		}
	}
	switch operator.kind {
	case tokenLess:
		return boolValue(comparison < 0), nil
	case tokenLessEqual:
		return boolValue(comparison <= 0), nil
	case tokenGreater:
		return boolValue(comparison > 0), nil
	default:
		return boolValue(comparison >= 0), nil
	}
}

func (v value) String() string {
	switch v.kind {
	case valueNumber:
		return strconv.FormatFloat(v.number, 'f', -1, 64)
	case valueString:
		return v.text
	case valueBool:
		return strconv.FormatBool(v.boolean)
	default:
		return ""
	}
}

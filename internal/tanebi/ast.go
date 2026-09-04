package tanebi

type expression interface{ expressionNode() }

type literalExpression struct{ value value }
type variableExpression struct{ name token }
type unaryExpression struct {
	operator token
	right    expression
}
type binaryExpression struct {
	left     expression
	operator token
	right    expression
}

func (literalExpression) expressionNode()  {}
func (variableExpression) expressionNode() {}
func (unaryExpression) expressionNode()    {}
func (binaryExpression) expressionNode()   {}

type statement interface{ statementNode() }

type letStatement struct {
	name  token
	value expression
}
type assignmentStatement struct {
	name  token
	value expression
}
type printStatement struct{ value expression }
type repeatStatement struct {
	count expression
	body  []statement
}
type ifStatement struct {
	condition expression
	thenBody  []statement
	elseBody  []statement
}

func (letStatement) statementNode()        {}
func (assignmentStatement) statementNode() {}
func (printStatement) statementNode()      {}
func (repeatStatement) statementNode()     {}
func (ifStatement) statementNode()         {}

package evaluator

import (
	"monkey/ast"
	"monkey/object"
	"monkey/token"
)

var (
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
	NULL  = &object.Null{}
)

func Eval(node ast.Node) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node.Statements)
	case *ast.BlockStatement:
		return evalBlockStatement(node.Statements)
	case *ast.ExpressionStatement:
		return Eval(node.Expression)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right)
		return evalPrefixExpression(node.Operator, right)
	case *ast.IfExpression:
		return evalIfExpression(node)
	case *ast.InfixExpression:
		left := Eval(node.Left)
		right := Eval(node.Right)
		return evalInfixExpression(left, right, node.Operator)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue)
		return &object.ReturnValue{Value: val}
	}

	return nil
}

func evalIfExpression(node *ast.IfExpression) object.Object {
	condition := Eval(node.Condition)
	var returnValue object.Object
	if isTruthy(condition) {
		returnValue = Eval(node.Consequence)
	} else if node.Alternative != nil {
		returnValue = Eval(node.Alternative)
	} else {
		return NULL
	}
	return returnValue
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func evalInfixExpression(left object.Object, right object.Object, operator string) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		leftValue := left.(*object.Integer)
		rightValue := right.(*object.Integer)
		return evalIntegerInfixExpression(leftValue, rightValue, operator)
	case operator == token.EQ:
		return nativeBoolToBooleanObject(left == right)
	case operator == token.NOT_EQ:
		return nativeBoolToBooleanObject(left != right)
	default:
		return NULL
	}
}

func evalIntegerInfixExpression(left *object.Integer, right *object.Integer, operator string) object.Object {
	switch operator {
	case token.PLUS:
		return &object.Integer{Value: left.Value + right.Value}
	case token.MINUS:
		return &object.Integer{Value: left.Value - right.Value}
	case token.ASTERISK:
		return &object.Integer{Value: left.Value * right.Value}
	case token.SLASH:
		return &object.Integer{Value: left.Value / right.Value}
	case token.EQ:
		return nativeBoolToBooleanObject(left.Value == right.Value)
	case token.NOT_EQ:
		return nativeBoolToBooleanObject(left.Value != right.Value)
	case token.LT:
		return nativeBoolToBooleanObject(left.Value < right.Value)
	case token.GT:
		return nativeBoolToBooleanObject(left.Value > right.Value)
	default:
		return NULL
	}
}

func evalPrefixExpression(operator string, object object.Object) object.Object {
	switch operator {
	case token.BANG:
		return evalBangOperator(object)
	case token.MINUS:
		return evalMinusPrefixOperator(object)
	default:
		return NULL
	}
}

func evalMinusPrefixOperator(o object.Object) object.Object {
	if o.Type() != object.INTEGER_OBJ {
		return NULL
	}

	return &object.Integer{
		Value: -o.(*object.Integer).Value,
	}
}

func evalBangOperator(o object.Object) object.Object {
	switch o {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func nativeBoolToBooleanObject(input bool) object.Object {
	if input {
		return TRUE
	}
	return FALSE
}

func evalBlockStatement(statements []ast.Statement) object.Object {
	var result object.Object

	for _, statement := range statements {
		result = Eval(statement)

		if result != nil && result.Type() == object.RETURN_VALUE_OBJ {
			return result
		}
	}

	return result
}

func evalProgram(statements []ast.Statement) object.Object {
	var result object.Object

	for _, statement := range statements {
		result = Eval(statement)

		if returnValue, ok := result.(*object.ReturnValue); ok {
			return returnValue.Value
		}
	}

	return result
}
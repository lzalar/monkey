package evaluator

import (
	"fmt"
	"monkey/ast"
	"monkey/object"
	"monkey/token"
)

var (
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
	NULL  = &object.Null{}
)

var builtins = map[string]*object.Builtin{
	"len": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *object.String:
				return &object.Integer{
					Value: int64(len(arg.Value)),
				}
			default:
				return newError("argument to `len` not supported, got %s", arg.Type())

			}
		},
	},
}

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node.Statements, env)
	case *ast.BlockStatement:
		return evalBlockStatement(node.Statements, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(left, right, node.Operator)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
	case *ast.Identifier:
		val, ok := env.Get(node.Value)
		if ok {
			return val
		}

		if builtin, ok := builtins[node.Value]; ok {
			return builtin
		}

		if !ok {
			return newError("identifier not found: %s", node.Value)
		}
		return val
	case *ast.FunctionLiteral:
		return &object.Function{
			Parameters: node.Parameters,
			Body:       node.Body,
			Env:        env,
		}
	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}

		args := evalExpressions(node.Arguments, env)

		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return applyFunction(function, args)
	case *ast.ArrayLiteral:
		elems := evalExpressions(node.Elements, env)

		if len(elems) == 1 && isError(elems[0]) {
			return elems[0]
		}

		return &object.Array{
			Elements: elems,
		}
	case *ast.IndexExpression:
		array := Eval(node.Left, env)
		if isError(array) {
			return array
		}

		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}

		return applyIndex(array, index)
	}

	return nil
}

func applyIndex(left object.Object, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)
	if idx < 0 || idx > max {
		return NULL
	}
	return arrayObject.Elements[idx]
}

func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		extendedEnv := extendFunctionEnvironment(fn, args)
		result := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(result)
	case *object.Builtin:
		return fn.Fn(args...)
	default:
		return newError("not a function: %s", fn.Type())
	}
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}

	return obj
}

func extendFunctionEnvironment(fn *object.Function, args []object.Object) *object.Environment {
	env := object.ExtendEnvironment(fn.Env)

	for i, parameter := range fn.Parameters {
		env.Set(parameter.Value, args[i])
	}

	return env
}

func evalExpressions(arguments []ast.Expression, env *object.Environment) []object.Object {
	results := []object.Object{}

	for _, argument := range arguments {
		result := Eval(argument, env)
		if isError(result) {
			return []object.Object{result}
		}
		results = append(results, result)
	}

	return results
}

func evalIfExpression(node *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(node.Condition, env)
	var returnValue object.Object
	if isTruthy(condition) {
		returnValue = Eval(node.Consequence, env)
	} else if node.Alternative != nil {
		returnValue = Eval(node.Alternative, env)
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
	case left.Type() != right.Type():
		return newError("type mismatch: %s + %s", left.Type(), right.Type())
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		leftValue := left.(*object.String)
		rightValue := right.(*object.String)
		return evalStringInfixExpression(leftValue, rightValue, operator)
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		leftValue := left.(*object.Integer)
		rightValue := right.(*object.Integer)
		return evalIntegerInfixExpression(leftValue, rightValue, operator)
	case operator == token.EQ:
		return nativeBoolToBooleanObject(left == right)
	case operator == token.NOT_EQ:
		return nativeBoolToBooleanObject(left != right)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(left *object.String, right *object.String, operator string) object.Object {
	switch operator {
	case token.PLUS:
		return &object.String{Value: left.Value + right.Value}
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
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
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalPrefixExpression(operator string, object object.Object) object.Object {
	switch operator {
	case token.BANG:
		return evalBangOperator(object)
	case token.MINUS:
		return evalMinusPrefixOperator(object)
	default:
		return newError("unknown operator: %s%s", operator, object.Type())
	}
}

func evalMinusPrefixOperator(o object.Object) object.Object {
	if o.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: %s%s", token.MINUS, o.Type())
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

func evalBlockStatement(statements []ast.Statement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range statements {
		result = Eval(statement, env)

		if result != nil {
			switch result.Type() {
			case object.RETURN_VALUE_OBJ, object.ERROR_OBJ:
				return result
			}
		}
	}

	return result
}

func evalProgram(statements []ast.Statement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range statements {
		result = Eval(statement, env)
		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func newError(format string, a ...any) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

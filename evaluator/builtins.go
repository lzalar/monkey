package evaluator

import "monkey/object"

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
			case *object.Array:
				return &object.Integer{
					Value: int64(len(arg.Elements)),
				}
			default:
				return newError("argument to `len` not supported, got %s", arg.Type())

			}
		},
	},
	"first": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *object.Array:
				if len(arg.Elements) > 0 {
					return arg.Elements[0]
				}
				return NULL
			default:
				return newError("argument to `first` must be ARRAY, got %s", arg.Type())

			}
		},
	},
	"last": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *object.Array:
				if len(arg.Elements) > 0 {
					return arg.Elements[len(arg.Elements)-1]
				}
				return NULL
			default:
				return newError("argument to `last` must be ARRAY, got %s", arg.Type())

			}
		},
	},
	"rest": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *object.Array:
				length := len(arg.Elements)
				if length > 0 {
					arr := make([]object.Object, length-1)
					copy(arr, arg.Elements[1:length])
					return &object.Array{Elements: arr}
				}
				return NULL
			default:
				return newError("argument to `rest` must be ARRAY, got %s", arg.Type())

			}
		},
	},
	"push": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}

			switch arg := args[0].(type) {
			case *object.Array:
				length := len(arg.Elements)

				arr := make([]object.Object, length, length+1)
				copy(arr, arg.Elements[:])
				arr = append(arr, args[1])
				return &object.Array{Elements: arr}
			default:
				return newError("argument to `push` must be ARRAY, got %s", arg.Type())
			}
		},
	},
}

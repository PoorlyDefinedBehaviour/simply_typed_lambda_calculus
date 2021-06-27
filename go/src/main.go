package main

import (
	"fmt"
	"log"

	"github.com/pkg/errors"
)

type Context map[string]interface{}

type Expression interface{}

type EInt int

type EVariable string

func (e *EVariable) String() string {
	return string(*e)
}

type EAbstraction struct {
	Parameter     string
	ParameterType Type
	Body          Expression
}

type EApplication struct {
	Function Expression
	Argument Expression
}

type Value interface{}

type VInt int

type VClosure struct {
	Context   Context
	Parameter string
	Body      Expression
}

type VNative func(Value) Value

type Type interface{}

type TInt struct{}

type TArrow struct {
	ParameterType Type
	BodyType      Type
}

func isTInt(typ Type) bool {
	_, ok := typ.(TInt)
	return ok
}

func typesEqual(a, b Type) bool {
	if a == b {
		return true
	}

	if isTInt(a) && isTInt(b) {
		return true
	}

	tArrowA, ok := a.(TArrow)
	if !ok {
		return false
	}

	tArrowB, ok := b.(TArrow)
	if !ok {
		return false
	}

	return typesEqual(tArrowA.ParameterType, tArrowB.ParameterType) &&
		typesEqual(tArrowA.BodyType, tArrowB.BodyType)
}

var ErrTypeError = errors.New("type error")

func Infer(context Context, expression Expression) (Type, error) {
	if _, ok := expression.(EInt); ok {
		return TInt{}, nil
	}

	if variable, ok := expression.(EVariable); ok {
		valueType, ok := context[string(variable)]
		if !ok {
			return nil, errors.WithStack(ErrTypeError)
		}

		return valueType, nil
	}

	if abstraction, ok := expression.(EAbstraction); ok {
		context[abstraction.Parameter] = abstraction.ParameterType
		bodyType, err := Infer(context, abstraction.Body)
		if err != nil {
			return nil, errors.WithStack(ErrTypeError)
		}
		return TArrow{
			ParameterType: abstraction.ParameterType,
			BodyType:      bodyType,
		}, nil
	}

	if application, ok := expression.(EApplication); ok {
		functionType, err := Infer(context, application.Function)
		if err != nil {
			return nil, errors.WithStack(ErrTypeError)
		}

		argumentType, err := Infer(context, application.Argument)
		if err != nil {
			return nil, errors.WithStack(ErrTypeError)
		}

		if tArrow, ok := functionType.(TArrow); ok {
			if typesEqual(tArrow.ParameterType, argumentType) {
				return tArrow.BodyType, nil
			}
		}
	}

	return nil, errors.WithStack(ErrTypeError)
}

func Interpret(context Context, expression Expression) (Value, error) {
	if value, ok := expression.(EInt); ok {
		return VInt(value), nil
	}

	if variable, ok := expression.(EVariable); ok {
		return context[string(variable)], nil
	}

	if abstraction, ok := expression.(EAbstraction); ok {
		return VClosure{
			Context:   context,
			Parameter: abstraction.Parameter,
			Body:      abstraction.Body,
		}, nil
	}

	if application, ok := expression.(EApplication); ok {
		argument, err := Interpret(context, application.Argument)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		function, err := Interpret(context, application.Function)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		if closure, ok := function.(VClosure); ok {
			closure.Context[closure.Parameter] = argument
			return Interpret(closure.Context, closure.Body)
		}

		if function, ok := function.(VNative); ok {
			return function(argument), nil
		}
	}

	return errors.WithStack(ErrTypeError), nil
}

func main() {
	typ, err := Infer(make(Context), EAbstraction{
		Parameter:     "a",
		ParameterType: TInt{},
		Body: EAbstraction{
			Parameter: "b",
			ParameterType: TArrow{
				ParameterType: TInt{},
				BodyType:      TInt{},
			},
			Body: EVariable("b"),
		},
	})
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%T\n", typ)
}

// Package evaluator takes s-expressions and runs it as a LISP program.
package evaluator

import (
	"fmt"

	"github.com/soishi1/toylisp/sexpressions"
)

var Nil = &Value{
	valueType: SExp,
	SExp: &sexpressions.SExp{
		Type:  sexpressions.ListType,
		Value: nil,
	},
}

type ValueType int

const (
	SExp ValueType = iota
	Lambda
	Primitive
)

type Value struct {
	*sexpressions.SExp
	valueType ValueType
	value     interface{}
}

type LambdaValue struct {
	args []string
	body []ast
	env  *Env
}

type PrimitiveFunc func(args []*Value) (*Value, error)

func (v *Value) String() string {
	switch v.valueType {
	case SExp:
		return v.SExp.String()
	case Lambda:
		return "#<lambda>"
	}
	return ""
}

type ast interface {
	Eval(e *Env) (*Value, error)
}

type literalAST struct {
	value *Value
}

func (a *literalAST) Eval(e *Env) (*Value, error) {
	return a.value, nil
}

type lookupAST struct {
	symbol string
}

func (a *lookupAST) Eval(e *Env) (*Value, error) {
	value, ok := e.Lookup(a.symbol)
	if !ok {
		return nil, fmt.Errorf("undefined variable %v", a.symbol)
	}
	return value, nil
}

type ifAST struct {
	condAST, thenAST, elseAST ast
}

func (a *ifAST) Eval(e *Env) (*Value, error) {
	condValue, err := a.condAST.Eval(e)
	if err != nil {
		return nil, err
	}
	if condValue.IsNil() {
		return a.elseAST.Eval(e)
	} else {
		return a.thenAST.Eval(e)
	}
}

type setAST struct {
	symbol   string
	valueAST ast
}

func (a *setAST) Eval(e *Env) (*Value, error) {
	value, err := a.valueAST.Eval(e)
	if err != nil {
		return nil, err
	}
	e.Set(a.symbol, value)
	return value, nil
}

type lambdaAST struct {
	symbols  []string
	bodyASTs []ast
}

func (a *lambdaAST) Eval(e *Env) (*Value, error) {
	return &Value{
		valueType: Lambda,
		value: &LambdaValue{
			args: a.symbols,
			body: a.bodyASTs,
			env:  newEnvWithParent(e),
		},
	}, nil
}

type applicationAST struct {
	funcAST ast
	argASTs []ast
}

func (a *applicationAST) Eval(e *Env) (*Value, error) {
	funcValue, err := a.funcAST.Eval(e)
	if err != nil {
		return nil, err
	}
	if funcValue.valueType == Lambda {
		lambda := funcValue.value.(*LambdaValue)
		return a.EvalLambdaApplication(e, lambda)
	}
	if funcValue.valueType == Primitive {
		primitive := funcValue.value.(PrimitiveFunc)
		return a.EvalPrimitiveApplication(e, primitive)
	}
	return nil, fmt.Errorf("Unsupported application function: %+v", funcValue)
}

func (a *applicationAST) EvalLambdaApplication(e *Env, lambda *LambdaValue) (*Value, error) {
	if len(lambda.args) != len(a.argASTs) {
		return nil, fmt.Errorf("%+v requires %v arguments, but got %v", lambda, len(lambda.args), len(a.argASTs))
	}
	applicationEnv := lambda.env
	for i := range lambda.args {
		arg, err := a.argASTs[i].Eval(e)
		if err != nil {
			return nil, err
		}
		applicationEnv.Set(lambda.args[i], arg)
	}
	var value *Value
	for i := range lambda.body {
		var err error
		value, err = lambda.body[i].Eval(applicationEnv)
		if err != nil {
			return nil, err
		}
	}
	return value, nil
}

func (a *applicationAST) EvalPrimitiveApplication(e *Env, primitive PrimitiveFunc) (*Value, error) {
	var args []*Value
	for i := range a.argASTs {
		arg, err := a.argASTs[i].Eval(e)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}
	return primitive(args)
}

type Env struct {
	vars   map[string]*Value
	parent *Env
}

// makeAST parses a s-expression and turn it into AST.
func makeAST(sexp *sexpressions.SExp) (ast, error) {
	switch sexp.Type {
	case sexpressions.StringType, sexpressions.IntType:
		return &literalAST{
			value: &Value{
				valueType: SExp,
				SExp:      sexp,
			},
		}, nil
	case sexpressions.ListType:
		list, _ := sexp.AsList()
		return makeASTFromList(list)
	case sexpressions.SymbolType:
		symbol, _ := sexp.AsSymbol()
		return &lookupAST{
			symbol: symbol,
		}, nil
	}
	return nil, fmt.Errorf("failed to evaluate %v (unknown sexpression type)", sexp)
}

func makeASTFromList(sexps []*sexpressions.SExp) (ast, error) {
	if len(sexps) == 0 {
		return &literalAST{value: Nil}, nil
	}

	first := sexps[0]
	if symbol, ok := first.AsSymbol(); ok {
		switch symbol {
		case "if":
			return makeIfAST(sexps)
		case "set":
			return makeSetAST(sexps)
		case "quote":
			return makeQuoteAST(sexps)
		case "lambda":
			return makeLambdaAST(sexps)
		}
	}
	return makeApplicationAST(sexps)
}

func makeIfAST(sexps []*sexpressions.SExp) (ast, error) {
	if len(sexps) != 3 && len(sexps) != 4 {
		return nil, fmt.Errorf("if requires 2 or 3 args: %+v", sexps)
	}

	var elseAST ast = &literalAST{value: Nil}
	if len(sexps) == 4 {
		var err error
		elseAST, err = makeAST(sexps[3])
		if err != nil {
			return nil, err
		}
	}

	thenAST, err := makeAST(sexps[2])
	if err != nil {
		return nil, err
	}

	condAST, err := makeAST(sexps[1])
	if err != nil {
		return nil, err
	}

	return &ifAST{
		condAST: condAST,
		thenAST: thenAST,
		elseAST: elseAST,
	}, nil
}

func makeSetAST(sexps []*sexpressions.SExp) (ast, error) {
	if len(sexps) != 3 {
		return nil, fmt.Errorf("set requires 2 args: %+v", sexps)
	}

	symbol, ok := sexps[1].AsSymbol()
	if !ok {
		return nil, fmt.Errorf("1st argument to set must be a symbol: %+v", sexps)
	}

	valueAST, err := makeAST(sexps[2])
	if err != nil {
		return nil, err
	}

	return &setAST{
		symbol:   symbol,
		valueAST: valueAST,
	}, nil
}

func makeQuoteAST(sexps []*sexpressions.SExp) (ast, error) {
	if len(sexps) != 2 {
		return nil, fmt.Errorf("quote requires 1 arg: %+v", sexps)
	}
	return &literalAST{
		value: &Value{
			valueType: SExp,
			SExp:      sexps[1],
		},
	}, nil
}

func makeLambdaAST(sexps []*sexpressions.SExp) (ast, error) {
	if len(sexps) < 3 {
		return nil, fmt.Errorf("lambda requires at least 2 arguments: %+v", sexps)
	}

	args, ok := sexps[1].AsList()
	if !ok {
		return nil, fmt.Errorf("1st argument to lambda must be a list of symbols: %+v", sexps)
	}
	var symbols []string
	for i := range args {
		symbol, ok := args[i].AsSymbol()
		if !ok {
			return nil, fmt.Errorf("1st argument to lambda must be a list of symbols: %+v", sexps)
		}
		symbols = append(symbols, symbol)
	}

	var bodyASTs []ast
	for i := 2; i < len(sexps); i++ {
		ast, err := makeAST(sexps[i])
		if err != nil {
			return nil, err
		}
		bodyASTs = append(bodyASTs, ast)
	}

	return &lambdaAST{
		symbols:  symbols,
		bodyASTs: bodyASTs,
	}, nil
}

func makeApplicationAST(sexps []*sexpressions.SExp) (ast, error) {
	if len(sexps) == 0 {
		return nil, fmt.Errorf("function application requires at least 1 argument: %+v", sexps)
	}

	funcAST, err := makeAST(sexps[0])
	if err != nil {
		return nil, err
	}

	var argASTs []ast
	for i := 1; i < len(sexps); i++ {
		argAST, err := makeAST(sexps[i])
		if err != nil {
			return nil, err
		}
		argASTs = append(argASTs, argAST)

	}

	return &applicationAST{
		funcAST: funcAST,
		argASTs: argASTs,
	}, nil
}

func NewEnv() *Env {
	return &Env{
		vars: map[string]*Value{
			"nil": Nil,
			"add": makePrimitive(func(args []*Value) (*Value, error) {
				sum := 0
				for i := range args {
					x, ok := args[i].AsInt()
					if !ok {
						return nil, fmt.Errorf("add argument[%v] is not int: %v", i, args[i])
					}
					sum += x
				}
				return &Value{
					valueType: SExp,
					SExp: &sexpressions.SExp{
						Type:  sexpressions.IntType,
						Value: sum,
					},
				}, nil
			}),
		},
		parent: nil,
	}
}

func makePrimitive(p PrimitiveFunc) *Value {
	return &Value{
		valueType: Primitive,
		value:     p,
	}
}

func newEnvWithParent(parent *Env) *Env {
	return &Env{
		vars:   make(map[string]*Value),
		parent: parent,
	}
}

func (e *Env) Lookup(symbol string) (result *Value, ok bool) {
	for cursor := e; cursor != nil; cursor = cursor.parent {
		value, ok := cursor.vars[symbol]
		if ok {
			return value, true
		}
	}
	return nil, false
}

func (e *Env) Set(symbol string, value *Value) {
	e.vars[symbol] = value
}

func (e *Env) String() string {
	if e.parent != nil {
		return fmt.Sprintf("%+v parent: %+v", e.vars, e.parent)
	}
	return fmt.Sprintf("%+v", e.vars)
}

func (e *Env) Eval(sexp *sexpressions.SExp) (result *Value, err error) {
	ast, err := makeAST(sexp)
	if err != nil {
		return nil, fmt.Errorf("makeAst(%v): %v", sexp, err)
	}
	return ast.Eval(e)
}

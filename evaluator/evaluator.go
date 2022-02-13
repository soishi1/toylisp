// Package evaluator takes syntactic tree and runs it as a LISP program.
package evaluator

import (
	"fmt"

	"github.com/soishi1/toylisp/sexpressions"
)

var Nil = &Value{
	valueType: SExp,
	value: &sexpressions.SExp{
		Type:  sexpressions.ListType,
		Value: nil,
	},
}

type ValueType int

const (
	SExp ValueType = iota
	Lambda
)

type Value struct {
	*sexpressions.SExp
	valueType ValueType
	value     interface{}
}

type LambdaValue struct {
	args []string
	body []*sexpressions.SExp
	env  *Env
}

func (v *Value) String() string {
	switch v.valueType {
	case SExp:
		return v.value.(*sexpressions.SExp).String()
	case Lambda:
		return "#<lambda>"
	}
	return ""
}

type Env struct {
	vars   map[string]*Value
	parent *Env
}

func NewEnv() *Env {
	return &Env{
		vars: map[string]*Value{
			"nil": Nil,
		},
		parent: nil,
	}
}

func newEnvWithParent(parent *Env) *Env {
	return &Env{
		vars:   make(map[string]*Value),
		parent: parent,
	}
}

func (e *Env) Lookup(symbol string) (result *Value, ok bool) {
	for cursor := e; e != nil; e = e.parent {
		value, ok := cursor.vars[symbol]
		if ok {
			return value, true
		}
	}
	return nil, false
}

func (e *Env) String() string {
	if e.parent != nil {
		return fmt.Sprintf("%+v parent: %+v", e.vars, e.parent)
	}
	return fmt.Sprintf("%+v", e.vars)
}

func (e *Env) Eval(sexp *sexpressions.SExp) (result *Value, err error) {
	switch sexp.Type {
	case sexpressions.StringType, sexpressions.IntType:
		return &Value{valueType: SExp, value: sexp}, nil
	case sexpressions.ListType:
		return Nil, nil
	case sexpressions.SymbolType:
		symbol, _ := sexp.AsSymbol()
		value, ok := e.Lookup(symbol)
		if !ok {
			return nil, fmt.Errorf("Undefined var %v, env = %v", symbol, e)
		}
		return value, nil
	}
	return nil, fmt.Errorf("failed to evaluate %v (unknown sexpression type)", sexp)
}

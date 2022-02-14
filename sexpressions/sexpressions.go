// Package sexpressions defines S-Expressions.
package sexpressions

import (
	"fmt"
	"strings"
)

type Type int

const (
	ListType = iota
	SymbolType
	IntType
	StringType
)

type SExp struct {
	Type  Type
	Value interface{}
}

func (s *SExp) AsList() (value []*SExp, ok bool) {
	if s.Type != ListType {
		return nil, false
	}
	if s.Value == nil {
		return nil, true
	}
	return s.Value.([]*SExp), true
}

func (s *SExp) AsSymbol() (value string, ok bool) {
	if s.Type != SymbolType {
		return "", false
	}
	return s.Value.(string), true
}

func (s *SExp) AsInt() (value int, ok bool) {
	if s.Type != IntType {
		return 0, false
	}
	return s.Value.(int), true
}

func (s *SExp) AsString() (value string, ok bool) {
	if s.Type != StringType {
		return "", false
	}
	return s.Value.(string), true
}

func (s *SExp) IsNil() bool {
	list, ok := s.AsList()
	return ok && len(list) == 0
}

func (s *SExp) String() string {
	if list, ok := s.AsList(); ok {
		var strs []string
		for i := range list {
			strs = append(strs, list[i].String())
		}
		return fmt.Sprintf("(%s)", strings.Join(strs, " "))
	} else if value, ok := s.AsInt(); ok {
		return fmt.Sprintf("%v", value)
	} else if value, ok := s.AsSymbol(); ok {
		return fmt.Sprintf("%v", value)
	} else if value, ok := s.AsString(); ok {
		return fmt.Sprintf("%q", value)
	} else {
		return fmt.Sprintf("%+v", value)
	}
}

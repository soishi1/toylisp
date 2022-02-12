// Parser turns list of tokens into S-expressions.
package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/soishi1/toylisp/tokenizer"
)

type SExpType int

const (
	ListSExpType = iota
	SymbolSExpType
	IntSExpType
	StringSExpType
)

type SExp struct {
	Type  SExpType
	Value interface{}
}

func (s *SExp) AsList() (value []*SExp, ok bool) {
	if s.Type != ListSExpType {
		return nil, false
	}
	return s.Value.([]*SExp), true
}

func (s *SExp) AsSymbol() (value string, ok bool) {
	if s.Type != SymbolSExpType {
		return "", false
	}
	return s.Value.(string), true
}

func (s *SExp) AsInt() (value int, ok bool) {
	if s.Type != IntSExpType {
		return 0, false
	}
	return s.Value.(int), true
}

func (s *SExp) AsString() (value string, ok bool) {
	if s.Type != StringSExpType {
		return "", false
	}
	return s.Value.(string), true
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

func Parse(tokens []*tokenizer.Token) ([]*SExp, error) {
	var result []*SExp
	rest := tokens
	needSpace := false
	for len(rest) > 0 {
		if needSpace {
			var err error
			rest, err = consume(tokenizer.Space, rest)
			if err != nil {
				return nil, err
			}
		}

		sexp, nextRest, err := parse1(rest)
		if err != nil {
			return nil, err
		}
		if sexp == nil && len(rest) != 0 {
			return nil, fmt.Errorf("Failed to parse: tokens: %v", rest)
		}
		if len(rest) == len(nextRest) {
			return nil, fmt.Errorf("parse1 didn't consume any token. tokens: %v", rest)
		}
		result = append(result, sexp)
		rest = nextRest
		needSpace = true
	}
	return result, nil
}

func parse1(tokens []*tokenizer.Token) (sexp *SExp, rest []*tokenizer.Token, err error) {
	firstToken := tokens[0]
	switch firstToken.Type {
	case tokenizer.OpenParen:
		return parseList(tokens)
	case tokenizer.Symbol:
		return &SExp{Type: SymbolSExpType, Value: firstToken.Str}, tokens[1:], nil
	case tokenizer.StringLiteral:
		// TODO(soishi): handle escaped characters.
		return &SExp{Type: StringSExpType, Value: firstToken.Str[1 : len(firstToken.Str)-1]}, tokens[1:], nil
	case tokenizer.NumberLiteral:
		// TODO(soishi): handle non integers.
		value, err := strconv.ParseInt(firstToken.Str, 10, 64)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse token %v as int", firstToken)
		}
		return &SExp{Type: IntSExpType, Value: int(value)}, tokens[1:], nil
	default:
		return nil, nil, fmt.Errorf("unexpected token at %v", tokens)
	}
}

func parseList(tokens []*tokenizer.Token) (sexp *SExp, rest []*tokenizer.Token, err error) {
	rest, err = consume(tokenizer.OpenParen, tokens)
	if err != nil {
		return nil, nil, err
	}
	needSpace := false

	var list []*SExp
	for len(rest) > 0 {
		var hasSpace bool
		rest, hasSpace = consumeIf(tokenizer.Space, rest)

		var ok bool
		rest, ok = consumeIf(tokenizer.CloseParen, rest)
		if ok {
			sexp := &SExp{
				Type:  ListSExpType,
				Value: list,
			}
			return sexp, rest, nil
		}

		if needSpace && !hasSpace {
			rest, err = consume(tokenizer.Space, rest)
			return nil, nil, err
		}

		sexp, nextRest, err := parse1(rest)
		if err != nil {
			return nil, nil, err
		}
		list = append(list, sexp)
		rest = nextRest
		needSpace = true
	}
	return nil, nil, fmt.Errorf("unmatched parens: tokens: %+v", tokens)
}

func consume(tokenType tokenizer.Type, tokens []*tokenizer.Token) (rest []*tokenizer.Token, err error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("unexpected end of tokens while expecting token %v", tokenType)
	}
	if got := tokens[0].Type; got != tokenType {
		return nil, fmt.Errorf("got unexpected token %v while expecting token %v at %v", got, tokenType, tokens)
	}
	return tokens[1:], nil
}

func consumeIf(tokenType tokenizer.Type, tokens []*tokenizer.Token) (rest []*tokenizer.Token, ok bool) {
	rest, err := consume(tokenType, tokens)
	if err != nil {
		return tokens, false
	}
	return tokens[1:], true
}

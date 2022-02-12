// Package tokenizer provides functions and definitions for splitting string into tokens.
package tokenizer

import (
	"fmt"
	"regexp"
)

// Type represents type of token (for example, space, or open paren).
type Type int

const (
	// Space represents whitespace, newlines, and so on.
	Space Type = iota
	// OpenParen represents '('.
	OpenParen
	// CloseParen represents '('.
	CloseParen
	// Symbol represents unquoted identifiers.
	Symbol
	// StringLiteral represents quoted strings.
	StringLiteral
	// NumberLiteral represents numbers (currently only supports decimal integers).
	NumberLiteral
)

// Token is one meaningful chunk of substring.
type Token struct {
	// Type tells which type this token is.
	Type Type
	// Str is the original substring that corresponds to this token.
	Str string
}

// String returns a description string of a token for debugging.
func (t *Token) String() string {
	return fmt.Sprintf("<%s>", t.Str)
}

// Tokenize splits s into tokens.
func Tokenize(s string) ([]*Token, error) {
	res := []*Token{}
	rest := s
	for len(rest) > 0 {
		t, nextRest, ok := tokenize1(rest)
		if !ok {
			return nil, fmt.Errorf("tokenize failed at %s", rest)
		}
		if len(nextRest) >= len(rest) {
			return nil, fmt.Errorf("tokenizers must consume at least 1 character: current head: %s", rest)
		}
		res = append(res, t)
		rest = nextRest
	}
	return res, nil
}

type tokenizer interface {
	Tokenize(s string) (t *Token, rest string, ok bool)
}

func tokenize1(s string) (t *Token, rest string, ok bool) {
	for _, tok := range subTokenizers {
		if t, rest, ok := tok.Tokenize(s); ok {
			return t, rest, ok
		}
	}
	return nil, s, false
}

var subTokenizers = []tokenizer{
	newRegexpTokenizer(Space, regexp.MustCompile(`\s+`)),
	newRegexpTokenizer(OpenParen, regexp.MustCompile(`\(`)),
	newRegexpTokenizer(CloseParen, regexp.MustCompile(`\)`)),
	newRegexpTokenizer(Symbol, regexp.MustCompile(`[a-zA-Z][a-zA-Z_]*`)),
	newRegexpTokenizer(StringLiteral, regexp.MustCompile(`"([^"\\]|\\"|\\\\)*"`)),
	newRegexpTokenizer(NumberLiteral, regexp.MustCompile(`[1-9][0-9]*`)),
}

type regexpTokenizer struct {
	tokenType Type
	re        *regexp.Regexp
}

func newRegexpTokenizer(tokenType Type, re *regexp.Regexp) *regexpTokenizer {
	return &regexpTokenizer{
		tokenType: tokenType,
		re:        re,
	}
}

func (rt *regexpTokenizer) Tokenize(s string) (t *Token, rest string, ok bool) {
	match := rt.re.FindIndex([]byte(s))
	if len(match) < 2 {
		return nil, s, false
	}
	start, end := match[0], match[1]
	if start != 0 {
		return nil, s, false
	}
	tok := &Token{
		Type: rt.tokenType,
		Str:  s[start:end],
	}
	return tok, s[end:], true
}

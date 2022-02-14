package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/soishi1/toylisp/evaluator"
	"github.com/soishi1/toylisp/parser"
	"github.com/soishi1/toylisp/tokenizer"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	env := evaluator.NewEnv()
	for scanner.Scan() {
		tokens, err := tokenizer.Tokenize(scanner.Text())
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(tokens)
		sexps, err := parser.Parse(tokens)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(sexps)
		for i := range sexps {
			value, err := env.Eval(sexps[i])
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println(value)
		}
	}
}

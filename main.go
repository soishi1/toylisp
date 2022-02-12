package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/soishi1/toylisp/parser"
	"github.com/soishi1/toylisp/tokenizer"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		tokens, err := tokenizer.Tokenize(scanner.Text())
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(tokens)
		sexps, err := parser.Parse(tokens)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(sexps)
	}
}

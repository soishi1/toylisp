package main

import (
	"fmt"

	"github.com/soishi1/toylisp/tokenizer"
)

func main() {
	fmt.Println(tokenizer.Tokenize(`(a(((((((abc(("""" )"a\"")))(10((1 a`))
}

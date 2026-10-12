package main

import (
	"os"
)

func main() {
	fileName := os.Args[1]
	input, err := os.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	tokens, err := tokenize(fileName, string(input))
	if err != nil {
		panic(err)
	}

	ast, err := parse(tokens)
	if err != nil {
		panic(err)
	}

	v := &codegenVisitor{}
	if err := walk(v, ast); err != nil {
		panic(err)
	}
}

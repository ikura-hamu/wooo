package main

import (
	"fmt"
	"io"
	"os"
)

func CompileBytes(w io.Writer, fileName string, data []byte, dumpAst bool) error {
	tokens, err := tokenize(fileName, string(data))
	if err != nil {
		return fmt.Errorf("tokenize: %w", err)
	}

	ast, err := parse(tokens)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	if dumpAst {
		dump(ast)
	}

	v := &codegenVisitor{w: w}
	if err := walk(v, ast); err != nil {
		return fmt.Errorf("codegen: %w", err)
	}

	return nil
}

func main() {
	fileName := os.Args[1]
	input, err := os.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	err = CompileBytes(os.Stdout, fileName, input, true)
	if err != nil {
		panic(err)
	}
}

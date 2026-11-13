package main

import (
	"fmt"
)

type codegenVisitor struct{}

func (v *codegenVisitor) visit(n node) (visitor, postVisit, error) {
	switch n := n.(type) {
	case *sourceFile:
		fmt.Printf("(module\n")
		return v, func() error {
			fmt.Printf(")\n")
			return nil
		}, nil
	case *functionDecl:
		fmt.Printf("(export %q (func $%s))\n", n.name.name, n.name.name)
		fmt.Printf("(func $%s ", n.name.name)
		if n.signature.result != nil {
			fmt.Printf("(result i32)\n") // TODO 正しい型
		} else {
			fmt.Println()
		}

		return v, func() error {
			fmt.Printf(")\n")
			return nil
		}, nil
	case *expression:
		return v, func() error {
			if n.isBinary {
				if n.binExpr.binaryOp.value == "+" {
					fmt.Printf("i32.add\n")
				}
				if n.binExpr.binaryOp.value == "-" {
					fmt.Printf("i32.sub\n")
				}
				if n.binExpr.binaryOp.value == "*" {
					fmt.Printf("i32.mul\n")
				}
				if n.binExpr.binaryOp.value == "/" {
					fmt.Printf("i32.div_s\n") // TODO: unsigned に対応
				}
			}
			return nil
		}, nil
	case *intLit:
		fmt.Printf("i32.const %d\n", n.n)
		return v, nil, nil
	default:
		return v, nil, nil
	}
}

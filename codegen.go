package main

import (
	"fmt"
	"strings"
)

type codegenVisitor struct {
	indentLevel int
}

func (v *codegenVisitor) visit(n node) (visitor, postVisit, error) {
	switch n := n.(type) {
	case *sourceFile:
		v.emitf("(module\n")
		return v, func() error {
			v.emitf(")\n")
			return nil
		}, nil
	case *functionDecl:
		v.incrementIndent()
		v.emitf("(export %q (func $%s))\n", n.name.name, n.name.name)
		v.decrementIndent()

		v.incrementIndent()
		v.emitf("(func $%s ", n.name.name)
		if n.signature.result != nil {
			v.emitf("(result i32)\n") // TODO 正しい型
		} else {
			fmt.Println()
		}

		return v, func() error {
			v.emitf(")\n")
			v.decrementIndent()
			return nil
		}, nil
	case *expression:
		if n.isBinary {
			v.incrementIndent()
			if n.binExpr.binaryOp.value == "+" {
				v.emitf("(i32.add\n")
			}
			if n.binExpr.binaryOp.value == "-" {
				v.emitf("(i32.sub\n")
			}
			if n.binExpr.binaryOp.value == "*" {
				v.emitf("(i32.mul\n")
			}
			if n.binExpr.binaryOp.value == "/" {
				v.emitf("(i32.div_s\n") // TODO: unsigned に対応
			}
		}
		return v, func() error {
			if n.isBinary {
				v.emitf(")\n")
				v.decrementIndent()
			}
			return nil
		}, nil
	case *intLit:
		v.incrementIndent()
		v.emitf("(i32.const %d)\n", n.n)
		v.decrementIndent()
		return v, nil, nil
	default:
		return v, nil, nil
	}
}

func (v *codegenVisitor) emitf(format string, args ...any) {
	fmt.Print(strings.Repeat("\t", v.indentLevel))
	fmt.Printf(format, args...)
}

func (v *codegenVisitor) incrementIndent() {
	v.indentLevel++
}

func (v *codegenVisitor) decrementIndent() {
	v.indentLevel--
}

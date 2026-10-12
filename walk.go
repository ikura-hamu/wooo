package main

import "fmt"

type postVisit func() error

type visitor interface {
	visit(n node) (visitor, postVisit, error)
}

func walk(v visitor, n node) error {
	if nextV, postVisit, err := v.visit(n); err != nil {
		return fmt.Errorf("visit: %w", err)
	} else if nextV == nil {
		return nil
	} else {
		v = nextV
		if postVisit != nil {
			defer postVisit()
		}
	}

	switch n := n.(type) {
	case *sourceFile:
		for _, topLevelDecl := range n.toplevelDecls {
			if err := walk(v, topLevelDecl); err != nil {
				return fmt.Errorf("walk topLevelDecl: %w", err)
			}
		}
	case *topLevelDecl:
		err := walk(v, n.functionDecl)
		if err != nil {
			return fmt.Errorf("walk functionDecl: %w", err)
		}
	case *functionDecl:
		if err := walk(v, n.name); err != nil {
			return fmt.Errorf("walk functionDecl name: %w", err)
		}
		if n.body != nil {
			if err := walk(v, n.body); err != nil {
				return fmt.Errorf("walk functionDecl body: %w", err)
			}
		}
	case *block:
		if err := walk(v, n.statementList); err != nil {
			return fmt.Errorf("walk block statementList: %w", err)
		}
	case *statementList:
		for _, statement := range n.statements {
			if err := walk(v, statement); err != nil {
				return fmt.Errorf("walk statementList statement: %w", err)
			}
		}
	case *statement:
		if n.statementKind == STATEMENT_RETURN {
			if err := walk(v, n.returnStatement); err != nil {
				return fmt.Errorf("walk statement returnStatement: %w", err)
			}
		}
	case *returnStatement:
		if n.expressions != nil {
			if err := walk(v, n.expressions); err != nil {
				return fmt.Errorf("walk returnStatement expressions: %w", err)
			}
		}
	case *expressionList:
		if err := walk(v, n.expression); err != nil {
			return fmt.Errorf("walk expressionList expression: %w", err)
		}
	case *expression:
		if err := walk(v, n.unaryExpr); err != nil {
			return fmt.Errorf("walk expression unaryExpr: %w", err)
		}
	case *unaryExpr:
		if err := walk(v, n.primaryExpr); err != nil {
			return fmt.Errorf("walk unaryExpr primaryExpr: %w", err)
		}
	case *primaryExpr:
		if err := walk(v, n.operand); err != nil {
			return fmt.Errorf("walk primaryExpr operand: %w", err)
		}
	case *operand:
		if err := walk(v, n.literal); err != nil {
			return fmt.Errorf("walk operand literal: %w", err)
		}
	case *literal:
		if err := walk(v, n.basicLit); err != nil {
			return fmt.Errorf("walk literal basicLit: %w", err)
		}
	case *basicLit:
		if err := walk(v, n.intLit); err != nil {
			return fmt.Errorf("walk basicLit intLit: %w", err)
		}
	case *intLit:
		return nil
	case *identifier:
		return nil
	default:
		return fmt.Errorf("walk: unexpected node type %T", n)
	}

	return nil
}

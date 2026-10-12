package main

import (
	"fmt"
	"strconv"
)

type node interface {
	_node()
}

type parser struct {
	tokens []*token
	cursor int
}

func (p *parser) currentToken() *token {
	return p.tokens[p.cursor]
}

func (p *parser) isOpPunc(op string) bool {
	return p.currentToken().kind == TK_OPERETORS_PUNCTUATION && p.currentToken().value == op
}

func (p *parser) isCurrentTokenKind(kind tokenKind) bool {
	return p.currentToken().kind == kind
}

func (p *parser) expectOpPunc(op string) error {
	if p.currentToken().kind != TK_OPERETORS_PUNCTUATION || p.currentToken().value != op {
		return fmt.Errorf("expected %q, got %q at cursor %d", op, p.currentToken().value, p.cursor)
	}
	p.cursor++
	return nil
}

func (p *parser) expectCurrentToken(kind tokenKind) error {
	if p.currentToken().kind != kind {
		return fmt.Errorf("expected token of kind %v at position %d", kind, p.cursor)
	}
	p.cursor++
	return nil
}

func parse(tokens []*token) (*sourceFile, error) {
	p := &parser{
		tokens: tokens,
		cursor: 0,
	}
	return p.sourceFile()
}

type sourceFile struct {
	toplevelDecls []*topLevelDecl
	node
}

func (p *parser) sourceFile() (*sourceFile, error) {
	toplevelDecls := make([]*topLevelDecl, 0)
	for p.currentToken().kind != TK_EOF {
		decl, err := p.topLevelDecl()
		if err != nil {
			return nil, fmt.Errorf("topLevelDecl: %w", err)
		}
		toplevelDecls = append(toplevelDecls, decl)

		if err := p.expectOpPunc(";"); err != nil {
			return nil, fmt.Errorf("expect ';': %w", err)
		}
	}

	return &sourceFile{toplevelDecls: toplevelDecls}, nil
}

type topLevelDecl struct {
	functionDecl *functionDecl
	node
}

func (p *parser) topLevelDecl() (*topLevelDecl, error) {
	if p.currentToken().kind == TK_FUNC {
		funcDeclNode, err := p.functionDecl()
		if err != nil {
			return nil, fmt.Errorf("functionDecl: %w", err)
		}
		return &topLevelDecl{functionDecl: funcDeclNode}, nil
	}

	return nil, fmt.Errorf("unexpected token")
}

type functionDecl struct {
	name      *identifier
	signature *signature
	body      *block
	node
}

func (p *parser) functionDecl() (*functionDecl, error) {
	if p.currentToken().kind != TK_FUNC {
		return nil, fmt.Errorf("expected function declaration")
	}
	p.cursor++

	name, err := p.identifier()
	if err != nil {
		return nil, fmt.Errorf("expected ident: %w", err)
	}

	signature, err := p.signature()
	if err != nil {
		return nil, fmt.Errorf("expected signature: %w", err)
	}

	if ok := p.isOpPunc("{"); !ok {
		return &functionDecl{name: name, signature: signature}, nil
	}

	block, err := p.block()
	if err != nil {
		return nil, err
	}

	return &functionDecl{name: name, signature: signature, body: block}, nil
}

type signature struct {
	parameters *parameters
	result     *result
	node
}

func (p *parser) signature() (*signature, error) {
	parameters, err := p.parameters()
	if err != nil {
		return nil, fmt.Errorf("parameters: %w", err)
	}

	var result *result
	if p.isOpPunc("(") || p.isCurrentTokenKind(TK_IDENT) {
		result, err = p.result()
		if err != nil {
			return nil, fmt.Errorf("result: %w", err)
		}
	}

	return &signature{parameters: parameters, result: result}, nil
}

type parameters struct {
	node
}

func (p *parser) parameters() (*parameters, error) {
	p.cursor += 2
	return &parameters{}, nil
}

type resultKind int

const (
	RESULT_PARAMETERS resultKind = iota + 1
	RESULT_TYPE
)

type result struct {
	kind       resultKind
	parameters *parameters
	typ        *typeNode
}

func (p *parser) result() (*result, error) {
	if p.isOpPunc("(") {
		parameters, err := p.parameters()
		if err != nil {
			return nil, fmt.Errorf("parameters: %w", err)
		}
		return &result{kind: RESULT_PARAMETERS, parameters: parameters}, nil
	}

	typ, err := p.typeNode()
	if err != nil {
		return nil, fmt.Errorf("typeNode: %w", err)
	}
	return &result{kind: RESULT_TYPE, typ: typ}, nil
}

type typeNode struct {
	typeName *typeName
}

func (p *parser) typeNode() (*typeNode, error) {
	typeName, err := p.typeName()
	if err != nil {
		return nil, fmt.Errorf("expected type name: %w", err)
	}
	return &typeNode{typeName: typeName}, nil
}

type typeName struct {
	identifier *identifier
}

func (p *parser) typeName() (*typeName, error) {
	identifier, err := p.identifier()
	if err != nil {
		return nil, fmt.Errorf("expected identifier: %w", err)
	}
	return &typeName{identifier: identifier}, nil
}

type block struct {
	statementList *statementList
	node
}

func (p *parser) block() (*block, error) {
	if err := p.expectOpPunc("{"); err != nil {
		return nil, fmt.Errorf("expected '{': %w", err)
	}

	statementList, err := p.statementList()
	if err != nil {
		return nil, fmt.Errorf("statementList: %w", err)
	}

	if err := p.expectOpPunc("}"); err != nil {
		return nil, fmt.Errorf("expected '}': %w", err)
	}

	return &block{statementList: statementList}, nil
}

type statementList struct {
	statements []*statement
	node
}

func (p *parser) statementList() (*statementList, error) {
	statements := make([]*statement, 0)
	for !p.isOpPunc("}") {
		statement, err := p.statement()
		if err != nil {
			return nil, fmt.Errorf("statement: %w", err)
		}
		if err := p.expectOpPunc(";"); err != nil {
			return nil, fmt.Errorf("expected ';': %w", err)
		}
		statements = append(statements, statement)
	}

	return &statementList{statements: statements}, nil
}

func (p *parser) statement() (*statement, error) {
	if p.currentToken().kind == TK_RETURN {
		returnStatement, err := p.returnStatement()
		if err != nil {
			return nil, fmt.Errorf("returnStatement: %w", err)
		}
		return &statement{
			statementKind:   STATEMENT_RETURN,
			returnStatement: returnStatement,
		}, nil
	}

	return nil, fmt.Errorf("unknown statement")
}

type statementKind int

const (
	STATEMENT_RETURN = iota + 1
)

type statement struct {
	statementKind   statementKind
	returnStatement *returnStatement
	node
}

type returnStatement struct {
	expressions *expressionList
	node
}

func (p *parser) returnStatement() (*returnStatement, error) {
	if err := p.expectCurrentToken(TK_RETURN); err != nil {
		return nil, fmt.Errorf("`return` expected: %w", err)
	}

	if p.isOpPunc(";") {
		return &returnStatement{}, nil
	}

	expressions, err := p.expressionList()
	if err != nil {
		return nil, fmt.Errorf("expressionList: %w", err)
	}

	return &returnStatement{expressions: expressions}, nil
}

type expressionList struct {
	expression *expression
	node
}

func (p *parser) expressionList() (*expressionList, error) {
	exp, err := p.expression()
	if err != nil {
		return nil, fmt.Errorf("expression: %w", err)
	}
	return &expressionList{expression: exp}, nil
}

type expression struct {
	unaryExpr *unaryExpr
	node
}

func (p *parser) expression() (*expression, error) {
	unaryExpr, err := p.unaryExpr()
	if err != nil {
		return nil, fmt.Errorf("unaryExpr: %w", err)
	}
	return &expression{unaryExpr: unaryExpr}, nil
}

type unaryExpr struct {
	primaryExpr *primaryExpr
	node
}

func (p *parser) unaryExpr() (*unaryExpr, error) {
	primaryExpr, err := p.primaryExpr()
	if err != nil {
		return nil, fmt.Errorf("primaryExpr: %w", err)
	}
	return &unaryExpr{primaryExpr: primaryExpr}, nil
}

type primaryExpr struct {
	operand *operand
	node
}

func (p *parser) primaryExpr() (*primaryExpr, error) {
	operand, err := p.operand()
	if err != nil {
		return nil, fmt.Errorf("operand: %w", err)
	}
	return &primaryExpr{operand: operand}, nil
}

type operand struct {
	literal *literal
	node
}

func (p *parser) operand() (*operand, error) {
	literal, err := p.literal()
	if err != nil {
		return nil, fmt.Errorf("literal: %w", err)
	}
	return &operand{literal: literal}, nil
}

type literal struct {
	basicLit *basicLit
	node
}

func (p *parser) literal() (*literal, error) {
	bl, err := p.basicLit()
	if err != nil {
		return nil, fmt.Errorf("basicLit: %w", err)
	}
	return &literal{basicLit: bl}, nil
}

type basicLit struct {
	intLit *intLit
	node
}

func (p *parser) basicLit() (*basicLit, error) {
	intL, err := p.intLit()
	if err != nil {
		return nil, err
	}
	return &basicLit{intLit: intL}, nil
}

type intLit struct {
	n int
	node
}

func (p *parser) intLit() (*intLit, error) {
	n, err := strconv.Atoi(p.currentToken().value)
	if err != nil {
		return nil, fmt.Errorf("parse int: %w", err)
	}
	p.cursor++
	return &intLit{n: n}, nil
}

type identifier struct {
	name string
	node
}

func (p *parser) identifier() (*identifier, error) {
	if ok := p.isCurrentTokenKind(TK_IDENT); !ok {
		return nil, fmt.Errorf("expected identifier, got %v", p.currentToken().kind)
	}
	name := p.currentToken().value
	p.cursor++
	return &identifier{name: name}, nil
}

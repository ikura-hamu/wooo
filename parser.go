package main

import (
	"fmt"
	"strconv"
)

type node interface {
	_node()
	id() int
}

type parser struct {
	tokens     []*token
	cursor     int
	lastNodeID int
}

type nodeData struct {
	nodeID int
	node
}

func (n *nodeData) id() int {
	return n.nodeID
}

func (p *parser) nodeData() *nodeData {
	p.lastNodeID++
	return &nodeData{nodeID: p.lastNodeID}
}

func (p *parser) consumeToken() *token {
	token := p.currentToken()
	p.cursor++
	return token
}

func (p *parser) currentToken() *token {
	return p.tokens[p.cursor]
}

func (p *parser) currentTokens(n int) []*token {
	return p.tokens[p.cursor : p.cursor+n]
}

func (p *parser) isPunctuation(op string) bool {
	return p.currentToken().kind == TK_PUNCTUATION && p.currentToken().value == op
}

func (p *parser) isCurrentTokenKind(kind tokenKind) bool {
	return p.currentToken().kind == kind
}

func (p *parser) expectPunctuation(op string) error {
	if p.currentToken().kind != TK_PUNCTUATION || p.currentToken().value != op {
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
	packageClause *packageClause
	toplevelDecls []*topLevelDecl
	node
}

func (p *parser) sourceFile() (*sourceFile, error) {
	packageClause, err := p.packageClause()
	if err != nil {
		return nil, fmt.Errorf("packageClause: %w", err)
	}

	if err := p.expectPunctuation(";"); err != nil {
		return nil, fmt.Errorf("expect ';': %w", err)
	}

	toplevelDecls := make([]*topLevelDecl, 0)
	for p.currentToken().kind != TK_EOF {
		decl, err := p.topLevelDecl()
		if err != nil {
			return nil, fmt.Errorf("topLevelDecl: %w", err)
		}
		toplevelDecls = append(toplevelDecls, decl)

		if err := p.expectPunctuation(";"); err != nil {
			return nil, fmt.Errorf("expect ';': %w", err)
		}
	}

	return &sourceFile{
		packageClause: packageClause,
		toplevelDecls: toplevelDecls,
		node:          p.nodeData(),
	}, nil
}

type declaration struct {
	varDecl *varDecl
	node
}

func (p *parser) declaration() (*declaration, error) {
	if err := p.expectCurrentToken(TK_VAR); err != nil {
		return nil, fmt.Errorf("expected var token: %w", err)
	}

	varDecl, err := p.varDecl()
	if err != nil {
		return nil, fmt.Errorf("varDecl: %w", err)
	}
	return &declaration{varDecl: varDecl, node: p.nodeData()}, nil
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
		return &topLevelDecl{functionDecl: funcDeclNode, node: p.nodeData()}, nil
	}

	return nil, fmt.Errorf("unexpected token")
}

type identifierList struct {
	identifier *identifier
	node
}

func (p *parser) identifierList() (*identifierList, error) {
	identifier, err := p.identifier()
	if err != nil {
		return nil, fmt.Errorf("expected identifier: %w", err)
	}

	return &identifierList{identifier: identifier, node: p.nodeData()}, nil
}

type varDecl struct {
	varSpec *varSpec
	node
}

func (p *parser) varDecl() (*varDecl, error) {
	varSpec, err := p.varSpec()
	if err != nil {
		return nil, fmt.Errorf("expected var spec: %w", err)
	}
	return &varDecl{varSpec: varSpec, node: p.nodeData()}, nil
}

type varSpec struct {
	identifierList *identifierList
	typeNode       *typeNode
	node
}

func (p *parser) varSpec() (*varSpec, error) {
	identifierList, err := p.identifierList()
	if err != nil {
		return nil, fmt.Errorf("expected identifier list: %w", err)
	}

	typeNode, err := p.typeNode()
	if err != nil {
		return nil, fmt.Errorf("expected type node: %w", err)
	}

	return &varSpec{
		identifierList: identifierList,
		typeNode:       typeNode,
		node:           p.nodeData(),
	}, nil
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

	if ok := p.isPunctuation("{"); !ok {
		return &functionDecl{name: name, signature: signature, node: p.nodeData()}, nil
	}

	block, err := p.block()
	if err != nil {
		return nil, err
	}

	return &functionDecl{name: name, signature: signature, body: block, node: p.nodeData()}, nil
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
	if p.isPunctuation("(") || p.isCurrentTokenKind(TK_IDENT) {
		result, err = p.result()
		if err != nil {
			return nil, fmt.Errorf("result: %w", err)
		}
	}

	return &signature{parameters: parameters, result: result, node: p.nodeData()}, nil
}

type parameters struct {
	node
}

func (p *parser) parameters() (*parameters, error) {
	p.cursor += 2
	return &parameters{node: p.nodeData()}, nil
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
	node
}

func (p *parser) result() (*result, error) {
	if p.isPunctuation("(") {
		parameters, err := p.parameters()
		if err != nil {
			return nil, fmt.Errorf("parameters: %w", err)
		}
		return &result{kind: RESULT_PARAMETERS, parameters: parameters, node: p.nodeData()}, nil
	}

	typ, err := p.typeNode()
	if err != nil {
		return nil, fmt.Errorf("typeNode: %w", err)
	}
	return &result{kind: RESULT_TYPE, typ: typ, node: p.nodeData()}, nil
}

type typeNode struct {
	typeName *typeName
	node
}

func (p *parser) typeNode() (*typeNode, error) {
	typeName, err := p.typeName()
	if err != nil {
		return nil, fmt.Errorf("expected type name: %w", err)
	}
	return &typeNode{typeName: typeName, node: p.nodeData()}, nil
}

type typeName struct {
	identifier *identifier
	node
}

func (p *parser) typeName() (*typeName, error) {
	identifier, err := p.identifier()
	if err != nil {
		return nil, fmt.Errorf("expected identifier: %w", err)
	}
	return &typeName{identifier: identifier, node: p.nodeData()}, nil
}

type block struct {
	statementList *statementList
	node
}

func (p *parser) block() (*block, error) {
	if err := p.expectPunctuation("{"); err != nil {
		return nil, fmt.Errorf("expected '{': %w", err)
	}

	statementList, err := p.statementList()
	if err != nil {
		return nil, fmt.Errorf("statementList: %w", err)
	}

	if err := p.expectPunctuation("}"); err != nil {
		return nil, fmt.Errorf("expected '}': %w", err)
	}

	return &block{statementList: statementList, node: p.nodeData()}, nil
}

type statementList struct {
	statements []*statement
	node
}

func (p *parser) statementList() (*statementList, error) {
	statements := make([]*statement, 0)
	for !p.isPunctuation("}") {
		statement, err := p.statement()
		if err != nil {
			return nil, fmt.Errorf("statement: %w", err)
		}
		if err := p.expectPunctuation(";"); err != nil {
			return nil, fmt.Errorf("expected ';': %w", err)
		}
		statements = append(statements, statement)
	}

	return &statementList{statements: statements, node: p.nodeData()}, nil
}

func (p *parser) statement() (*statement, error) {
	if p.currentToken().kind == TK_VAR {
		declaration, err := p.declaration()
		if err != nil {
			return nil, fmt.Errorf("declaration: %w", err)
		}
		return &statement{
			statementKind: STATEMENT_DECLARATION,
			declaration:   declaration,
			node:          p.nodeData(),
		}, nil
	}
	if p.currentToken().kind == TK_RETURN {
		returnStatement, err := p.returnStatement()
		if err != nil {
			return nil, fmt.Errorf("returnStatement: %w", err)
		}
		return &statement{
			statementKind:   STATEMENT_RETURN,
			returnStatement: returnStatement,
			node:            p.nodeData(),
		}, nil
	}

	return nil, fmt.Errorf("unknown statement")
}

type statementKind int

const (
	STATEMENT_DECLARATION statementKind = iota + 1
	STATEMENT_RETURN
)

type statement struct {
	statementKind   statementKind
	declaration     *declaration
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

	if p.isPunctuation(";") {
		return &returnStatement{node: p.nodeData()}, nil
	}

	expressions, err := p.expressionList()
	if err != nil {
		return nil, fmt.Errorf("expressionList: %w", err)
	}

	return &returnStatement{expressions: expressions, node: p.nodeData()}, nil
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
	return &expressionList{expression: exp, node: p.nodeData()}, nil
}

type expressionKind int

const (
	EXPRESSION_KIND_UNARY_EXPRESSION expressionKind = iota + 1
	EXPRESSION_KIND_BINARY_OP_EXPRESSION
)

type expression struct {
	isBinary  bool
	unaryExpr *unaryExpr
	binExpr   *binExpr
	node
}

type binExpr struct {
	lhs      *expression
	binaryOp *token
	rhs      *expression
}

func (p *parser) expression() (*expression, error) {
	lhs, err := p.unaryExpr()
	if err != nil {
		return nil, fmt.Errorf("unary expr: %w", err)
	}

	e := &expression{
		unaryExpr: lhs,
		node:      p.nodeData(),
	}

	for p.isBinOperator() {
		binaryOp := p.consumeToken()

		rhs, err := p.unaryExpr()
		if err != nil {
			return nil, fmt.Errorf("unary expr: %w", err)
		}

		if !e.isBinary {
			e.binExpr = &binExpr{
				lhs:      &expression{unaryExpr: lhs, node: p.nodeData()},
				binaryOp: binaryOp,
				rhs:      &expression{unaryExpr: rhs, node: p.nodeData()},
			}
			e.unaryExpr = nil
		} else {
			newBinOpStrength := binOpStrength[binaryOp.value]
			eBinOpStrength := binOpStrength[e.binExpr.binaryOp.value]

			if eBinOpStrength >= newBinOpStrength {
				e.binExpr = &binExpr{
					lhs: &expression{
						isBinary: true,
						binExpr: &binExpr{
							lhs:      e.binExpr.lhs,
							binaryOp: e.binExpr.binaryOp,
							rhs:      e.binExpr.rhs,
						},
						node: p.nodeData(),
					},
					binaryOp: binaryOp,
					rhs:      &expression{unaryExpr: rhs, node: p.nodeData()},
				}
			} else {
				e.binExpr = &binExpr{
					lhs:      e.binExpr.lhs,
					binaryOp: e.binExpr.binaryOp,
					rhs: &expression{
						isBinary: true,
						binExpr: &binExpr{
							lhs:      e.binExpr.rhs,
							binaryOp: binaryOp,
							rhs:      &expression{unaryExpr: rhs, node: p.nodeData()},
						},
						node: p.nodeData(),
					},
				}
			}
			e.unaryExpr = nil
		}
		e.isBinary = true
	}

	return e, nil
}

var binOpStrength map[string]int = map[string]int{
	"+": 4,
	"-": 4,
	"*": 5,
	"/": 5,
}

func (p *parser) isBinOperator() bool {
	if !p.isCurrentTokenKind(TK_OPERATOR) {
		return false
	}
	v := p.currentToken().value

	switch v {
	case "+", "-", "*", "/":
		return true
	default:
		return false
	}
}

func (p *parser) isUnaryOperator() bool {
	if !p.isCurrentTokenKind(TK_OPERATOR) {
		return false
	}
	v := p.currentToken().value

	switch v {
	case "+", "-":
		return true
	default:
		return false
	}
}

type unaryExprKind int

const (
	UNARY_EXPR_KIND_PRIMARY unaryExprKind = iota
	UNARY_EXPR_KIND_WITH_UNARY_OP
)

type unaryExpr struct {
	kind        unaryExprKind
	primaryExpr *primaryExpr
	withUnaryOp *unaryExprWithUnaryOp
	node
}

type unaryExprWithUnaryOp struct {
	operator  string
	unaryExpr *unaryExpr
}

func (p *parser) unaryExpr() (*unaryExpr, error) {
	if p.isUnaryOperator() {
		op := p.consumeToken()
		operand, err := p.unaryExpr()
		if err != nil {
			return nil, fmt.Errorf("unary expr: %w", err)
		}

		return &unaryExpr{
			kind: UNARY_EXPR_KIND_WITH_UNARY_OP,
			withUnaryOp: &unaryExprWithUnaryOp{
				operator:  op.value,
				unaryExpr: operand,
			},
			node: p.nodeData(),
		}, nil
	}

	primaryExpr, err := p.primaryExpr()
	if err != nil {
		return nil, fmt.Errorf("primaryExpr: %w", err)
	}
	return &unaryExpr{
		kind:        UNARY_EXPR_KIND_PRIMARY,
		primaryExpr: primaryExpr,
		node:        p.nodeData()}, nil
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
	return &primaryExpr{operand: operand, node: p.nodeData()}, nil
}

type operandKind int

const (
	OPERAND_KIND_LITERAL operandKind = iota + 1
	OPERAND_KIND_OPERAND_NAME
	OPERAND_KIND_EXPRESSION_WITH_PARENTHESES
)

type operand struct {
	kind        operandKind
	literal     *literal
	operandName *operandName
	expression  *expression
	node
}

func (p *parser) operand() (*operand, error) {
	if p.currentToken().kind == TK_PUNCTUATION && p.currentToken().value == "(" {
		if err := p.expectPunctuation("("); err != nil {
			return nil, fmt.Errorf("expect punctuation: %w", err)
		}

		expr, err := p.expression()
		if err != nil {
			return nil, fmt.Errorf("exprettion: %w", err)
		}

		if err := p.expectPunctuation(")"); err != nil {
			return nil, fmt.Errorf("expect punctuation: %w", err)
		}

		return &operand{
			kind:       OPERAND_KIND_EXPRESSION_WITH_PARENTHESES,
			expression: expr,
			node:       p.nodeData(),
		}, nil
	}

	tokens2 := p.currentTokens(2)
	if tokens2[0].kind != TK_IDENT || tokens2[1].value == "{" {
		// まだ composite literal には対応していないが、{ で区別できるのでする。
		literal, err := p.literal()
		if err != nil {
			return nil, fmt.Errorf("literal: %w", err)
		}
		return &operand{
			kind:    OPERAND_KIND_LITERAL,
			literal: literal,
			node:    p.nodeData(),
		}, nil
	}

	operandName, err := p.operandName()
	if err != nil {
		return nil, fmt.Errorf("operand name: %w", err)
	}
	return &operand{kind: OPERAND_KIND_OPERAND_NAME, operandName: operandName, node: p.nodeData()}, nil
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
	return &literal{basicLit: bl, node: p.nodeData()}, nil
}

type basicLit struct {
	intLit *intLit
	node
}

func (p *parser) basicLit() (*basicLit, error) {
	intL, err := p.intLit()
	if err != nil {
		return nil, fmt.Errorf("intLit: %w", err)
	}
	return &basicLit{intLit: intL, node: p.nodeData()}, nil
}

type operandName struct {
	operandName *identifier
	node
}

func (p *parser) operandName() (*operandName, error) {
	identifier, err := p.identifier()
	if err != nil {
		return nil, fmt.Errorf("expected identifier: %w", err)
	}
	return &operandName{operandName: identifier, node: p.nodeData()}, nil
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
	return &intLit{n: n, node: p.nodeData()}, nil
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
	return &identifier{name: name, node: p.nodeData()}, nil
}

type packageClause struct {
	packageName *packageName
	node
}

func (p *parser) packageClause() (*packageClause, error) {
	if err := p.expectCurrentToken(TK_PACKAGE); err != nil {
		return nil, fmt.Errorf("expected \"package\": %w", err)
	}
	packageName, err := p.packageName()
	if err != nil {
		return nil, fmt.Errorf("expected package name: %w", err)
	}
	return &packageClause{packageName: packageName, node: p.nodeData()}, nil
}

type packageName struct {
	identifier *identifier
	node
}

func (p *parser) packageName() (*packageName, error) {
	identifier, err := p.identifier()
	if err != nil {
		return nil, fmt.Errorf("expected identifier: %w", err)
	}
	return &packageName{identifier: identifier, node: p.nodeData()}, nil
}

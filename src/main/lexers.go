package main

import "fmt"

type keyword string

const (
	selectKeyword keyword = "select"
	fromKeyword   keyword = "from"
	whereKeyword  keyword = "where"
	asKeyword     keyword = "as"

	createKeyword keyword = "create"
	tableKeyword  keyword = "table"

	insertKeyword keyword = "insert"
	intoKeyword   keyword = "into"
	valuesKeyword keyword = "values"

	updateKeyword keyword = "update"
	deleteKeyword keyword = "delete"

	intKeyword  keyword = "int"
	textKeyword keyword = "text"
)

type symbol string

const (
	semicolonSymbol  symbol = ";"
	commaSymbol      symbol = ","
	asteriskSymbol   symbol = "*"
	leftparenSymbol  symbol = "("
	rightparenSymbol symbol = ")"
	concatSymbol     symbol = "+"
)

type tokenKind uint

const (
	keywordKind tokenKind = iota
	symbolKind
	identifierKind

	stringKind
	numericKind
	boolKind
)

type location struct {
	row uint
	col uint
}

func (l *location) String() string {
	return fmt.Sprintf("%d: %d", l.row, l.col)
}

type token struct {
	value string
	kind  tokenKind
	loc   location
}

type cursor struct {
	pointer uint
	loc     location
}

func (t *token) equals(other *token) bool {
	return t.kind == other.kind && t.value == other.value
}

type lexer func(string, cursor) (*token, cursor, bool)

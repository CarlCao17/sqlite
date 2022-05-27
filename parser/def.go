package parser

import "fmt"

type keyword string

const (
	selectKeyword keyword = "select"
	fromKeyword   keyword = "from"
	whereKeyword  keyword = "where"
	asKeyword     keyword = "as"
	tableKeyword  keyword = "table"
	createKeyword keyword = "create"
	insertKeyword keyword = "insert"
	valuesKeyword keyword = "values"
	updateKeyword keyword = "update"
	deleteKeyword keyword = "delete"

	intKeyword  keyword = "int"
	textKeyword keyword = "text"
)

type symbol string

const (
	semiconlonSymbol symbol = ";"
	commaSymbol      symbol = ","
	asteriskSymbol   symbol = "*"
	leftparenSymbol  symbol = "("
	rightparenSymbol symbol = ")"
)

type tokenKind uint

const (
	keywordKind tokenKind = iota
	symbolKind
	identifierKind
	stringKind
	numericKind
)

type location struct {
	row int
	col int
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
	p   uint
	loc location
}

func (t *token) equals(other *token) bool {
	return t.kind == other.kind && t.value == other.value
}

type lexer func(string, cursor) (*token, cursor, bool)

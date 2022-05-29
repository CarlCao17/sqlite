package main

type AST struct {
	Statements []*Statement
}

type ASTKind int

const (
	SelectKind ASTKind = iota
	CreateTableKind
	InsertKind
)

type Statement struct {
	SelectStatement      *SelectStatement
	CreateTableStatement *CreateTableStatement
	InsertStatement      *InsertStatement
	Kind                 ASTKind
}

type InsertStatement struct {
	table  token
	values *[]*expression
}

type expressionKind int

const (
	literalKind expressionKind = iota
)

type expression struct {
	literal *token
	kind    expressionKind
}

type CreateTableStatement struct {
	name token
	cols *[]*columnDefinition
}

type columnDefinition struct {
	name     token
	datatype token
}

type SelectStatement struct {
	item []*expression
	from *token
}

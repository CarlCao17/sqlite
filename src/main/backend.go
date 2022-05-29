package main

import "errors"

type ColumnType uint

const (
	TextType ColumnType = iota
	IntType
)

type Cell interface {
	AsText() string
	AsInt() int32
}

type Results struct {
	Columns []struct {
		Type ColumnType
		Name string
	}
	Rows [][]Cell
}

var (
	ErrTableNotExist     = errors.New("Table does not exist")
	ErrColumnNotExist    = errors.New("Column does not exist")
	ErrInvalidSelectItem = errors.New("Select item is not valid")
	ErrInvalidDataType   = errors.New("Invalid data type")
	ErrMissingValues     = errors.New("Missing values")
)

type Backend interface {
	CreateTable(statement *CreateTableStatement) error
	Insert(statement *InsertStatement) (int, error)
	Select(statement *SelectStatement) (*Results, error)
}

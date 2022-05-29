package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
)

type MemoryCell []byte

func (mc MemoryCell) AsInt() int32 {
	var i int32
	err := binary.Read(bytes.NewBuffer(mc), binary.BigEndian, &i)
	if err != nil {
		panic(err)
	}
	return i
}

func (mc MemoryCell) AsText() string {
	return string(mc)
}

type table struct {
	columns     []string
	columnTypes []ColumnType
	rows        [][]MemoryCell
}

type MemoryBackend struct {
	tables map[string]*table
}

func NewMemoryBackend() *MemoryBackend {
	return &MemoryBackend{
		tables: map[string]*table{},
	}
}

func (mb *MemoryBackend) CreateTable(crt *CreateTableStatement) error {
	t := table{}
	mb.tables[crt.name.value] = &t
	if crt.cols == nil {
		return nil
	}

	for _, col := range *crt.cols {
		t.columns = append(t.columns, col.name.value)

		var dt ColumnType
		switch col.datatype.value {
		case "int":
			dt = IntType
		case "text":
			dt = TextType
		default:
			return ErrInvalidDataType
		}
		t.columnTypes = append(t.columnTypes, dt)
	}
	return nil
}

func (mb *MemoryBackend) Insert(inst *InsertStatement) (int, error) {
	table, ok := mb.tables[inst.table.value]
	if !ok {
		return 0, ErrTableNotExist
	}

	if inst.values == nil {
		return 0, nil
	}

	row := []MemoryCell{}

	if len(*inst.values) != len(table.columns) {
		return 0, ErrMissingValues
	}

	for _, value := range *inst.values {
		if value.kind != literalKind {
			fmt.Println("Skipping non-literal.")
			continue
		}

		row = append(row, mb.tokenToCell(value.literal))
	}
	table.rows = append(table.rows, row)
	return 1, nil
}

func (mb *MemoryBackend) Select(slct *SelectStatement) (*Results, error) {
	table, ok := mb.tables[slct.from.value]
	if !ok {
		return nil, ErrTableNotExist
	}

	results := [][]Cell{}
	columns := []struct {
		Type ColumnType
		Name string
	}{}

	for i, row := range table.rows {
		result := []Cell{}
		isFirstRow := i == 0

		for _, exp := range slct.item {
			// Unsupported
			if exp.kind != literalKind {
				fmt.Println("Skipping non-literal expression")
				continue
			}

			lit := exp.literal
			if lit.kind == identifierKind {
				found := false
				for i, tableCol := range table.columns {
					if tableCol == lit.value {
						if isFirstRow {
							columns = append(columns, struct {
								Type ColumnType
								Name string
							}{
								Type: table.columnTypes[i],
								Name: lit.value,
							})
						}
						result = append(result, row[i])
						found = true
						break
					}
				}
				if !found {
					return nil, ErrColumnNotExist
				}
				continue
			}
			return nil, ErrColumnNotExist
		}
		results = append(results, result)
	}
	return &Results{
		Columns: columns,
		Rows:    results,
	}, nil
}

func (mb *MemoryBackend) tokenToCell(t *token) MemoryCell {
	if t.kind == numericKind {
		buf := new(bytes.Buffer)
		i, err := strconv.Atoi(t.value)
		if err != nil {
			panic(err)
		}

		err = binary.Write(buf, binary.BigEndian, int32(i))
		if err != nil {
			panic(err)
		}
		return MemoryCell(buf.Bytes())
	}

	if t.kind == stringKind {
		return MemoryCell(t.value)
	}

	return nil
}

package main

import (
	"fmt"
	"strings"
)

func Spaces(n int) string {
	b := strings.Builder{}
	for i := 0; i < n; i++ {
		b.WriteString(" ")
	}
	return b.String()
}

func PrintSQLQueryResult(results *Results) {
	for _, col := range results.Columns {
		fmt.Printf("| %s ", col.Name)
	}
	fmt.Printf("|\n")

	for i := 0; i < 20; i++ {
		fmt.Printf("=")
	}
	fmt.Println()

	for _, result := range results.Rows {
		fmt.Printf("|")

		for i, cell := range result {
			typ := results.Columns[i].Type
			s := ""
			switch typ {
			case IntType:
				s = fmt.Sprintf("%d", cell.AsInt())
			case TextType:
				s = fmt.Sprintf("%s", cell.AsText())
			}
			fmt.Printf(" %s | ", s)
		}
		fmt.Println()
	}
	fmt.Println("ok")
}

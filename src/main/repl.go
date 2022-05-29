package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	mb := NewMemoryBackend()

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Welcom to sqlite.")
	for {
		fmt.Print("$ ")
		text, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		text = strings.TrimSpace(text)
		text = strings.Replace(text, "\n", "", -1)

		ast, err := Parse(text)
		if err != nil {
			panic(err)
		}

		for _, stmt := range ast.Statements {
			switch stmt.Kind {
			case CreateTableKind:
				err = mb.CreateTable(stmt.CreateTableStatement)
				if err != nil {
					panic(err)
				}
				fmt.Println("ok")
			case InsertKind:
				rows, err := mb.Insert(stmt.InsertStatement)
				if err != nil {
					panic(err)
				}
				fmt.Printf("%d Rows has been effected\n", rows)
			case SelectKind:
				results, err := mb.Select(stmt.SelectStatement)
				if err != nil {
					panic(err)
				}
				PrintSQLQueryResult(results)
			}
		}
	}
}

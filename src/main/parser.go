package main

import (
	"errors"
	"fmt"
)

func Parse(source string) (*AST, error) {
	tokens, err := lex(source)
	if err != nil {
		return nil, err
	}

	ast := AST{}
	cursor := uint(0)
	for cursor < uint(len(tokens)) {
		stmt, newCursor, ok := parseStatement(tokens, cursor, tokenFromSymbol(semicolonSymbol))
		if !ok {
			helpMessage(tokens, cursor, "Expect statement")
			return nil, errors.New("failed to parse, expect statement")
		}
		cursor = newCursor

		ast.Statements = append(ast.Statements, stmt)
		atLeastOneSemicolon := false
		for expectToken(tokens, cursor, tokenFromSymbol(semicolonSymbol)) {
			cursor++
			atLeastOneSemicolon = true
		}
		if !atLeastOneSemicolon {
			helpMessage(tokens, cursor, "Expected semi-colon delimiter between statements")
			return nil, errors.New("missing semi-colon between statements")
		}
	}
	return &ast, nil
}

func parseStatement(tokens []*token, initialCursor uint, delimiter token) (*Statement, uint, bool) {
	cursor := initialCursor

	// Look for a SELECT statement
	slct, newCursor, ok := parseSelectStatement(tokens, cursor, delimiter)
	if ok {
		return &Statement{
			Kind:            SelectKind,
			SelectStatement: slct,
		}, newCursor, true
	}

	// Look for a INSERT statement
	inst, newCursor, ok := parseInsertStatement(tokens, cursor, delimiter)
	if ok {
		return &Statement{
			Kind:            InsertKind,
			InsertStatement: inst,
		}, newCursor, true
	}

	// Look for a CREATE statement
	crtTbl, newCursor, ok := parseCreateTableStatement(tokens, cursor, delimiter)
	if ok {
		return &Statement{
			Kind:                 CreateTableKind,
			CreateTableStatement: crtTbl,
		}, newCursor, true
	}
	return nil, initialCursor, false
}

// parseSelectStatement parse SELECT statment, it will follow the pattern:
// SELECT
// $expression [, ...]
// FROM
// $table_name
func parseSelectStatement(tokens []*token, initialCursor uint, delimiter token) (*SelectStatement, uint, bool) {
	//if len(tokens) == 0 || initialCursor >= uint(len(tokens)) {
	//	return nil, initialCursor, false
	//}
	cursor := initialCursor
	if !expectToken(tokens, cursor, tokenFromKeyword(selectKeyword)) {
		return nil, initialCursor, false
	}
	slct := SelectStatement{}

	exps, newCursor, ok := parseExpressions(tokens, cursor, []token{tokenFromKeyword(fromKeyword)})
	if !ok {
		return nil, initialCursor, false
	}
	cursor = newCursor
	if expectToken(tokens, cursor, tokenFromKeyword(fromKeyword)) {
		cursor++

		from, newCursor, ok := parseToken(tokens, cursor, identifierKind)
		if !ok {
			helpMessage(tokens, cursor, "Expected FROM token")
			return nil, initialCursor, false
		}
		slct.from = from
		slct.item = *exps
		cursor = newCursor
	}
	return &slct, cursor, true

}

// parseInsertStatement parse INSERT statement, it will follow the pattern:
// INSERT
// INTO
// $table_name
// VALUES
// (
// $expression [, ...]
// )
func parseInsertStatement(tokens []*token, initialCursor uint, delimiter token) (*InsertStatement, uint, bool) {
	//if len(tokens) == 0 || initialCursor > uint(len(tokens)) {
	//	return nil, initialCursor, false
	//}
	cursor := initialCursor
	// Look for INSERT
	if !expectToken(tokens, cursor, tokenFromKeyword(insertKeyword)) {
		return nil, initialCursor, false
	}
	cursor++

	// Look for INTO
	if !expectToken(tokens, cursor, tokenFromKeyword(intoKeyword)) {
		helpMessage(tokens, cursor, "Expected into")
		return nil, initialCursor, false
	}
	cursor++

	// Look for table name
	table, newCursor, ok := parseToken(tokens, cursor, identifierKind)
	if !ok {
		helpMessage(tokens, cursor, "Expected table name")
		return nil, initialCursor, false
	}
	cursor = newCursor

	// Look for VALUES
	if !expectToken(tokens, cursor, tokenFromKeyword(valuesKeyword)) {
		helpMessage(tokens, cursor, "Expected values")
		return nil, initialCursor, false
	}
	cursor++

	// Look for left paren
	if !expectToken(tokens, cursor, tokenFromSymbol(leftparenSymbol)) {
		helpMessage(tokens, cursor, "Expected '('")
	}
	cursor++

	// Look for expression list
	values, newCursor, ok := parseExpressions(tokens, cursor, []token{tokenFromSymbol(rightparenSymbol)})
	if !ok {
		return nil, initialCursor, false
	}
	cursor = newCursor

	// Look for right paren
	if !expectToken(tokens, cursor, tokenFromSymbol(rightparenSymbol)) {
		helpMessage(tokens, cursor, "Expected ')'")
		return nil, initialCursor, false
	}
	cursor++

	return &InsertStatement{
		table:  *table,
		values: values,
	}, cursor, true
}

// parseCreateTableStatement parse CREATE statement, it will follow the pattern
// CREATE TABLE
// $table_name
// (
// [$column_name $column_type [,...]]
// )
func parseCreateTableStatement(tokens []*token, initialCursor uint, delimiters token) (*CreateTableStatement, uint, bool) {
	//if len(tokens) == 0 || initialCursor > uint(len(tokens)) {
	//	return nil, initialCursor, false
	//}
	cursor := initialCursor

	// Look for CREATE TABLE
	if !expectToken(tokens, cursor, tokenFromKeyword(createKeyword)) {
		return nil, initialCursor, false
	}
	cursor++
	if !expectToken(tokens, cursor, tokenFromKeyword(tableKeyword)) {
		return nil, initialCursor, false
	}
	cursor++

	// Look for table name
	tableName, newCursor, ok := parseToken(tokens, cursor, identifierKind)
	if !ok {
		helpMessage(tokens, cursor, "Expected table name")
		return nil, initialCursor, false
	}
	cursor = newCursor

	// Look for left paren
	if !expectToken(tokens, cursor, tokenFromSymbol(leftparenSymbol)) {
		helpMessage(tokens, cursor, "Expected left parenthesis '('")
		return nil, initialCursor, false
	}
	cursor++

	// Look for column definitions
	cols, newCursor, ok := parseColumnDefinitions(tokens, cursor, tokenFromSymbol(rightparenSymbol))
	if !ok {
		return nil, initialCursor, false
	}
	cursor = newCursor

	if !expectToken(tokens, cursor, tokenFromSymbol(rightparenSymbol)) {
		helpMessage(tokens, cursor, "Expected right parenthesis ')'")
		return nil, initialCursor, false
	}
	cursor++

	return &CreateTableStatement{
		name: *tableName,
		cols: cols,
	}, cursor, true
}

func parseExpressions(tokens []*token, initialCursor uint, delimiters []token) (*[]*expression, uint, bool) {
	cursor := initialCursor

	exps := []*expression{}

outer:
	for cursor < uint(len(tokens)) {
		currentToken := tokens[cursor]
		for _, delimiter := range delimiters {
			if delimiter.equals(currentToken) {
				break outer
			}
		}
		if len(exps) > 0 {
			if !expectToken(tokens, cursor, tokenFromSymbol(commaSymbol)) {
				return nil, initialCursor, false
			}
			cursor++
		}
		// Look for expression
		exp, newCursor, ok := parseExpression(tokens, cursor, tokenFromSymbol(commaSymbol))
		if !ok {
			helpMessage(tokens, cursor, "Expected expression")
			return nil, initialCursor, false
		}
		cursor = newCursor
		exps = append(exps, exp)
	}
	return &exps, cursor, true
}

func parseExpression(tokens []*token, initialCursor uint, _ token) (*expression, uint, bool) {
	cursor := initialCursor

	kinds := []tokenKind{identifierKind, numericKind, stringKind}
	for _, kind := range kinds {
		t, newCursor, ok := parseToken(tokens, cursor, kind)
		if ok {
			return &expression{
				literal: t,
				kind:    literalKind,
			}, newCursor, true
		}
	}

	return nil, initialCursor, false
}

func parseToken(tokens []*token, initialCursor uint, kind tokenKind) (*token, uint, bool) {
	cursor := initialCursor

	if cursor >= uint(len(tokens)) {
		return nil, initialCursor, false
	}

	current := tokens[cursor]
	if current.kind == kind {
		return current, cursor + 1, true
	}
	return nil, initialCursor, false
}

func parseColumnDefinitions(tokens []*token, initialCursor uint, delimiter token) (*[]*columnDefinition, uint, bool) {
	cursor := initialCursor

	cds := []*columnDefinition{}
	for {
		if cursor >= uint(len(tokens)) {
			return nil, initialCursor, false
		}

		// Look for a delimiter
		current := tokens[cursor]
		if delimiter.equals(current) {
			break
		}

		// Look for a comma
		if len(cds) > 0 {
			if !expectToken(tokens, cursor, tokenFromSymbol(commaSymbol)) {
				helpMessage(tokens, cursor, "Expected comma")
				return nil, initialCursor, false
			}
			cursor++
		}

		// Look for a column name
		id, newCursor, ok := parseToken(tokens, cursor, identifierKind)
		if !ok {
			helpMessage(tokens, cursor, "Expect column name")
			return nil, initialCursor, false
		}
		cursor = newCursor

		// Look for a column type
		ty, newCursor, ok := parseToken(tokens, cursor, keywordKind)
		if !ok {
			helpMessage(tokens, cursor, "Expected column type")
			return nil, initialCursor, false
		}
		cursor = newCursor

		cds = append(cds, &columnDefinition{
			name:     *id,
			datatype: *ty,
		})
	}
	return &cds, cursor, true
}

func tokenFromKeyword(k keyword) token {
	return token{
		kind:  keywordKind,
		value: string(k),
	}
}

func tokenFromSymbol(s symbol) token {
	return token{
		kind:  symbolKind,
		value: string(s),
	}
}

func expectToken(tokens []*token, cursor uint, t token) bool {
	if cursor >= uint(len(tokens)) {
		return false
	}
	return t.equals(tokens[cursor])
}

func helpMessage(tokens []*token, cursor uint, msg string) {
	var c *token
	if cursor < uint(len(tokens)) {
		c = tokens[cursor]
	} else {
		c = tokens[cursor-1]
	}
	fmt.Printf("[%d:%d]: %s, got: %s\n", c.loc.row, c.loc.col, msg, c.value)
}

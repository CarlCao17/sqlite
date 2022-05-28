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
		stmt, newCursor, ok := parseStatement(tokens, cursor, tokenFromSymbol(semiconlonSymbol))
		if !ok {
			helpMessage(tokens, cursor, "Expect statement")
			return nil, errors.New("failed to parse, expect statement")
		}
		cursor = newCursor

		ast.Statements = append(ast.Statements, stmt)
		atLeastOneSemicolon := false
		for expectToken(tokens, cursor, tokenFromSymbol(semiconlonSymbol)) {
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
	if len(tokens) == 0 || initialCursor >= uint(len(tokens)) {
		return nil, initialCursor, false
	}
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
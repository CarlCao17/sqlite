package main

import (
	"fmt"
	"strings"
	"unicode"
)

var (
	lexers  = []lexer{lexKeyword, lexSymbol, lexIdentifier, lexString, lexNumeric}
	symbols = []string{
		string(commaSymbol),
		string(leftparenSymbol),
		string(rightparenSymbol),
		string(semicolonSymbol),
		string(asteriskSymbol),
	}
	keywords = []string{
		string(selectKeyword),
		string(fromKeyword),
		string(insertKeyword),
		string(intoKeyword),
		string(tableKeyword),
		string(createKeyword),
		string(valuesKeyword),

		string(intKeyword),
		string(textKeyword),

		string(updateKeyword),
		string(deleteKeyword),
		string(asKeyword),
		string(whereKeyword),
	}
)

func lex(source string) ([]*token, error) {
	tokens := []*token{}
	cur := cursor{}
lex:
	for cur.pointer < uint(len(source)) {
		for _, l := range lexers {
			if t, c, ok := l(source, cur); ok {
				if t != nil {
					tokens = append(tokens, t)
				}
				cur = c
				continue lex
			}
		}
		hint := source
		hint += "\n" + Spaces(int(cur.pointer)) + "^" + fmt.Sprintf("\nat: %v", cur.loc)
		return nil, fmt.Errorf("unable to lex tokens: %s", hint)
	}
	return tokens, nil
}

func lexKeyword(source string, ic cursor) (*token, cursor, bool) {
	cur := ic
	match := longestMatch(source, ic, keywords)
	if match == "" {
		return nil, ic, false
	}
	cur.pointer = ic.pointer + uint(len(match))
	cur.loc.col = ic.loc.col + uint(len(match))

	return &token{
		value: match,
		kind:  keywordKind,
		loc:   ic.loc,
	}, cur, true

}

func lexSymbol(source string, ic cursor) (*token, cursor, bool) {
	c := source[ic.pointer]
	cur := ic
	cur.pointer++
	cur.loc.col++

	switch c {
	case '\n':
		cur.loc.row++
		cur.loc.col = 0
		fallthrough
	case '\t':
		fallthrough
	case ' ':
		return nil, cur, true
	}

	match := longestMatch(source, ic, symbols)
	if match == "" {
		return nil, ic, false
	}

	cur.pointer = ic.pointer + uint(len(match))
	cur.loc.col = ic.loc.col + uint(len(match))

	return &token{
		value: match,
		kind:  symbolKind,
		loc:   ic.loc,
	}, cur, true
}

func lexIdentifier(source string, ic cursor) (*token, cursor, bool) {
	// Handle seperately if is a double-quoted identifier
	if token, newCursor, ok := lexCharacterDelimited(source, ic, '"'); ok {
		return token, newCursor, true
	}
	cur := ic
	// identifier should start with alphabet
	c := source[cur.pointer]
	isAlphabetical := c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z'
	if !isAlphabetical {
		return nil, ic, false
	}

	var value []byte
	for cur.pointer < uint(len(source)) {
		c = source[cur.pointer]
		isAlphabetical = c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z'
		isDigit := c >= '0' && c <= '9'
		isValidSymbol := c == '$' || c == '_'
		if isAlphabetical || isDigit || isValidSymbol {
			value = append(value, c)
			cur.pointer++
			cur.loc.col++
			continue
		}
		break
	}
	if len(value) == 0 {
		return nil, ic, false
	}
	return &token{
		// Unquoted identifier are case-insensitive
		value: strings.ToLower(string(value)),
		kind:  identifierKind,
		loc:   ic.loc,
	}, cur, true
}

func lexNumeric(source string, ic cursor) (*token, cursor, bool) {
	cur := ic
	periodFound := false
	expMarkFound := false

	for ; cur.pointer < uint(len(source)); cur.pointer++ {
		c := source[cur.pointer]
		cur.loc.col++

		isDigit := unicode.IsDigit(rune(c))
		isPeriod := c == '.'
		isExpMarker := c == 'e' || c == 'E'

		// Must start with a digit or period
		if cur.pointer == ic.pointer {
			if !isDigit && !isPeriod {
				return nil, ic, false
			}
		}
		if isPeriod {
			if periodFound {
				return nil, ic, false
			}
			periodFound = isPeriod
			continue
		}
		if isExpMarker {
			if expMarkFound {
				return nil, ic, false
			}
			// No periods allowed after expMarker
			periodFound = true
			expMarkFound = true

			if cur.pointer == uint(len(source)-1) {
				return nil, ic, false
			}
			cNext := source[cur.pointer+1]
			if cNext == '-' || cNext == '+' {
				cur.pointer++
				cur.loc.col++
			}
			continue
		}
		if !isDigit {
			// should back one byte to the current
			cur.loc.col--
			break
		}
	}
	// No character accumulated
	if cur.pointer == ic.pointer {
		return nil, ic, false
	}

	return &token{
		value: source[ic.pointer:cur.pointer],
		kind:  numericKind,
		loc:   ic.loc,
	}, cur, true
}

func lexString(source string, ic cursor) (*token, cursor, bool) {
	return lexCharacterDelimited(source, ic, '\'')
}

// lexCharacterDelimited does not allow nested delimter
// for example: source is a = 'c''bc''b', then will recognize that string equals to c"bc
//						"a = 'c"bc''"'
func lexCharacterDelimited(source string, ic cursor, delimiter byte) (*token, cursor, bool) {
	cur := ic

	if len(source[cur.pointer:]) == 0 {
		return nil, ic, false
	}

	if source[cur.pointer] != delimiter {
		return nil, ic, false
	}
	cur.pointer++
	cur.loc.col++
	// delimit the left quote
	ic = cur

	var value []byte
	for ; cur.pointer < uint(len(source)); cur.pointer++ {
		c := source[cur.pointer]

		if c == delimiter {
			// SQL escapes are via double characters, not backslash.
			if cur.pointer+1 >= uint(len(source)) || source[cur.pointer+1] != delimiter {
				// delimit the right quote
				cur.pointer++
				cur.loc.col++
				return &token{
					value: string(value),
					kind:  stringKind,
					loc:   ic.loc,
				}, cur, true
			}
			value = append(value, delimiter)
			cur.pointer++
			cur.loc.col++
		}
		value = append(value, c)
		cur.loc.col++
	}
	return nil, ic, false
}

func longestMatch(source string, ic cursor, options []string) string {
	var value []byte
	var match string
	skipOptions := make(map[string]bool)

	cur := ic
	for cur.pointer < uint(len(source)) {
		ch := byte(unicode.ToLower(rune(source[cur.pointer])))
		value = append(value, ch)
		cur.pointer++

		for _, option := range options {
			if _, exist := skipOptions[option]; exist {
				continue
			}

			if string(value) == option {
				skipOptions[option] = true
				if len(option) > len(match) {
					match = option
				}
			}

			hasCommonPrefix := string(value) == option[:len(value)]
			tooLong := len(value) > len(option)
			if tooLong || !hasCommonPrefix {
				skipOptions[option] = true
			}
		}
		if len(skipOptions) == len(options) {
			break
		}
	}
	return match
}

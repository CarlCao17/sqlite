package parser

import "fmt"

var (
	lexers = []lexer{lexKeyword, lexSymbol, lexIdentifier, lexString, lexNumeric}
)

func lex(source string) ([]*token, error) {
	tokens := []*token{}
	cur := cursor{}
lex:
	for cur.p < uint(len(source)) {
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
		hint += "\n" + Spaces(int(cur.p)) + "|" + fmt.Sprintf("\nat: %s", cur.loc)
		return nil, fmt.Errorf("unable to lex tokens: %s", hint)
	}
}

//
//func lexKeyword()

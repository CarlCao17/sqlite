package main

import "strings"

func Spaces(n int) string {
	b := strings.Builder{}
	for i := 0; i < n; i++ {
		b.WriteString(" ")
	}
	return b.String()
}

package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	type args struct {
		source string
	}
	tests := []struct {
		name    string
		args    args
		want    *AST
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args.source)
			if !tt.wantErr(t, err, fmt.Sprintf("Parse(%v)", tt.args.source)) {
				return
			}
			assert.Equalf(t, tt.want, got, "Parse(%v)", tt.args.source)
		})
	}
}

package main

import (
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_lex(t *testing.T) {
	Convey("edge case", t, func() {
		source := ""
		expectTokens := []*token{}
		got, err := lex(source)
		So(got, ShouldResemble, expectTokens)
		So(err, ShouldBeNil)
	})

	Convey("normal case", t, func() {
		testCases := map[string]struct {
			t   []*token
			err error
		}{}
		for source, expect := range testCases {
			gotTokens, err := lex(source)
			So(gotTokens, ShouldResemble, expect.t)
			So(err, ShouldBeNil)
		}
	})
}

func Test_lexNumeric(t *testing.T) {
	type args struct {
		source string
		ic     cursor
	}
	type testCase struct {
		name  string
		args  args
		want  *token
		want1 cursor
		want2 bool
	}
	start := cursor{}
	Convey("edge case", t, func() {
		cursorSpace := cursor{
			pointer: 5,
			loc: location{
				row: 0,
				col: 5,
			},
		}
		cases := []testCase{
			{
				name: "empty source",
				args: args{
					source: "",
					ic:     start,
				},
				want:  nil,
				want1: start,
				want2: false,
			},
			{
				name: "start with alphabet",
				args: args{
					source: "v123",
					ic:     start,
				},
				want:  nil,
				want1: start,
				want2: false,
			},
			{
				name: "spaces",
				args: args{
					source: Spaces(10),
					ic:     cursorSpace,
				},
				want:  nil,
				want1: cursorSpace,
				want2: false,
			},
		}
		for _, c := range cases {
			got, got1, got2 := lexNumeric(c.args.source, c.args.ic)
			So(got, ShouldResemble, c.want)
			So(got1, ShouldResemble, c.want1)
			So(got2, ShouldEqual, c.want2)
		}
	})

	Convey("normal case", t, func() {
		cases := []testCase{
			{
				name: "decimal",
				args: args{
					source: "3.1415926",
					ic:     start,
				},
				want: &token{
					value: "3.1415926",
					kind:  numericKind,
					loc:   start.loc,
				},
				want1: cursor{
					pointer: 9,
					loc: location{
						row: 0,
						col: 9,
					},
				},
				want2: true,
			},
			{
				name: "decimal2",
				args: args{
					source: "4. ",
					ic:     start,
				},
				want: &token{
					value: "4.",
					kind:  numericKind,
					loc:   start.loc,
				},
				want1: cursor{
					pointer: 2,
					loc: location{
						row: 0,
						col: 2,
					},
				},
				want2: true,
			},
			{
				name: "decimal3",
				args: args{
					source: "   .0340",
					ic: cursor{
						pointer: 3,
						loc: location{
							row: 1,
							col: 3,
						},
					},
				},
				want: &token{
					value: ".0340",
					kind:  numericKind,
					loc: location{
						row: 1,
						col: 3,
					},
				},
				want1: cursor{
					pointer: 8,
					loc: location{
						row: 1,
						col: 8,
					},
				},
				want2: true,
			},
			{
				name: "integer",
				args: args{
					source: "123456789abc",
					ic: cursor{
						pointer: 3,
						loc: location{
							row: 1,
							col: 3,
						},
					},
				},
				want: &token{
					value: "456789",
					kind:  numericKind,
					loc: location{
						row: 1,
						col: 3,
					},
				},
				want1: cursor{
					pointer: 9,
					loc: location{
						row: 1,
						col: 9,
					},
				},
				want2: true,
			},
			{
				name: "integer2",
				args: args{
					source: "0 0",
					ic:     start,
				},
				want: &token{
					value: "0",
					kind:  numericKind,
					loc:   start.loc,
				},
				want1: cursor{
					pointer: 1,
					loc: location{
						row: 0,
						col: 1,
					},
				},
				want2: true,
			},
			{
				name: "exponent",
				args: args{
					source: "1.23e12abc",
					ic:     start,
				},
				want: &token{
					value: "1.23e12",
					kind:  numericKind,
					loc:   start.loc,
				},
				want1: cursor{
					pointer: 7,
					loc: location{
						row: 0,
						col: 7,
					},
				},
				want2: true,
			},
			{
				name: "scientific notation",
				args: args{
					source: ".23456789E+12",
					ic:     start,
				},
				want: &token{
					value: ".23456789E+12",
					kind:  numericKind,
					loc:   start.loc,
				},
				want1: cursor{
					pointer: 13,
					loc: location{
						row: 0,
						col: 13,
					},
				},
				want2: true,
			},
		}
		for _, c := range cases {
			got, got1, got2 := lexNumeric(c.args.source, c.args.ic)
			So(got, ShouldResemble, c.want)
			So(got1, ShouldResemble, c.want1)
			So(got2, ShouldEqual, c.want2)
		}

		Convey("multi rows", func() {
			cases := []testCase{
				{
					name: "two row scientific notation",
					args: args{
						source: "select\n\t20.45e-13 ",
						ic: cursor{
							pointer: 8,
							loc: location{
								row: 1,
								col: 1,
							},
						},
					},
					want: &token{
						value: "20.45e-13",
						kind:  numericKind,
						loc: location{
							row: 1,
							col: 1,
						},
					},
					want1: cursor{
						pointer: 17,
						loc: location{
							row: 1,
							col: 10,
						},
					},
					want2: true,
				},
				{
					name: "multi-rows decimal",
					args: args{
						source: "select\n\t*\nfrom products\nwhere total = 13041.56\n",
						ic: cursor{
							pointer: 38,
							loc: location{
								row: 3,
								col: 14,
							},
						},
					},
					want: &token{
						value: "13041.56",
						kind:  numericKind,
						loc: location{
							row: 3,
							col: 14,
						},
					},
					want1: cursor{
						pointer: 46,
						loc: location{
							row: 3,
							col: 22,
						},
					},
					want2: true,
				},
			}
			for _, c := range cases {
				got, got1, got2 := lexNumeric(c.args.source, c.args.ic)
				So(got, ShouldResemble, c.want)
				So(got1, ShouldResemble, c.want1)
				So(got2, ShouldEqual, c.want2)
			}
		})
	})
}

func Test_lexKeyword(t *testing.T) {
	type args struct {
		source string
		ic     cursor
	}
	type testCase struct {
		name  string
		args  args
		want  *token
		want1 cursor
		want2 bool
	}
	start := cursor{}

	Convey("edge case", t, func() {
		testCases := []testCase{
			{
				name: "empty input",
				args: args{
					source: "",
					ic:     start,
				},
				want:  nil,
				want1: start,
				want2: false,
			},
			{
				name: "error prefix with space",
				args: args{
					source: " source",
					ic:     start,
				},
				want:  nil,
				want1: start,
				want2: false,
			},
			{
				name: "error prefix",
				args: args{
					source: "a select",
					ic:     start,
				},
				want:  nil,
				want1: start,
				want2: false,
			},
			{
				name: "error keyword",
				args: args{
					source: "salect",
					ic:     start,
				},
				want:  nil,
				want1: start,
				want2: false,
			},
		}
		for _, c := range testCases {
			got, got1, got2 := lexKeyword(c.args.source, c.args.ic)
			So(got, ShouldResemble, c.want)
			So(got1, ShouldResemble, c.want1)
			So(got2, ShouldEqual, c.want2)
		}
	})

	Convey("normal case", t, func() {
		testCases := []testCase{
			{
				name: "select keyword starting from pos 1",
				args: args{
					source: " select ",
					ic: cursor{
						pointer: 1,
						loc:     location{0, 1},
					},
				},
				want: &token{
					value: "select",
					kind:  keywordKind,
					loc:   location{0, 1},
				},
				want1: cursor{pointer: 7, loc: location{0, 7}},
				want2: true,
			},
			{
				name: "normal SQL query statement",
				args: args{
					source: ";\n select * from table where primary_key = 1",
					ic:     cursor{pointer: 3, loc: location{1, 2}},
				},
				want: &token{
					value: "select",
					kind:  keywordKind,
					loc:   location{1, 2},
				},
				want1: cursor{pointer: 9, loc: location{1, 8}},
				want2: true,
			},
		}
		for _, c := range testCases {
			got, got1, got2 := lexKeyword(c.args.source, c.args.ic)
			So(got, ShouldResemble, c.want)
			So(got1, ShouldResemble, c.want1)
			So(got2, ShouldEqual, c.want2)
		}
	})
}

func Test_lexSymbol(t *testing.T) {
	type args struct {
		source string
		ic     cursor
	}
	type testCase struct {
		name  string
		args  args
		want  *token
		want1 cursor
		want2 bool
	}

	Convey("edge case", t, func() {
		testCases := []testCase{}
		for _, c := range testCases {
			got, got1, got2 := lexSymbol(c.args.source, c.args.ic)
			So(got, ShouldResemble, c.want)
			So(got1, ShouldResemble, c.want1)
			So(got2, ShouldEqual, c.want2)
		}
	})

	Convey("normal case", t, func() {
		testCases := []testCase{}
		for _, c := range testCases {
			got, got1, got2 := lexSymbol(c.args.source, c.args.ic)
			So(got, ShouldResemble, c.want)
			So(got1, ShouldResemble, c.want1)
			So(got2, ShouldEqual, c.want2)
		}
	})
}

func Test_lexIdentifier(t *testing.T) {
	type args struct {
		source string
		ic     cursor
	}
	type testCase struct {
		name  string
		args  args
		want  *token
		want1 cursor
		want2 bool
	}

	Convey("edge case", t, func() {
		testCases := []testCase{}
		for _, c := range testCases {
			got, got1, got2 := lexIdentifier(c.args.source, c.args.ic)
			So(got, ShouldResemble, c.want)
			So(got1, ShouldResemble, c.want1)
			So(got2, ShouldEqual, c.want2)
		}
	})

	Convey("normal case", t, func() {
		testCases := []testCase{}
		for _, c := range testCases {
			got, got1, got2 := lexIdentifier(c.args.source, c.args.ic)
			So(got, ShouldResemble, c.want)
			So(got1, ShouldResemble, c.want1)
			So(got2, ShouldEqual, c.want2)
		}
	})

}

func Test_lexCharacterDelimited(t *testing.T) {
	type args struct {
		source    string
		ic        cursor
		delimiter byte
	}
	tests := []struct {
		name  string
		args  args
		want  *token
		want1 cursor
		want2 bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := lexCharacterDelimited(tt.args.source, tt.args.ic, tt.args.delimiter)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("lexCharacterDelimited() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("lexCharacterDelimited() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("lexCharacterDelimited() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func Test_longestMatch(t *testing.T) {
	type args struct {
		source  string
		ic      cursor
		options []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := longestMatch(tt.args.source, tt.args.ic, tt.args.options); got != tt.want {
				t.Errorf("longestMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

package main

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestSpaces(t *testing.T) {
	convey.Convey("test", t, func() {
		testCases := map[int]string{
			0:  "",
			1:  " ",
			2:  "  ",
			3:  "   ",
			4:  "    ",
			8:  "        ",
			20: "                    ",
		}
		for n, spaceStr := range testCases {
			got := Spaces(n)
			convey.So(got, convey.ShouldEqual, spaceStr)
		}
	})
}

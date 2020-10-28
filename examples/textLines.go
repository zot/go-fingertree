//Text lines example: tracks text offsets by both line and character
package main

import (
	"fmt"

	ft "github.com/zot/go-fingertree"
)

//lineMeasure accumulates both line and character offsets
type lineMeasure struct {
	line int
	char int
}

//newLineMeasurer track both line and character offsets
func newLineMeasurer() *ft.Measurer {
	return ft.NewMeasurer(
		func() ft.MeasureValue {
			return &lineMeasure{0, 0}
		}, func(item ft.TreeItem) ft.MeasureValue {
			return &lineMeasure{1, len(item.(string))}
		}, func(v1 ft.MeasureValue, v2 ft.MeasureValue) ft.MeasureValue {
			l1 := v1.(*lineMeasure)
			l2 := v2.(*lineMeasure)
			return &lineMeasure{l1.line + l2.line, l1.char + l2.char}
		})
}

func characterOffset(t ft.Fingertree, offset int) {
	s := t.Split(func(m ft.MeasureValue) bool {
		return m.(*lineMeasure).char > offset
	})
	m1 := s[0].Measure().(*lineMeasure)
	line := s[1].PeekFirst()
	if line == nil {
		line = "EOF"
	}
	fmt.Printf("offset %d, line %d:%d: %s\n", offset, m1.line, m1.char, line)
}

func lineOffset(t ft.Fingertree, offset int) {
	s := t.Split(func(m ft.MeasureValue) bool {
		return m.(*lineMeasure).line > offset
	})
	m1 := s[0].Measure().(*lineMeasure)
	line := s[1].PeekFirst()
	if line == nil {
		line = "EOF"
	}
	fmt.Printf("line %d:%d: %s\n", m1.line, m1.char, line)
}

func main() {
	t := ft.With(newLineMeasurer(), "this", "is", "a", "test")
	characterOffset(t, 0)
	characterOffset(t, 3)
	characterOffset(t, 4)
	characterOffset(t, 5)
	characterOffset(t, 6)
	characterOffset(t, 7)
	characterOffset(t, 10)
	characterOffset(t, 11)
	lineOffset(t, 0)
	lineOffset(t, 1)
	lineOffset(t, 2)
	lineOffset(t, 3)
	lineOffset(t, 4)
}

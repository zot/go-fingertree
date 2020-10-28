package main

import (
	. "fingertree"
	"fmt"
	"math"
)

func testTree(t Fingertree) {
	t.Each(func(item TreeItem) bool {
		if _, ok := item.(int); !ok {
			if _, ok = item.(Traversable); ok {
				panic(fmt.Sprintf("found traversable item: %#v", item))
			} else {
				panic(fmt.Sprintf("found non-int, non-traversable: %#v", item))
			}
		}
		return true
	})
}

func treeString(t Fingertree) string {
	s := fmt.Sprintf("Tree(%v)[", t.Measure())
	first := true

	for _, v := range Items(t) {
		if first {
			first = false
		} else {
			s += ", "
		}
		s += fmt.Sprint(v)
	}
	s += "]"
	return s
}

func printTree(t Fingertree, p Predicate) {
	fmt.Printf("%s %v\n", treeString(t.TakeUntil(p)), t.Find(p))
}

func main() {
	m := NewMeasurer(
		func() MeasureValue { return 0 },
		func(i TreeItem) MeasureValue { return 1 },
		func(m1 MeasureValue, m2 MeasureValue) MeasureValue {
			return m1.(int) + m2.(int)
		})
	t := With(m)

	for i := 1; i <= 15; i++ {
		t = t.AddLast(i)
	}

	testTree(t)
	assertRange(1, 0, With(m, 1, 2).RemoveFirst().RemoveFirst())
	assertRange(1, 15, t)
	assertRange(2, 15, t.RemoveFirst())
	assertRange(1, 14, t.RemoveLast())
	assertSplitRange(1, 10, 15, t, func(m MeasureValue) bool { return m.(int) > 10 })
	assertSplitRange(1, 0, 2, With(m, 1, 2), func(m MeasureValue) bool { return m.(int) > -1 })
	assertSplitRange(1, 0, 2, With(m, 1, 2), func(m MeasureValue) bool { return m.(int) > 0 })
	assertSplitRange(1, 1, 2, With(m, 1, 2), func(m MeasureValue) bool { return m.(int) > 1 })
	assertSplitRange(1, 2, 1, With(m, 1, 2), func(m MeasureValue) bool { return m.(int) > 2 })
	assertSplitRange(1, 2, 1, With(m, 1, 2), func(m MeasureValue) bool { return m.(int) > 3 })
	assertSplitRange(1, 2, 1, With(m, 1, 2), func(m MeasureValue) bool { return m.(int) > 4 })
	t = With(m)
	for i := 1; i <= 100; i++ {
		t = t.AddLast(i)
	}
	for i := 0; i <= 101; i++ {
		items := Items(t.Split(func(m MeasureValue) bool { return m.(int) > i })[0])
		f := t.Find(func(m MeasureValue) bool { return m.(int) > i })
		if i >= 100 {
			assertEqual(100, len(items), "Bad number of items in tree")
		} else {
			assertEqual(i, len(items), "Bad number of items in tree")
		}
		if i == 0 {
			assertEqual(nil, f[0], "Bad first result in find")
		} else if i >= 100 {
			assertEqual(100, f[0], "Bad first result in find")
		} else {
			assertEqual(i, f[0], "Bad first result in find")
		}
		if i >= 100 {
			assertEqual(nil, f[1], "Bad second result in find")
		} else {
			assertEqual(i+1, f[1], "Bad second result in find")
		}
	}
}

func assertSplitRange(start, mid, end int, t Fingertree, pred Predicate) {
	split := t.Split(pred)
	f := t.Find(pred)
	assertRange(start, mid, split[0])
	assertRange(mid+1, end, split[1])
	if mid < start {
		assertEqual(nil, f[0], "Bad find first value")
	} else {
		assertEqual(mid, f[0], "Bad find first value")
	}
	if end < mid+1 {
		assertEqual(nil, f[1], "Bad find second value")
	} else {
		assertEqual(mid+1, f[1], "Bad find second value")
	}
}

func assertRange(start, end int, t Fingertree) {
	items := Items(t)
	dead := func() {
		panic(fmt.Sprintf("Bad tree, expected contiguous items in [%d, %d] but got: %v", start, end, items))
	}
	if len(items) != int(math.Max(0, float64(end-start+1))) {
		dead()
	}
	for _, i := range items {
		if i != start {
			dead()
		}
		start++
	}
}

func assertEqual(expected interface{}, got interface{}, msg string) {
	if expected != got {
		panic(fmt.Sprintf("%s: expected %v but got %v", msg, expected, got))
	}
}

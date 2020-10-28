// Package fingertree implements finger trees in Go.
//
//   Wikipedia: https://en.wikipedia.org/wiki/Finger_tree
//
// The finger tree is an incredibly versatile, customizable data
// structure that functions more or less like a random-access list
// except that, rather than using indexes to access the items, you
// define "measurements" for that. This key distinction provides a
// large amount of power and flexibility. Measurements give you a way
// to define a sort of "width-space" for items, specifying how "wide"
// each item is and accessing items by "offsets in the
// width-space". Measurements can be simple numbers or they can be
// complex structures, allowing you to index items on multiple aspects
// simultaneously (see examples/textLines.go which lets you find lines
// of text by either line number or character offset). I can't say
// enough about how interesting and powerful finger trees are.
//
// Finger trees are reasonably performant although specialized data
// structures will perform better for their targeted tasks.
// Nevertheless, finger trees are very easy to use and I often use
// them for a first-cut before writing a custom data structure -- I've
// found that it's better to debug and maintain fewer pieces of
// complex code than more of them.
//
// This is a modified port of Xueqiao Xu's JavaScript Fingertree code
//   https://github.com/qiao/fingertree.js
//   <xueqiaoxu@gmail.com>
//
// Which is based on:
//   Ralf Hinze and Ross Paterson,
//   "Finger trees: a simple general-purpose data structure"
//   http://www.soi.city.ac.uk/~ross/papers/FingerTree.html
//
//
// COPYRIGHT
//
// © 2020 William R. Burdick Jr. (Bill Burdick) <bill.burdick@gmail.com>
//
// MIT License
//
// Permission is hereby granted, free of charge, to any person
// obtaining a copy of this software and associated documentation
// files (the "Software"), to deal in the Software without
// restriction, including without limitation the rights to use, copy,
// modify, merge, publish, distribute, sublicense, and/or sell copies
// of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS
// BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN
// ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package fingertree

//TreeItem an item in a tree (interface{})
type TreeItem = interface{}

//A function that returns whether a measurement matches
type Predicate = func(MeasureValue) bool

//A function that returns whether a tree item matches
type Code = func(TreeItem) bool

//MeasureValue the result of a measurement. It's a good idea to keep
//these immutable because parts of a Fingertree store MeasureValues.
type MeasureValue interface{}

//Traversable a traversable node (a Fingertree, or node)
type Traversable interface {
	// Measure return the measurement of the tree
	Measure() MeasureValue
	// Each execute c on each item in the tree until it returns false
	// or all the items have been processed. Returns whether all of
	// the items were processed.
	Each(c Code) bool
	// Each execute c on each item in the tree in reverse until it
	// returns false or all the items have been processed. Returns
	// whether all of the items were processed.
	EachReverse(p Code) bool
	findable
}

//Fingertree interface
type Fingertree interface {
	Traversable
	// IsEmpty return whether the tree is empty
	IsEmpty() bool
	// PeekFirst return the first item in the tree or nil if the tree is empty
	PeekFirst() TreeItem
	// PeekLast return the last item in the tree or nil if the tree is empty
	PeekLast() TreeItem
	// AddFirst return a tree containing i, followed by all of this tree's items
	AddFirst(i TreeItem) Fingertree
	// AddLast return a tree containing all of this tree's items followed by i
	AddLast(i TreeItem) Fingertree
	// RemoveFirst return a tree without the first item
	RemoveFirst() Fingertree
	// RemoveLast return a tree without the last item
	RemoveLast() Fingertree
	// Concat return a tree containing all of this tree's items,
	// followed by all of tree's items
	Concat(tree Fingertree) Fingertree
	// Split return two trees, the first one containing all of the
	// initial items that satisfy p and the second containing the
	// items that follow them
	Split(p Predicate) []Fingertree
	// TakeUntil return a tree containing the initial items that do not satisfy p
	TakeUntil(p Predicate) Fingertree
	// DropUntil return a tree with the initial items removed that do not satisfy p
	DropUntil(p Predicate) Fingertree
	// Find returnss a pair with the last item that does not satisfy p
	// and the first item that satisfies p. This much lighter weight
	// than Split()
	Find(p Predicate) []TreeItem
}

//Measurer measures items in a fingertree
type Measurer struct {
	// Identity return a "zero" measure value
	Identity func() MeasureValue
	// Measure return the measurement for i
	Measure func(i TreeItem) MeasureValue
	// Sum return the sum of two measurements
	Sum func(measurement1 MeasureValue, measurement2 MeasureValue) MeasureValue
}

//NewMeasurer create a measurer
func NewMeasurer(identity func() MeasureValue, measure func(TreeItem) MeasureValue, sum func(MeasureValue, MeasureValue) MeasureValue) *Measurer {
	return &Measurer{
		identity,
		measure,
		sum,
	}
}

// With makes a tree for some items
func With(m *Measurer, xs ...interface{}) Fingertree {
	return fromArray(m, xs)
}

//Items returns an array of the items in t
func Items(t Fingertree) []TreeItem {
	var items treeItems

	t.Each(func(item TreeItem) bool {
		items = append(items, item)
		return true
	})
	return items
}

type findable interface {
	first() TreeItem
	last() TreeItem
	find(p Predicate, m MeasureValue, l, r TreeItem) []TreeItem
}

type splittable interface {
	Fingertree
	force() splittable
	splitTree(p Predicate, initial MeasureValue) *treeSplit
}

type treeItems = []interface{}

// the result of a low-level tree treeSplit
type treeSplit struct {
	// elements that do not satisfy the predicate -- either an array or a Fingertree
	left splittable
	// the first element that satisfies the predicate
	mid TreeItem
	// the rest of the elements -- either an array or a Fingertree
	right splittable
}

// the result of a low-level tree split
type digitSplit struct {
	// elements that do not satisfy the predicate -- either an array or a Fingertree
	left treeItems
	// the first element that satisfies the predicate
	mid TreeItem
	// the rest of the elements -- either an array or a Fingertree
	right treeItems
}

//digit is an internal part of a Fingertree, it implements the findable interface
type digit struct {
	measurer    *Measurer
	items       treeItems
	measurement MeasureValue
}

//node is an internal part of a Fingertree, it implements the Traversable interface
type node struct {
	measurer    *Measurer
	items       treeItems
	measurement MeasureValue
}

//empty an empty Fingertree
type empty struct {
	measurer *Measurer
}

//single a Fingertree with one element
type single struct {
	measurer    *Measurer
	item        TreeItem
	measurement MeasureValue
}

//deep a Fingertree with more than one element
type deep struct {
	measurer    *Measurer
	left        *digit
	middle      splittable
	right       *digit
	measurement MeasureValue
}

//delayedFingertree a computed Fingertree
type delayedFingertree struct {
	thunk func() splittable
	tree  splittable
}

func newSplit(left Fingertree, mid TreeItem, right Fingertree) *treeSplit {
	return &treeSplit{left.(splittable), mid, right.(splittable)}
}

func newDigit(measurer *Measurer, items treeItems) *digit {
	m := measurer.Identity()
	for _, item := range items {
		if t, ok := item.(Traversable); ok {
			m = measurer.Sum(m, t.Measure())
		} else {
			m = measurer.Sum(m, measurer.Measure(item))
		}
	}
	return &digit{measurer, items, m}
}

func newNode(measurer *Measurer, items treeItems) *node {
	m := measurer.Identity()
	for _, item := range items {
		if t, ok := item.(Traversable); ok {
			m = measurer.Sum(m, t.Measure())
		} else {
			m = measurer.Sum(m, measurer.Measure(item))
		}
	}
	return &node{measurer, items, m}
}

func newEmpty(mesurer *Measurer) *empty {
	return &empty{mesurer}
}

func newSingle(measurer *Measurer, item TreeItem) *single {
	if t, ok := item.(Traversable); ok {return &single{measurer, item, t.Measure()}
	}
	return &single{measurer, item, measurer.Measure(item)}
}

func newDeep(measurer *Measurer, left *digit, middle splittable, right *digit) *deep {
	if left.count() == 0 {
		panic("Creating deep with empty left!")
	}
	if right.count() == 0 {
		panic("Creating deep with empty right!")
	}
	return &deep{measurer, left, middle, right, nil}
}

func newDelayedFingerTree(f func() splittable) *delayedFingertree {
	return &delayedFingertree{thunk: f}
}

func traverse(items treeItems, c Code) bool {
	for _, item := range items {
		if !traverseItem(item, c) {return false}
	}
	return true
}

func traverseItem(item TreeItem, c Code) bool {
	if t, ok := item.(Traversable); ok {return t.Each(c)}
	return c(item)
}

func traverseReverse(items treeItems, c Code) bool {
	for i := len(items) - 1; i >= 0; i-- {
		if !traverseItemReverse(items[i], c) {return false}
	}
	return true
}

func traverseItemReverse(item TreeItem, c Code) bool {
	if t, ok := item.(Traversable); ok {return t.EachReverse(c)}
	return c(item)
}

func (d *digit) each(c Code) bool            { return traverse(d.items, c) }
func (d *digit) eachReverse(c Code) bool     { return traverseReverse(d.items, c) }
func (d *digit) count() int                  { return len(d.items) }
func (d *digit) first() TreeItem             { return d.items[0] }
func (d *digit) last() TreeItem              { return d.items[len(d.items)-1] }
func (d *digit) removeFirst() *digit         { return d.slice(1, len(d.items)) }
func (d *digit) removeLast() *digit          { return d.slice(0, len(d.items)-1) }
func (d *digit) slice(start, end int) *digit { return newDigit(d.measurer, d.items[start:end]) }
func (d *digit) split(p Predicate, m MeasureValue) *digitSplit {
	var item TreeItem
	i := 0

	if len(d.items) == 1 {return &digitSplit{nil, d.items[0], nil}
	}
	for i, item = range d.items {
		m = d.measurer.Sum(m, d.measurer.Measure(item))
		if p(m) {break}
	}
	return &digitSplit{d.items[0:i], item, d.items[i+1:]}
}
func (d *digit) find(p Predicate, m MeasureValue, l, r TreeItem) []TreeItem {
	for i, item := range d.items {
		newM := d.measurer.Sum(m, d.measurer.Measure(item))
		if p(newM) {
			if i > 0 {
				l = last(d.items[i-1])
			}
			if t, ok := item.(findable); ok {
				if i+1 < len(d.items) {
					r = first(d.items[i+1])
				}
				return t.find(p, m, l, r)
			}
			return []TreeItem{l, item}
		}
		m = newM
	}
	return []TreeItem{last(d), r}
}

func first(item TreeItem) TreeItem {
	for {
		if t, ok := item.(findable); ok {
			item = t.first()
		} else {
			return item
		}
	}
}

func last(item TreeItem) TreeItem {
	for {
		if t, ok := item.(findable); ok {
			item = t.last()
		} else {
			return item
		}
	}
}

//node is a Traversable
func (n *node) Each(c Code) bool        { return traverse(n.items, c) }
func (n *node) EachReverse(c Code) bool { return traverseReverse(n.items, c) }
func (n *node) Measure() MeasureValue   { return n.measurement }
func (n *node) first() TreeItem         { return n.items[0] }
func (n *node) last() TreeItem          { return n.items[len(n.items)-1] }

func (n *node) find(p Predicate, m MeasureValue, l, r TreeItem) []TreeItem {
	for i, item := range n.items {
		newM := n.measurer.Sum(m, n.measurer.Measure(item))
		if p(newM) {
			if i > 0 {
				l = last(n.items[i-1])
			}
			if t, ok := item.(findable); ok {
				if i+1 < len(n.items) {
					r = first(n.items[i+1])
				}
				return t.find(p, m, l, r)
			}
			return []TreeItem{l, item}
		}
		m = newM
	}
	return []TreeItem{last(n), r}
}
func (n *node) toDigit() *digit                                   { return newDigit(n.measurer, n.items) }
func (e *empty) Measure() MeasureValue                            { return e.measurer.Identity() }
func (e *empty) IsEmpty() bool                                    { return true }
func (e *empty) PeekFirst() TreeItem                              { return nil }
func (e *empty) PeekLast() TreeItem                               { return nil }
func (e *empty) first() TreeItem                                  { return nil }
func (e *empty) last() TreeItem                                   { return nil }
func (e *empty) AddFirst(i TreeItem) Fingertree                   { return newSingle(e.measurer, i) }
func (e *empty) AddLast(i TreeItem) Fingertree                    { return newSingle(e.measurer, i) }
func (e *empty) RemoveFirst() Fingertree                          { return e }
func (e *empty) RemoveLast() Fingertree                           { return e }
func (e *empty) Concat(tree Fingertree) Fingertree                { return tree }
func (e *empty) Split(p Predicate) []Fingertree                   { return []Fingertree{e, e} }
func (e *empty) TakeUntil(p Predicate) Fingertree                 { return e }
func (e *empty) DropUntil(p Predicate) Fingertree                 { return e }
func (e *empty) Each(p Code) bool                                 { return true }
func (e *empty) EachReverse(p Code) bool                          { return true }
func (e *empty) force() splittable                                { return e }
func (e *empty) splitTree(p Predicate, i MeasureValue) *treeSplit { return newSplit(e, e, e) }
func (e *empty) Find(p Predicate) []TreeItem                      { return []TreeItem{nil, nil} }
func (e *empty) find(p Predicate, i MeasureValue, l, r TreeItem) []TreeItem {
	return []TreeItem{nil, nil}
}

func (s *single) Measure() MeasureValue { return s.measurement }
func (s *single) IsEmpty() bool         { return false }
func (s *single) first() TreeItem       { return s.item }
func (s *single) last() TreeItem        { return s.item }
func (s *single) PeekFirst() TreeItem   { return s.item }
func (s *single) PeekLast() TreeItem    { return s.item }
func (s *single) AddFirst(item TreeItem) Fingertree {
	return newDeep(s.measurer,
		newDigit(s.measurer, treeItems{item}),
		newEmpty(makeNodeMeasurer(s.measurer)),
		newDigit(s.measurer, treeItems{s.item}))
}
func (s *single) AddLast(item TreeItem) Fingertree {
	return newDeep(s.measurer,
		newDigit(s.measurer, treeItems{s.item}),
		newEmpty(makeNodeMeasurer(s.measurer)),
		newDigit(s.measurer, treeItems{item}))
}
func (s *single) RemoveFirst() Fingertree            { return newEmpty(s.measurer) }
func (s *single) RemoveLast() Fingertree             { return newEmpty(s.measurer) }
func (s *single) Concat(other Fingertree) Fingertree { return other.AddFirst(s.item) }
func (s *single) Split(p Predicate) []Fingertree {
	if p(s.measurement) {return []Fingertree{newEmpty(s.measurer), s}
	}
	return []Fingertree{s, newEmpty(s.measurer)}
}
func (s *single) Find(p Predicate) []TreeItem { return s.find(p, s.measurer.Identity(), nil, nil) }
func (s *single) TakeUntil(p Predicate) Fingertree {
	if p(s.measurement) {return newEmpty(s.measurer)}
	return s
}
func (s *single) DropUntil(p Predicate) Fingertree {
	if p(s.measurement) {return s}
	return newEmpty(s.measurer)
}
func (s *single) Each(p Code) bool        { return traverseItem(s.item, p) }
func (s *single) EachReverse(p Code) bool { return traverseItemReverse(s.item, p) }
func (s *single) force() splittable       { return s }
func (s *single) splitTree(p Predicate, initial MeasureValue) *treeSplit {
	return newSplit(newEmpty(s.measurer), s.item, newEmpty(s.measurer))
}
func (s *single) find(p Predicate, i MeasureValue, l, r TreeItem) []TreeItem {
	if p(s.measurement) {
		if t, ok := s.item.(findable); ok {return t.find(p, i, l, r)}
		return []TreeItem{l, s.item}
	}
	return []TreeItem{last(s.item), r}
}

func (d *deep) Measure() MeasureValue {
	if d.measurement == nil {
		m := d.measurer
		d.measurement = m.Sum(m.Sum(d.left.measurement, d.middle.Measure()), d.right.measurement)
	}
	return d.measurement
}
func (d *deep) IsEmpty() bool       { return false }
func (d *deep) PeekFirst() TreeItem { return d.left.first() }
func (d *deep) PeekLast() TreeItem  { return d.right.last() }
func (d *deep) first() TreeItem     { return d.left.first() }
func (d *deep) last() TreeItem      { return d.right.last() }
func (d *deep) force() splittable   { return d }
func (d *deep) AddFirst(item TreeItem) Fingertree {
	if d.left.count() == 4 {return newDeep(d.measurer,
			newDigit(d.measurer, treeItems{item, d.left.items[0]}),
			d.middle.AddFirst(newNode(d.measurer, d.left.items[1:])).(splittable),
			d.right)
	}
	return newDeep(d.measurer, newDigit(d.measurer, append(treeItems{item}, d.left.items...)), d.middle, d.right)
}
func (d *deep) AddLast(item TreeItem) Fingertree {
	if d.right.count() == 4 {
		return newDeep(d.measurer,
			d.left,
			d.middle.AddLast(newNode(d.measurer, d.right.items[0:3])).(splittable),
			newDigit(d.measurer, treeItems{d.right.items[3], item}))
	}
	return newDeep(d.measurer,
		d.left,
		d.middle,
		newDigit(d.measurer, append(d.right.items, item)))
}
func (d *deep) RemoveFirst() Fingertree {
	if d.left.count() > 1 {return newDeep(d.measurer, d.left.removeFirst(), d.middle, d.right)}
	if !d.middle.IsEmpty() {
		newMid := newDelayedFingerTree(func() splittable { return d.middle.RemoveFirst().(splittable) })
		return newDeep(d.measurer, d.middle.PeekFirst().(*node).toDigit(), newMid, d.right)
	}
	if d.right.count() == 1 {return newSingle(d.measurer, d.right.items[0])}
	return newDeep(d.measurer, d.right.slice(0, 1), d.middle, d.right.removeFirst())
}
func (d *deep) RemoveLast() Fingertree {
	if d.right.count() > 1 {return newDeep(d.measurer, d.left, d.middle, d.right.removeLast())}
	if !d.middle.IsEmpty() {
		newMid := newDelayedFingerTree(func() splittable { return d.middle.RemoveLast().(splittable) })
		return newDeep(d.measurer, d.left, newMid, d.middle.PeekLast().(*node).toDigit())
	}
	l := d.left
	if l.count() == 1 {return newSingle(d.measurer, d.left.items[0])}
	return newDeep(d.measurer, l.removeLast(), d.middle, l.slice(l.count()-1, l.count()))
}
func (d *deep) Concat(other Fingertree) Fingertree {
	other = other.(splittable).force()
	if _, ok := other.(*empty); ok {return d}
	if o, ok := other.(*single); ok {return d.AddLast(o.item)}
	return app3(d, nil, other.(splittable))
}
func (d *deep) splitTree(p Predicate, initial MeasureValue) *treeSplit {
	leftMeasure := d.measurer.Sum(initial, d.left.measurement)
	if p(leftMeasure) {
		dsplit := d.left.split(p, initial)
		return newSplit(fromArray(d.measurer, dsplit.left),
			dsplit.mid,
			deepLeft(d.measurer, dsplit.right, d.middle, d.right))
	}
	midMeasure := d.measurer.Sum(leftMeasure, d.middle.Measure())
	if p(midMeasure) {
		midSplit := d.middle.splitTree(p, leftMeasure)
		split := midSplit.mid.(*node).toDigit().split(p, d.measurer.Sum(leftMeasure, midSplit.left.(Fingertree).Measure()))
		return newSplit(deepRight(d.measurer, d.left, midSplit.left, split.left),
			split.mid,
			deepLeft(d.measurer, split.right, midSplit.right, d.right))
	}
	dsplit := d.right.split(p, midMeasure)
	return newSplit(deepRight(d.measurer, d.left, d.middle, dsplit.left),
		dsplit.mid,
		fromArray(d.measurer, dsplit.right))
}
func (d *deep) Split(p Predicate) []Fingertree {
	if p(d.Measure()) {
		split := d.splitTree(p, d.measurer.Identity())
		return []Fingertree{split.left, split.right.AddFirst(split.mid)}
	}
	return []Fingertree{d, newEmpty(d.measurer)}
}
func (d *deep) Find(p Predicate) []TreeItem { return d.find(p, d.measurer.Identity(), nil, nil) }
func (d *deep) find(p Predicate, i MeasureValue, l, r TreeItem) []TreeItem {
	leftMeasure := d.measurer.Sum(i, d.left.measurement)
	if p(leftMeasure) {return d.left.find(p, i, l, first(d.middle))}
	midMeasure := d.measurer.Sum(leftMeasure, d.middle.Measure())
	l = last(d.left)
	if p(midMeasure) {return d.middle.find(p, leftMeasure, l, first(d.right))}
	if !d.middle.IsEmpty() {
		l = last(d.middle)
	}
	return d.right.find(p, midMeasure, l, r)
}
func (d *deep) TakeUntil(p Predicate) Fingertree {
	return d.Split(p)[0]
}
func (d *deep) DropUntil(p Predicate) Fingertree {
	return d.Split(p)[1]
}

func (d *deep) Each(p Code) bool {
	if !d.left.each(p) {return false}
	if !d.middle.Each(p) {return false}
	return d.right.each(p)
}
func (d *deep) EachReverse(p Code) bool {
	if !d.right.eachReverse(p) {return false}
	if !d.middle.EachReverse(p) {return false}
	return d.left.eachReverse(p)
}

func (d *delayedFingertree) Measure() MeasureValue            { return d.force().Measure() }
func (d *delayedFingertree) IsEmpty() bool                    { return d.force().IsEmpty() }
func (d *delayedFingertree) PeekFirst() TreeItem              { return d.force().PeekFirst() }
func (d *delayedFingertree) PeekLast() TreeItem               { return d.force().PeekLast() }
func (d *delayedFingertree) first() TreeItem                  { return d.force().first() }
func (d *delayedFingertree) last() TreeItem                   { return d.force().last() }
func (d *delayedFingertree) AddFirst(i TreeItem) Fingertree   { return d.force().AddFirst(i) }
func (d *delayedFingertree) AddLast(i TreeItem) Fingertree    { return d.force().AddLast(i) }
func (d *delayedFingertree) RemoveFirst() Fingertree          { return d.force().RemoveFirst() }
func (d *delayedFingertree) RemoveLast() Fingertree           { return d.force().RemoveLast() }
func (d *delayedFingertree) Concat(t Fingertree) Fingertree   { return d.force().Concat(t) }
func (d *delayedFingertree) Split(p Predicate) []Fingertree   { return d.force().Split(p) }
func (d *delayedFingertree) TakeUntil(p Predicate) Fingertree { return d.force().TakeUntil(p) }
func (d *delayedFingertree) DropUntil(p Predicate) Fingertree { return d.force().DropUntil(p) }
func (d *delayedFingertree) Each(p Code) bool                 { return d.force().Each(p) }
func (d *delayedFingertree) EachReverse(p Code) bool          { return d.force().EachReverse(p) }

func (d *delayedFingertree) Find(p Predicate) []TreeItem { return d.force().Find(p) }
func (d *delayedFingertree) find(p Predicate, i MeasureValue, l, r TreeItem) []TreeItem {
	return d.force().find(p, i, l, r)
}

func (d *delayedFingertree) force() splittable {
	if d.tree == nil {
		d.tree = d.thunk()
	}
	return d.tree
}

func (d *delayedFingertree) splitTree(p Predicate, i MeasureValue) *treeSplit {
	return d.force().splitTree(p, i)
}

func deepLeft(m *Measurer, left treeItems, mid splittable, right *digit) splittable {
	if len(left) > 0 {return newDeep(m, newDigit(m, left), mid, right)}
	if mid.IsEmpty() {return fromArray(m, right.items)}
	return newDelayedFingerTree(func() splittable {
		return newDeep(m, mid.PeekFirst().(*node).toDigit(), mid.RemoveFirst().(splittable), right)
	})
}

func deepRight(m *Measurer, left *digit, mid splittable, right treeItems) splittable {
	if len(right) > 0 {return newDeep(m, left, mid, newDigit(m, right))}
	if mid.IsEmpty() {return fromArray(m, left.items)}
	return newDelayedFingerTree(func() splittable {
		return newDeep(m, left, mid.RemoveLast().(splittable), mid.PeekLast().(*node).toDigit())
	})
}

//concatenate two fingertrees with additional elements in between
func app3(t1 splittable, items treeItems, t2 splittable) splittable {
	t1 = t1.force()
	t2 = t2.force()
	if _, ok := t1.(*empty); ok {return prependItems(t2, items)}
	if _, ok := t2.(*empty); ok {return appendItems(t1, items)}
	if s, ok := t1.(*single); ok {return prependItems(t2, items).AddFirst(s.item).(splittable)}
	if s, ok := t2.(*single); ok {return appendItems(t1, items).AddLast(s.item).(splittable)}
	d1 := t1.(*deep)
	d2 := t2.(*deep)
	return newDeep(d1.measurer,
		d1.left,
		newDelayedFingerTree(func() splittable {
			newNodes := make(treeItems, 0, len(d1.right.items)+len(items)+len(d2.left.items))
			return app3(d1.middle,
				nodes(d1.measurer,
					append(append(append(newNodes, d1.right.items...), items...), d1.left.items...),
					nil),
				d2.middle)
		}),
		d2.right)
}

func nodes(m *Measurer, xs treeItems, result treeItems) treeItems {
	switch len(xs) {
	case 2, 3:
		return append(result, newNode(m, xs))
	case 4:
		return append(result, newNode(m, xs[0:2]), newNode(m, xs[2:]))
	default:
		return nodes(m, xs[3:], append(result, newNode(m, xs[0:3])))
	}
}

func prependItems(tree Fingertree, items treeItems) splittable {
	for i := len(items) - 1; i >= 0; i-- {
		tree = tree.AddFirst(items[i])
	}
	return tree.(splittable)
}

func appendItems(tree Fingertree, items treeItems) splittable {
	for _, i := range items {
		tree = tree.AddLast(i)
	}
	return tree.(splittable)
}

func fromArray(m *Measurer, xs treeItems) splittable {
	return prependItems(newEmpty(m), xs)
}

func makeNodeMeasurer(measurer *Measurer) *Measurer {
	return NewMeasurer(measurer.Identity, func(n TreeItem) MeasureValue {
		return n.(Traversable).Measure()
	}, measurer.Sum)
}

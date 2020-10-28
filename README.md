# fingertree

[![GoDoc](https://img.shields.io/badge/pkg.go.dev-doc-blue)](http://pkg.go.dev/github.com/zot/go-fingertree)

Package fingertree implements finger trees in Go.

```
Wikipedia: [https://en.wikipedia.org/wiki/Finger_tree](https://en.wikipedia.org/wiki/Finger_tree)
```

The finger tree is an incredibly versatile, customizable data
structure that functions more or less like a random-access list
except that, rather than using indexes to access the items, you
define "measurements" for that. This key distinction provides a
large amount of power and flexibility. Measurements give you a way
to define a sort of "width-space" for items, specifying how "wide"
each item is and accessing items by "offsets in the
width-space". Measurements can be simple numbers or they can be
complex structures, allowing you to index items on multiple aspects
simultaneously (see examples/textLines.go which lets you find lines
of text by either line number or character offset). I can't say
enough about how interesting and powerful finger trees are.

Finger trees are reasonably performant although specialized data
structures will perform better for their targeted tasks.
Nevertheless, finger trees are very easy to use and I often use
them for a first-cut before writing a custom data structure -- I've
found that it's better to debug and maintain fewer pieces of
complex code than more of them.

This is a modified port of Xueqiao Xu's JavaScript Fingertree code

```
[https://github.com/qiao/fingertree.js](https://github.com/qiao/fingertree.js)
<xueqiaoxu@gmail.com>
```

Which is based on:

```
Ralf Hinze and Ross Paterson,
"Finger trees: a simple general-purpose data structure"
[http://www.soi.city.ac.uk/~ross/papers/FingerTree.html](http://www.soi.city.ac.uk/~ross/papers/FingerTree.html)
```

## COPYRIGHT

© 2020 William R. Burdick Jr. (Bill Burdick) <bill.burdick@gmail.com>

## MIT License

Permission is hereby granted, free of charge, to any person
obtaining a copy of this software and associated documentation
files (the "Software"), to deal in the Software without
restriction, including without limitation the rights to use, copy,
modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS
BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN
ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

## Sub Packages

* [examples](./examples): Text lines example: tracks text offsets by both line and character

* [test](./test)

---
Readme created from Go doc with [goreadme](https://github.com/posener/goreadme)

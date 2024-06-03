// Modified from go built-in ring
// + Add generic for type Ring and its functions

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ring implements operations on circular lists.
package limiter

// A Ring is an element of a circular list, or ring.
// Rings do not have a beginning or end; a pointer to any ring element
// serves as reference to the entire ring. Empty rings are represented
// as nil Ring pointers. The zero value for a Ring is a one-element
// ring with a nil Value.
type Ring[T comparable] struct {
	n, p  *Ring[T]
	Value T // for use by client; untouched by this library
}

func (r *Ring[T]) init() *Ring[T] {
	r.n = r
	r.p = r
	return r
}

// next returns the next ring element. r must not be empty.
func (r *Ring[T]) next() *Ring[T] {
	if r.n == nil {
		return r.init()
	}
	return r.n
}

// prev returns the previous ring element. r must not be empty.
func (r *Ring[T]) prev() *Ring[T] {
	if r.n == nil {
		return r.init()
	}
	return r.p
}

// move moves n % r.Len() elements backward (n < 0) or forward (n >= 0)
// in the ring and returns that ring element. r must not be empty.
func (r *Ring[T]) move(n int) *Ring[T] {
	if r.n == nil {
		return r.init()
	}
	switch {
	case n < 0:
		for ; n < 0; n++ {
			r = r.p
		}
	case n > 0:
		for ; n > 0; n-- {
			r = r.n
		}
	}
	return r
}

// newRing creates a ring of n elements.
func newRing[T comparable](n int) *Ring[T] {
	if n <= 0 {
		return nil
	}
	r := new(Ring[T])
	p := r
	for i := 1; i < n; i++ {
		p.n = &Ring[T]{p: p}
		p = p.n
	}
	p.n = r
	r.p = p
	return r
}

// link connects ring r with ring s such that r.Next()
// becomes s and returns the original value for r.Next().
// r must not be empty.
//
// If r and s point to the same ring, linking
// them removes the elements between r and s from the ring.
// The removed elements form a subring and the result is a
// reference to that subring (if no elements were removed,
// the result is still the original value for r.Next(),
// and not nil).
//
// If r and s point to different rings, linking
// them creates a single ring with the elements of s inserted
// after r. The result points to the element following the
// last element of s after insertion.
func (r *Ring[T]) link(s *Ring[T]) *Ring[T] {
	n := r.next()
	if s != nil {
		p := s.prev()
		// Note: Cannot use multiple assignment because
		// evaluation order of LHS is not specified.
		r.n = s
		s.p = r
		n.p = p
		p.n = n
	}
	return n
}

// unlink removes n % r.Len() elements from the ring r, starting
// at r.Next(). If n % r.Len() == 0, r remains unchanged.
// The result is the removed subring. r must not be empty.
func (r *Ring[T]) unlink(n int) *Ring[T] {
	if n <= 0 {
		return nil
	}
	return r.link(r.move(n + 1))
}

// len computes the number of elements in ring r.
// It executes in time proportional to the number of elements.
func (r *Ring[T]) len() int {
	n := 0
	if r != nil {
		n = 1
		for p := r.next(); p != r; p = p.n {
			n++
		}
	}
	return n
}

// do calls function f on each element of the ring, in forward order.
// The behavior of do is undefined if f changes *r.
func (r *Ring[T]) do(f func(any)) {
	if r != nil {
		f(r.Value)
		for p := r.next(); p != r; p = p.n {
			f(p.Value)
		}
	}
}

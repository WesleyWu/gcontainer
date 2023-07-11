// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with l file,
// You can obtain one at https://github.com/gogf/gf.
//

// Package glist provides most commonly used doubly linked list container which also supports concurrent-safe/unsafe switch feature.
package glist

import (
	"bytes"
	json2 "encoding/json"

	"github.com/wesleywu/gcontainer/garray"
	"github.com/wesleywu/gcontainer/internal/deepcopy"
	"github.com/wesleywu/gcontainer/internal/json"
	"github.com/wesleywu/gcontainer/internal/rwmutex"
	"github.com/wesleywu/gcontainer/utils/gconv"
)

// List represents a doubly linked list.
// The zero value for List is an empty list ready to use.
type List[T comparable] struct {
	mu   rwmutex.RWMutex
	root Element[T] // sentinel list element, only &root, root.prev, and root.next are used
	len  int        // current list length excluding (this) sentinel element
}

// Element is an element of a linked list.
type Element[T comparable] struct {
	// Next and previous pointers in the doubly-linked list of elements.
	// To simplify the implementation, internally a list l is implemented
	// as a ring, such that &l.root is both the next element of the last
	// list element (l.Back()) and the previous element of the first list
	// element (l.Front()).
	next, prev *Element[T]

	// The list to which this element belongs.
	list *List[T]

	// The value stored with this element.
	Value T
}

// Init initializes or clears list l.
func (l *List[T]) Init() *List[T] {
	l.root.next = &l.root
	l.root.prev = &l.root
	l.len = 0
	return l
}

// New returns an initialized list.
func New[T comparable](safe ...bool) *List[T] {
	l := new(List[T]).Init()
	l.mu = rwmutex.Create(safe...)
	return l
}

// NewFrom creates and returns a list from a copy of given slice `array`.
// The parameter `safe` is used to specify whether using list in concurrent-safety,
// which is false in default.
func NewFrom[T comparable](array []T, safe ...bool) *List[T] {
	l := New[T](safe...)
	for _, v := range array {
		l.PushBack(v)
	}
	return l
}

// Add append a new element e with value v at the back of list l and returns true.
func (l *List[T]) Add(values ...T) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	l.lazyInit()
	for _, value := range values {
		_ = l.insertValue(value, l.root.prev)
	}
	return true
}

// AddAll adds all the elements in the specified collection to this list.
// Returns true if this collection changed as a result of the call
func (l *List[T]) AddAll(values garray.Collection[T]) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	l.lazyInit()
	values.ForEach(func(value T) bool {
		_ = l.insertValue(value, l.root.prev)
		return true
	})
	return true
}

// Contains returns true if this collection contains the specified element.
func (l *List[T]) Contains(value T) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	l.lazyInit()
	found := false
	length := l.len
	if length > 0 {
		for i, e := 0, l.root.next; i < length; i, e = i+1, e.Next() {
			if e.Value == value {
				found = true
				break
			}
		}
	}
	return found
}

// ContainsAll returns true if this collection contains all the elements in the specified collection.
func (l *List[T]) ContainsAll(values garray.Collection[T]) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	l.lazyInit()
	foundMap := make(map[T]bool, 0)
	values.ForEach(func(t T) bool {
		foundMap[t] = false
		return true
	})
	length := l.len
	if length > 0 {
		for i, e := 0, l.root.next; i < length; i, e = i+1, e.Next() {
			if _, ok := foundMap[e.Value]; ok {
				foundMap[e.Value] = true
				break
			}
		}
	}
	for _, found := range foundMap {
		if !found {
			return false
		}
	}
	return true
}

// ForEach iterates all elements in this collection readonly with custom callback function `f`.
// If `f` returns true, then it continues iterating; or false to stop.
func (l *List[T]) ForEach(f func(T) bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	l.lazyInit()
	length := l.len
	if length > 0 {
		for i, e := 0, l.root.next; i < length; i, e = i+1, e.Next() {
			if !f(e.Value) {
				break
			}
		}
	}
}

// IsEmpty returns true if this collection contains no elements.
func (l *List[T]) IsEmpty() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	l.lazyInit()
	return l.len == 0
}

// Slice returns an array containing shadow copy of all the elements in this list.
func (l *List[T]) Slice() []T {
	return l.FrontAll()
}

// search returns the matching element in this list, or nil if the element can not be found.
func (l *List[T]) search(value T) *Element[T] {
	if l.len > 0 {
		for i, e := 0, l.root.next; i < l.len; i, e = i+1, e.Next() {
			if e.Value == value {
				return e
			}
		}
	}
	return nil
}

// Next returns the next list element or nil.
func (e *Element[T]) Next() *Element[T] {
	if p := e.next; e.list != nil && p != &e.list.root {
		return p
	}
	return nil
}

// Prev returns the previous list element or nil.
func (e *Element[T]) Prev() *Element[T] {
	if p := e.prev; e.list != nil && p != &e.list.root {
		return p
	}
	return nil
}

// Len returns the number of elements of list l.
// The complexity is O(1).
func (l *List[T]) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.len
}

// Size is alias of Len.
func (l *List[T]) Size() int {
	return l.Len()
}

// Front returns the first element of list l or nil if the list is empty.
func (l *List[T]) Front() *Element[T] {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.len == 0 {
		return nil
	}
	return l.root.next
}

// Back returns the last element of list l or nil if the list is empty.
func (l *List[T]) Back() *Element[T] {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}

// lazyInit lazily initializes a zero List value.
func (l *List[T]) lazyInit() {
	if l.root.next == nil {
		l.Init()
	}
}

// insert inserts e after at, increments l.len, and returns e.
func (l *List[T]) insert(e, at *Element[T]) *Element[T] {
	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
	e.list = l
	l.len++
	return e
}

// insertValue is a convenience wrapper for insert(&Element{Value: v}, at).
func (l *List[T]) insertValue(v T, at *Element[T]) *Element[T] {
	return l.insert(&Element[T]{Value: v}, at)
}

// remove removes e from its list, decrements l.len
func (l *List[T]) remove(e *Element[T]) {
	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = nil // avoid memory leaks
	e.prev = nil // avoid memory leaks
	e.list = nil
	l.len--
}

// move moves e to next to at.
func (l *List[T]) move(e, at *Element[T]) {
	if e == at {
		return
	}
	e.prev.next = e.next
	e.next.prev = e.prev

	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
}

// Remove removes all of this list's elements that are also contained in the specified slice
// if it is present.
// Returns true if this collection changed as a result of the call
func (l *List[T]) Remove(values ...T) (changed bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	changed = false
	for _, value := range values {
		existing := l.search(value)
		if existing != nil {
			l.remove(existing)
			changed = true
		}
	}
	return
}

// RemoveAll removes all of this list's elements that are also contained in the specified collection
// Returns true if this collection changed as a result of the call
func (l *List[T]) RemoveAll(values garray.Collection[T]) (changed bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	changed = false
	values.ForEach(func(value T) bool {
		existing := l.search(value)
		if existing != nil {
			l.remove(existing)
			changed = true
		}
		return true
	})
	return
}

// PushBack inserts a new element e with value v at the back of list l and returns e.
func (l *List[T]) PushBack(v T) *Element[T] {
	l.mu.RLock()
	defer l.mu.RUnlock()
	l.lazyInit()
	return l.insertValue(v, l.root.prev)
}

// PushFront inserts a new element e with value v at the front of list l and returns e.
func (l *List[T]) PushFront(v T) *Element[T] {
	l.mu.RLock()
	defer l.mu.RUnlock()
	l.lazyInit()
	return l.insertValue(v, &l.root)
}

// PushBacks inserts multiple new elements with values `values` at the back of list `l`.
func (l *List[T]) PushBacks(values []T) {
	l.mu.Lock()
	l.mu.Unlock()
	l.lazyInit()
	for _, v := range values {
		l.PushBack(v)
	}
}

// PushFronts inserts multiple new elements with values `values` at the front of list `l`.
func (l *List[T]) PushFronts(values []T) {
	l.mu.Lock()
	defer l.mu.RUnlock()
	l.lazyInit()
	for _, v := range values {
		l.PushFront(v)
	}
}

// PopBack removes the element from back of `l` and returns the value of the element.
func (l *List[T]) PopBack() (value T) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lazyInit()
	if e := l.root.prev; e != nil {
		value = e.Value
		if e.list == l {
			// if e.list == l, l must have been initialized when e was inserted
			// in l or l == nil (e is a zero Element) and l.remove will crash
			l.remove(e)
		}
	}
	return
}

// PopFront removes the element from front of `l` and returns the value of the element.
func (l *List[T]) PopFront() (value T) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lazyInit()
	if e := l.root.next; e != nil {
		value = e.Value
		if e.list == l {
			// if e.list == l, l must have been initialized when e was inserted
			// in l or l == nil (e is a zero Element) and l.remove will crash
			l.remove(e)
		}
	}
	return
}

// PopBacks removes `max` elements from back of `l`
// and returns values of the removed elements as slice.
func (l *List[T]) PopBacks(max int) (values []T) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lazyInit()
	length := l.Len()
	if length > 0 {
		if max > 0 && max < length {
			length = max
		}
		values = make([]T, length)
		for i := 0; i < length; i++ {
			back := l.root.prev
			values[i] = back.Value
			if back.list == l {
				// if e.list == l, l must have been initialized when e was inserted
				// in l or l == nil (e is a zero Element) and l.remove will crash
				l.remove(back)
			}
		}
	}
	return
}

// PopFronts removes `max` elements from front of `l`
// and returns values of the removed elements as slice.
func (l *List[T]) PopFronts(max int) (values []T) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lazyInit()
	length := l.Len()
	if length > 0 {
		if max > 0 && max < length {
			length = max
		}
		values = make([]T, length)
		for i := 0; i < length; i++ {
			front := l.root.next
			values[i] = front.Value
			if front.list == l {
				// if e.list == l, l must have been initialized when e was inserted
				// in l or l == nil (e is a zero Element) and l.remove will crash
				l.remove(front)
			}
		}
	}
	return
}

// PopBackAll removes all elements from back of `l`
// and returns values of the removed elements as slice.
func (l *List[T]) PopBackAll() []T {
	return l.PopBacks(-1)
}

// PopFrontAll removes all elements from front of `l`
// and returns values of the removed elements as slice.
func (l *List[T]) PopFrontAll() []T {
	return l.PopFronts(-1)
}

// FrontAll copies and returns values of all elements from front of `l` as slice.
func (l *List[T]) FrontAll() (values []T) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	l.lazyInit()
	length := l.Len()
	if length > 0 {
		values = make([]T, length)
		for i, e := 0, l.root.next; i < length; i, e = i+1, e.Next() {
			values[i] = e.Value
		}
	}
	return
}

// BackAll copies and returns values of all elements from back of `l` as slice.
func (l *List[T]) BackAll() (values []T) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	l.lazyInit()
	length := l.Len()
	if length > 0 {
		values = make([]T, length)
		for i, e := 0, l.root.prev; i < length; i, e = i+1, e.Prev() {
			values[i] = e.Value
		}
	}
	return
}

// FrontValue returns value of the first element of `l` or nil if the list is empty.
func (l *List[T]) FrontValue() (value T) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	l.lazyInit()
	if e := l.root.next; e != nil {
		value = e.Value
	}
	return
}

// BackValue returns value of the last element of `l` or nil if the list is empty.
func (l *List[T]) BackValue() (value T) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	l.lazyInit()
	if e := l.root.prev; e != nil {
		value = e.Value
	}
	return
}

// InsertBefore inserts a new element e with value v immediately before mark and returns e.
// If mark is not an element of l, the list is not modified.
// The mark must not be nil.
func (l *List[T]) InsertBefore(mark *Element[T], v T) *Element[T] {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if mark.list != l {
		return nil
	}
	// see comment in List.Remove about initialization of l
	return l.insertValue(v, mark.prev)
}

// InsertAfter inserts a new element e with value v immediately after mark and returns e.
// If mark is not an element of l, the list is not modified.
// The mark must not be nil.
func (l *List[T]) InsertAfter(mark *Element[T], v T) *Element[T] {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if mark.list != l {
		return nil
	}
	// see comment in List.Remove about initialization of l
	return l.insertValue(v, mark)
}

// MoveToFront moves element e to the front of list l.
// If e is not an element of l, the list is not modified.
// The element must not be nil.
func (l *List[T]) MoveToFront(e *Element[T]) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if e.list != l || l.root.next == e {
		return
	}
	// see comment in List.Remove about initialization of l
	l.move(e, &l.root)
}

// MoveToBack moves element e to the back of list l.
// If e is not an element of l, the list is not modified.
// The element must not be nil.
func (l *List[T]) MoveToBack(e *Element[T]) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if e.list != l || l.root.prev == e {
		return
	}
	// see comment in List.Remove about initialization of l
	l.move(e, l.root.prev)
}

// MoveBefore moves element e to its new position before mark.
// If e or mark is not an element of l, or e == mark, the list is not modified.
// The element and mark must not be nil.
func (l *List[T]) MoveBefore(e, mark *Element[T]) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if e.list != l || e == mark || mark.list != l {
		return
	}
	l.move(e, mark.prev)
}

// MoveAfter moves element e to its new position after mark.
// If e or mark is not an element of l, or e == mark, the list is not modified.
// The element and mark must not be nil.
func (l *List[T]) MoveAfter(e, mark *Element[T]) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if e.list != l || e == mark || mark.list != l {
		return
	}
	l.move(e, mark)
}

// PushBackList inserts a copy of another list at the back of list l.
// The lists l and other may be the same. They must not be nil.
func (l *List[T]) PushBackList(other *List[T]) {
	if l != other {
		other.mu.RLock()
		defer other.mu.RUnlock()
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	l.lazyInit()
	for i, e := other.len, other.root.next; i > 0; i, e = i-1, e.Next() {
		l.insertValue(e.Value, l.root.prev)
	}
}

// PushFrontList inserts a copy of another list at the front of list l.
// The lists l and other may be the same. They must not be nil.
func (l *List[T]) PushFrontList(other *List[T]) {
	if l != other {
		other.mu.RLock()
		defer other.mu.RUnlock()
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lazyInit()
	for i, e := other.len, other.root.prev; i > 0; i, e = i-1, e.Prev() {
		l.insertValue(e.Value, &l.root)
	}
}

// Clear removes all the elements from this collection.
func (l *List[T]) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Init()
}

func (l *List[T]) Clone() garray.Collection[T] {
	l.mu.RLock()
	defer l.mu.RUnlock()
	l.lazyInit()
	values := make([]T, l.len)
	for i, e := 0, l.root.next; i < l.len; i, e = i+1, e.Next() {
		values[i] = e.Value
	}
	return NewFrom(values, l.mu.IsSafe())
}

func (l *List[T]) Equals(another garray.Collection[T]) bool {
	if l == another {
		return true
	}
	var (
		ano *List[T]
		ok  bool
	)
	if ano, ok = another.(*List[T]); !ok {
		return false
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	ano.mu.RLock()
	defer ano.mu.RUnlock()
	if l.len != ano.len {
		return false
	}
	values := make([]T, l.len)
	for i, e := 0, l.root.next; i < l.len; i, e = i+1, e.Next() {
		values[i] = e.Value
	}
	valuesAno := make([]T, l.len)
	for i, e := 0, ano.root.next; i < ano.len; i, e = i+1, e.Next() {
		valuesAno[i] = e.Value
	}
	for i := 0; i < l.len; i++ {
		if values[i] != valuesAno[i] {
			return false
		}
	}
	return true
}

// Iterator is alias of IteratorAsc.
func (l *List[T]) Iterator(f func(e *Element[T]) bool) {
	l.IteratorAsc(f)
}

// IteratorAsc iterates the list readonly in ascending order with given callback function `f`.
// If `f` returns true, then it continues iterating; or false to stop.
func (l *List[T]) IteratorAsc(f func(e *Element[T]) bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	l.lazyInit()
	length := l.Len()
	if length > 0 {
		for i, e := 0, l.root.next; i < length; i, e = i+1, e.Next() {
			if !f(e) {
				break
			}
		}
	}
}

// IteratorDesc iterates the list readonly in descending order with given callback function `f`.
// If `f` returns true, then it continues iterating; or false to stop.
func (l *List[T]) IteratorDesc(f func(e *Element[T]) bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	l.lazyInit()
	length := l.Len()
	if length > 0 {
		for i, e := 0, l.root.prev; i < length; i, e = i+1, e.Prev() {
			if !f(e) {
				break
			}
		}
	}
}

// Join joins list elements with a string `glue`.
func (l *List[T]) Join(glue string) string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	l.lazyInit()
	buffer := bytes.NewBuffer(nil)
	for i, e := 0, l.root.next; i < l.len; i, e = i+1, e.Next() {
		buffer.WriteString(gconv.String(e.Value))
		if i != l.len-1 {
			buffer.WriteString(glue)
		}
	}
	return buffer.String()
}

// String returns current list as a string.
func (l *List[T]) String() string {
	l.lazyInit()
	return "[" + l.Join(",") + "]"
}

// Sum returns the sum of values in an array.
func (l *List[T]) Sum() (sum int) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	for i, e := 0, l.root.next; i < l.len; i, e = i+1, e.Next() {
		sum += gconv.Int(e.Value)
	}
	return
}

// MarshalJSON implements the interface MarshalJSON for json.Marshal.
func (l List[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.FrontAll())
}

// UnmarshalJSON implements the interface UnmarshalJSON for json.Unmarshal.
func (l *List[T]) UnmarshalJSON(b []byte) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lazyInit()
	var array []T
	if err := json.UnmarshalUseNumber(b, &array); err != nil {
		return err
	}
	for _, v := range array {
		l.PushBack(v)
	}
	return nil
}

// UnmarshalValue is an interface implement which sets any type of value for list.
func (l *List[T]) UnmarshalValue(value interface{}) (err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lazyInit()
	var array []T
	switch value.(type) {
	case string, []byte, json2.Number:
		err = json.UnmarshalUseNumber(gconv.Bytes(value), &array)
	default:
		array = gconv.SliceAny[T](value)
	}
	for _, v := range array {
		l.PushBack(v)
	}
	return err
}

// DeepCopy implements interface for deep copy of current type.
func (l *List[T]) DeepCopy() garray.Collection[T] {
	if l == nil {
		return nil
	}

	l.mu.RLock()
	defer l.mu.RUnlock()

	var (
		length = l.Len()
		values = make([]T, length)
	)
	if length > 0 {
		for i, e := 0, l.root.next; i < length; i, e = i+1, e.Next() {
			values[i] = deepcopy.Copy(e.Value).(T)
		}
	}
	return NewFrom[T](values, l.mu.IsSafe())
}

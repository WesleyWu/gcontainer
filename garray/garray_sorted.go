// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package garray

import (
	"bytes"
	json2 "encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/wesleywu/gcontainer/internal/deepcopy"
	"github.com/wesleywu/gcontainer/internal/json"
	"github.com/wesleywu/gcontainer/internal/rwmutex"
	"github.com/wesleywu/gcontainer/utils/comparator"
	"github.com/wesleywu/gcontainer/utils/empty"
	"github.com/wesleywu/gcontainer/utils/gconv"
	"github.com/wesleywu/gcontainer/utils/grand"
	"github.com/wesleywu/gcontainer/utils/gstr"
)

// SortedArray is a golang sorted array with rich features.
// It is using increasing order in default, which can be changed by
// setting it a custom comparator.
// It contains a concurrent-safe/unsafe switch, which should be set
// when its initialization and cannot be changed then.
type SortedArray[T comparable] struct {
	mu         rwmutex.RWMutex
	array      []T
	unique     bool                     // Whether enable unique feature(false)
	comparator comparator.Comparator[T] // Comparison function(it returns -1: a < b; 0: a == b; 1: a > b)
}

// NewSortedArray creates and returns an empty sorted array.
// The parameter `safe` is used to specify whether using array in concurrent-safety, which is false in default.
// The parameter `comparator` used to compare values to sort in array,
// if it returns value < 0, means `a` < `b`; the `a` will be inserted before `b`;
// if it returns value = 0, means `a` = `b`; the `a` will be replaced by     `b`;
// if it returns value > 0, means `a` > `b`; the `a` will be inserted after  `b`;
func NewSortedArray[T comparable](comparator func(a, b T) int, safe ...bool) *SortedArray[T] {
	return NewSortedArraySize[T](0, comparator, safe...)
}

// NewSortedArraySize create and returns an sorted array with given size and cap.
// The parameter `safe` is used to specify whether using array in concurrent-safety,
// which is false in default.
func NewSortedArraySize[T comparable](cap int, comparator func(a, b T) int, safe ...bool) *SortedArray[T] {
	return &SortedArray[T]{
		mu:         rwmutex.Create(safe...),
		array:      make([]T, 0, cap),
		comparator: comparator,
	}
}

// NewSortedArrayRange creates and returns an array by a range from `start` to `end`
// with step value `step`.
func NewSortedArrayRange(start, end, step int, comparator func(a, b int) int, safe ...bool) *SortedArray[int] {
	if step == 0 {
		panic(fmt.Sprintf(`invalid step value: %d`, step))
	}
	slice := make([]int, 0)
	index := 0
	for i := start; i <= end; i += step {
		slice = append(slice, i)
		index++
	}
	return NewSortedArrayFrom(slice, comparator, safe...)
}

// NewSortedArrayFrom creates and returns an sorted array with given slice `array`.
// The parameter `safe` is used to specify whether using array in concurrent-safety,
// which is false in default.
func NewSortedArrayFrom[T comparable](array []T, comparator func(a, b T) int, safe ...bool) *SortedArray[T] {
	a := NewSortedArraySize(0, comparator, safe...)
	a.array = array
	sort.Slice(a.array, func(i, j int) bool {
		return a.getComparator()(a.array[i], a.array[j]) < 0
	})
	return a
}

// NewSortedArrayFromCopy creates and returns an sorted array from a copy of given slice `array`.
// The parameter `safe` is used to specify whether using array in concurrent-safety,
// which is false in default.
func NewSortedArrayFromCopy[T comparable](array []T, comparator func(a, b T) int, safe ...bool) *SortedArray[T] {
	newArray := make([]T, len(array))
	copy(newArray, array)
	return NewSortedArrayFrom(newArray, comparator, safe...)
}

// At returns the value by the specified index.
// If the given `index` is out of range of the array, it returns `nil`.
func (a *SortedArray[T]) At(index int) (value T) {
	value, _ = a.Get(index)
	return
}

// SetArray sets the underlying slice array with the given `array`.
func (a *SortedArray[T]) SetArray(array []T) Array[T] {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.array = array
	sort.Slice(a.array, func(i, j int) bool {
		return a.getComparator()(a.array[i], a.array[j]) < 0
	})
	return a
}

// SetComparator sets/changes the comparator for sorting.
// It resorts the array as the comparator is changed.
func (a *SortedArray[T]) SetComparator(comparator func(a, b T) int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.comparator = comparator
	sort.Slice(a.array, func(i, j int) bool {
		return a.getComparator()(a.array[i], a.array[j]) < 0
	})
}

// Sort sorts the array in increasing order.
// The parameter `reverse` controls whether sort
// in increasing order(default) or decreasing order
func (a *SortedArray[T]) Sort() Array[T] {
	a.mu.Lock()
	defer a.mu.Unlock()
	sort.Slice(a.array, func(i, j int) bool {
		return a.getComparator()(a.array[i], a.array[j]) < 0
	})
	return a
}

// Add adds one or multiple values to sorted array, the array always keeps sorted.
// It's alias of function Append, see Append.
func (a *SortedArray[T]) Add(values ...T) Array[T] {
	return a.Append(values...)
}

// Append adds one or multiple values to sorted array, the array always keeps sorted.
func (a *SortedArray[T]) Append(values ...T) Array[T] {
	if len(values) == 0 {
		return a
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, value := range values {
		index, cmp := a.binSearch(value, false)
		if a.unique && cmp == 0 {
			continue
		}
		if index < 0 {
			a.array = append(a.array, value)
			continue
		}
		if cmp > 0 {
			index++
		}
		a.array = append(a.array[:index], append([]T{value}, a.array[index:]...)...)
	}
	return a
}

// Get returns the value by the specified index.
// If the given `index` is out of range of the array, the `found` is false.
func (a *SortedArray[T]) Get(index int) (value T, found bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if index < 0 || index >= len(a.array) {
		found = false
		return
	}
	return a.array[index], true
}

// Remove removes an item by index.
// If the given `index` is out of range of the array, the `found` is false.
func (a *SortedArray[T]) Remove(index int) (value T, found bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.doRemoveWithoutLock(index)
}

// doRemoveWithoutLock removes an item by index without lock.
func (a *SortedArray[T]) doRemoveWithoutLock(index int) (value T, found bool) {
	if index < 0 || index >= len(a.array) {
		found = false
		return
	}
	// Determine array boundaries when deleting to improve deletion efficiency.
	if index == 0 {
		value := a.array[0]
		a.array = a.array[1:]
		return value, true
	} else if index == len(a.array)-1 {
		value := a.array[index]
		a.array = a.array[:index]
		return value, true
	}
	// If it is a non-boundary delete,
	// it will involve the creation of an array,
	// then the deletion is less efficient.
	value = a.array[index]
	a.array = append(a.array[:index], a.array[index+1:]...)
	return value, true
}

// RemoveValue removes an item by value.
// It returns true if value is found in the array, or else false if not found.
func (a *SortedArray[T]) RemoveValue(value T) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if i, r := a.binSearch(value, false); r == 0 {
		_, res := a.doRemoveWithoutLock(i)
		return res
	}
	return false
}

// RemoveValues removes an item by `values`.
func (a *SortedArray[T]) RemoveValues(values ...T) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, value := range values {
		if i, r := a.binSearch(value, false); r == 0 {
			a.doRemoveWithoutLock(i)
		}
	}
}

// PopLeft pops and returns an item from the beginning of array.
// Note that if the array is empty, the `found` is false.
func (a *SortedArray[T]) PopLeft() (value T, found bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.array) == 0 {
		found = false
		return
	}
	value = a.array[0]
	a.array = a.array[1:]
	return value, true
}

// PopRight pops and returns an item from the end of array.
// Note that if the array is empty, the `found` is false.
func (a *SortedArray[T]) PopRight() (value T, found bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	index := len(a.array) - 1
	if index < 0 {
		found = false
		return
	}
	value = a.array[index]
	a.array = a.array[:index]
	return value, true
}

// PopRand randomly pops and return an item out of array.
// Note that if the array is empty, the `found` is false.
func (a *SortedArray[T]) PopRand() (value T, found bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.doRemoveWithoutLock(grand.Intn(len(a.array)))
}

// PopRands randomly pops and returns `size` items out of array.
func (a *SortedArray[T]) PopRands(size int) []T {
	a.mu.Lock()
	defer a.mu.Unlock()
	if size <= 0 || len(a.array) == 0 {
		return nil
	}
	if size >= len(a.array) {
		size = len(a.array)
	}
	array := make([]T, size)
	for i := 0; i < size; i++ {
		array[i], _ = a.doRemoveWithoutLock(grand.Intn(len(a.array)))
	}
	return array
}

// PopLefts pops and returns `size` items from the beginning of array.
func (a *SortedArray[T]) PopLefts(size int) []T {
	a.mu.Lock()
	defer a.mu.Unlock()
	if size <= 0 || len(a.array) == 0 {
		return nil
	}
	if size >= len(a.array) {
		array := a.array
		a.array = a.array[:0]
		return array
	}
	value := a.array[0:size]
	a.array = a.array[size:]
	return value
}

// PopRights pops and returns `size` items from the end of array.
func (a *SortedArray[T]) PopRights(size int) []T {
	a.mu.Lock()
	defer a.mu.Unlock()
	if size <= 0 || len(a.array) == 0 {
		return nil
	}
	index := len(a.array) - size
	if index <= 0 {
		array := a.array
		a.array = a.array[:0]
		return array
	}
	value := a.array[index:]
	a.array = a.array[:index]
	return value
}

// Range picks and returns items by range, like array[start:end].
// Notice, if in concurrent-safe usage, it returns a copy of slice;
// else a pointer to the underlying data.
//
// If `end` is negative, then the offset will start from the end of array.
// If `end` is omitted, then the sequence will have everything from start up
// until the end of the array.
func (a *SortedArray[T]) Range(start int, end ...int) []T {
	a.mu.RLock()
	defer a.mu.RUnlock()
	offsetEnd := len(a.array)
	if len(end) > 0 && end[0] < offsetEnd {
		offsetEnd = end[0]
	}
	if start > offsetEnd {
		return nil
	}
	if start < 0 {
		start = 0
	}
	array := ([]T)(nil)
	if a.mu.IsSafe() {
		array = make([]T, offsetEnd-start)
		copy(array, a.array[start:offsetEnd])
	} else {
		array = a.array[start:offsetEnd]
	}
	return array
}

// SubSlice returns a slice of elements from the array as specified
// by the `offset` and `size` parameters.
// If in concurrent safe usage, it returns a copy of the slice; else a pointer.
//
// If offset is non-negative, the sequence will start at that offset in the array.
// If offset is negative, the sequence will start that far from the end of the array.
//
// If length is given and is positive, then the sequence will have up to that many elements in it.
// If the array is shorter than the length, then only the available array elements will be present.
// If length is given and is negative then the sequence will stop that many elements from the end of the array.
// If it is omitted, then the sequence will have everything from offset up until the end of the array.
//
// Any possibility crossing the left border of array, it will fail.
func (a *SortedArray[T]) SubSlice(offset int, length ...int) []T {
	a.mu.RLock()
	defer a.mu.RUnlock()
	size := len(a.array)
	if len(length) > 0 {
		size = length[0]
	}
	if offset > len(a.array) {
		return nil
	}
	if offset < 0 {
		offset = len(a.array) + offset
		if offset < 0 {
			return nil
		}
	}
	if size < 0 {
		offset += size
		size = -size
		if offset < 0 {
			return nil
		}
	}
	end := offset + size
	if end > len(a.array) {
		end = len(a.array)
		size = len(a.array) - offset
	}
	if a.mu.IsSafe() {
		s := make([]T, size)
		copy(s, a.array[offset:])
		return s
	} else {
		return a.array[offset:end]
	}
}

// Sum returns the sum of values in an array.
func (a *SortedArray[T]) Sum() (sum int) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	for _, v := range a.array {
		sum += gconv.Int(v)
	}
	return
}

// Len returns the length of array.
func (a *SortedArray[T]) Len() int {
	a.mu.RLock()
	length := len(a.array)
	a.mu.RUnlock()
	return length
}

// Slice returns the underlying data of array.
// Note that, if it's in concurrent-safe usage, it returns a copy of underlying data,
// or else a pointer to the underlying data.
func (a *SortedArray[T]) Slice() []T {
	var array []T
	if a.mu.IsSafe() {
		a.mu.RLock()
		defer a.mu.RUnlock()
		array = make([]T, len(a.array))
		copy(array, a.array)
	} else {
		array = a.array
	}
	return array
}

// Interfaces returns current array as []T.
func (a *SortedArray[T]) Interfaces() []T {
	return a.Slice()
}

// Contains checks whether a value exists in the array.
func (a *SortedArray[T]) Contains(value T) bool {
	return a.Search(value) != -1
}

// Contains checks whether a value exists in the array.
func (a *SortedArray[T]) ContainsI(value T) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if len(a.array) == 0 {
		return false
	}
	if s, ok := any(value).(string); ok {
		for _, v := range a.array {
			if strings.EqualFold(any(v).(string), s) {
				return true
			}
		}
		return false
	}
	return a.Contains(value)
}

// Search searches array by `value`, returns the index of `value`,
// or returns -1 if not exists.
func (a *SortedArray[T]) Search(value T) (index int) {
	if i, r := a.binSearch(value, true); r == 0 {
		return i
	}
	return -1
}

// Binary search.
// It returns the last compared index and the result.
// If `result` equals to 0, it means the value at `index` is equals to `value`.
// If `result` lesser than 0, it means the value at `index` is lesser than `value`.
// If `result` greater than 0, it means the value at `index` is greater than `value`.
func (a *SortedArray[T]) binSearch(value T, lock bool) (index int, result int) {
	if lock {
		a.mu.RLock()
		defer a.mu.RUnlock()
	}
	if len(a.array) == 0 {
		return -1, -2
	}
	min := 0
	max := len(a.array) - 1
	mid := 0
	cmp := -2
	for min <= max {
		mid = min + (max-min)/2
		cmp = a.getComparator()(value, a.array[mid])
		switch {
		case cmp < 0:
			max = mid - 1
		case cmp > 0:
			min = mid + 1
		default:
			return mid, cmp
		}
	}
	return mid, cmp
}

// SetUnique sets unique mark to the array,
// which means it does not contain any repeated items.
// It also does unique check, remove all repeated items.
func (a *SortedArray[T]) SetUnique(unique bool) Array[T] {
	oldUnique := a.unique
	a.unique = unique
	if unique && oldUnique != unique {
		a.Unique()
	}
	return a
}

// Unique uniques the array, clear repeated items.
func (a *SortedArray[T]) Unique() Array[T] {
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.array) == 0 {
		return a
	}
	i := 0
	for {
		if i == len(a.array)-1 {
			break
		}
		if a.getComparator()(a.array[i], a.array[i+1]) == 0 {
			a.array = append(a.array[:i+1], a.array[i+1+1:]...)
		} else {
			i++
		}
	}
	return a
}

// Clone returns a new array, which is a copy of current array.
func (a *SortedArray[T]) Clone() (newArray Array[T]) {
	a.mu.RLock()
	array := make([]T, len(a.array))
	copy(array, a.array)
	a.mu.RUnlock()
	return NewSortedArrayFrom(array, a.comparator, a.mu.IsSafe())
}

// Clear deletes all items of current array.
func (a *SortedArray[T]) Clear() Array[T] {
	a.mu.Lock()
	if len(a.array) > 0 {
		a.array = make([]T, 0)
	}
	a.mu.Unlock()
	return a
}

// LockFunc locks writing by callback function `f`.
func (a *SortedArray[T]) LockFunc(f func(array []T)) Array[T] {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Keep the array always sorted.
	defer sort.Slice(a.array, func(i, j int) bool {
		return a.getComparator()(a.array[i], a.array[j]) < 0
	})

	f(a.array)
	return a
}

// RLockFunc locks reading by callback function `f`.
func (a *SortedArray[T]) RLockFunc(f func(array []T)) Array[T] {
	a.mu.RLock()
	defer a.mu.RUnlock()
	f(a.array)
	return a
}

// Merge merges `array` into current array.
// The parameter `array` can be any garray or slice type.
// The difference between Merge and Append is Append supports only specified slice type,
// but Merge supports more parameter types.
func (a *SortedArray[T]) Merge(array Array[T]) Array[T] {
	return a.Add(array.Slice()...)
}

// MergeSlice merges slice into current array.
// The parameter `array` can be another StdArray
// The difference between Merge and Append is Append supports only specified slice type,
// but Merge supports more parameter types.
func (a *SortedArray[T]) MergeSlice(slice []T) Array[T] {
	if len(slice) == 0 {
		return a
	}
	return a.Append(slice...)
}

// Chunk splits an array into multiple arrays,
// the size of each array is determined by `size`.
// The last chunk may contain less than size elements.
func (a *SortedArray[T]) Chunk(size int) [][]T {
	if size < 1 {
		return nil
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	length := len(a.array)
	chunks := int(math.Ceil(float64(length) / float64(size)))
	var n [][]T
	for i, end := 0, 0; chunks > 0; chunks-- {
		end = (i + 1) * size
		if end > length {
			end = length
		}
		n = append(n, a.array[i*size:end])
		i++
	}
	return n
}

// Rand randomly returns one item from array(no deleting).
func (a *SortedArray[T]) Rand() (value T, found bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if len(a.array) == 0 {
		found = false
		return
	}
	return a.array[grand.Intn(len(a.array))], true
}

// Rands randomly returns `size` items from array(no deleting).
func (a *SortedArray[T]) Rands(size int) []T {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if size <= 0 || len(a.array) == 0 {
		return nil
	}
	array := make([]T, size)
	for i := 0; i < size; i++ {
		array[i] = a.array[grand.Intn(len(a.array))]
	}
	return array
}

// Join joins array elements with a string `glue`.
func (a *SortedArray[T]) Join(glue string) string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if len(a.array) == 0 {
		return ""
	}
	buffer := bytes.NewBuffer(nil)
	for k, v := range a.array {
		buffer.WriteString(gconv.String(v))
		if k != len(a.array)-1 {
			buffer.WriteString(glue)
		}
	}
	return buffer.String()
}

// CountValues counts the number of occurrences of all values in the array.
func (a *SortedArray[T]) CountValues() map[T]int {
	m := make(map[T]int)
	a.mu.RLock()
	defer a.mu.RUnlock()
	for _, v := range a.array {
		m[v]++
	}
	return m
}

// Iterator is alias of IteratorAsc.
func (a *SortedArray[T]) Iterator(f func(k int, v T) bool) {
	a.IteratorAsc(f)
}

// IteratorAsc iterates the array readonly in ascending order with given callback function `f`.
// If `f` returns true, then it continues iterating; or false to stop.
func (a *SortedArray[T]) IteratorAsc(f func(k int, v T) bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	for k, v := range a.array {
		if !f(k, v) {
			break
		}
	}
}

// IteratorDesc iterates the array readonly in descending order with given callback function `f`.
// If `f` returns true, then it continues iterating; or false to stop.
func (a *SortedArray[T]) IteratorDesc(f func(k int, v T) bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	for i := len(a.array) - 1; i >= 0; i-- {
		if !f(i, a.array[i]) {
			break
		}
	}
}

// String returns current array as a string, which implements like json.Marshal does.
func (a *SortedArray[T]) String() string {
	if a == nil {
		return ""
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	buffer := bytes.NewBuffer(nil)
	buffer.WriteByte('[')
	s := ""
	for k, v := range a.array {
		s = gconv.String(v)
		if gstr.IsNumeric(s) {
			buffer.WriteString(s)
		} else {
			buffer.WriteString(`"` + gstr.QuoteMeta(s, `"\`) + `"`)
		}
		if k != len(a.array)-1 {
			buffer.WriteByte(',')
		}
	}
	buffer.WriteByte(']')
	return buffer.String()
}

// MarshalJSON implements the interface MarshalJSON for json.Marshal.
// Note that do not use pointer as its receiver here.
func (a SortedArray[T]) MarshalJSON() ([]byte, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return json.Marshal(a.array)
}

// UnmarshalJSON implements the interface UnmarshalJSON for json.Unmarshal.
// Note that the comparator is set as string comparator in default.
func (a *SortedArray[T]) UnmarshalJSON(b []byte) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := json.UnmarshalUseNumber(b, &a.array); err != nil {
		return err
	}
	if a.comparator == nil {
		a.comparator = comparator.ComparatorAny[T]
	}
	if a.array != nil {
		sort.Slice(a.array, func(i, j int) bool {
			return a.comparator(a.array[i], a.array[j]) < 0
		})
	}
	return nil
}

// UnmarshalValue is an interface implement which sets any type of value for array.
func (a *SortedArray[T]) UnmarshalValue(value interface{}) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	switch value.(type) {
	case string, []byte, json2.Number:
		if err := json.UnmarshalUseNumber(gconv.Bytes(value), &a.array); err != nil {
			return err
		}
	default:
		a.array = gconv.SliceAny[T](value)
	}
	if a.comparator == nil {
		a.comparator = comparator.ComparatorAny[T]
	}
	if a.array != nil {
		sort.Slice(a.array, func(i, j int) bool {
			return a.comparator(a.array[i], a.array[j]) < 0
		})
	}
	return nil
}

// FilterNil removes all nil value of the array.
func (a *SortedArray[T]) FilterNil() Array[T] {
	a.mu.Lock()
	defer a.mu.Unlock()
	for i := 0; i < len(a.array); {
		if empty.IsNil(a.array[i]) {
			a.array = append(a.array[:i], a.array[i+1:]...)
		} else {
			break
		}
	}
	for i := len(a.array) - 1; i >= 0; {
		if empty.IsNil(a.array[i]) {
			a.array = append(a.array[:i], a.array[i+1:]...)
		} else {
			break
		}
	}
	return a
}

// Filter iterates array and filters elements using custom callback function.
// It removes the element from array if callback function `filter` returns true,
// it or else does nothing and continues iterating.
func (a *SortedArray[T]) Filter(filter func(index int, value T) bool) Array[T] {
	a.mu.Lock()
	defer a.mu.Unlock()
	for i := 0; i < len(a.array); {
		if filter(i, a.array[i]) {
			a.array = append(a.array[:i], a.array[i+1:]...)
		} else {
			i++
		}
	}
	return a
}

// FilterEmpty removes all empty value of the array.
// Values like: 0, nil, false, "", len(slice/map/chan) == 0 are considered empty.
func (a *SortedArray[T]) FilterEmpty() Array[T] {
	a.mu.Lock()
	defer a.mu.Unlock()
	for i := 0; i < len(a.array); {
		if empty.IsEmpty(a.array[i]) {
			a.array = append(a.array[:i], a.array[i+1:]...)
		} else {
			break
		}
	}
	for i := len(a.array) - 1; i >= 0; {
		if empty.IsEmpty(a.array[i]) {
			a.array = append(a.array[:i], a.array[i+1:]...)
		} else {
			break
		}
	}
	return a
}

// Walk applies a user supplied function `f` to every item of array.
func (a *SortedArray[T]) Walk(f func(value T) T) Array[T] {
	a.mu.Lock()
	defer a.mu.Unlock()
	// Keep the array always sorted.
	defer sort.Slice(a.array, func(i, j int) bool {
		return a.getComparator()(a.array[i], a.array[j]) < 0
	})
	for i, v := range a.array {
		a.array[i] = f(v)
	}
	return a
}

// IsEmpty checks whether the array is empty.
func (a *SortedArray[T]) IsEmpty() bool {
	return a.Len() == 0
}

// getComparator returns the comparator if it's previously set,
// or else it panics.
func (a *SortedArray[T]) getComparator() func(a, b T) int {
	if a.comparator == nil {
		panic("comparator is missing for sorted array")
	}
	return a.comparator
}

// DeepCopy implements interface for deep copy of current type.
func (a *SortedArray[T]) DeepCopy() Array[T] {
	if a == nil {
		return nil
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	newSlice := make([]T, len(a.array))
	for i, v := range a.array {
		newSlice[i] = deepcopy.Copy(v).(T)
	}
	return NewSortedArrayFrom[T](newSlice, a.comparator, a.mu.IsSafe())
}

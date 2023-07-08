// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

// go test *.go

package gset_test

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/wesleywu/gcontainer/garray"
	"github.com/wesleywu/gcontainer/gset"
	"github.com/wesleywu/gcontainer/internal/gtest"
	"github.com/wesleywu/gcontainer/internal/json"
	"github.com/wesleywu/gcontainer/utils/gconv"
)

type exampleElement struct {
}

func TestSet_Var(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		var s gset.HashSet[int]
		s.Add(1, 1, 2)
		s.Add([]int{3, 4}...)
		t.Assert(s.Size(), 4)
		t.AssertIN(1, s.Slice())
		t.AssertIN(2, s.Slice())
		t.AssertIN(3, s.Slice())
		t.AssertIN(4, s.Slice())
		t.AssertNI(0, s.Slice())
		t.Assert(s.Contains(4), true)
		t.Assert(s.Contains(5), false)
		s.Remove(1)
		t.Assert(s.Size(), 3)
		s.Clear()
		t.Assert(s.Size(), 0)
	})
}

func TestSet_New(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		s := gset.New[int]()
		s.Add(1, 1, 2)
		s.Add([]int{3, 4}...)
		t.Assert(s.Size(), 4)
		t.AssertIN(1, s.Slice())
		t.AssertIN(2, s.Slice())
		t.AssertIN(3, s.Slice())
		t.AssertIN(4, s.Slice())
		t.AssertNI(0, s.Slice())
		t.Assert(s.Contains(4), true)
		t.Assert(s.Contains(5), false)
		s.Remove(1)
		t.Assert(s.Size(), 3)
		s.Clear()
		t.Assert(s.Size(), 0)
	})
}

func TestSet_Basic(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		s := gset.NewSet[int]()
		s.Add(1, 1, 2)
		s.Add([]int{3, 4}...)
		t.Assert(s.Size(), 4)
		t.AssertIN(1, s.Slice())
		t.AssertIN(2, s.Slice())
		t.AssertIN(3, s.Slice())
		t.AssertIN(4, s.Slice())
		t.AssertNI(0, s.Slice())
		t.Assert(s.Contains(4), true)
		t.Assert(s.Contains(5), false)
		s.Remove(1)
		t.Assert(s.Size(), 3)
		s.Clear()
		t.Assert(s.Size(), 0)
	})
}

func TestSet_Iterator(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		s := gset.NewSet[int]()
		s.Add(1, 2, 3)
		t.Assert(s.Size(), 3)

		a1 := garray.New[int](true)
		a2 := garray.New[int](true)
		s.Iterator(func(v int) bool {
			a1.Append(1)
			return false
		})
		s.Iterator(func(v int) bool {
			a2.Append(1)
			return true
		})
		t.Assert(a1.Len(), 1)
		t.Assert(a2.Len(), 3)
	})
}

func TestSet_LockFunc(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		s := gset.NewSet[int]()
		s.Add(1, 2, 3)
		t.Assert(s.Size(), 3)
		s.LockFunc(func(m map[int]struct{}) {
			delete(m, 1)
		})
		t.Assert(s.Size(), 2)
		s.RLockFunc(func(m map[int]struct{}) {
			t.Assert(m, map[int]struct{}{
				3: {},
				2: {},
			})
		})
	})
}

func TestSet_Equal(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		s1 := gset.NewSet[int]()
		s2 := gset.NewSet[int]()
		s3 := gset.NewSet[int]()
		s4 := gset.NewSet[int]()
		s1.Add(1, 2, 3)
		s2.Add(1, 2, 3)
		s3.Add(1, 2, 3, 4)
		s4.Add(4, 5, 6)
		t.Assert(s1.Equal(s2), true)
		t.Assert(s1.Equal(s3), false)
		t.Assert(s1.Equal(s4), false)
		s5 := s1
		t.Assert(s1.Equal(s5), true)
	})
}

func TestSet_IsSubsetOf(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		s1 := gset.NewSet[int]()
		s2 := gset.NewSet[int]()
		s3 := gset.NewSet[int]()
		s1.Add(1, 2)
		s2.Add(1, 2, 3)
		s3.Add(1, 2, 3, 4)
		t.Assert(s1.IsSubsetOf(s2), true)
		t.Assert(s2.IsSubsetOf(s3), true)
		t.Assert(s1.IsSubsetOf(s3), true)
		t.Assert(s2.IsSubsetOf(s1), false)
		t.Assert(s3.IsSubsetOf(s2), false)

		s4 := s1
		t.Assert(s1.IsSubsetOf(s4), true)
	})
}

func TestSet_Union(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		s1 := gset.NewSet[int]()
		s2 := gset.NewSet[int]()
		s1.Add(1, 2)
		s2.Add(3, 4)
		s3 := s1.Union(s2)
		t.Assert(s3.Contains(1), true)
		t.Assert(s3.Contains(2), true)
		t.Assert(s3.Contains(3), true)
		t.Assert(s3.Contains(4), true)
	})
}

func TestSet_Diff(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		s1 := gset.NewSet[int]()
		s2 := gset.NewSet[int]()
		s1.Add(1, 2, 3)
		s2.Add(3, 4, 5)
		s3 := s1.Diff(s2)
		t.Assert(s3.Contains(1), true)
		t.Assert(s3.Contains(2), true)
		t.Assert(s3.Contains(3), false)
		t.Assert(s3.Contains(4), false)

		s4 := s1
		s5 := s1.Diff(s2, s4)
		t.Assert(s5.Contains(1), true)
		t.Assert(s5.Contains(2), true)
		t.Assert(s5.Contains(3), false)
		t.Assert(s5.Contains(4), false)
	})
}

func TestSet_Intersect(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		s1 := gset.NewSet[int]()
		s2 := gset.NewSet[int]()
		s1.Add(1, 2, 3)
		s2.Add(3, 4, 5)
		s3 := s1.Intersect(s2)
		t.Assert(s3.Contains(1), false)
		t.Assert(s3.Contains(2), false)
		t.Assert(s3.Contains(3), true)
		t.Assert(s3.Contains(4), false)
	})
}

func TestSet_Complement(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		s1 := gset.NewSet[int]()
		s2 := gset.NewSet[int]()
		s1.Add(1, 2, 3)
		s2.Add(3, 4, 5)
		s3 := s1.Complement(s2)
		t.Assert(s3.Contains(1), false)
		t.Assert(s3.Contains(2), false)
		t.Assert(s3.Contains(4), true)
		t.Assert(s3.Contains(5), true)
	})
}

func TestNewFrom(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		s1 := gset.NewFrom[string]([]string{"a"})
		s2 := gset.NewFrom[string]([]string{"b"}, false)
		s3 := gset.NewFrom[string]([]string{"3"}, true)
		s4 := gset.NewFrom[string]([]string{"s1", "s2"}, true)
		t.Assert(s1.Contains("a"), true)
		t.Assert(s2.Contains("b"), true)
		t.Assert(s3.Contains("3"), true)
		t.Assert(s4.Contains("s1"), true)
		t.Assert(s4.Contains("s3"), false)

	})
}

func TestNew(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		s1 := gset.New[string]()
		s1.Add("a", "2")
		s2 := gset.New[string](true)
		s2.Add("b", "3")
		t.Assert(s1.Contains("a"), true)

	})
}

func TestSet_Join(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		s1 := gset.New[string](true)
		s1.Add("a", "a1", "b", "c")
		str1 := s1.Join(",")
		t.Assert(strings.Contains(str1, "a1"), true)
	})
	gtest.C(t, func(t *gtest.T) {
		s1 := gset.New[string](true)
		s1.Add("a", `"b"`, `\c`)
		str1 := s1.Join(",")
		t.Assert(strings.Contains(str1, `"b"`), true)
		t.Assert(strings.Contains(str1, `\c`), true)
		t.Assert(strings.Contains(str1, `a`), true)
	})
	gtest.C(t, func(t *gtest.T) {
		s1 := gset.HashSet[int]{}
		t.Assert(s1.Join(","), "")
	})
}

func TestSet_String(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		s1 := gset.New[string](true)
		s1.Add("a", "a2", "b", "c")
		str1 := s1.String()
		t.Assert(strings.Contains(str1, "["), true)
		t.Assert(strings.Contains(str1, "]"), true)
		t.Assert(strings.Contains(str1, "a2"), true)

		s1 = nil
		t.Assert(s1.String(), "")

		s2 := gset.New[int]()
		s2.Add(1)
		t.Assert(s2.String(), "[1]")
	})
}

func TestSet_Merge(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		s1 := gset.New[string](true)
		s2 := gset.New[string](true)
		s1.Add("a", "a2", "b", "c")
		s2.Add("b", "b1", "e", "f")
		ss := s1.Merge(s2)
		t.Assert(ss.Contains("a2"), true)
		t.Assert(ss.Contains("b1"), true)

	})
}

func TestSet_Sum(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		s1 := gset.New[int](true)
		s1.Add(1, 2, 3, 4)
		t.Assert(s1.Sum(), int(10))

	})
}

func TestSet_Pop(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		s := gset.New[int](true)
		t.Assert(s.Pop(), 0)
		s.Add(1, 2, 3, 4)
		t.Assert(s.Size(), 4)
		t.AssertIN(s.Pop(), []int{1, 2, 3, 4})
		t.Assert(s.Size(), 3)
	})
}

func TestSet_Pops(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		s := gset.New[int](true)
		s.Add(1, 2, 3, 4)
		t.Assert(s.Size(), 4)
		t.Assert(s.Pops(0), nil)
		t.AssertIN(s.Pops(1), []int{1, 2, 3, 4})
		t.Assert(s.Size(), 3)
		a := s.Pops(6)
		t.Assert(len(a), 3)
		t.AssertIN(a, []int{1, 2, 3, 4})
		t.Assert(s.Size(), 0)
	})

	gtest.C(t, func(t *gtest.T) {
		s := gset.New[int](true)
		a := []int{1, 2, 3, 4}
		s.Add(a...)
		t.Assert(s.Size(), 4)
		t.Assert(s.Pops(-2), nil)
		t.AssertIN(s.Pops(-1), a)
	})
}

func TestSet_Json(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		s1 := []string{"a", "b", "d", "c"}
		a1 := gset.NewFrom(s1)
		b1, err1 := json.Marshal(a1)
		b2, err2 := json.Marshal(s1)
		t.Assert(len(b1), len(b2))
		t.Assert(err1, err2)

		a2 := gset.New[string]()
		err2 = json.UnmarshalUseNumber(b2, &a2)
		t.Assert(err2, nil)
		t.Assert(a2.Contains("a"), true)
		t.Assert(a2.Contains("b"), true)
		t.Assert(a2.Contains("c"), true)
		t.Assert(a2.Contains("d"), true)
		t.Assert(a2.Contains("e"), false)

		var a3 gset.HashSet[string]
		err := json.UnmarshalUseNumber(b2, &a3)
		t.AssertNil(err)
		t.Assert(a3.Contains("a"), true)
		t.Assert(a3.Contains("b"), true)
		t.Assert(a3.Contains("c"), true)
		t.Assert(a3.Contains("d"), true)
		t.Assert(a3.Contains("e"), false)
	})
}

func TestSet_AddIfNotExist(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		s := gset.New[int](true)
		s.Add(1)
		t.Assert(s.Contains(1), true)
		t.Assert(s.AddIfNotExist(1), false)
		t.Assert(s.AddIfNotExist(2), true)
		t.Assert(s.Contains(2), true)
		t.Assert(s.AddIfNotExist(2), false)
		t.Assert(s.Contains(2), true)
	})
}

func TestSet_AddIfNotExistFunc(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		s := gset.New[int](true)
		s.Add(1)
		t.Assert(s.Contains(1), true)
		t.Assert(s.Contains(2), false)
		t.Assert(s.AddIfNotExistFunc(2, func() bool { return false }), false)
		t.Assert(s.Contains(2), false)
		t.Assert(s.AddIfNotExistFunc(2, func() bool { return true }), true)
		t.Assert(s.Contains(2), true)
		t.Assert(s.AddIfNotExistFunc(2, func() bool { return true }), false)
		t.Assert(s.Contains(2), true)
	})
	gtest.C(t, func(t *gtest.T) {
		s := gset.New[int](true)
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := s.AddIfNotExistFunc(1, func() bool {
				time.Sleep(100 * time.Millisecond)
				return true
			})
			t.Assert(r, false)
		}()
		s.Add(1)
		wg.Wait()
	})
	gtest.C(t, func(t *gtest.T) {
		s := gset.HashSet[int]{}
		t.Assert(s.AddIfNotExistFunc(1, func() bool { return true }), true)
	})
}

func TestSet_Walk(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		var set gset.HashSet[int]
		set.Add([]int{1, 2}...)
		set.Walk(func(item int) int {
			return gconv.Int(item) + 10
		})
		t.Assert(set.Size(), 2)
		t.Assert(set.Contains(11), true)
		t.Assert(set.Contains(12), true)
	})
}

func TestSet_AddIfNotExistFuncLock(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		s := gset.New[int](true)
		wg := sync.WaitGroup{}
		wg.Add(2)
		go func() {
			defer wg.Done()
			r := s.AddIfNotExistFuncLock(1, func() bool {
				time.Sleep(500 * time.Millisecond)
				return true
			})
			t.Assert(r, true)
		}()
		time.Sleep(100 * time.Millisecond)
		go func() {
			defer wg.Done()
			r := s.AddIfNotExistFuncLock(1, func() bool {
				return true
			})
			t.Assert(r, false)
		}()
		wg.Wait()
	})
	gtest.C(t, func(t *gtest.T) {
		s := gset.New[*exampleElement](true)
		t.Assert(s.AddIfNotExistFuncLock(nil, func() bool { return true }), false)
		s1 := gset.HashSet[int]{}
		t.Assert(s1.AddIfNotExistFuncLock(1, func() bool { return true }), true)
	})
}

func TestSet_UnmarshalValue(t *testing.T) {
	type V struct {
		Name string
		Set  *gset.HashSet[string]
	}
	// JSON
	gtest.C(t, func(t *gtest.T) {
		var v *V
		err := gconv.Struct(map[string]interface{}{
			"name": "john",
			"set":  []byte(`["k1","k2","k3"]`),
		}, &v)
		t.AssertNil(err)
		t.Assert(v.Name, "john")
		t.Assert(v.Set.Size(), 3)
		t.Assert(v.Set.Contains("k1"), true)
		t.Assert(v.Set.Contains("k2"), true)
		t.Assert(v.Set.Contains("k3"), true)
		t.Assert(v.Set.Contains("k4"), false)
	})
	// Map
	gtest.C(t, func(t *gtest.T) {
		var v *V
		err := gconv.Struct(map[string]interface{}{
			"name": "john",
			"set":  []string{"k1", "k2", "k3"},
		}, &v)
		t.AssertNil(err)
		t.Assert(v.Name, "john")
		t.Assert(v.Set.Size(), 3)
		t.Assert(v.Set.Contains("k1"), true)
		t.Assert(v.Set.Contains("k2"), true)
		t.Assert(v.Set.Contains("k3"), true)
		t.Assert(v.Set.Contains("k4"), false)
	})
}

func TestSet_DeepCopy(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		set := gset.New[int]()
		set.Add(1, 2, 3)

		copySet := set.DeepCopy().(*gset.HashSet[int])
		copySet.Add(4)
		t.AssertNE(set.Size(), copySet.Size())
		t.AssertNE(set.String(), copySet.String())

		set = nil
		t.AssertNil(set.DeepCopy())
	})
}

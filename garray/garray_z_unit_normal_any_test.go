// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

// go test *.go

package garray_test

import (
	"testing"
	"time"

	"github.com/wesleywu/gcontainer/utils/empty"
	"github.com/wesleywu/gcontainer/utils/gconv"

	"github.com/wesleywu/gcontainer/garray"
	"github.com/wesleywu/gcontainer/internal/gtest"
	"github.com/wesleywu/gcontainer/internal/json"
)

func Test_Array_Basic(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		expect := []int{0, 1, 2, 3}
		array := garray.NewArrayFrom[int](expect)
		array2 := garray.NewArrayFrom[int](expect)
		array3 := garray.NewArrayFrom[int]([]int{})
		array4 := garray.NewArrayRange(1, 5, 1)

		t.Assert(array.Slice(), expect)
		t.Assert(array.Interfaces(), expect)
		err := array.Set(0, 100)
		t.AssertNil(err)

		err = array.Set(100, 100)
		t.AssertNE(err, nil)

		t.Assert(array.IsEmpty(), false)

		copyArray := array.DeepCopy()
		ca := copyArray.(*garray.StdArray[int])
		ca.Set(0, 1)
		cval, _ := ca.Get(0)
		val, _ := array.Get(0)
		t.AssertNE(cval, val)

		v, ok := array.Get(0)
		t.Assert(v, 100)
		t.Assert(ok, true)

		v, ok = array.Get(1)
		t.Assert(v, 1)
		t.Assert(ok, true)

		v, ok = array.Get(4)
		t.Assert(v, 0)
		t.Assert(ok, false)

		t.Assert(array.Search(100), 0)
		t.Assert(array3.Search(100), -1)
		t.Assert(array.Contains(100), true)

		ok = array.Remove(0) // element 0->0 was replaced by 0->100
		t.Assert(ok, false)

		ok = array.Remove(100) // element 0->0 was replaced by 0->100
		t.Assert(ok, true)

		ok = array.Remove(-1)
		t.Assert(ok, false)

		ok = array.Remove(100000)
		t.Assert(ok, false)

		ok = array2.Remove(3)
		t.Assert(ok, true)

		ok = array2.Remove(1)
		t.Assert(ok, true)

		t.Assert(array.Contains(100), false)
		array.Add(4)
		t.Assert(array.Len(), 4)
		array.InsertBefore(0, 100)
		array.InsertAfter(0, 200)
		t.Assert(array.Slice(), []int{100, 200, 2, 2, 3, 4})
		array.InsertBefore(5, 300)
		array.InsertAfter(6, 400)
		t.Assert(array.Slice(), []int{100, 200, 2, 2, 3, 300, 4, 400})
		array.Clear()
		t.Assert(array.Size(), 0)
		err = array.InsertBefore(99, 9900)
		t.AssertNE(err, nil)
		err = array.InsertAfter(99, 9900)
		t.AssertNE(err, nil)

		t.Assert(array4.String(), "[1,2,3,4,5]")
	})
}

func TestArray_Sort(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		expect1 := []int{0, 1, 2, 3}
		expect2 := []int{3, 2, 1, 0}
		array := garray.NewArray[int]()
		for i := 3; i >= 0; i-- {
			array.Add(i)
		}
		array.SortFunc(func(v1, v2 int) bool {
			return v1 < v2
		})
		t.Assert(array.Slice(), expect1)
		array.SortFunc(func(v1, v2 int) bool {
			return v1 > v2
		})
		t.Assert(array.Slice(), expect2)
	})
}

func TestArray_Unique(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		expect := []int{1, 2, 3, 4, 5, 3, 2, 2, 3, 5, 5}
		array := garray.NewArrayFrom(expect)
		t.Assert(array.Unique().Slice(), []int{1, 2, 3, 4, 5})
	})
	gtest.C(t, func(t *gtest.T) {
		expect := []int{}
		array := garray.NewArrayFrom(expect)
		t.Assert(array.Unique().Slice(), []int{})
	})
}

func TestArray_PushAndPop(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		expect := []int{0, 1, 2, 3}
		array := garray.NewArrayFrom[int](expect)
		t.Assert(array.Slice(), expect)

		v, ok := array.PopLeft()
		t.Assert(v, 0)
		t.Assert(ok, true)

		v, ok = array.PopRight()
		t.Assert(v, 3)
		t.Assert(ok, true)

		v, ok = array.PopRand()
		t.AssertIN(v, []int{1, 2})
		t.Assert(ok, true)

		v, ok = array.PopRand()
		t.AssertIN(v, []int{1, 2})
		t.Assert(ok, true)

		t.Assert(array.Len(), 0)
		array.PushLeft(1).(*garray.StdArray[int]).PushRight(2)
		t.Assert(array.Slice(), []int{1, 2})
	})
}

func TestArray_PopRands(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		a1 := []int{100, 200, 300, 400, 500, 600}
		array := garray.NewFromCopy(a1)
		t.AssertIN(array.PopRands(2), []int{100, 200, 300, 400, 500, 600})
	})
}

func TestArray_PopLeft(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		array := garray.NewFrom[int]([]int{1, 2, 3})
		v, ok := array.PopLeft()
		t.Assert(v, 1)
		t.Assert(ok, true)
		t.Assert(array.Len(), 2)
		v, ok = array.PopLeft()
		t.Assert(v, 2)
		t.Assert(ok, true)
		t.Assert(array.Len(), 1)
		v, ok = array.PopLeft()
		t.Assert(v, 3)
		t.Assert(ok, true)
		t.Assert(array.Len(), 0)
	})
}

func TestArray_PopRight(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		array := garray.NewFrom[int]([]int{1, 2, 3})

		v, ok := array.PopRight()
		t.Assert(v, 3)
		t.Assert(ok, true)
		t.Assert(array.Len(), 2)

		v, ok = array.PopRight()
		t.Assert(v, 2)
		t.Assert(ok, true)
		t.Assert(array.Len(), 1)

		v, ok = array.PopRight()
		t.Assert(v, 1)
		t.Assert(ok, true)
		t.Assert(array.Len(), 0)
	})
}

func TestArray_PopLefts(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		array := garray.NewFrom[int]([]int{1, 2, 3})
		t.Assert(array.PopLefts(2), []int{1, 2})
		t.Assert(array.Len(), 1)
		t.Assert(array.PopLefts(2), []int{3})
		t.Assert(array.Len(), 0)
	})
}

func TestArray_PopRights(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		array := garray.NewFrom[int]([]int{1, 2, 3})
		t.Assert(array.PopRights(2), []int{2, 3})
		t.Assert(array.Len(), 1)
		t.Assert(array.PopLefts(2), []int{1})
		t.Assert(array.Len(), 0)
	})
}

func TestArray_PopLeftsAndPopRights(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		array := garray.New[int]()
		v, ok := array.PopLeft()
		t.Assert(v, 0)
		t.Assert(ok, false)
		poppedLefts := array.PopLefts(10)
		t.Assert(poppedLefts, nil)

		v, ok = array.PopRight()
		t.Assert(v, 0)
		t.Assert(ok, false)
		t.Assert(array.PopRights(10), nil)

		v, ok = array.PopRand()
		t.Assert(v, 0)
		t.Assert(ok, false)
		t.Assert(array.PopRands(10), nil)
	})

	gtest.C(t, func(t *gtest.T) {
		value1 := []int{0, 1, 2, 3, 4, 5, 6}
		value2 := []int{0, 1, 2, 3, 4, 5, 6}
		array1 := garray.NewArrayFrom(value1)
		array2 := garray.NewArrayFrom(value2)
		t.Assert(array1.PopLefts(2), []int{0, 1})
		t.Assert(array1.Slice(), []int{2, 3, 4, 5, 6})
		t.Assert(array1.PopRights(2), []int{5, 6})
		t.Assert(array1.Slice(), []int{2, 3, 4})
		t.Assert(array1.PopRights(20), []int{2, 3, 4})
		t.Assert(array1.Slice(), []int{})
		t.Assert(array2.PopLefts(20), []int{0, 1, 2, 3, 4, 5, 6})
		t.Assert(array2.Slice(), []int{})
	})
}

func TestArray_Range(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		value1 := []int{0, 1, 2, 3, 4, 5, 6}
		array1 := garray.NewArrayFrom(value1)
		array2 := garray.NewArrayFrom(value1, true)
		t.Assert(array1.Range(0, 1), []int{0})
		t.Assert(array1.Range(1, 2), []int{1})
		t.Assert(array1.Range(0, 2), []int{0, 1})
		t.Assert(array1.Range(-1, 10), value1)
		t.Assert(array1.Range(10, 2), nil)
		t.Assert(array2.Range(1, 3), []int{1, 2})
	})
}

func TestArray_Merge(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		i1 := []int{0, 1, 2, 3}
		i2 := []int{4, 5, 6, 7}
		array1 := garray.NewArrayFrom[int](i1)
		array2 := garray.NewArrayFrom[int](i2)
		array1.AddAll(array2)
		t.Assert(array1.Slice(), []int{0, 1, 2, 3, 4, 5, 6, 7})

		s1 := garray.NewArrayFrom[string]([]string{"a", "b", "c", "d"})
		s2 := []int{8, 9}
		i3 := garray.NewArrayFrom[int]([]int{1, 2, 3})
		i4 := garray.NewArrayFrom[int]([]int{3})
		s3 := garray.NewArrayFrom[string]([]string{"g", "h"})
		a1 := garray.NewArrayFrom[int](i1)

		s1.AddAll(s3)
		t.Assert(s1.Len(), 6)
		a1.AddAll(i3)
		t.Assert(a1.Len(), 7)
		a1.AddAll(i4)
		t.Assert(a1.Len(), 8)
		a1.Add(s2...)
		t.Assert(a1.Len(), 10)
	})
}

func TestArray_Fill(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		a1 := []int{0}
		a2 := []int{0}
		array1 := garray.NewArrayFrom(a1)
		array2 := garray.NewArrayFrom(a2, true)

		t.Assert(array1.Fill(1, 2, 100), nil)
		t.Assert(array1.Slice(), []int{0, 100, 100})

		t.Assert(array2.Fill(0, 2, 100), nil)
		t.Assert(array2.Slice(), []int{100, 100})

		t.AssertNE(array2.Fill(-1, 2, 100), nil)
		t.Assert(array2.Slice(), []int{100, 100})
	})
}

func TestArray_Chunk(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		a1 := []int{1, 2, 3, 4, 5}
		array1 := garray.NewArrayFrom(a1)
		chunks := array1.Chunk(2)
		t.Assert(len(chunks), 3)
		t.Assert(chunks[0], []int{1, 2})
		t.Assert(chunks[1], []int{3, 4})
		t.Assert(chunks[2], []int{5})
		t.Assert(array1.Chunk(0), nil)
	})
	gtest.C(t, func(t *gtest.T) {
		a1 := []int{1, 2, 3, 4, 5}
		array1 := garray.NewArrayFrom(a1)
		chunks := array1.Chunk(3)
		t.Assert(len(chunks), 2)
		t.Assert(chunks[0], []int{1, 2, 3})
		t.Assert(chunks[1], []int{4, 5})
		t.Assert(array1.Chunk(0), nil)
	})
	gtest.C(t, func(t *gtest.T) {
		a1 := []int{1, 2, 3, 4, 5, 6}
		array1 := garray.NewArrayFrom(a1)
		chunks := array1.Chunk(2)
		t.Assert(len(chunks), 3)
		t.Assert(chunks[0], []int{1, 2})
		t.Assert(chunks[1], []int{3, 4})
		t.Assert(chunks[2], []int{5, 6})
		t.Assert(array1.Chunk(0), nil)
	})
	gtest.C(t, func(t *gtest.T) {
		a1 := []int{1, 2, 3, 4, 5, 6}
		array1 := garray.NewArrayFrom(a1)
		chunks := array1.Chunk(3)
		t.Assert(len(chunks), 2)
		t.Assert(chunks[0], []int{1, 2, 3})
		t.Assert(chunks[1], []int{4, 5, 6})
		t.Assert(array1.Chunk(0), nil)
	})
}

func TestArray_Pad(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		a1 := []int{0}
		array1 := garray.NewArrayFrom(a1)
		t.Assert(array1.Pad(3, 1).Slice(), []int{0, 1, 1})
		t.Assert(array1.Pad(-4, 1).Slice(), []int{1, 0, 1, 1})
		t.Assert(array1.Pad(3, 1).Slice(), []int{1, 0, 1, 1})
	})
}

func TestArray_SubSlice(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		a1 := []int{0, 1, 2, 3, 4, 5, 6}
		array1 := garray.NewArrayFrom(a1)
		array2 := garray.NewArrayFrom(a1, true)
		t.Assert(array1.SubSlice(0, 2), []int{0, 1})
		t.Assert(array1.SubSlice(2, 2), []int{2, 3})
		t.Assert(array1.SubSlice(5, 8), []int{5, 6})
		t.Assert(array1.SubSlice(9, 1), nil)
		t.Assert(array1.SubSlice(-2, 2), []int{5, 6})
		t.Assert(array1.SubSlice(-9, 2), nil)
		t.Assert(array1.SubSlice(1, -2), nil)
		t.Assert(array2.SubSlice(0, 2), []int{0, 1})
	})
}

func TestArray_Rand(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		a1 := []int{0, 1, 2, 3, 4, 5, 6}
		array1 := garray.NewArrayFrom(a1)
		t.Assert(len(array1.Rands(2)), 2)
		t.Assert(len(array1.Rands(10)), 10)
		t.AssertIN(array1.Rands(1)[0], a1)
	})

	gtest.C(t, func(t *gtest.T) {
		s1 := []string{"a", "b", "c", "d"}
		a1 := garray.NewArrayFrom(s1)
		i1, ok := a1.Rand()
		t.Assert(ok, true)
		t.Assert(a1.Contains(i1), true)
		t.Assert(a1.Len(), 4)
	})

	gtest.C(t, func(t *gtest.T) {
		a1 := []int{}
		array1 := garray.NewArrayFrom(a1)
		rand, found := array1.Rand()
		t.Assert(rand, 0)
		t.Assert(found, false)
	})

	gtest.C(t, func(t *gtest.T) {
		a1 := []int{}
		array1 := garray.NewArrayFrom(a1)
		rand := array1.Rands(1)
		t.AssertNil(rand)
	})
}

func TestArray_Shuffle(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		a1 := []int{0, 1, 2, 3, 4, 5, 6}
		array1 := garray.NewArrayFrom(a1)
		t.Assert(array1.Shuffle().Size(), 7)
	})
}

func TestArray_Reverse(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		a1 := []int{0, 1, 2, 3, 4, 5, 6}
		array1 := garray.NewArrayFrom(a1)
		t.Assert(array1.Reverse().Slice(), []int{6, 5, 4, 3, 2, 1, 0})
	})
}

func TestArray_Join(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		a1 := []int{0, 1, 2, 3, 4, 5, 6}
		array1 := garray.NewArrayFrom(a1)
		t.Assert(array1.Join("."), `0.1.2.3.4.5.6`)
	})

	gtest.C(t, func(t *gtest.T) {
		a1 := []int{0, 1, 2, 3}
		array1 := garray.NewArrayFrom(a1)
		t.Assert(array1.Join("."), `0.1.2.3`)
	})

	gtest.C(t, func(t *gtest.T) {
		a1 := []int{}
		array1 := garray.NewArrayFrom(a1)
		t.Assert(len(array1.Join(".")), 0)
	})
}

func TestArray_String(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		a1 := []int{0, 1, 2, 3, 4, 5, 6}
		array1 := garray.NewArrayFrom(a1)
		t.Assert(array1.String(), `[0,1,2,3,4,5,6]`)
		array1 = nil
		t.Assert(array1.String(), "")
	})
}

func TestArray_Sum(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		a1 := []int{0, 1, 2, 3}
		a2 := []string{"a", "b", "c"}
		a3 := []string{"a", "1", "2"}

		array1 := garray.NewArrayFrom(a1)
		array2 := garray.NewArrayFrom(a2)
		array3 := garray.NewArrayFrom(a3)

		t.Assert(array1.Sum(), 6)
		t.Assert(array2.Sum(), 0)
		t.Assert(array3.Sum(), 3)

	})
}

func TestArray_Clone(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		a1 := []int{0, 1, 2, 3}
		array1 := garray.NewArrayFrom(a1)
		array2 := array1.Clone()

		t.Assert(array1.Len(), 4)
		t.Assert(array2.Sum(), 6)
		t.AssertEQ(array1, array2)

	})
}

func TestArray_CountValues(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		a1 := []string{"a", "b", "c", "d", "e", "d"}
		array1 := garray.NewArrayFrom(a1)
		array2 := array1.CountValues()
		t.Assert(len(array2), 5)
		t.Assert(array2["b"], 1)
		t.Assert(array2["d"], 2)
	})
}

func TestArray_LockFunc(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		s1 := []string{"a", "b", "c", "d"}
		a1 := garray.NewArrayFrom(s1, true)

		ch1 := make(chan int64, 3)
		ch2 := make(chan int64, 3)
		// go1
		go a1.LockFunc(func(n1 garray.Array[string]) { // 读写锁
			time.Sleep(2 * time.Second) // 暂停2秒
			err := n1.Set(2, "g")
			t.AssertNE(err, nil)
			ch2 <- gconv.Int64(time.Now().UnixNano() / 1000 / 1000)
		})

		// go2
		go func() {
			time.Sleep(100 * time.Millisecond) // 故意暂停0.01秒,等go1执行锁后，再开始执行.
			ch1 <- gconv.Int64(time.Now().UnixNano() / 1000 / 1000)
			a1.Len()
			ch1 <- gconv.Int64(time.Now().UnixNano() / 1000 / 1000)
		}()

		t1 := <-ch1
		t2 := <-ch1
		<-ch2 // 等待go1完成

		// 防止ci抖动,以豪秒为单位
		t.AssertGT(t2-t1, 20) // go1加的读写互斥锁，所go2读的时候被阻塞。
		t.Assert(a1.Contains("g"), true)
	})
}

func TestArray_RLockFunc(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		s1 := []string{"a", "b", "c", "d"}
		a1 := garray.NewArrayFrom(s1, true)

		ch1 := make(chan int64, 3)
		ch2 := make(chan int64, 1)
		// go1
		go a1.RLockFunc(func(n1 garray.Array[string]) { // 读锁
			time.Sleep(2 * time.Second) // 暂停1秒
			err := n1.Set(2, "g")
			t.AssertNE(err, nil)
			ch2 <- gconv.Int64(time.Now().UnixNano() / 1000 / 1000)
		})

		// go2
		go func() {
			time.Sleep(100 * time.Millisecond) // 故意暂停0.01秒,等go1执行锁后，再开始执行.
			ch1 <- gconv.Int64(time.Now().UnixNano() / 1000 / 1000)
			a1.Len()
			ch1 <- gconv.Int64(time.Now().UnixNano() / 1000 / 1000)
		}()

		t1 := <-ch1
		t2 := <-ch1
		<-ch2 // 等待go1完成

		// 防止ci抖动,以豪秒为单位
		t.AssertLT(t2-t1, 20) // go1加的读锁，所go2读的时候，并没有阻塞。
		t.Assert(a1.Contains("g"), true)
	})
}

func TestArray_Json(t *testing.T) {
	// pointer
	gtest.C(t, func(t *gtest.T) {
		s1 := []string{"a", "b", "d", "c"}
		a1 := garray.NewArrayFrom(s1)
		b1, err1 := json.Marshal(a1)
		b2, err2 := json.Marshal(s1)
		t.Assert(b1, b2)
		t.Assert(err1, err2)

		a2 := garray.New[string]()
		err2 = json.UnmarshalUseNumber(b2, &a2)
		t.Assert(err2, nil)
		t.Assert(a2.Slice(), s1)

		var a3 garray.StdArray[string]
		err := json.UnmarshalUseNumber(b2, &a3)
		t.AssertNil(err)
		t.Assert(a3.Slice(), s1)
	})
	// value.
	gtest.C(t, func(t *gtest.T) {
		s1 := []string{"a", "b", "d", "c"}
		a1 := *garray.NewArrayFrom(s1)
		b1, err1 := json.Marshal(a1)
		b2, err2 := json.Marshal(s1)
		t.Assert(b1, b2)
		t.Assert(err1, err2)

		a2 := garray.New[string]()
		err2 = json.UnmarshalUseNumber(b2, &a2)
		t.Assert(err2, nil)
		t.Assert(a2.Slice(), s1)

		var a3 garray.StdArray[string]
		err := json.UnmarshalUseNumber(b2, &a3)
		t.AssertNil(err)
		t.Assert(a3.Slice(), s1)
	})
	// pointer
	gtest.C(t, func(t *gtest.T) {
		type User struct {
			Name   string
			Scores *garray.StdArray[int]
		}
		data := map[string]interface{}{
			"Name":   "john",
			"Scores": []int{99, 100, 98},
		}
		b, err := json.Marshal(data)
		t.AssertNil(err)

		user := new(User)
		err = json.UnmarshalUseNumber(b, user)
		t.AssertNil(err)
		t.Assert(user.Name, data["Name"])
		t.Assert(user.Scores, data["Scores"])
	})
	// value
	gtest.C(t, func(t *gtest.T) {
		type User struct {
			Name   string
			Scores garray.StdArray[int]
		}
		data := map[string]interface{}{
			"Name":   "john",
			"Scores": []int{99, 100, 98},
		}
		b, err := json.Marshal(data)
		t.AssertNil(err)

		user := new(User)
		err = json.UnmarshalUseNumber(b, user)
		t.AssertNil(err)
		t.Assert(user.Name, data["Name"])
		t.Assert(user.Scores, data["Scores"])
	})
}

func TestArray_Iterator(t *testing.T) {
	slice := []string{"a", "b", "d", "c"}
	array := garray.NewArrayFrom[string](slice)
	gtest.C(t, func(t *gtest.T) {
		array.ForEachAsc(func(k int, v string) bool {
			t.Assert(v, slice[k])
			return true
		})
	})
	gtest.C(t, func(t *gtest.T) {
		array.ForEachAsc(func(k int, v string) bool {
			t.Assert(v, slice[k])
			return true
		})
	})
	gtest.C(t, func(t *gtest.T) {
		array.ForEachDesc(func(k int, v string) bool {
			t.Assert(v, slice[k])
			return true
		})
	})
	gtest.C(t, func(t *gtest.T) {
		index := 0
		array.ForEachAsc(func(k int, v string) bool {
			index++
			return false
		})
		t.Assert(index, 1)
	})
	gtest.C(t, func(t *gtest.T) {
		index := 0
		array.ForEachAsc(func(k int, v string) bool {
			index++
			return false
		})
		t.Assert(index, 1)
	})
	gtest.C(t, func(t *gtest.T) {
		index := 0
		array.ForEachDesc(func(k int, v string) bool {
			index++
			return false
		})
		t.Assert(index, 1)
	})
}

func TestArray_RemoveValue(t *testing.T) {
	slice := []string{"a", "b", "d", "c"}
	array := garray.NewArrayFrom[string](slice)
	gtest.C(t, func(t *gtest.T) {
		t.Assert(array.RemoveValue("e"), false)
		t.Assert(array.RemoveValue("b"), true)
		t.Assert(array.RemoveValue("a"), true)
		t.Assert(array.RemoveValue("c"), true)
		t.Assert(array.RemoveValue("f"), false)
	})
}

func TestArray_RemoveValues(t *testing.T) {
	slice := []string{"a", "b", "d", "c"}
	array := garray.NewArrayFrom[string](slice)
	gtest.C(t, func(t *gtest.T) {
		array.Remove("a", "b", "c")
		t.Assert(array.Slice(), []string{"d"})
	})
}

func TestArray_UnmarshalValue(t *testing.T) {
	type V struct {
		Name  string
		Array *garray.StdArray[byte]
	}
	type VInt struct {
		Name  string
		Array *garray.StdArray[int]
	}
	// JSON
	gtest.C(t, func(t *gtest.T) {
		var v *V
		err := gconv.Struct(map[string]interface{}{
			"name":  "john",
			"array": []byte(`[1,2,3]`),
		}, &v)
		t.AssertNil(err)
		t.Assert(v.Name, "john")
		t.Assert(v.Array.Slice(), []byte{1, 2, 3})
	})
	// Map
	gtest.C(t, func(t *gtest.T) {
		var v *VInt
		err := gconv.Struct(map[string]interface{}{
			"name":  "john",
			"array": []int{1, 2, 3},
		}, &v)
		t.AssertNil(err)
		t.Assert(v.Name, "john")
		t.Assert(v.Array.Slice(), []int{1, 2, 3})
	})
}

func TestArray_FilterNil(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		values := []*exampleElement{
			{code: 0},
			{code: 1},
			{code: 2},
		}
		array := garray.NewArrayFromCopy[*exampleElement](values)
		t.Assert(array.FilterNil().Slice(), values)
	})
	gtest.C(t, func(t *gtest.T) {
		values := []*exampleElement{
			nil,
			{code: 0},
			{code: 1},
			{code: 2},
			nil,
		}
		array := garray.NewArrayFromCopy(values)
		t.Assert(array.FilterNil().Slice(), []*exampleElement{
			{code: 0},
			{code: 1},
			{code: 2},
		})
	})
}

func TestArray_Filter(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		values := []*exampleElement{
			{code: 0},
			{code: 1},
			{code: 2},
		}
		array := garray.NewArrayFromCopy(values)
		t.Assert(array.Filter(func(index int, value *exampleElement) bool {
			return empty.IsNil(value)
		}).Slice(), values)
	})
	gtest.C(t, func(t *gtest.T) {
		values := []*exampleElement{
			nil,
			{code: 0},
			{code: 1},
			{code: 2},
			nil,
		}
		array := garray.NewArrayFromCopy(values)
		t.Assert(array.Filter(func(index int, value *exampleElement) bool {
			return empty.IsNil(value)
		}).Slice(), []*exampleElement{
			{code: 0},
			{code: 1},
			{code: 2},
		})
	})
	gtest.C(t, func(t *gtest.T) {
		values := []*exampleElement{
			{},
			{code: 0},
			{code: 1},
			{code: 2},
		}
		array := garray.NewArrayFromCopy(values)

		t.Assert(array.Filter(func(index int, value *exampleElement) bool {
			return empty.IsEmpty(value)
		}).Slice(), []*exampleElement{
			{code: 1},
			{code: 2},
		})
	})
	gtest.C(t, func(t *gtest.T) {
		values := []*exampleElement{
			{code: 1},
			{code: 2},
		}
		array := garray.NewArrayFromCopy(values)

		t.Assert(array.Filter(func(index int, value *exampleElement) bool {
			return empty.IsEmpty(value)
		}).Slice(), values)
	})
}

func TestArray_FilterEmpty(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		values := []*exampleElement{
			{},
			{code: 0},
			{code: 1},
			{code: 2},
		}
		array := garray.NewArrayFromCopy(values)
		t.Assert(array.FilterEmpty().Slice(), []*exampleElement{
			{code: 1},
			{code: 2},
		})
	})
	gtest.C(t, func(t *gtest.T) {
		values := []*exampleElement{
			{code: 3},
			{code: 1},
			{code: 2},
		}
		array := garray.NewArrayFromCopy(values)
		t.Assert(array.FilterEmpty().Slice(), values)
	})
}

func TestArray_Walk(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		array := garray.NewArrayFrom[string]([]string{"1", "2"})
		t.Assert(array.Walk(func(value string) string {
			return "key-" + value
		}), []string{"key-1", "key-2"})
	})
}

// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

// go test *.go -bench=".*" -benchmem

package glist_test

import (
	"testing"

	"github.com/wesleywu/gcontainer/glist"
)

var (
	l = glist.New[int](true)
)

func Benchmark_PushBack(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			l.PushBack(i)
			i++
		}
	})
}

func Benchmark_PushFront(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			l.PushFront(i)
			i++
		}
	})
}

func Benchmark_Len(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Len()
		}
	})
}

func Benchmark_PopFront(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.PopFront()
		}
	})
}

func Benchmark_PopBack(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.PopBack()
		}
	})
}

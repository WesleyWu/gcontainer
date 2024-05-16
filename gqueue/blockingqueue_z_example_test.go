// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package gqueue_test

import (
	"context"
	"fmt"
	"time"

	"github.com/wesleywu/gcontainer/gqueue"
	"github.com/wesleywu/gcontainer/gtimer"
)

type exampleQueueItem struct {
	index int
}

func ExampleNew() {
	n := 10
	q := gqueue.New[int](100)

	// Producer
	for i := 0; i < n; i++ {
		q.Push(i)
	}

	// Close the queue in three seconds.
	gtimer.SetTimeout(context.Background(), time.Second*3, func(ctx context.Context) error {
		q.Close()
		return nil
	})

	// The consumer constantly reads the queue data.
	// If there is no data in the queue, it will block.
	// The queue is read using the queue.C property exposed
	// by the queue object and the selectIO multiplexing syntax
	// example:
	// for {
	//    select {
	//        case v := <-queue.C:
	//            if v != nil {
	//                fmt.Println(v)
	//            } else {
	//                return
	//            }
	//    }
	// }
	for {
		v, ok := q.Pop()
		if ok {
			fmt.Print(v)
		} else {
			break
		}
	}

	// Output:
	// 0123456789
}

func ExampleBlockingQueue_Push() {
	q := gqueue.New[int]()

	for i := 0; i < 10; i++ {
		q.Push(i)
	}

	fmt.Println(q.MustPop())
	fmt.Println(q.MustPop())
	fmt.Println(q.MustPop())

	// Output:
	// 0
	// 1
	// 2
}

func ExampleBlockingQueue_Pop() {
	q := gqueue.New[int]()

	for i := 0; i < 10; i++ {
		q.Push(i)
	}

	v, ok := q.Pop()
	fmt.Println(v, ok)
	v, ok = q.Pop()
	fmt.Println(v, ok)
	v, ok = q.Pop()
	fmt.Println(v, ok)
	q.Close()
	v, ok = q.Pop()
	fmt.Println(v, ok)

	// Output:
	// 0 true
	// 1 true
	// 2 true
	// 0 false
}

func ExampleBlockingQueue_MustPop() {
	q := gqueue.New[*exampleQueueItem]()

	for i := 0; i < 10; i++ {
		q.Push(&exampleQueueItem{index: i})
	}

	fmt.Println(q.MustPop())
	fmt.Println(q.MustPop())
	fmt.Println(q.MustPop())
	q.Close()
	fmt.Println(q.MustPop())

	// Output:
	// &{0}
	// &{1}
	// &{2}
	// <nil>
}

func ExampleBlockingQueue_Close() {
	q := gqueue.New[int]()

	for i := 0; i < 10; i++ {
		q.Push(i)
	}

	time.Sleep(time.Millisecond)
	q.Close()

	fmt.Println(q.Len())
	fmt.Println(q.Pop())

	// May Output:
	// 0
	// <nil>
}

func ExampleBlockingQueue_Len() {
	q := gqueue.New[int]()

	q.Push(1)
	q.Push(2)

	fmt.Println(q.Len())

	// May Output:
	// 2
}

func ExampleBlockingQueue_Size() {
	q := gqueue.New[int]()

	q.Push(1)
	q.Push(2)

	// Size is alias of Len.
	fmt.Println(q.Size())

	// May Output:
	// 2
}

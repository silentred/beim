package lib

import (
	"container/list"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestLinkedList(t *testing.T) {
	list := createDBList(4)

	i, _ := list.Pop()
	fmt.Println(i, list.Len())

	printList(list)
}

func test(t *testing.T) {
	mutex := &sync.Mutex{}
	cond := sync.NewCond(mutex)

	list := createDBList(4)
	//ch := make(chan int, 1)

	go func() {
		// Unshift node to List
		for range time.Tick(time.Second / 2) {
			cond.L.Lock()

			if list.Len() > 10 {
				cond.Wait()
			}

			node := &ListNode{Data: rand.Int()}
			fmt.Println("unshifting one node with data", node.Data)
			list.Unshift(node)

			cond.Signal()

			cond.L.Unlock()
		}
	}()

	for {
		cond.L.Lock()
		for list.tail == nil {
			cond.Wait()
		}
		data, err := list.Pop()
		fmt.Println("pop one data", data)
		cond.Signal()
		if err == nil {

		}
		cond.L.Unlock()

		fmt.Println("sleep 2 sec to pop next node from list")
		time.Sleep(time.Second * 2)
	}

}

func TestList(t *testing.T) {
	l := list.New()

	ele := &list.Element{Value: 1}
	l.PushBack(ele)
	fmt.Println(l.Front().Value, l.Back().Value, l.Len())

	ele2 := &list.Element{Value: 2}
	l.PushBack(ele2)
	ele3 := &list.Element{Value: 3}
	fmt.Println(l.PushBack(ele3).Value)

	fmt.Println(l.Front().Value, l.Back().Value)

	fmt.Println(l.Len())
	fmt.Println(l.Remove(l.Back()))

	fmt.Println(l.Front().Value, l.Back().Value)
}

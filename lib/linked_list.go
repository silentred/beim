package lib

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyList = errors.New("list is empty")
)

const (
	FlagInQueue = iota
	FlagInflight
)

type DbLinkedList struct {
	head *ListNode
	tail *ListNode
}

type ListNode struct {
	Flag byte // in queue 0 ; inflight 1
	Data int
	prev *ListNode
	next *ListNode
}

// Unshift append node to head
func (list *DbLinkedList) Unshift(node *ListNode) {
	if list.head != nil {
		list.head.prev = node
	}

	node.next = list.head
	list.head = node

	if list.tail == nil {
		list.tail = node
	}
}

// Pop out the tail node's data
func (list *DbLinkedList) Pop() (int, error) {
	if list.tail == nil {
		return 0, ErrEmptyList
	}

	pop := list.tail
	list.tail = pop.prev
	pop.prev = nil

	if list.tail != nil {
		list.tail.next = nil
	} else {
		list.head = nil
	}

	return pop.Data, nil
}

func (list *DbLinkedList) Len() int {
	var n int
	var curr *ListNode
	curr = list.head
	for curr != nil {
		n++
		curr = curr.next
	}

	return n
}

func printList(list *DbLinkedList) {
	var currNode *ListNode
	currNode = list.head
	for currNode != nil {
		fmt.Printf("%d ", currNode.Data)
		currNode = currNode.next
	}

	fmt.Println()

	currNode = list.tail
	for currNode != nil {
		fmt.Printf("%d ", currNode.Data)
		currNode = currNode.prev
	}
}

func createDBList(n int) *DbLinkedList {
	list := &DbLinkedList{}

	for i := 0; i < n; i++ {
		node := &ListNode{Data: i}
		list.Unshift(node)
	}

	return list
}

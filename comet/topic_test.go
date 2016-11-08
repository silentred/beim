package comet

import (
	"fmt"
	"testing"
)

func TestParse(t *testing.T) {
	topic := "user/123/+/test"

	res, err := parseTopic(topic)
	fmt.Println(res, err, len(res))
}

func TestSub(t *testing.T) {
	info := &SubInfo{"user_123", 1}
	info2 := &SubInfo{"user_456", 1}
	root := newNode("")

	topic := "user/123"
	res, _ := parseTopic(topic)
	root.subscribe(res, info, actAdd)
	root.subscribe(res, info2, actAdd)

	//root.subscribe(res, info2, actDel)

	printNodes(root)

	topic = "user/+"
	res, _ = parseTopic(topic)
	var ret []*SubInfo
	fmt.Printf("%p \n", &ret)
	root.getInfos(res, &ret)

	fmt.Println(ret, len(ret))
}

func printNodes(node *topicNode) {
	fmt.Printf("%s - %d | ", node.Name, len(node.children))

	if node.children != nil && len(node.children) > 0 {
		for _, n := range node.children {
			printNodes(n)
		}
	}

	if len(node.Infos) > 0 {
		for _, info := range node.Infos {
			fmt.Println(node.Name, info)
		}
	}
}

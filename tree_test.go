package cob

import (
	"fmt"
	"testing"
)

func TestTreeNode(t *testing.T) {

	root := &treeNode{name: "/", children: make([]*treeNode, 0)}

	root.Put("/user/get/:id")
	root.Put("/user/get/aa")

	root.Put("/user/create/hello")
	root.Put("/order/create/aa")

	node := root.Get("/user/get/1")
	fmt.Println(node)

	node = root.Get("/user/create/hello")
	fmt.Println(node)

	node = root.Get("/order/create/aa")
	fmt.Println(node)

	node = root.Get("/user/get/aa")
	fmt.Println(node)

}

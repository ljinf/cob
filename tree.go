package cob

import (
	"strings"
)

type treeNode struct {
	name       string
	children   []*treeNode
	routerName string
	isEnd      bool
}

//put path:/user/name/:id

func (t *treeNode) Put(path string) {
	root := t
	strs := strings.Split(path, "/")
	for index, name := range strs {
		if index != 0 {
			isMatch := false
			for _, node := range t.children {
				if node.name == name {
					isMatch = true
					t = node
					break
				}
			}
			if !isMatch {
				isEnd := false
				if index == len(strs)-1 {
					isEnd = true
				}
				node := &treeNode{
					name:     name,
					children: make([]*treeNode, 0),
					isEnd:    isEnd,
				}
				t.children = append(t.children, node)
				t = node
			}
		}
	}
	t = root
}

//get path:/user/name/1
func (t *treeNode) Get(path string) *treeNode {
	strs := strings.Split(path, "/")
	routerName := ""
	for index, name := range strs {
		if index != 0 {
			//isMatch := false
			for _, node := range t.children {
				if node.name == name || strings.Contains(node.name, ":") || node.name == "*" {
					routerName += "/" + node.name
					node.routerName = routerName
					//isMatch = true
					t = node
					if index == len(strs)-1 {
						return node
					}
					break
				}
			}
			//if !isMatch {
			//	for _, node := range t.children {
			//		// /user/**
			//		// /user/namm/fsdf
			//		// /user/aa/bb
			//		if node.name == "**" {
			//			return node
			//		}
			//	}
			//}
		}
	}
	return nil
}

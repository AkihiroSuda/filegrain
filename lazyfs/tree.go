package lazyfs

import (
	"fmt"
	"strings"
)

type node struct {
	m map[string]*node // key: basename
	x interface{}
}

func newNode() *node {
	return &node{
		m: make(map[string]*node, 0),
		x: nil,
	}
}

type nodeManager struct {
	root *node
	sep  string
}

func newNodeManager(sep string) *nodeManager {
	return &nodeManager{
		root: newNode(),
		sep:  sep,
	}
}

func (nm *nodeManager) insert(path string, x interface{}) {
	n := nm.root
	for _, s := range strings.Split(path, nm.sep) {
		if s == "" {
			continue
		}
		if s == "." || s == ".." {
			panic(fmt.Errorf("disallowed path element: %q", s))
		}
		nn, ok := n.m[s]
		if !ok {
			nn = newNode()
			n.m[s] = nn
		}
		n = nn
	}
	n.x = x
}

// lookup returns node if found or nil.
// note that lookup("") returns nm.root (== lookup(nm.sep))
func (nm *nodeManager) lookup(path string) *node {
	n := nm.root
	for _, s := range strings.Split(path, nm.sep) {
		if s == "" {
			continue
		}
		if s == "." || s == ".." {
			panic(fmt.Errorf("disallowed path element: %q", s))
		}
		nn, ok := n.m[s]
		if !ok {
			return nil
		}
		n = nn
	}
	return n
}

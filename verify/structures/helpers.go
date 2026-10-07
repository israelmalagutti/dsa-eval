// Package structures verifies the "Stacks and queues", "Linked lists",
// "Heaps" and "Trees" entries of questions.json by differential testing.
package structures

import (
	"math/rand/v2"
	"strconv"
	"strings"
)

func newRNG(seed uint64) *rand.Rand { return rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15)) }

func randInts(r *rand.Rand, n, lo, hi int) []int {
	a := make([]int, n)
	for i := range a {
		a[i] = lo + r.IntN(hi-lo+1)
	}
	return a
}

// ---------- linked lists ----------

type ListNode struct {
	Val  int
	Next *ListNode
}

func fromSlice(a []int) *ListNode {
	d := &ListNode{}
	t := d
	for _, v := range a {
		t.Next = &ListNode{Val: v}
		t = t.Next
	}
	return d.Next
}

// toSlice walks an acyclic list (panics-safe cap to catch accidental cycles).
func toSlice(h *ListNode) []int {
	var a []int
	for n := 0; h != nil; h = h.Next {
		a = append(a, h.Val)
		n++
		if n > 1_000_000 {
			panic("cycle while converting list")
		}
	}
	return a
}

// ---------- trees ----------

type TreeNode struct {
	Val         int
	Left, Right *TreeNode
}

// randTree builds a random binary tree with n nodes and values in [lo,hi].
// Shape is random (includes skewed trees by chance; skew helpers are separate).
func randTree(r *rand.Rand, n, lo, hi int) *TreeNode {
	if n == 0 {
		return nil
	}
	leftN := r.IntN(n)
	root := &TreeNode{Val: lo + r.IntN(hi-lo+1)}
	root.Left = randTree(r, leftN, lo, hi)
	root.Right = randTree(r, n-1-leftN, lo, hi)
	return root
}

// randBST inserts random values (duplicates skipped) into a BST.
func randBST(r *rand.Rand, n, lo, hi int) (*TreeNode, []int) {
	var root *TreeNode
	seen := map[int]bool{}
	var vals []int
	for len(vals) < n && len(seen) < hi-lo+1 {
		v := lo + r.IntN(hi-lo+1)
		if seen[v] {
			continue
		}
		seen[v] = true
		vals = append(vals, v)
		root = bstInsert(root, v)
	}
	return root, vals
}

func bstInsert(n *TreeNode, v int) *TreeNode {
	if n == nil {
		return &TreeNode{Val: v}
	}
	if v < n.Val {
		n.Left = bstInsert(n.Left, v)
	} else {
		n.Right = bstInsert(n.Right, v)
	}
	return n
}

func allNodes(n *TreeNode) []*TreeNode {
	if n == nil {
		return nil
	}
	return append(append([]*TreeNode{n}, allNodes(n.Left)...), allNodes(n.Right)...)
}

func treeEqual(a, b *TreeNode) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Val == b.Val && treeEqual(a.Left, b.Left) && treeEqual(a.Right, b.Right)
}

// treeString is a structural dump used only in failure messages.
func treeString(n *TreeNode) string {
	if n == nil {
		return "#"
	}
	var sb strings.Builder
	sb.WriteString("(" + strconv.Itoa(n.Val) + " " + treeString(n.Left) + " " + treeString(n.Right) + ")")
	return sb.String()
}

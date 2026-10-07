package structures

import (
	"math"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// ===== tr-max-path-sum =====

// Literal (patched): best = -infinity; gain(null) = 0; gain(n): gl, gr = gain(children);
// best = max(best, n.val + max(0,gl) + max(0,gr)); return n.val + max(0, gl, gr).
func maxPathLiteral(root *TreeNode) int {
	best := math.MinInt
	var gain func(n *TreeNode) int
	gain = func(n *TreeNode) int {
		if n == nil {
			return 0
		}
		gl, gr := gain(n.Left), gain(n.Right)
		best = max(best, n.Val+max(0, gl)+max(0, gr))
		return n.Val + max(0, gl, gr)
	}
	gain(root)
	return best
}

// Brute: for every ordered pair of nodes (u,v), sum the unique u-v path via parent pointers.
func maxPathBrute(root *TreeNode) int {
	nodes := allNodes(root)
	parent := map[*TreeNode]*TreeNode{}
	depth := map[*TreeNode]int{root: 0}
	for _, n := range nodes {
		for _, c := range []*TreeNode{n.Left, n.Right} {
			if c != nil {
				parent[c] = n
				depth[c] = depth[n] + 1
			}
		}
	}
	best := math.MinInt
	for _, u := range nodes {
		for _, v := range nodes {
			a, b, s := u, v, 0
			for depth[a] > depth[b] {
				s += a.Val
				a = parent[a]
			}
			for depth[b] > depth[a] {
				s += b.Val
				b = parent[b]
			}
			for a != b {
				s += a.Val + b.Val
				a, b = parent[a], parent[b]
			}
			best = max(best, s+a.Val)
		}
	}
	return best
}

func TestMaxPathSum(t *testing.T) {
	r := newRNG(30)
	cases := []*TreeNode{
		{Val: -3},
		{Val: 1, Left: &TreeNode{Val: 2}, Right: &TreeNode{Val: 3}},
		{Val: -10, Left: &TreeNode{Val: 9}, Right: &TreeNode{Val: 20, Left: &TreeNode{Val: 15}, Right: &TreeNode{Val: 7}}},
		{Val: -2, Left: &TreeNode{Val: -1}},
		{Val: 1, Left: &TreeNode{Val: -2, Left: &TreeNode{Val: 3}}}, // skewed
	}
	for range 5000 {
		cases = append(cases, randTree(r, 1+r.IntN(10), -6, 6))
	}
	for _, c := range cases {
		if got, want := maxPathLiteral(c), maxPathBrute(c); got != want {
			t.Fatalf("%s: got %d want %d", treeString(c), got, want)
		}
	}
}

// ===== tr-kth-smallest-bst =====

// Literal (patched): stack empty, cur = root; while cur or stack: push lefts; pop; k -= 1;
// k == 0 -> return; cur = cur.right.
func kthSmallestLiteral(root *TreeNode, k int) int {
	var st []*TreeNode
	cur := root
	for cur != nil || len(st) > 0 {
		for cur != nil {
			st = append(st, cur)
			cur = cur.Left
		}
		cur = st[len(st)-1]
		st = st[:len(st)-1]
		k--
		if k == 0 {
			return cur.Val
		}
		cur = cur.Right
	}
	panic("k out of range")
}

func TestKthSmallestBST(t *testing.T) {
	r := newRNG(31)
	for it := range 3000 {
		root, vals := randBST(r, 1+r.IntN(15), -20, 20)
		if it == 0 { // right-skewed
			root, vals = nil, nil
			for v := range 8 {
				root = bstInsert(root, v)
				vals = append(vals, v)
			}
		}
		if it == 1 { // left-skewed
			root, vals = nil, nil
			for v := 8; v > 0; v-- {
				root = bstInsert(root, -v)
				vals = append(vals, -v)
			}
		}
		s := slices.Clone(vals)
		slices.Sort(s)
		for k := 1; k <= len(s); k++ {
			if got := kthSmallestLiteral(root, k); got != s[k-1] {
				t.Fatalf("%s k=%d: got %d want %d", treeString(root), k, got, s[k-1])
			}
		}
	}
}

// ===== tr-right-view =====

// Literal (patched) BFS: null -> []; per level process n = len(queue) nodes, record i == n-1;
// enqueue non-null left then right.
func rightViewBFS(root *TreeNode) []int {
	var out []int
	if root == nil {
		return out
	}
	q := []*TreeNode{root}
	for len(q) > 0 {
		n := len(q)
		for i := range n {
			x := q[0]
			q = q[1:]
			if i == n-1 {
				out = append(out, x.Val)
			}
			if x.Left != nil {
				q = append(q, x.Left)
			}
			if x.Right != nil {
				q = append(q, x.Right)
			}
		}
	}
	return out
}

// Literal DFS alternative: visit right first, record first node at each new depth.
func rightViewDFS(root *TreeNode) []int {
	var out []int
	var dfs func(n *TreeNode, d int)
	dfs = func(n *TreeNode, d int) {
		if n == nil {
			return
		}
		if d == len(out) {
			out = append(out, n.Val)
		}
		dfs(n.Right, d+1)
		dfs(n.Left, d+1)
	}
	dfs(root, 0)
	return out
}

// Brute: compute depth and preorder position of every node; per depth pick the
// node with the greatest "horizontal order" (preorder index among nodes at that depth
// visited left-to-right is the last one).
func rightViewBrute(root *TreeNode) []int {
	last := map[int]int{}
	maxD := -1
	var walk func(n *TreeNode, d int)
	walk = func(n *TreeNode, d int) { // left-to-right preorder: last write per depth is rightmost
		if n == nil {
			return
		}
		last[d] = n.Val
		maxD = max(maxD, d)
		walk(n.Left, d+1)
		walk(n.Right, d+1)
	}
	walk(root, 0)
	var out []int
	for d := 0; d <= maxD; d++ {
		out = append(out, last[d])
	}
	return out
}

func TestRightView(t *testing.T) {
	r := newRNG(32)
	cases := []*TreeNode{nil, {Val: 1}, {Val: 1, Left: &TreeNode{Val: 2, Right: &TreeNode{Val: 5}}, Right: &TreeNode{Val: 3}},
		{Val: 1, Left: &TreeNode{Val: 2, Left: &TreeNode{Val: 4}}}}
	for range 5000 {
		cases = append(cases, randTree(r, r.IntN(12), -3, 3))
	}
	for _, c := range cases {
		want := rightViewBrute(c)
		if got := rightViewBFS(c); !slices.Equal(got, want) {
			t.Fatalf("BFS %s: got %v want %v", treeString(c), got, want)
		}
		if got := rightViewDFS(c); !slices.Equal(got, want) {
			t.Fatalf("DFS %s: got %v want %v", treeString(c), got, want)
		}
	}
}

// ===== tr-validate-bst =====

// Literal (patched): valid(n, lo, hi), lo/hi = null means no bound: null -> true;
// lo != null and n.val <= lo -> false; hi != null and n.val >= hi -> false;
// recurse (left, lo, n.val) and (right, n.val, hi). Start with (root, null, null).
func validBSTLiteral(root *TreeNode) bool {
	var valid func(n *TreeNode, lo, hi *int) bool
	valid = func(n *TreeNode, lo, hi *int) bool {
		if n == nil {
			return true
		}
		if lo != nil && n.Val <= *lo {
			return false
		}
		if hi != nil && n.Val >= *hi {
			return false
		}
		return valid(n.Left, lo, &n.Val) && valid(n.Right, &n.Val, hi)
	}
	return valid(root, nil, nil)
}

// Brute: definition check — every node in left subtree < node < every node in right subtree.
func validBSTBrute(n *TreeNode) bool {
	if n == nil {
		return true
	}
	for _, x := range allNodes(n.Left) {
		if x.Val >= n.Val {
			return false
		}
	}
	for _, x := range allNodes(n.Right) {
		if x.Val <= n.Val {
			return false
		}
	}
	return validBSTBrute(n.Left) && validBSTBrute(n.Right)
}

func TestValidateBST(t *testing.T) {
	r := newRNG(33)
	cases := []*TreeNode{nil, {Val: 1},
		{Val: 5, Left: &TreeNode{Val: 1}, Right: &TreeNode{Val: 4, Left: &TreeNode{Val: 3}, Right: &TreeNode{Val: 6}}}, // false
		{Val: 5, Left: &TreeNode{Val: 4}, Right: &TreeNode{Val: 6, Left: &TreeNode{Val: 3}, Right: &TreeNode{Val: 7}}}, // grandchild violation
		{Val: 2, Left: &TreeNode{Val: 2}},                      // duplicate -> false (strict)
		{Val: math.MinInt, Right: &TreeNode{Val: math.MaxInt}}, // boundary values -> true
	}
	nValid := 0
	for range 8000 {
		if r.IntN(2) == 0 {
			b, _ := randBST(r, r.IntN(10), -10, 10)
			// randomly perturb one value to create near-misses
			if ns := allNodes(b); len(ns) > 0 && r.IntN(2) == 0 {
				ns[r.IntN(len(ns))].Val += r.IntN(7) - 3
			}
			cases = append(cases, b)
		} else {
			cases = append(cases, randTree(r, r.IntN(6), -3, 3))
		}
	}
	for _, c := range cases {
		want := validBSTBrute(c)
		if want {
			nValid++
		}
		if got := validBSTLiteral(c); got != want {
			t.Fatalf("%s: got %v want %v", treeString(c), got, want)
		}
	}
	if nValid < 1000 {
		t.Fatalf("too few valid cases generated: %d", nValid)
	}
}

// ===== tr-lca =====

// Literal: null/p/q -> node; l, r = recurse; both non-null -> node; else non-null one.
func lcaLiteral(n, p, q *TreeNode) *TreeNode {
	if n == nil || n == p || n == q {
		return n
	}
	l, r := lcaLiteral(n.Left, p, q), lcaLiteral(n.Right, p, q)
	if l != nil && r != nil {
		return n
	}
	if l != nil {
		return l
	}
	return r
}

// Brute: ancestor chains via parent map; deepest common entry.
func lcaBrute(root, p, q *TreeNode) *TreeNode {
	parent := map[*TreeNode]*TreeNode{}
	for _, n := range allNodes(root) {
		if n.Left != nil {
			parent[n.Left] = n
		}
		if n.Right != nil {
			parent[n.Right] = n
		}
	}
	anc := map[*TreeNode]bool{}
	for x := p; x != nil; x = parent[x] {
		anc[x] = true
	}
	for x := q; x != nil; x = parent[x] {
		if anc[x] {
			return x
		}
	}
	return nil
}

func TestLCA(t *testing.T) {
	r := newRNG(34)
	for range 3000 {
		root := randTree(r, 1+r.IntN(12), 0, 2) // duplicate values: identity matters
		ns := allNodes(root)
		for _, p := range ns {
			for _, q := range ns { // includes p == q and ancestor/descendant pairs
				if got, want := lcaLiteral(root, p, q), lcaBrute(root, p, q); got != want {
					t.Fatalf("%s: wrong LCA", treeString(root))
				}
			}
		}
	}
}

// ===== tr-serialize =====

// Literal (patched): preorder, str(val) or '#', joined by ','; deser splits on ',' and
// build() consumes tokens[i] with i += 1.
func serialize(n *TreeNode) string {
	var toks []string
	var walk func(n *TreeNode)
	walk = func(n *TreeNode) {
		if n == nil {
			toks = append(toks, "#")
			return
		}
		toks = append(toks, strconv.Itoa(n.Val))
		walk(n.Left)
		walk(n.Right)
	}
	walk(n)
	return strings.Join(toks, ",")
}

func deserialize(s string) *TreeNode {
	toks := strings.Split(s, ",")
	i := 0
	var build func() *TreeNode
	build = func() *TreeNode {
		tok := toks[i]
		i++
		if tok == "#" {
			return nil
		}
		v, err := strconv.Atoi(tok)
		if err != nil {
			panic(err)
		}
		n := &TreeNode{Val: v}
		n.Left = build()
		n.Right = build()
		return n
	}
	return build()
}

func TestSerialize(t *testing.T) {
	r := newRNG(35)
	cases := []*TreeNode{nil, {Val: -7}, {Val: math.MinInt, Left: &TreeNode{Val: math.MaxInt}},
		{Val: 1, Left: &TreeNode{Val: 1, Left: &TreeNode{Val: 1}}}} // skewed duplicates
	for range 5000 {
		cases = append(cases, randTree(r, r.IntN(15), -100, 100))
	}
	for _, c := range cases {
		if got := deserialize(serialize(c)); !treeEqual(got, c) {
			t.Fatalf("%s round-tripped to %s", treeString(c), treeString(got))
		}
		if got := deserializeBFS(serializeBFS(c)); !treeEqual(got, c) {
			t.Fatalf("BFS: %s round-tripped to %s", treeString(c), treeString(got))
		}
	}
}

// Alternative listed in the patched technique: BFS level order with null markers.
func serializeBFS(root *TreeNode) string {
	var toks []string
	q := []*TreeNode{root}
	for len(q) > 0 {
		n := q[0]
		q = q[1:]
		if n == nil {
			toks = append(toks, "#")
			continue
		}
		toks = append(toks, strconv.Itoa(n.Val))
		q = append(q, n.Left, n.Right)
	}
	return strings.Join(toks, ",")
}

func deserializeBFS(s string) *TreeNode {
	toks := strings.Split(s, ",")
	mk := func(tok string) *TreeNode {
		if tok == "#" {
			return nil
		}
		v, err := strconv.Atoi(tok)
		if err != nil {
			panic(err)
		}
		return &TreeNode{Val: v}
	}
	root := mk(toks[0])
	q := []*TreeNode{root}
	i := 1
	for len(q) > 0 {
		n := q[0]
		q = q[1:]
		if n == nil {
			continue
		}
		n.Left, n.Right = mk(toks[i]), mk(toks[i+1])
		i += 2
		q = append(q, n.Left, n.Right)
	}
	return root
}

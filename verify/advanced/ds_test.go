package advanced

import (
	"math"
	"slices"
	"strings"
	"testing"
)

// ---------- ds-word-dictionary ----------

type trieNode struct {
	children [26]*trieNode
	end      bool
}

type wordDict struct{ root *trieNode }

func newWordDict() *wordDict { return &wordDict{root: &trieNode{}} }

func (d *wordDict) addWord(w string) {
	n := d.root
	for i := 0; i < len(w); i++ {
		c := w[i] - 'a'
		if n.children[c] == nil {
			n.children[c] = &trieNode{}
		}
		n = n.children[c]
	}
	n.end = true
}

func (d *wordDict) search(w string) bool {
	var rec func(n *trieNode, i int) bool
	rec = func(n *trieNode, i int) bool {
		if i == len(w) {
			return n.end
		}
		if w[i] == '.' {
			for _, ch := range n.children {
				if ch != nil && rec(ch, i+1) {
					return true
				}
			}
			return false
		}
		ch := n.children[w[i]-'a']
		return ch != nil && rec(ch, i+1)
	}
	return rec(d.root, 0)
}

func naiveMatch(words []string, p string) bool {
	for _, w := range words {
		if len(w) != len(p) {
			continue
		}
		ok := true
		for i := range len(p) {
			if p[i] != '.' && p[i] != w[i] {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}

func TestWordDictionary(t *testing.T) {
	r := newRng(20)
	for range 300 {
		d := newWordDict()
		var model []string
		for range 200 {
			if r.IntN(2) == 0 {
				w := randStr(r, 1+r.IntN(4), "abc")
				d.addWord(w)
				model = append(model, w)
			} else {
				p := randStr(r, r.IntN(5), "abc..")
				if strings.Count(p, ".") > 2 { // patched prompt: at most 2 dots
					continue
				}
				if g, w := d.search(p), naiveMatch(model, p); g != w {
					t.Fatalf("search(%q) got=%v want=%v words=%v", p, g, w, model)
				}
			}
		}
	}
}

// ---------- ds-word-search-ii ----------

type wsNode struct {
	children [26]*wsNode
	word     string
}

func findWordsLiteral(g [][]byte, words []string) []string {
	root := &wsNode{}
	for _, w := range words {
		n := root
		for i := 0; i < len(w); i++ {
			c := w[i] - 'a'
			if n.children[c] == nil {
				n.children[c] = &wsNode{}
			}
			n = n.children[c]
		}
		n.word = w
	}
	m, nc := len(g), len(g[0])
	var res []string
	var dfs func(r, c int, parent *wsNode)
	dfs = func(r, c int, parent *wsNode) {
		ch := g[r][c]
		if ch == '#' {
			return
		}
		node := parent.children[ch-'a']
		if node == nil {
			return
		}
		if node.word != "" {
			res = append(res, node.word)
			node.word = "" // clear to avoid duplicates
		}
		g[r][c] = '#'
		for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			nr, ncc := r+d[0], c+d[1]
			if nr >= 0 && nr < m && ncc >= 0 && ncc < nc {
				dfs(nr, ncc, node)
			}
		}
		g[r][c] = ch
		// prune emptied leaf
		empty := node.word == ""
		for _, k := range node.children {
			if k != nil {
				empty = false
				break
			}
		}
		if empty {
			parent.children[ch-'a'] = nil
		}
	}
	for r := range m {
		for c := range nc {
			dfs(r, c, root)
		}
	}
	return res
}

func TestWordSearchII(t *testing.T) {
	r := newRng(21)
	check := func(g [][]byte, words []string) {
		orig := cloneGrid(g)
		got := findWordsLiteral(g, words)
		var want []string
		seen := map[string]bool{}
		for _, w := range words {
			if !seen[w] && bruteGridHas(orig, w) {
				want = append(want, w)
			}
			seen[w] = true
		}
		slices.Sort(got)
		slices.Sort(want)
		if !slices.Equal(got, want) {
			t.Fatalf("g=%q words=%v got=%v want=%v", orig, words, got, want)
		}
	}
	check([][]byte{[]byte("oaan"), []byte("etae"), []byte("ihkr"), []byte("iflv")},
		[]string{"oath", "pea", "eat", "rain"})
	check([][]byte{[]byte("ab"), []byte("cd")}, []string{"abcb"})
	check([][]byte{[]byte("a")}, []string{"a", "a", "aa"})
	for range 3000 {
		g := randGrid(r, 1+r.IntN(3), 1+r.IntN(4), "abc")
		k := 1 + r.IntN(8)
		words := make([]string, k)
		for i := range words {
			words[i] = randStr(r, 1+r.IntN(5), "abc")
		}
		check(g, words)
	}
}

// ---------- ds-range-min-updates ----------

type segTree struct {
	n    int
	tree []int
}

func newSegTree(a []int) *segTree {
	s := &segTree{n: len(a), tree: make([]int, 4*len(a))}
	var build func(node, l, r int)
	build = func(node, l, r int) {
		if l == r {
			s.tree[node] = a[l]
			return
		}
		m := (l + r) / 2
		build(2*node, l, m)
		build(2*node+1, m+1, r)
		s.tree[node] = min(s.tree[2*node], s.tree[2*node+1])
	}
	build(1, 0, s.n-1)
	return s
}

func (s *segTree) update(i, v int) {
	var rec func(node, l, r int)
	rec = func(node, l, r int) {
		if l == r {
			s.tree[node] = v
			return
		}
		m := (l + r) / 2
		if i <= m {
			rec(2*node, l, m)
		} else {
			rec(2*node+1, m+1, r)
		}
		s.tree[node] = min(s.tree[2*node], s.tree[2*node+1])
	}
	rec(1, 0, s.n-1)
}

func (s *segTree) query(ql, qr int) int {
	var rec func(node, l, r int) int
	rec = func(node, l, r int) int {
		if qr < l || r < ql {
			return math.MaxInt
		}
		if ql <= l && r <= qr {
			return s.tree[node]
		}
		m := (l + r) / 2
		return min(rec(2*node, l, m), rec(2*node+1, m+1, r))
	}
	return rec(1, 0, s.n-1)
}

func TestRangeMinUpdates(t *testing.T) {
	r := newRng(22)
	for range 500 {
		n := 1 + r.IntN(20)
		a := make([]int, n)
		for i := range a {
			a[i] = r.IntN(201) - 100
		}
		s := newSegTree(a)
		for range 200 {
			if r.IntN(2) == 0 {
				i, v := r.IntN(n), r.IntN(201)-100
				a[i] = v
				s.update(i, v)
			} else {
				l := r.IntN(n)
				rr := l + r.IntN(n-l)
				if g, w := s.query(l, rr), slices.Min(a[l:rr+1]); g != w {
					t.Fatalf("query(%d,%d) on %v got=%d want=%d", l, rr, a, g, w)
				}
			}
		}
	}
}

// ---------- ds-count-smaller-right ----------

func countSmallerLiteral(a []int) []int {
	n := len(a)
	sorted := slices.Clone(a)
	slices.Sort(sorted)
	sorted = slices.Compact(sorted)
	rank := make([]int, n)
	for i, v := range a {
		p, _ := slices.BinarySearch(sorted, v)
		rank[i] = p + 1 // 1-based, equal values share a rank
	}
	bit := make([]int, len(sorted)+1)
	query := func(i int) int {
		s := 0
		for ; i > 0; i -= i & -i {
			s += bit[i]
		}
		return s
	}
	update := func(i, d int) {
		for ; i < len(bit); i += i & -i {
			bit[i] += d
		}
	}
	ans := make([]int, n)
	for i := n - 1; i >= 0; i-- {
		ans[i] = query(rank[i] - 1)
		update(rank[i], 1)
	}
	return ans
}

func TestCountSmallerRight(t *testing.T) {
	r := newRng(23)
	check := func(a []int) {
		want := make([]int, len(a))
		for i := range a {
			for j := i + 1; j < len(a); j++ {
				if a[j] < a[i] {
					want[i]++
				}
			}
		}
		if g := countSmallerLiteral(a); !slices.Equal(g, want) {
			t.Fatalf("a=%v got=%v want=%v", a, g, want)
		}
	}
	check(nil)
	check([]int{1})
	check([]int{5, 2, 6, 1})
	check([]int{-1, -1})
	check([]int{2, 2, 2, 1})
	for range 5000 {
		n := r.IntN(12)
		a := make([]int, n)
		for i := range a {
			a[i] = r.IntN(9) - 4
		}
		check(a)
	}
}

// ---------- ds-lru-cache ----------
// Patched key_idea: put on an existing key updates the value and moves it after head.

type lruNode struct {
	key, val   int
	prev, next *lruNode
}

type lru struct {
	cap        int
	m          map[int]*lruNode
	head, tail *lruNode // sentinels; head.next is most recent
}

func newLRU(c int) *lru {
	h, t := &lruNode{}, &lruNode{}
	h.next, t.prev = t, h
	return &lru{cap: c, m: map[int]*lruNode{}, head: h, tail: t}
}

func (l *lru) unlink(n *lruNode) { n.prev.next, n.next.prev = n.next, n.prev }
func (l *lru) pushFront(n *lruNode) {
	n.prev, n.next = l.head, l.head.next
	l.head.next.prev = n
	l.head.next = n
}

func (l *lru) get(k int) int {
	n, ok := l.m[k]
	if !ok {
		return -1
	}
	l.unlink(n)
	l.pushFront(n)
	return n.val
}

func (l *lru) put(k, v int) {
	if n, ok := l.m[k]; ok {
		n.val = v
		l.unlink(n)
		l.pushFront(n)
		return
	}
	n := &lruNode{key: k, val: v}
	l.m[k] = n
	l.pushFront(n)
	if len(l.m) > l.cap {
		victim := l.tail.prev
		l.unlink(victim)
		delete(l.m, victim.key)
	}
}

// naive model: slice ordered most-recent-first.
type naiveLRU struct {
	cap  int
	keys []int
	vals map[int]int
}

func (n *naiveLRU) touch(k int) {
	i := slices.Index(n.keys, k)
	n.keys = slices.Delete(n.keys, i, i+1)
	n.keys = slices.Insert(n.keys, 0, k)
}

func (n *naiveLRU) get(k int) int {
	v, ok := n.vals[k]
	if !ok {
		return -1
	}
	n.touch(k)
	return v
}

func (n *naiveLRU) put(k, v int) {
	if _, ok := n.vals[k]; ok {
		n.vals[k] = v
		n.touch(k)
		return
	}
	n.vals[k] = v
	n.keys = slices.Insert(n.keys, 0, k)
	if len(n.keys) > n.cap {
		delete(n.vals, n.keys[len(n.keys)-1])
		n.keys = n.keys[:len(n.keys)-1]
	}
}

func TestLRUCache(t *testing.T) {
	r := newRng(24)
	for range 500 {
		c := 1 + r.IntN(4)
		a, b := newLRU(c), &naiveLRU{cap: c, vals: map[int]int{}}
		for op := range 300 {
			k := r.IntN(7)
			if r.IntN(2) == 0 {
				if g, w := a.get(k), b.get(k); g != w {
					t.Fatalf("cap=%d op=%d get(%d) got=%d want=%d", c, op, k, g, w)
				}
			} else {
				v := r.IntN(100)
				a.put(k, v)
				b.put(k, v)
			}
			if len(a.m) != len(b.keys) {
				t.Fatalf("size mismatch")
			}
		}
	}
}

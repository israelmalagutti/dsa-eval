package graphs

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"
)

// ---------- gr-word-ladder ----------

// Literal (patched key_idea): set of list words; queue (beginWord, 1); pop (w, d),
// return d if w == endWord; enqueue (candidate, d+1) for each one-letter candidate
// in the set, removing it; return 0 when the queue empties.
func ladderLiteral(begin, end string, list []string) int {
	set := map[string]bool{}
	for _, w := range list {
		set[w] = true
	}
	type item struct {
		w string
		d int
	}
	q := []item{{begin, 1}}
	for len(q) > 0 {
		it := q[0]
		q = q[1:]
		if it.w == end {
			return it.d
		}
		b := []byte(it.w)
		for i := range b {
			orig := b[i]
			for ch := byte('a'); ch <= 'z'; ch++ {
				b[i] = ch
				cand := string(b)
				if set[cand] {
					delete(set, cand)
					q = append(q, item{cand, it.d + 1})
				}
			}
			b[i] = orig
		}
	}
	return 0
}

// Oracle: explicit graph over {begin} ∪ list, Floyd-Warshall.
func ladderBrute(begin, end string, list []string) int {
	nodes := []string{begin}
	for _, w := range list {
		if !slices.Contains(nodes, w) {
			nodes = append(nodes, w)
		}
	}
	endIdx := slices.Index(nodes[1:], end)
	if endIdx < 0 {
		return 0
	}
	endIdx++
	n := len(nodes)
	d := make([][]int, n)
	for i := range d {
		d[i] = make([]int, n)
		for j := range d[i] {
			diff := 0
			for k := range nodes[i] {
				if nodes[i][k] != nodes[j][k] {
					diff++
				}
			}
			switch {
			case i == j:
				d[i][j] = 0
			case diff == 1:
				d[i][j] = 1
			default:
				d[i][j] = inf
			}
		}
	}
	for k := range n {
		for i := range n {
			for j := range n {
				d[i][j] = min(d[i][j], d[i][k]+d[k][j])
			}
		}
	}
	if d[0][endIdx] >= inf {
		return 0
	}
	return d[0][endIdx] + 1
}

func randWord(rng *rand.Rand, l, alpha int) string {
	b := make([]byte, l)
	for i := range b {
		b[i] = byte('a' + rng.IntN(alpha))
	}
	return string(b)
}

func TestWordLadder(t *testing.T) {
	type tc struct {
		b, e string
		l    []string
	}
	edge := []tc{
		{"hit", "cog", []string{"hot", "dot", "dog", "lot", "log", "cog"}},
		{"hit", "cog", []string{"hot", "dot", "dog", "lot", "log"}}, // end missing
		{"a", "c", []string{"a", "b", "c"}},
		{"ab", "cd", []string{}},
		{"ab", "ad", []string{"ad", "ad"}},
		{"hot", "dog", []string{"hot", "dog"}},   // unreachable
		{"aa", "bb", []string{"aa", "ab", "bb"}}, // begin in list
	}
	for i, c := range edge {
		if got, want := ladderLiteral(c.b, c.e, c.l), ladderBrute(c.b, c.e, c.l); got != want {
			t.Errorf("edge %d: got %d want %d", i, got, want)
		}
	}
	rng := newRng()
	for range 5000 {
		l, alpha := 1+rng.IntN(3), 2+rng.IntN(3)
		b := randWord(rng, l, alpha)
		e := randWord(rng, l, alpha)
		if b == e {
			continue // prompt/LeetCode: beginWord != endWord
		}
		var list []string
		for range rng.IntN(10) {
			list = append(list, randWord(rng, l, alpha))
		}
		if rng.IntN(4) != 0 {
			list = append(list, e)
		}
		rng.Shuffle(len(list), func(i, j int) { list[i], list[j] = list[j], list[i] })
		if got, want := ladderLiteral(b, e, list), ladderBrute(b, e, list); got != want {
			t.Fatalf("%s->%s %v: got %d want %d", b, e, list, got, want)
		}
	}
}

// ---------- gr-course-schedule ----------

// Literal Kahn's algorithm.
func courseLiteral(n int, pre [][2]int) []int {
	adj := make([][]int, n)
	indeg := make([]int, n)
	for _, p := range pre {
		a, b := p[0], p[1]
		adj[b] = append(adj[b], a)
		indeg[a]++
	}
	var q []int
	for i := range n {
		if indeg[i] == 0 {
			q = append(q, i)
		}
	}
	order := []int{}
	for len(q) > 0 {
		u := q[0]
		q = q[1:]
		order = append(order, u)
		for _, v := range adj[u] {
			indeg[v]--
			if indeg[v] == 0 {
				q = append(q, v)
			}
		}
	}
	if len(order) < n {
		return []int{}
	}
	return order
}

func validOrder(n int, pre [][2]int, order []int) bool {
	if len(order) != n {
		return false
	}
	pos := make([]int, n)
	seen := make([]bool, n)
	for i, c := range order {
		if c < 0 || c >= n || seen[c] {
			return false
		}
		seen[c] = true
		pos[c] = i
	}
	for _, p := range pre {
		if pos[p[1]] >= pos[p[0]] {
			return false
		}
	}
	return true
}

func TestCourseSchedule(t *testing.T) {
	rng := newRng()
	type tc struct {
		n   int
		pre [][2]int
	}
	cases := []tc{
		{1, nil}, {1, [][2]int{{0, 0}}}, {2, [][2]int{{1, 0}}}, {2, [][2]int{{1, 0}, {0, 1}}},
		{3, [][2]int{{1, 0}, {1, 0}, {2, 1}}}, {4, nil}, {3, [][2]int{{2, 2}, {1, 0}}},
	}
	for range 5000 {
		n := 1 + rng.IntN(6)
		var pre [][2]int
		for range rng.IntN(2 * n) {
			pre = append(pre, [2]int{rng.IntN(n), rng.IntN(n)})
		}
		cases = append(cases, tc{n, pre})
	}
	for _, c := range cases {
		possible := permutations(c.n, func(p []int) bool { return validOrder(c.n, c.pre, p) })
		got := courseLiteral(c.n, c.pre)
		if possible && !validOrder(c.n, c.pre, got) {
			t.Fatalf("n=%d %v: invalid order %v", c.n, c.pre, got)
		}
		if !possible && len(got) != 0 {
			t.Fatalf("n=%d %v: expected empty, got %v", c.n, c.pre, got)
		}
	}
}

// ---------- gr-redundant-connection ----------

// Literal: DSU with path compression + union by rank; return first edge with find(u)==find(v).
func redundantLiteral(n int, edges [][2]int) [2]int {
	par := make([]int, n+1)
	rank := make([]int, n+1)
	for i := range par {
		par[i] = i
	}
	var find func(x int) int
	find = func(x int) int {
		if par[x] != x {
			par[x] = find(par[x])
		}
		return par[x]
	}
	for _, e := range edges {
		a, b := find(e[0]), find(e[1])
		if a == b {
			return e
		}
		if rank[a] < rank[b] {
			a, b = b, a
		}
		par[b] = a
		if rank[a] == rank[b] {
			rank[a]++
		}
	}
	return [2]int{}
}

// Oracle: last edge whose removal leaves a connected graph (n nodes, n-1 edges => tree).
func redundantBrute(n int, edges [][2]int) [2]int {
	zero := make([][2]int, len(edges))
	for i, e := range edges {
		zero[i] = [2]int{e[0] - 1, e[1] - 1}
	}
	for i := len(edges) - 1; i >= 0; i-- {
		if connected(n, zero, i) {
			return edges[i]
		}
	}
	return [2]int{}
}

func TestRedundantConnection(t *testing.T) {
	rng := newRng()
	for range 5000 {
		n := 3 + rng.IntN(8)
		var edges [][2]int
		has := map[[2]int]bool{}
		perm := rng.Perm(n)
		for i := 1; i < n; i++ {
			u, v := perm[i], perm[rng.IntN(i)]
			has[[2]int{min(u, v), max(u, v)}] = true
			edges = append(edges, [2]int{u + 1, v + 1})
		}
		for {
			u, v := rng.IntN(n), rng.IntN(n)
			if u != v && !has[[2]int{min(u, v), max(u, v)}] {
				edges = append(edges, [2]int{u + 1, v + 1})
				break
			}
		}
		rng.Shuffle(len(edges), func(i, j int) { edges[i], edges[j] = edges[j], edges[i] })
		if got, want := redundantLiteral(n, edges), redundantBrute(n, edges); got != want {
			t.Fatalf("n=%d %v: got %v want %v", n, edges, got, want)
		}
	}
}

// ---------- gr-directed-cycle ----------

// Literal: three-color DFS.
func directedCycleLiteral(adj [][]int) bool {
	color := make([]int, len(adj))
	var dfs func(u int) bool
	dfs = func(u int) bool {
		color[u] = 1
		for _, v := range adj[u] {
			if color[v] == 1 {
				return true
			}
			if color[v] == 0 && dfs(v) {
				return true
			}
		}
		color[u] = 2
		return false
	}
	for u := range adj {
		if color[u] == 0 && dfs(u) {
			return true
		}
	}
	return false
}

// Oracle: acyclic iff some permutation puts every edge forward.
func directedCycleBrute(adj [][]int) bool {
	n := len(adj)
	acyclic := permutations(n, func(p []int) bool {
		pos := make([]int, n)
		for i, v := range p {
			pos[v] = i
		}
		for u := range adj {
			for _, v := range adj[u] {
				if pos[v] <= pos[u] {
					return false
				}
			}
		}
		return true
	})
	return !acyclic
}

// Literal undirected: visited neighbor that is not the parent => cycle.
func undirectedCycleParent(n int, edges [][2]int) bool {
	adj := make([][]int, n)
	for _, e := range edges {
		adj[e[0]] = append(adj[e[0]], e[1])
		adj[e[1]] = append(adj[e[1]], e[0])
	}
	vis := make([]bool, n)
	var dfs func(u, p int) bool
	dfs = func(u, p int) bool {
		vis[u] = true
		for _, v := range adj[u] {
			if v == p {
				continue
			}
			if vis[v] || dfs(v, u) {
				return true
			}
		}
		return false
	}
	for u := range n {
		if !vis[u] && dfs(u, -1) {
			return true
		}
	}
	return false
}

// Literal undirected DSU: find(u)==find(v) => cycle.
func undirectedCycleDSU(n int, edges [][2]int) bool {
	par := make([]int, n)
	for i := range par {
		par[i] = i
	}
	find := func(x int) int {
		for par[x] != x {
			x = par[x]
		}
		return x
	}
	for _, e := range edges {
		a, b := find(e[0]), find(e[1])
		if a == b {
			return true
		}
		par[a] = b
	}
	return false
}

// Oracle: forest iff E == V - components (components via BFS).
func undirectedCycleBrute(n int, edges [][2]int) bool {
	adj := make([][]int, n)
	for _, e := range edges {
		adj[e[0]] = append(adj[e[0]], e[1])
		adj[e[1]] = append(adj[e[1]], e[0])
	}
	seen := make([]bool, n)
	comps := 0
	for s := range n {
		if seen[s] {
			continue
		}
		comps++
		seen[s] = true
		q := []int{s}
		for len(q) > 0 {
			u := q[0]
			q = q[1:]
			for _, v := range adj[u] {
				if !seen[v] {
					seen[v] = true
					q = append(q, v)
				}
			}
		}
	}
	return len(edges) > n-comps
}

func TestDirectedCycle(t *testing.T) {
	edge := [][][]int{
		{{}},
		{{0}},             // self-loop
		{{1}, {}},         // single edge
		{{1}, {0}},        // 2-cycle
		{{1, 1}, {}},      // duplicate edge, no cycle
		{{1, 2}, {2}, {}}, // diamond-ish, cross edge to finished node
		{{}, {2}, {1}},    // disconnected with cycle
	}
	for i, adj := range edge {
		if got, want := directedCycleLiteral(adj), directedCycleBrute(adj); got != want {
			t.Errorf("edge %d %v: got %v want %v", i, adj, got, want)
		}
	}
	rng := newRng()
	for range 5000 {
		n := 1 + rng.IntN(6)
		adj := make([][]int, n)
		for range rng.IntN(2 * n) {
			u := rng.IntN(n)
			adj[u] = append(adj[u], rng.IntN(n))
		}
		if got, want := directedCycleLiteral(adj), directedCycleBrute(adj); got != want {
			t.Fatalf("%v: got %v want %v", adj, got, want)
		}
	}
}

// Also run on multigraphs (self-loops, parallel edges): the parent-by-node check
// still detects a doubled edge, from the endpoint that sees it second.
func TestUndirectedCycle(t *testing.T) {
	rng := newRng()
	for i := range 6000 {
		multi := i%2 == 1
		n := 1 + rng.IntN(7)
		has := map[[2]int]bool{}
		var edges [][2]int
		for range rng.IntN(2 * n) {
			u, v := rng.IntN(n), rng.IntN(n)
			k := [2]int{min(u, v), max(u, v)}
			if !multi && (u == v || has[k]) {
				continue
			}
			has[k] = true
			edges = append(edges, [2]int{u, v})
		}
		want := undirectedCycleBrute(n, edges)
		if got := undirectedCycleParent(n, edges); got != want {
			t.Fatalf("parent n=%d %v: got %v want %v", n, edges, got, want)
		}
		if got := undirectedCycleDSU(n, edges); got != want {
			t.Fatalf("dsu n=%d %v: got %v want %v", n, edges, got, want)
		}
	}
}

// ---------- gr-possible-bipartition ----------

// Literal: BFS 2-coloring.
func bipartitionLiteral(n int, dislikes [][2]int) bool {
	adj := make([][]int, n+1)
	for _, d := range dislikes {
		adj[d[0]] = append(adj[d[0]], d[1])
		adj[d[1]] = append(adj[d[1]], d[0])
	}
	color := make([]int, n+1)
	for i := range color {
		color[i] = -1
	}
	for s := 1; s <= n; s++ {
		if color[s] != -1 {
			continue
		}
		color[s] = 0
		q := []int{s}
		for len(q) > 0 {
			u := q[0]
			q = q[1:]
			for _, v := range adj[u] {
				if color[v] == color[u] {
					return false
				}
				if color[v] == -1 {
					color[v] = 1 - color[u]
					q = append(q, v)
				}
			}
		}
	}
	return true
}

// Oracle: try all 2^n colorings.
func bipartitionBrute(n int, dislikes [][2]int) bool {
	for mask := 0; mask < 1<<n; mask++ {
		ok := true
		for _, d := range dislikes {
			if (mask>>(d[0]-1))&1 == (mask>>(d[1]-1))&1 {
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

func TestPossibleBipartition(t *testing.T) {
	type tc struct {
		n int
		d [][2]int
	}
	cases := []tc{
		{1, nil}, {2, [][2]int{{1, 2}}}, {3, [][2]int{{1, 2}, {2, 3}, {1, 3}}},
		{4, [][2]int{{1, 2}, {1, 3}, {2, 4}}}, {5, [][2]int{{1, 2}, {2, 3}, {3, 4}, {4, 5}, {1, 5}}},
		{4, [][2]int{{1, 2}, {1, 2}, {3, 4}}},
	}
	rng := newRng()
	for range 5000 {
		n := 1 + rng.IntN(9)
		var d [][2]int
		for range rng.IntN(2 * n) {
			a, b := 1+rng.IntN(n), 1+rng.IntN(n)
			if a != b {
				d = append(d, [2]int{a, b})
			}
		}
		cases = append(cases, tc{n, d})
	}
	for _, c := range cases {
		if got, want := bipartitionLiteral(c.n, c.d), bipartitionBrute(c.n, c.d); got != want {
			t.Fatalf("n=%d %v: got %v want %v", c.n, c.d, got, want)
		}
	}
}

// ---------- gr-critical-connections ----------

// Literal Tarjan bridges (parent skipped by node).
func bridgesLiteral(n int, edges [][2]int) map[[2]int]bool {
	adj := make([][]int, n)
	for _, e := range edges {
		adj[e[0]] = append(adj[e[0]], e[1])
		adj[e[1]] = append(adj[e[1]], e[0])
	}
	disc := make([]int, n)
	low := make([]int, n)
	for i := range disc {
		disc[i] = -1
	}
	time := 0
	out := map[[2]int]bool{}
	var dfs func(u, p int)
	dfs = func(u, p int) {
		disc[u], low[u] = time, time
		time++
		for _, v := range adj[u] {
			if v == p {
				continue
			}
			if disc[v] == -1 {
				dfs(v, u)
				low[u] = min(low[u], low[v])
				if low[v] > disc[u] {
					out[[2]int{min(u, v), max(u, v)}] = true
				}
			} else {
				low[u] = min(low[u], disc[v])
			}
		}
	}
	for u := range n {
		if disc[u] == -1 {
			dfs(u, -1)
		}
	}
	return out
}

// Oracle: remove each edge, check connectivity.
func bridgesBrute(n int, edges [][2]int) map[[2]int]bool {
	out := map[[2]int]bool{}
	for i, e := range edges {
		if !connected(n, edges, i) {
			out[[2]int{min(e[0], e[1]), max(e[0], e[1])}] = true
		}
	}
	return out
}

func TestCriticalConnections(t *testing.T) {
	type tc struct {
		n int
		e [][2]int
	}
	cases := []tc{
		{1, nil}, {2, [][2]int{{0, 1}}}, {4, [][2]int{{0, 1}, {1, 2}, {2, 0}, {1, 3}}},
		{3, [][2]int{{0, 1}, {1, 2}, {2, 0}}},
	}
	rng := newRng()
	for range 5000 {
		n := 1 + rng.IntN(9)
		cases = append(cases, tc{n, randomConnected(rng, n, rng.IntN(n+1))})
	}
	for _, c := range cases {
		got, want := bridgesLiteral(c.n, c.e), bridgesBrute(c.n, c.e)
		if fmt.Sprint(got) != fmt.Sprint(want) {
			t.Fatalf("n=%d %v: got %v want %v", c.n, c.e, got, want)
		}
	}
}

// Why the patched prompt says "no repeated connections": with parallel edges,
// skipping the parent by node wrongly reports a doubled edge as a bridge.
func TestCriticalConnectionsMultigraphCaveat(t *testing.T) {
	e := [][2]int{{0, 1}, {0, 1}}
	if len(bridgesLiteral(2, e)) != 1 || len(bridgesBrute(2, e)) != 0 {
		t.Fatal("expected parent-by-node to misreport doubled edge")
	}
}

// ---------- alternative techniques added to `technique` ----------

// gr-course-schedule alt: DFS reverse postorder (three colors; gray hit => cycle).
func courseDFS(n int, pre [][2]int) []int {
	adj := make([][]int, n)
	for _, p := range pre {
		adj[p[1]] = append(adj[p[1]], p[0])
	}
	color := make([]int, n)
	var post []int
	var dfs func(u int) bool
	dfs = func(u int) bool {
		color[u] = 1
		for _, v := range adj[u] {
			if color[v] == 1 || (color[v] == 0 && !dfs(v)) {
				return false
			}
		}
		color[u] = 2
		post = append(post, u)
		return true
	}
	for u := range n {
		if color[u] == 0 && !dfs(u) {
			return []int{}
		}
	}
	slices.Reverse(post)
	return post
}

func TestCourseScheduleDFS(t *testing.T) {
	rng := newRng()
	for range 5000 {
		n := 1 + rng.IntN(6)
		var pre [][2]int
		for range rng.IntN(2 * n) {
			pre = append(pre, [2]int{rng.IntN(n), rng.IntN(n)})
		}
		possible := permutations(n, func(p []int) bool { return validOrder(n, pre, p) })
		got := courseDFS(n, pre)
		if possible != validOrder(n, pre, got) || (!possible && len(got) != 0) {
			t.Fatalf("n=%d %v: got %v possible=%v", n, pre, got, possible)
		}
	}
}

// gr-directed-cycle alt: Kahn's; cycle iff fewer than V nodes come out.
func directedCycleKahn(adj [][]int) bool {
	n := len(adj)
	indeg := make([]int, n)
	for u := range adj {
		for _, v := range adj[u] {
			indeg[v]++
		}
	}
	var q []int
	for i := range n {
		if indeg[i] == 0 {
			q = append(q, i)
		}
	}
	out := 0
	for len(q) > 0 {
		u := q[0]
		q = q[1:]
		out++
		for _, v := range adj[u] {
			indeg[v]--
			if indeg[v] == 0 {
				q = append(q, v)
			}
		}
	}
	return out < n
}

func TestDirectedCycleKahn(t *testing.T) {
	rng := newRng()
	for range 5000 {
		n := 1 + rng.IntN(6)
		adj := make([][]int, n)
		for range rng.IntN(2 * n) {
			u := rng.IntN(n)
			adj[u] = append(adj[u], rng.IntN(n))
		}
		if got, want := directedCycleKahn(adj), directedCycleBrute(adj); got != want {
			t.Fatalf("%v: got %v want %v", adj, got, want)
		}
	}
}

// gr-possible-bipartition alt: DSU. All of u's enemies must share a set;
// conflict if u ends up in the same set as an enemy.
func bipartitionDSU(n int, dislikes [][2]int) bool {
	adj := make([][]int, n+1)
	for _, d := range dislikes {
		adj[d[0]] = append(adj[d[0]], d[1])
		adj[d[1]] = append(adj[d[1]], d[0])
	}
	par := make([]int, n+1)
	for i := range par {
		par[i] = i
	}
	var find func(x int) int
	find = func(x int) int {
		if par[x] != x {
			par[x] = find(par[x])
		}
		return par[x]
	}
	for u := 1; u <= n; u++ {
		for _, v := range adj[u] {
			if find(u) == find(v) {
				return false
			}
			par[find(v)] = find(adj[u][0])
		}
	}
	return true
}

func TestPossibleBipartitionDSU(t *testing.T) {
	rng := newRng()
	for range 5000 {
		n := 1 + rng.IntN(9)
		var d [][2]int
		for range rng.IntN(2 * n) {
			a, b := 1+rng.IntN(n), 1+rng.IntN(n)
			if a != b {
				d = append(d, [2]int{a, b})
			}
		}
		if got, want := bipartitionDSU(n, d), bipartitionBrute(n, d); got != want {
			t.Fatalf("n=%d %v: got %v want %v", n, d, got, want)
		}
	}
}

package graphs

import (
	"math/rand/v2"
	"slices"
)

const inf = 1 << 40

var dirs4 = [4][2]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}}

func newRng() *rand.Rand { return rand.New(rand.NewPCG(1, 2)) }

func copyGrid[T any](g [][]T) [][]T {
	out := make([][]T, len(g))
	for i := range g {
		out[i] = slices.Clone(g[i])
	}
	return out
}

// permutations calls f with every permutation of 0..n-1; stops early if f returns true.
func permutations(n int, f func([]int) bool) bool {
	p := make([]int, n)
	for i := range p {
		p[i] = i
	}
	var rec func(k int) bool
	rec = func(k int) bool {
		if k == n {
			return f(p)
		}
		for i := k; i < n; i++ {
			p[k], p[i] = p[i], p[k]
			if rec(k + 1) {
				return true
			}
			p[k], p[i] = p[i], p[k]
		}
		return false
	}
	return rec(0)
}

// connected reports whether undirected graph on nodes 0..n-1 is connected (BFS), skipping edge index skip.
func connected(n int, edges [][2]int, skip int) bool {
	if n == 0 {
		return true
	}
	adj := make([][]int, n)
	for i, e := range edges {
		if i == skip {
			continue
		}
		adj[e[0]] = append(adj[e[0]], e[1])
		adj[e[1]] = append(adj[e[1]], e[0])
	}
	seen := make([]bool, n)
	seen[0] = true
	q := []int{0}
	cnt := 1
	for len(q) > 0 {
		u := q[0]
		q = q[1:]
		for _, v := range adj[u] {
			if !seen[v] {
				seen[v] = true
				cnt++
				q = append(q, v)
			}
		}
	}
	return cnt == n
}

// randomConnected returns a simple connected undirected graph on n nodes (0-based).
func randomConnected(rng *rand.Rand, n, extra int) [][2]int {
	has := map[[2]int]bool{}
	var edges [][2]int
	add := func(u, v int) bool {
		if u == v {
			return false
		}
		a, b := min(u, v), max(u, v)
		if has[[2]int{a, b}] {
			return false
		}
		has[[2]int{a, b}] = true
		edges = append(edges, [2]int{u, v})
		return true
	}
	perm := rng.Perm(n)
	for i := 1; i < n; i++ {
		add(perm[i], perm[rng.IntN(i)])
	}
	for range extra {
		add(rng.IntN(n), rng.IntN(n))
	}
	rng.Shuffle(len(edges), func(i, j int) { edges[i], edges[j] = edges[j], edges[i] })
	return edges
}

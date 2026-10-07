package graphs

import (
	"container/heap"
	"math/rand/v2"
	"slices"
	"testing"
)

type wedge struct{ u, v, w int }

// ---------- gr-network-delay ----------

type pq [][2]int // (dist, node)

func (h pq) Len() int           { return len(h) }
func (h pq) Less(i, j int) bool { return h[i][0] < h[j][0] }
func (h pq) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *pq) Push(x any)        { *h = append(*h, x.([2]int)) }
func (h *pq) Pop() any {
	old := *h
	x := old[len(old)-1]
	*h = old[:len(old)-1]
	return x
}

// Literal Dijkstra with lazy deletion; nodes 1..n.
func networkDelayLiteral(times []wedge, n, k int) int {
	adj := make([][]wedge, n+1)
	for _, e := range times {
		adj[e.u] = append(adj[e.u], e)
	}
	dist := make([]int, n+1)
	for i := range dist {
		dist[i] = inf
	}
	dist[k] = 0
	h := &pq{{0, k}}
	for h.Len() > 0 {
		top := heap.Pop(h).([2]int)
		d, u := top[0], top[1]
		if d > dist[u] {
			continue
		}
		for _, e := range adj[u] {
			if d+e.w < dist[e.v] {
				dist[e.v] = d + e.w
				heap.Push(h, [2]int{dist[e.v], e.v})
			}
		}
	}
	ans := 0
	for i := 1; i <= n; i++ {
		if dist[i] >= inf {
			return -1
		}
		ans = max(ans, dist[i])
	}
	return ans
}

// Oracle: Bellman-Ford with n-1 full relaxation rounds.
func networkDelayBrute(times []wedge, n, k int) int {
	dist := make([]int, n+1)
	for i := range dist {
		dist[i] = inf
	}
	dist[k] = 0
	for range n {
		for _, e := range times {
			if dist[e.u] < inf {
				dist[e.v] = min(dist[e.v], dist[e.u]+e.w)
			}
		}
	}
	ans := 0
	for i := 1; i <= n; i++ {
		if dist[i] >= inf {
			return -1
		}
		ans = max(ans, dist[i])
	}
	return ans
}

func randWeighted(rng *rand.Rand, n, m, maxW, base int) []wedge {
	var es []wedge
	for range m {
		es = append(es, wedge{base + rng.IntN(n), base + rng.IntN(n), rng.IntN(maxW + 1)})
	}
	return es
}

func TestNetworkDelay(t *testing.T) {
	type tc struct {
		e    []wedge
		n, k int
	}
	cases := []tc{
		{nil, 1, 1},
		{nil, 2, 1},
		{[]wedge{{2, 1, 1}, {2, 3, 1}, {3, 4, 1}}, 4, 2},
		{[]wedge{{1, 2, 0}, {2, 1, 0}}, 2, 1},
		{[]wedge{{1, 2, 5}, {1, 2, 1}, {1, 1, 3}}, 2, 1},
		{[]wedge{{1, 2, 1}}, 2, 2},
	}
	rng := newRng()
	for range 5000 {
		n := 1 + rng.IntN(7)
		cases = append(cases, tc{randWeighted(rng, n, rng.IntN(3*n), 1+rng.IntN(10), 1), n, 1 + rng.IntN(n)})
	}
	for _, c := range cases {
		if got, want := networkDelayLiteral(c.e, c.n, c.k), networkDelayBrute(c.e, c.n, c.k); got != want {
			t.Fatalf("%v n=%d k=%d: got %d want %d", c.e, c.n, c.k, got, want)
		}
	}
}

// ---------- gr-cheapest-k-stops ----------

// Literal: k+1 rounds of copy-relax. inf = 1<<40 so inf+w cannot overflow int64;
// -1 if dist[dst] >= inf. Prices may be 0 (patched prompt: price >= 0).
func cheapestLiteral(n int, flights []wedge, src, dst, k int) int {
	dist := make([]int, n)
	for i := range dist {
		dist[i] = inf
	}
	dist[src] = 0
	for range k + 1 {
		tmp := slices.Clone(dist)
		for _, f := range flights {
			tmp[f.v] = min(tmp[f.v], dist[f.u]+f.w)
		}
		dist = tmp
	}
	if dist[dst] >= inf {
		return -1
	}
	return dist[dst]
}

// Oracle: enumerate every walk from src with at most k+1 edges.
func cheapestBrute(n int, flights []wedge, src, dst, k int) int {
	best := inf
	var rec func(u, cost, edges int)
	rec = func(u, cost, edges int) {
		if u == dst {
			best = min(best, cost)
		}
		if edges == k+1 {
			return
		}
		for _, f := range flights {
			if f.u == u {
				rec(f.v, cost+f.w, edges+1)
			}
		}
	}
	rec(src, 0, 0)
	if best >= inf {
		return -1
	}
	return best
}

func TestCheapestKStops(t *testing.T) {
	type tc struct {
		n       int
		f       []wedge
		s, d, k int
	}
	cases := []tc{
		{4, []wedge{{0, 1, 100}, {1, 2, 100}, {2, 0, 100}, {1, 3, 600}, {2, 3, 200}}, 0, 3, 1},
		{3, []wedge{{0, 1, 100}, {1, 2, 100}, {0, 2, 500}}, 0, 2, 0},
		{3, []wedge{{0, 1, 100}, {1, 2, 100}, {0, 2, 500}}, 0, 2, 1},
		{3, []wedge{{0, 1, 1}}, 0, 2, 2},
		// order-sensitivity check: chain listed in path order, k=0
		{4, []wedge{{0, 1, 1}, {1, 2, 1}, {2, 3, 1}, {0, 3, 100}}, 0, 3, 0},
	}
	rng := newRng()
	for range 4000 {
		n := 2 + rng.IntN(5)
		s, d := rng.IntN(n), rng.IntN(n)
		if s == d {
			continue
		}
		var f []wedge
		for _, e := range randWeighted(rng, n, rng.IntN(2*n+1), 9, 0) {
			if e.u != e.v {
				f = append(f, e)
			}
		}
		cases = append(cases, tc{n, f, s, d, rng.IntN(4)})
	}
	for _, c := range cases {
		if got, want := cheapestLiteral(c.n, c.f, c.s, c.d, c.k), cheapestBrute(c.n, c.f, c.s, c.d, c.k); got != want {
			t.Fatalf("%+v: got %d want %d", c, got, want)
		}
	}
}

// ---------- gr-threshold-city ----------

// Literal Floyd-Warshall (patched key_idea: min over parallel roads; scan i ascending,
// take i when count <= best so ties go to the larger index).
func thresholdLiteral(n int, roads []wedge, thr int) int {
	d := make([][]int, n)
	for i := range d {
		d[i] = make([]int, n)
		for j := range d[i] {
			d[i][j] = inf
		}
		d[i][i] = 0
	}
	for _, r := range roads {
		d[r.u][r.v] = min(d[r.u][r.v], r.w)
		d[r.v][r.u] = min(d[r.v][r.u], r.w)
	}
	for k := range n {
		for i := range n {
			for j := range n {
				d[i][j] = min(d[i][j], d[i][k]+d[k][j])
			}
		}
	}
	best, bestCnt := -1, inf
	for i := range n {
		cnt := 0
		for j := range n {
			if j != i && d[i][j] <= thr {
				cnt++
			}
		}
		if cnt <= bestCnt {
			best, bestCnt = i, cnt
		}
	}
	return best
}

// Oracle: shortest distances by enumerating all simple paths.
func thresholdBrute(n int, roads []wedge, thr int) int {
	best, bestCnt := -1, inf
	for s := range n {
		dist := make([]int, n)
		for i := range dist {
			dist[i] = inf
		}
		vis := make([]bool, n)
		var rec func(u, c int)
		rec = func(u, c int) {
			dist[u] = min(dist[u], c)
			vis[u] = true
			for _, r := range roads {
				for _, p := range [][2]int{{r.u, r.v}, {r.v, r.u}} {
					if p[0] == u && !vis[p[1]] {
						rec(p[1], c+r.w)
					}
				}
			}
			vis[u] = false
		}
		rec(s, 0)
		cnt := 0
		for j := range n {
			if j != s && dist[j] <= thr {
				cnt++
			}
		}
		if cnt <= bestCnt {
			best, bestCnt = s, cnt
		}
	}
	return best
}

func TestThresholdCity(t *testing.T) {
	type tc struct {
		n   int
		r   []wedge
		thr int
	}
	cases := []tc{
		{1, nil, 5},
		{2, nil, 5},
		{4, []wedge{{0, 1, 3}, {1, 2, 1}, {1, 3, 4}, {2, 3, 1}}, 4},
		{5, []wedge{{0, 1, 2}, {0, 4, 8}, {1, 2, 3}, {1, 4, 2}, {2, 3, 1}, {3, 4, 1}}, 2},
		{3, []wedge{{0, 1, 5}, {0, 1, 1}, {1, 2, 9}, {1, 2, 2}}, 2}, // parallel roads, cheaper one second
		{3, []wedge{{0, 1, 1}, {0, 1, 5}}, 1},
	}
	rng := newRng()
	for range 3000 {
		n := 1 + rng.IntN(6)
		has := map[[2]int]bool{}
		var r []wedge
		for _, e := range randWeighted(rng, n, rng.IntN(2*n+1), 6, 0) {
			e.w++ // positive weights
			k := [2]int{min(e.u, e.v), max(e.u, e.v)}
			if e.u == e.v || (has[k] && rng.IntN(2) == 0) {
				continue // parallel roads allowed by the patched prompt
			}
			has[k] = true
			r = append(r, e)
		}
		cases = append(cases, tc{n, r, rng.IntN(12)})
	}
	for _, c := range cases {
		if got, want := thresholdLiteral(c.n, c.r, c.thr), thresholdBrute(c.n, c.r, c.thr); got != want {
			t.Fatalf("%+v: got %d want %d", c, got, want)
		}
	}
}

// ---------- gr-min-cost-points ----------

func manhattan(a, b [2]int) int {
	return max(a[0]-b[0], b[0]-a[0]) + max(a[1]-b[1], b[1]-a[1])
}

// Literal Prim O(n^2).
func mstPrim(pts [][2]int) int {
	n := len(pts)
	best := make([]int, n)
	in := make([]bool, n)
	for i := range best {
		best[i] = inf
	}
	best[0] = 0
	total := 0
	for range n {
		u := -1
		for v := range n {
			if !in[v] && (u == -1 || best[v] < best[u]) {
				u = v
			}
		}
		in[u] = true
		total += best[u]
		for v := range n {
			if !in[v] {
				best[v] = min(best[v], manhattan(pts[u], pts[v]))
			}
		}
	}
	return total
}

// Literal Kruskal.
func mstKruskal(pts [][2]int) int {
	n := len(pts)
	var es []wedge
	for i := range n {
		for j := i + 1; j < n; j++ {
			es = append(es, wedge{i, j, manhattan(pts[i], pts[j])})
		}
	}
	slices.SortFunc(es, func(a, b wedge) int { return a.w - b.w })
	par := make([]int, n)
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
	total := 0
	for _, e := range es {
		a, b := find(e.u), find(e.v)
		if a != b {
			par[a] = b
			total += e.w
		}
	}
	return total
}

// Oracle: enumerate all (n-1)-subsets of edges, keep spanning ones (BFS connectivity).
func mstBrute(pts [][2]int) int {
	n := len(pts)
	if n <= 1 {
		return 0
	}
	var all [][2]int
	for i := range n {
		for j := i + 1; j < n; j++ {
			all = append(all, [2]int{i, j})
		}
	}
	best := inf
	chosen := make([][2]int, 0, n-1)
	var rec func(start, cost int)
	rec = func(start, cost int) {
		if len(chosen) == n-1 {
			if cost < best && connected(n, chosen, -1) {
				best = cost
			}
			return
		}
		for i := start; i < len(all); i++ {
			e := all[i]
			chosen = append(chosen, e)
			rec(i+1, cost+manhattan(pts[e[0]], pts[e[1]]))
			chosen = chosen[:len(chosen)-1]
		}
	}
	rec(0, 0)
	return best
}

func TestMinCostPoints(t *testing.T) {
	cases := [][][2]int{
		{{0, 0}},
		{{0, 0}, {0, 0}},
		{{0, 0}, {2, 2}, {3, 10}, {5, 2}, {7, 0}},
		{{3, 12}, {-2, 5}, {-4, 1}},
	}
	rng := newRng()
	for range 2000 {
		n := 1 + rng.IntN(6)
		pts := make([][2]int, n)
		for i := range pts {
			pts[i] = [2]int{rng.IntN(9) - 4, rng.IntN(9) - 4}
		}
		cases = append(cases, pts)
	}
	for _, pts := range cases {
		want := mstBrute(pts)
		if got := mstPrim(pts); got != want {
			t.Fatalf("prim %v: got %d want %d", pts, got, want)
		}
		if got := mstKruskal(pts); got != want {
			t.Fatalf("kruskal %v: got %d want %d", pts, got, want)
		}
	}
}

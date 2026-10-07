package graphs

import (
	"math/rand/v2"
	"testing"
)

// ---------- gr-islands ----------

// Literal: for each unvisited '1': count++, then flood fill it, setting cells to '0'.
func islandsLiteral(grid [][]byte) int {
	g := copyGrid(grid)
	m := len(g)
	count := 0
	var fill func(r, c int)
	fill = func(r, c int) {
		if r < 0 || r >= m || c < 0 || c >= len(g[r]) || g[r][c] != '1' {
			return
		}
		g[r][c] = '0'
		for _, d := range dirs4 {
			fill(r+d[0], c+d[1])
		}
	}
	for r := range g {
		for c := range g[r] {
			if g[r][c] == '1' {
				count++
				fill(r, c)
			}
		}
	}
	return count
}

// Oracle: DSU over adjacent land cells; components = land - successful unions.
func islandsBrute(g [][]byte) int {
	m := len(g)
	if m == 0 {
		return 0
	}
	n := len(g[0])
	par := make([]int, m*n)
	for i := range par {
		par[i] = i
	}
	var find func(x int) int
	find = func(x int) int {
		for par[x] != x {
			x = par[x]
		}
		return x
	}
	land := 0
	for r := range m {
		for c := range n {
			if g[r][c] != '1' {
				continue
			}
			land++
			for _, nb := range [][2]int{{r + 1, c}, {r, c + 1}} {
				if nb[0] < m && nb[1] < n && g[nb[0]][nb[1]] == '1' {
					a, b := find(r*n+c), find(nb[0]*n+nb[1])
					if a != b {
						par[a] = b
						land--
					}
				}
			}
		}
	}
	return land
}

func randByteGrid(rng *rand.Rand, m, n int, p float64) [][]byte {
	g := make([][]byte, m)
	for r := range g {
		g[r] = make([]byte, n)
		for c := range g[r] {
			g[r][c] = '0'
			if rng.Float64() < p {
				g[r][c] = '1'
			}
		}
	}
	return g
}

func TestIslands(t *testing.T) {
	edge := [][][]byte{
		{},
		{{'0'}},
		{{'1'}},
		{{'1', '0', '1', '0', '1'}},
		{{'1', '1'}, {'1', '1'}},
		{{'1', '0'}, {'0', '1'}}, // diagonal does not connect
		{{'1', '1', '1'}, {'1', '0', '1'}, {'1', '1', '1'}},
	}
	for i, g := range edge {
		if got, want := islandsLiteral(g), islandsBrute(g); got != want {
			t.Errorf("edge %d: got %d want %d", i, got, want)
		}
	}
	rng := newRng()
	for range 5000 {
		g := randByteGrid(rng, 1+rng.IntN(7), 1+rng.IntN(7), rng.Float64())
		if got, want := islandsLiteral(g), islandsBrute(g); got != want {
			t.Fatalf("%q: got %d want %d", g, got, want)
		}
	}
}

// ---------- gr-rotting-oranges ----------

// Literal (patched key_idea): while the queue is non-empty and fresh > 0:
// minutes++, pop exactly this round's cells, rot fresh neighbors (fresh--, enqueue).
// Return minutes if fresh == 0, else -1.
// (The original "answer is the number of levels" counted level 0: [[2,1]] gave 2, want 1.)
func rottingLiteral(grid [][]int) int {
	g := copyGrid(grid)
	var q [][2]int
	fresh := 0
	for r := range g {
		for c := range g[r] {
			switch g[r][c] {
			case 2:
				q = append(q, [2]int{r, c})
			case 1:
				fresh++
			}
		}
	}
	minutes := 0
	for len(q) > 0 && fresh > 0 {
		minutes++
		var next [][2]int
		for _, p := range q {
			for _, d := range dirs4 {
				r, c := p[0]+d[0], p[1]+d[1]
				if r >= 0 && r < len(g) && c >= 0 && c < len(g[r]) && g[r][c] == 1 {
					g[r][c] = 2
					fresh--
					next = append(next, [2]int{r, c})
				}
			}
		}
		q = next
	}
	if fresh == 0 {
		return minutes
	}
	return -1
}

// Oracle: literal minute-by-minute simulation of the prompt.
func rottingBrute(grid [][]int) int {
	g := copyGrid(grid)
	for minute := 0; ; minute++ {
		fresh := 0
		next := copyGrid(g)
		changed := false
		for r := range g {
			for c := range g[r] {
				if g[r][c] == 1 {
					fresh++
					for _, d := range dirs4 {
						rr, cc := r+d[0], c+d[1]
						if rr >= 0 && rr < len(g) && cc >= 0 && cc < len(g[rr]) && g[rr][cc] == 2 {
							next[r][c] = 2
							changed = true
						}
					}
				}
			}
		}
		if fresh == 0 {
			return minute
		}
		if !changed {
			return -1
		}
		g = next
	}
}

func randIntGrid(rng *rand.Rand, m, n, k int) [][]int {
	g := make([][]int, m)
	for r := range g {
		g[r] = make([]int, n)
		for c := range g[r] {
			g[r][c] = rng.IntN(k)
		}
	}
	return g
}

var rottingEdge = [][][]int{
	{{0}},
	{{1}},
	{{2}},
	{{2, 1}},
	{{2, 1, 1}, {1, 1, 0}, {0, 1, 1}},
	{{2, 1, 1}, {0, 1, 1}, {1, 0, 1}},
	{{0, 2}},
	{{2, 0, 1}},
	{{2, 1, 1, 1, 2}},
}

func TestRotting(t *testing.T) {
	for i, g := range rottingEdge {
		if got, want := rottingLiteral(g), rottingBrute(g); got != want {
			t.Errorf("edge %d %v: got %d want %d", i, g, got, want)
		}
	}
	rng := newRng()
	for range 5000 {
		g := randIntGrid(rng, 1+rng.IntN(6), 1+rng.IntN(6), 3)
		if got, want := rottingLiteral(g), rottingBrute(g); got != want {
			t.Fatalf("%v: got %d want %d", g, got, want)
		}
	}
}

// ---------- gr-01-bfs-arrows ----------
// Arrow codes (patched prompt, LeetCode 1368): 1 right, 2 left, 3 down, 4 up; code-1 indexes dirs4.

// Literal: deque; relaxing a 0-cost edge pushes front, a 1-cost edge pushes back.
func arrowsLiteral(g [][]int) int {
	m, n := len(g), len(g[0])
	dist := make([]int, m*n)
	for i := range dist {
		dist[i] = inf
	}
	dist[0] = 0
	dq := []int{0}
	for len(dq) > 0 {
		u := dq[0]
		dq = dq[1:]
		r, c := u/n, u%n
		for k, d := range dirs4 {
			rr, cc := r+d[0], c+d[1]
			if rr < 0 || rr >= m || cc < 0 || cc >= n {
				continue
			}
			w := 1
			if g[r][c]-1 == k {
				w = 0
			}
			v := rr*n + cc
			if dist[u]+w < dist[v] {
				dist[v] = dist[u] + w
				if w == 0 {
					dq = append([]int{v}, dq...)
				} else {
					dq = append(dq, v)
				}
			}
		}
	}
	return dist[m*n-1]
}

// Oracle: try every arrow assignment; cost = cells changed; valid if following
// arrows from (0,0) reaches (m-1,n-1).
func arrowsBrute(g [][]int) int {
	m, n := len(g), len(g[0])
	N := m * n
	cur := make([]int, N)
	best := inf
	var rec func(i, cost int)
	rec = func(i, cost int) {
		if cost >= best {
			return
		}
		if i == N {
			r, c := 0, 0
			for range N {
				if r == m-1 && c == n-1 {
					break
				}
				d := dirs4[cur[r*n+c]]
				r, c = r+d[0], c+d[1]
				if r < 0 || r >= m || c < 0 || c >= n {
					return
				}
			}
			if r == m-1 && c == n-1 {
				best = cost
			}
			return
		}
		orig := g[i/n][i%n]
		for a := range 4 {
			cur[i] = a
			add := 0
			if a != orig-1 {
				add = 1
			}
			rec(i+1, cost+add)
		}
	}
	rec(0, 0)
	return best
}

func TestArrows(t *testing.T) {
	edge := [][][]int{
		{{1}},
		{{2}},
		{{1, 1, 1}},
		{{2, 2, 2}},
		{{1, 3}, {4, 2}},
		{{3, 3}, {1, 1}},
		{{4, 4}, {4, 4}},
		{{1, 1, 3}, {3, 2, 2}, {1, 1, 4}},
		{{1, 1, 1, 1}, {2, 2, 2, 2}, {1, 1, 1, 1}, {2, 2, 2, 2}}, // LeetCode example 1: 3
	}
	for i, g := range edge {
		if got, want := arrowsLiteral(g), arrowsBrute(g); got != want {
			t.Errorf("edge %d %v: got %d want %d", i, g, got, want)
		}
	}
	rng := newRng()
	for i := range 3000 {
		m, n := 1+rng.IntN(3), 1+rng.IntN(3)
		if m*n > 6 && i%5 != 0 {
			m, n = 2, 3
		}
		g := randIntGrid(rng, m, n, 4)
		for r := range g {
			for c := range g[r] {
				g[r][c]++
			}
		}
		if got, want := arrowsLiteral(g), arrowsBrute(g); got != want {
			t.Fatalf("%v: got %d want %d", g, got, want)
		}
	}
}

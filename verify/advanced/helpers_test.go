package advanced

import (
	"math/rand/v2"
	"slices"
	"strings"
)

func newRng(seed uint64) *rand.Rand { return rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15)) }

// canon sorts each inner slice's representation and the outer list, for set comparison.
func canonLists(xs [][]int, sortInner bool) []string {
	out := make([]string, 0, len(xs))
	for _, x := range xs {
		c := slices.Clone(x)
		if sortInner {
			slices.Sort(c)
		}
		var b strings.Builder
		for _, v := range c {
			b.WriteString(itoa(v))
			b.WriteByte(',')
		}
		out = append(out, b.String())
	}
	slices.Sort(out)
	return out
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var buf [24]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func randStr(r *rand.Rand, n int, alpha string) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = alpha[r.IntN(len(alpha))]
	}
	return string(b)
}

func randGrid(r *rand.Rand, m, n int, alpha string) [][]byte {
	g := make([][]byte, m)
	for i := range g {
		g[i] = []byte(randStr(r, n, alpha))
	}
	return g
}

func cloneGrid(g [][]byte) [][]byte {
	c := make([][]byte, len(g))
	for i := range g {
		c[i] = slices.Clone(g[i])
	}
	return c
}

// bruteGridHas: exhaustive simple-path enumeration with an explicit visited set.
func bruteGridHas(g [][]byte, w string) bool {
	if len(w) == 0 {
		return true
	}
	m := len(g)
	if m == 0 {
		return false
	}
	n := len(g[0])
	vis := make([][]bool, m)
	for i := range vis {
		vis[i] = make([]bool, n)
	}
	var walk func(r, c int, path []byte) bool
	walk = func(r, c int, path []byte) bool {
		path = append(path, g[r][c])
		if len(path) == len(w) {
			return string(path) == w
		}
		vis[r][c] = true
		defer func() { vis[r][c] = false }()
		for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			nr, nc := r+d[0], c+d[1]
			if nr >= 0 && nr < m && nc >= 0 && nc < n && !vis[nr][nc] {
				if walk(nr, nc, path) {
					return true
				}
			}
		}
		return false
	}
	for r := range m {
		for c := range n {
			if walk(r, c, nil) {
				return true
			}
		}
	}
	return false
}

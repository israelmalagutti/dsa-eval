package advanced

import (
	"slices"
	"testing"
)

// ---------- bt-combination-sum ----------

func combSumLiteral(c []int, target int) [][]int {
	var res [][]int
	var path []int
	var dfs func(start, remain int)
	dfs = func(start, remain int) {
		if remain == 0 {
			res = append(res, slices.Clone(path))
			return
		}
		for i := start; i < len(c); i++ {
			if c[i] <= remain {
				path = append(path, c[i])
				dfs(i, remain-c[i])
				path = path[:len(path)-1]
			}
		}
	}
	dfs(0, target)
	return res
}

// brute: enumerate multiplicity vectors.
func combSumBrute(c []int, target int) [][]int {
	var res [][]int
	cnt := make([]int, len(c))
	var rec func(i, sum int)
	rec = func(i, sum int) {
		if i == len(c) {
			if sum == target {
				var x []int
				for j, k := range cnt {
					for range k {
						x = append(x, c[j])
					}
				}
				res = append(res, x)
			}
			return
		}
		for k := 0; sum+k*c[i] <= target; k++ {
			cnt[i] = k
			rec(i+1, sum+k*c[i])
		}
		cnt[i] = 0
	}
	rec(0, 0)
	return res
}

func TestCombinationSum(t *testing.T) {
	r := newRng(1)
	check := func(c []int, target int) {
		got := combSumLiteral(c, target)
		want := combSumBrute(c, target)
		// result must not contain duplicate multisets
		if g := canonLists(got, true); !slices.Equal(g, canonLists(want, true)) || len(slices.Compact(slices.Clone(g))) != len(g) {
			t.Fatalf("c=%v target=%d got=%v want=%v", c, target, got, want)
		}
	}
	check([]int{2, 3, 6, 7}, 7)
	check([]int{1}, 1)
	check([]int{2}, 1)
	check([]int{5}, 5)
	check([]int{1, 2}, 0) // target 0 -> [[]]
	for range 3000 {
		n := 1 + r.IntN(5)
		perm := r.Perm(12)[:n]
		c := make([]int, n)
		for i, p := range perm {
			c[i] = p + 1 // distinct positive
		}
		check(c, 1+r.IntN(20))
	}
}

// ---------- bt-permutations-dupes ----------

func permUniqueLiteral(a []int) [][]int {
	a = slices.Clone(a)
	slices.Sort(a)
	used := make([]bool, len(a))
	var res [][]int
	var path []int
	var dfs func()
	dfs = func() {
		if len(path) == len(a) {
			res = append(res, slices.Clone(path))
			return
		}
		for i := range a {
			if used[i] || (i > 0 && a[i] == a[i-1] && !used[i-1]) {
				continue
			}
			used[i] = true
			path = append(path, a[i])
			dfs()
			path = path[:len(path)-1]
			used[i] = false
		}
	}
	dfs()
	return res
}

// brute: all n! index permutations, dedupe by value.
func permUniqueBrute(a []int) [][]int {
	seen := map[string]bool{}
	var res [][]int
	idx := make([]int, len(a))
	for i := range idx {
		idx[i] = i
	}
	var heap func(k int)
	heap = func(k int) {
		if k == len(a) {
			p := make([]int, len(a))
			for i, j := range idx {
				p[i] = a[j]
			}
			key := canonLists([][]int{p}, false)[0]
			if !seen[key] {
				seen[key] = true
				res = append(res, p)
			}
			return
		}
		for i := k; i < len(a); i++ {
			idx[k], idx[i] = idx[i], idx[k]
			heap(k + 1)
			idx[k], idx[i] = idx[i], idx[k]
		}
	}
	heap(0)
	return res
}

func TestPermutationsDupes(t *testing.T) {
	r := newRng(2)
	check := func(a []int) {
		got := canonLists(permUniqueLiteral(a), false)
		want := canonLists(permUniqueBrute(a), false)
		if !slices.Equal(got, want) {
			t.Fatalf("a=%v got=%v want=%v", a, got, want)
		}
	}
	check(nil)
	check([]int{1})
	check([]int{1, 1})
	check([]int{1, 1, 2})
	check([]int{3, 3, 3, 3})
	check([]int{-1, 0, -1, 2})
	for range 2000 {
		n := r.IntN(7)
		a := make([]int, n)
		for i := range a {
			a[i] = r.IntN(4) - 1
		}
		check(a)
	}
}

// ---------- bt-n-queens ----------

func nQueensLiteral(n int) int {
	count := 0
	cols, d1, d2 := map[int]bool{}, map[int]bool{}, map[int]bool{}
	var dfs func(r int)
	dfs = func(r int) {
		if r == n {
			count++
			return
		}
		for c := range n {
			if !cols[c] && !d1[r-c] && !d2[r+c] {
				cols[c], d1[r-c], d2[r+c] = true, true, true
				dfs(r + 1)
				delete(cols, c)
				delete(d1, r-c)
				delete(d2, r+c)
			}
		}
	}
	dfs(0)
	return count
}

// brute: all n^n placements of one queen per row, pairwise check.
func nQueensBrute(n int) int {
	pos := make([]int, n)
	count := 0
	var rec func(r int)
	rec = func(r int) {
		if r == n {
			for i := range n {
				for j := i + 1; j < n; j++ {
					if pos[i] == pos[j] || abs(pos[i]-pos[j]) == j-i {
						return
					}
				}
			}
			count++
			return
		}
		for c := range n {
			pos[r] = c
			rec(r + 1)
		}
	}
	rec(0)
	return count
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func TestNQueens(t *testing.T) {
	known := []int{1, 1, 0, 0, 2, 10, 4, 40, 92, 352} // OEIS A000170, n=0..9
	for n := 0; n <= 9; n++ {
		if got := nQueensLiteral(n); got != known[n] {
			t.Fatalf("n=%d got=%d want=%d", n, got, known[n])
		}
		if n <= 7 {
			if b := nQueensBrute(n); b != known[n] {
				t.Fatalf("brute n=%d got=%d", n, b)
			}
		}
	}
}

// ---------- bt-word-search ----------

// Patched key_idea: mark the cell by overwriting it with a sentinel,
// so the grid[r][c]==w[i] check also rejects visited cells.
func wordSearchLiteral(g [][]byte, w string) bool {
	m, n := len(g), len(g[0])
	var dfs func(r, c, i int) bool
	dfs = func(r, c, i int) bool {
		if r < 0 || r >= m || c < 0 || c >= n || g[r][c] != w[i] {
			return false
		}
		if i == len(w)-1 {
			return true
		}
		ch := g[r][c]
		g[r][c] = '#'
		ok := dfs(r+1, c, i+1) || dfs(r-1, c, i+1) || dfs(r, c+1, i+1) || dfs(r, c-1, i+1)
		g[r][c] = ch
		return ok
	}
	for r := range m {
		for c := range n {
			if dfs(r, c, 0) {
				return true
			}
		}
	}
	return false
}

func TestWordSearch(t *testing.T) {
	r := newRng(3)
	check := func(g [][]byte, w string) {
		orig := cloneGrid(g)
		got := wordSearchLiteral(g, w)
		want := bruteGridHas(orig, w)
		if got != want {
			t.Fatalf("g=%q w=%q got=%v want=%v", orig, w, got, want)
		}
		for i := range g {
			if string(g[i]) != string(orig[i]) {
				t.Fatalf("grid not restored")
			}
		}
	}
	abce := [][]byte{[]byte("ABCE"), []byte("SFCS"), []byte("ADEE")}
	check(abce, "ABCCED")
	check(abce, "SEE")
	check(abce, "ABCB") // would need reuse
	check([][]byte{[]byte("a")}, "a")
	check([][]byte{[]byte("a")}, "aa")
	check([][]byte{[]byte("aa")}, "aaa")
	for range 5000 {
		m, n := 1+r.IntN(3), 1+r.IntN(4)
		g := randGrid(r, m, n, "ab")
		check(g, randStr(r, 1+r.IntN(7), "ab"))
	}
}

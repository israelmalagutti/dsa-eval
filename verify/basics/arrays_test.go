package basics

import (
	"math"
	"math/rand/v2"
	"slices"
	"sort"
	"strings"
	"testing"
)

// ---------- tp-sorted-two-sum ----------

func twoSumLiteral(a []int, t int) (int, int, bool) {
	l, r := 0, len(a)-1
	for l < r {
		s := a[l] + a[r]
		if s == t {
			return l + 1, r + 1, true
		}
		if s < t {
			l++
		} else {
			r--
		}
	}
	return 0, 0, false
}

// twoSumPairs returns every valid 1-based (i, j) pair, i < j.
func twoSumPairs(a []int, t int) [][2]int {
	var out [][2]int
	for i := range a {
		for j := i + 1; j < len(a); j++ {
			if a[i]+a[j] == t {
				out = append(out, [2]int{i + 1, j + 1})
			}
		}
	}
	return out
}

func TestTPSortedTwoSum(t *testing.T) {
	r := newRNG()
	// Prompt guarantees exactly one solution; only such inputs are tested.
	check := func(a []int, tg int) {
		want := twoSumPairs(a, tg)
		if len(want) != 1 {
			t.Fatalf("bad test input a=%v t=%d", a, tg)
		}
		i, j, ok := twoSumLiteral(a, tg)
		if !ok || [2]int{i, j} != want[0] {
			t.Fatalf("a=%v t=%d got [%d,%d] want %v", a, tg, i, j, want[0])
		}
	}
	check([]int{2, 7, 11, 15}, 9)
	check([]int{3, 3}, 6)
	check([]int{-3, -1, 0, 4}, -4)
	check([]int{-1, 0}, -1)
	for n := 0; n < iters; {
		a := sortedCopy(randInts(r, 2+r.IntN(9), -6, 6))
		tg := r.IntN(25) - 12
		if len(twoSumPairs(a, tg)) == 1 {
			check(a, tg)
			n++
		}
	}
}

// ---------- tp-container ----------

func containerLiteral(h []int) int {
	l, r, best := 0, len(h)-1, 0
	for l < r {
		best = max(best, min(h[l], h[r])*(r-l))
		if h[l] < h[r] {
			l++
		} else {
			r--
		}
	}
	return best
}

func containerBrute(h []int) int {
	best := 0
	for i := range h {
		for j := i + 1; j < len(h); j++ {
			best = max(best, min(h[i], h[j])*(j-i))
		}
	}
	return best
}

func TestTPContainer(t *testing.T) {
	r := newRNG()
	check := func(h []int) {
		if got, want := containerLiteral(h), containerBrute(h); got != want {
			t.Fatalf("h=%v got %d want %d", h, got, want)
		}
	}
	check(nil)
	check([]int{5})
	check([]int{1, 8, 6, 2, 5, 4, 8, 3, 7})
	check([]int{4, 4, 4, 4})
	for range iters {
		check(randInts(r, r.IntN(12), 0, 9))
	}
}

// ---------- sw-longest-unique ----------

func longestUniqueLiteral(s string) int {
	var last [256]int
	for i := range last {
		last[i] = -1
	}
	l, best := 0, 0
	for r := 0; r < len(s); r++ {
		c := s[r]
		if last[c] >= l {
			l = last[c] + 1
		}
		last[c] = r
		best = max(best, r-l+1)
	}
	return best
}

// Alternative accepted by technique: set of window chars, shrink left on repeat.
func longestUniqueSet(s string) int {
	in := map[byte]bool{}
	l, best := 0, 0
	for r := 0; r < len(s); r++ {
		for in[s[r]] {
			delete(in, s[l])
			l++
		}
		in[s[r]] = true
		best = max(best, r-l+1)
	}
	return best
}

func longestUniqueBrute(s string) int {
	best := 0
	for i := range len(s) {
		seen := map[byte]bool{}
		for j := i; j < len(s) && !seen[s[j]]; j++ {
			seen[s[j]] = true
			best = max(best, j-i+1)
		}
	}
	return best
}

func randStr(r *rand.Rand, n int, alpha string) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = alpha[r.IntN(len(alpha))]
	}
	return string(b)
}

func TestSWLongestUnique(t *testing.T) {
	r := newRNG()
	check := func(s string) {
		want := longestUniqueBrute(s)
		if got := longestUniqueLiteral(s); got != want {
			t.Fatalf("s=%q got %d want %d", s, got, want)
		}
		if got := longestUniqueSet(s); got != want {
			t.Fatalf("set s=%q got %d want %d", s, got, want)
		}
	}
	for _, s := range []string{"", "a", "aaaa", "abcabcbb", "pwwkew", "abba", "dvdf"} {
		check(s)
	}
	for range iters {
		check(randStr(r, r.IntN(14), "abcd"))
	}
}

// ---------- sw-min-window ----------

// minWindowLiteral follows the patched key_idea word for word.
func minWindowLiteral(s, t string) string {
	need := map[byte]int{}
	for i := range len(t) {
		need[t[i]]++
	}
	missing := len(t)
	l, bestL, bestLen := 0, 0, math.MaxInt
	for r := 0; r < len(s); r++ {
		if need[s[r]] > 0 {
			missing--
		}
		need[s[r]]--
		for missing == 0 {
			if r-l+1 < bestLen {
				bestL, bestLen = l, r-l+1
			}
			need[s[l]]++
			if need[s[l]] > 0 {
				missing++
			}
			l++
		}
	}
	if bestLen == math.MaxInt {
		return ""
	}
	return s[bestL : bestL+bestLen]
}

func covers(w, t string) bool {
	var c [256]int
	for i := range len(w) {
		c[w[i]]++
	}
	for i := range len(t) {
		c[t[i]]--
		if c[t[i]] < 0 {
			return false
		}
	}
	return true
}

// Brute: shortest covering window, leftmost on ties, "" if none (t non-empty).
func minWindowBrute(s, t string) string {
	for L := 1; L <= len(s); L++ {
		for i := 0; i+L <= len(s); i++ {
			if covers(s[i:i+L], t) {
				return s[i : i+L]
			}
		}
	}
	return ""
}

func TestSWMinWindow(t *testing.T) {
	r := newRNG()
	check := func(s, tt string) {
		if got, want := minWindowLiteral(s, tt), minWindowBrute(s, tt); got != want {
			t.Fatalf("s=%q t=%q got %q want %q", s, tt, got, want)
		}
	}
	check("ADOBECODEBANC", "ABC")
	check("a", "a")
	check("a", "aa")
	check("ab", "b")
	check("aa", "aa")
	check("", "a")
	check("abba", "a") // tie: leftmost
	check("bab", "ab") // tie: leftmost "ba"
	for range iters {
		check(randStr(r, r.IntN(12), "abc"), randStr(r, 1+r.IntN(4), "abcd"))
	}
}

// ---------- ps-subarray-sum-k ----------

func subarraySumLiteral(a []int, k int) int {
	cnt := map[int]int{0: 1}
	p, ans := 0, 0
	for _, x := range a {
		p += x
		ans += cnt[p-k]
		cnt[p]++
	}
	return ans
}

func subarraySumBrute(a []int, k int) int {
	ans := 0
	for i := range a {
		s := 0
		for j := i; j < len(a); j++ {
			s += a[j]
			if s == k {
				ans++
			}
		}
	}
	return ans
}

func TestPSSubarraySumK(t *testing.T) {
	r := newRNG()
	check := func(a []int, k int) {
		if got, want := subarraySumLiteral(a, k), subarraySumBrute(a, k); got != want {
			t.Fatalf("a=%v k=%d got %d want %d", a, k, got, want)
		}
	}
	check(nil, 0)
	check([]int{0, 0, 0}, 0)
	check([]int{1, 1, 1}, 2)
	check([]int{1, -1, 0}, 0)
	for range iters {
		check(randInts(r, r.IntN(12), -3, 3), r.IntN(7)-3)
	}
}

// ---------- ps-2d-range ----------

func build2D(a [][]int) [][]int {
	m := len(a)
	n := 0
	if m > 0 {
		n = len(a[0])
	}
	P := make([][]int, m+1)
	for i := range P {
		P[i] = make([]int, n+1)
	}
	for i := range m {
		for j := range n {
			P[i+1][j+1] = a[i][j] + P[i][j+1] + P[i+1][j] - P[i][j]
		}
	}
	return P
}

func query2D(P [][]int, r1, c1, r2, c2 int) int {
	return P[r2+1][c2+1] - P[r1][c2+1] - P[r2+1][c1] + P[r1][c1]
}

func TestPS2DRange(t *testing.T) {
	r := newRNG()
	for range iters / 5 {
		m, n := 1+r.IntN(6), 1+r.IntN(6)
		a := make([][]int, m)
		for i := range a {
			a[i] = randInts(r, n, -9, 9)
		}
		P := build2D(a)
		for r1 := range m {
			for r2 := r1; r2 < m; r2++ {
				for c1 := range n {
					for c2 := c1; c2 < n; c2++ {
						want := 0
						for i := r1; i <= r2; i++ {
							for j := c1; j <= c2; j++ {
								want += a[i][j]
							}
						}
						if got := query2D(P, r1, c1, r2, c2); got != want {
							t.Fatalf("a=%v q=(%d,%d)-(%d,%d) got %d want %d", a, r1, c1, r2, c2, got, want)
						}
					}
				}
			}
		}
	}
}

// ---------- diff-range-updates ----------

type upd struct{ l, r, v int }

// d must have length n+1 (key_idea writes d[r+1] with r up to n-1).
func diffLiteral(n int, us []upd) []int {
	d := make([]int, n+1)
	for _, u := range us {
		d[u.l] += u.v
		d[u.r+1] -= u.v
	}
	out := make([]int, n)
	s := 0
	for i := range n {
		s += d[i]
		out[i] = s
	}
	return out
}

func diffBrute(n int, us []upd) []int {
	out := make([]int, n)
	for _, u := range us {
		for i := u.l; i <= u.r; i++ {
			out[i] += u.v
		}
	}
	return out
}

func TestDiffRangeUpdates(t *testing.T) {
	r := newRNG()
	check := func(n int, us []upd) {
		if got, want := diffLiteral(n, us), diffBrute(n, us); !slices.Equal(got, want) {
			t.Fatalf("n=%d us=%v got %v want %v", n, us, got, want)
		}
	}
	check(0, nil)
	check(1, []upd{{0, 0, 5}})
	check(3, []upd{{0, 2, 1}, {2, 2, -4}})
	for range iters {
		n := 1 + r.IntN(10)
		var us []upd
		for range r.IntN(6) {
			a, b := r.IntN(n), r.IntN(n)
			us = append(us, upd{min(a, b), max(a, b), r.IntN(11) - 5})
		}
		check(n, us)
	}
}

// ---------- hash-group-anagrams ----------

func groupAnagramsLiteral(words []string) [][]string {
	groups := map[string][]string{}
	var order []string
	for _, w := range words {
		b := []byte(w)
		slices.Sort(b)
		k := string(b)
		if _, ok := groups[k]; !ok {
			order = append(order, k)
		}
		groups[k] = append(groups[k], w)
	}
	var out [][]string
	for _, k := range order {
		out = append(out, groups[k])
	}
	return out
}

func groupAnagramsCount(words []string) [][]string {
	groups := map[[26]int][]string{}
	var order [][26]int
	for _, w := range words {
		var k [26]int
		for i := range len(w) {
			k[w[i]-'a']++
		}
		if _, ok := groups[k]; !ok {
			order = append(order, k)
		}
		groups[k] = append(groups[k], w)
	}
	var out [][]string
	for _, k := range order {
		out = append(out, groups[k])
	}
	return out
}

func isAnagram(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	used := make([]bool, len(b))
outer:
	for i := range len(a) {
		for j := range len(b) {
			if !used[j] && b[j] == a[i] {
				used[j] = true
				continue outer
			}
		}
		return false
	}
	return true
}

// Brute: union-find style pairwise grouping by first-seen representative.
func groupAnagramsBrute(words []string) [][]string {
	var out [][]string
	for _, w := range words {
		placed := false
		for i := range out {
			if isAnagram(out[i][0], w) {
				out[i] = append(out[i], w)
				placed = true
				break
			}
		}
		if !placed {
			out = append(out, []string{w})
		}
	}
	return out
}

func canonGroups(g [][]string) []string {
	var keys []string
	for _, grp := range g {
		c := slices.Clone(grp)
		sort.Strings(c)
		keys = append(keys, strings.Join(c, ","))
	}
	sort.Strings(keys)
	return keys
}

func TestHashGroupAnagrams(t *testing.T) {
	r := newRNG()
	check := func(ws []string) {
		want := canonGroups(groupAnagramsBrute(ws))
		for name, f := range map[string]func([]string) [][]string{"sorted": groupAnagramsLiteral, "count": groupAnagramsCount} {
			if got := canonGroups(f(ws)); !slices.Equal(got, want) {
				t.Fatalf("%s ws=%v got %v want %v", name, ws, got, want)
			}
		}
	}
	check(nil)
	check([]string{""})
	check([]string{"", ""})
	check([]string{"eat", "tea", "tan", "ate", "nat", "bat"})
	for range iters {
		var ws []string
		for range r.IntN(8) {
			ws = append(ws, randStr(r, r.IntN(4), "abc"))
		}
		check(ws)
	}
}

// ---------- hash-longest-consecutive ----------

func longestConsecLiteral(nums []int) int {
	s := map[int]bool{}
	for _, x := range nums {
		s[x] = true
	}
	best := 0
	for x := range s {
		if !s[x-1] {
			y := x
			for s[y] {
				y++
			}
			best = max(best, y-x)
		}
	}
	return best
}

func longestConsecBrute(nums []int) int {
	best := 0
	for _, x := range nums {
		l := 0
		for slices.Contains(nums, x+l) {
			l++
		}
		best = max(best, l)
	}
	return best
}

func TestHashLongestConsecutive(t *testing.T) {
	r := newRNG()
	check := func(a []int) {
		if got, want := longestConsecLiteral(a), longestConsecBrute(a); got != want {
			t.Fatalf("a=%v got %d want %d", a, got, want)
		}
	}
	check(nil)
	check([]int{7})
	check([]int{100, 4, 200, 1, 3, 2})
	check([]int{1, 1, 1})
	check([]int{-2, -1, 0, 1})
	for range iters {
		check(randInts(r, r.IntN(12), -8, 8))
	}
}

// ---------- kadane-max-subarray ----------

func kadaneLiteral(a []int) int {
	cur, best := a[0], a[0]
	for _, x := range a[1:] {
		cur = max(x, cur+x)
		best = max(best, cur)
	}
	return best
}

func kadaneBrute(a []int) int {
	best := math.MinInt
	for i := range a {
		s := 0
		for j := i; j < len(a); j++ {
			s += a[j]
			best = max(best, s)
		}
	}
	return best
}

func TestKadane(t *testing.T) {
	r := newRNG()
	check := func(a []int) {
		if got, want := kadaneLiteral(a), kadaneBrute(a); got != want {
			t.Fatalf("a=%v got %d want %d", a, got, want)
		}
	}
	check([]int{-5})
	check([]int{-3, -1, -2})
	check([]int{-2, 1, -3, 4, -1, 2, 1, -5, 4})
	check([]int{0, 0})
	for range iters {
		check(randInts(r, 1+r.IntN(12), -6, 6))
	}
}

// ---------- dnf-sort-colors ----------

func dnfLiteral(a []int) {
	lo, mid, hi := 0, 0, len(a)-1
	for mid <= hi {
		switch a[mid] {
		case 0:
			a[lo], a[mid] = a[mid], a[lo]
			lo++
			mid++
		case 1:
			mid++
		case 2:
			a[mid], a[hi] = a[hi], a[mid]
			hi--
		}
	}
}

func TestDNFSortColors(t *testing.T) {
	r := newRNG()
	check := func(a []int) {
		b := slices.Clone(a)
		dnfLiteral(b)
		if want := sortedCopy(a); !slices.Equal(b, want) {
			t.Fatalf("a=%v got %v want %v", a, b, want)
		}
	}
	check(nil)
	check([]int{2})
	check([]int{2, 2, 2})
	check([]int{2, 0, 2, 1, 1, 0})
	for range iters {
		check(randInts(r, r.IntN(12), 0, 2))
	}
}

// ---------- cyclic-missing ----------

func missingMarkLiteral(in []int) []int {
	a := slices.Clone(in)
	abs := func(x int) int { return max(x, -x) }
	for _, x := range a {
		i := abs(x) - 1
		a[i] = -abs(a[i])
	}
	var out []int
	for i, x := range a {
		if x > 0 {
			out = append(out, i+1)
		}
	}
	return out
}

// Cyclic-sort alternative from the patched key_idea.
func missingCyclicLiteral(in []int) []int {
	a := slices.Clone(in)
	for i := range a {
		for a[i] != i+1 && a[a[i]-1] != a[i] {
			j := a[i] - 1
			a[i], a[j] = a[j], a[i]
		}
	}
	var out []int
	for i, x := range a {
		if x != i+1 {
			out = append(out, i+1)
		}
	}
	return out
}

func missingBrute(a []int) []int {
	var out []int
	for v := 1; v <= len(a); v++ {
		if !slices.Contains(a, v) {
			out = append(out, v)
		}
	}
	return out
}

func TestCyclicMissingMark(t *testing.T) {
	r := newRNG()
	check := func(a []int) {
		if got, want := missingMarkLiteral(a), missingBrute(a); !slices.Equal(got, want) {
			t.Fatalf("a=%v got %v want %v", a, got, want)
		}
		if got, want := missingCyclicLiteral(a), missingBrute(a); !slices.Equal(got, want) {
			t.Fatalf("cyclic a=%v got %v want %v", a, got, want)
		}
	}
	check(nil)
	check([]int{1})
	check([]int{4, 3, 2, 7, 8, 2, 3, 1})
	check([]int{1, 1})
	check([]int{3, 3, 3})
	for range iters {
		n := r.IntN(12)
		check(randInts(r, n, 1, max(n, 1))[:n])
	}
}

// ---------- cyclic-first-missing-positive ----------

func firstMissingPosLiteral(in []int) int {
	a := slices.Clone(in)
	n := len(a)
	for i := range n {
		for 1 <= a[i] && a[i] <= n && a[a[i]-1] != a[i] {
			j := a[i] - 1
			a[i], a[j] = a[j], a[i]
		}
	}
	for i := range n {
		if a[i] != i+1 {
			return i + 1
		}
	}
	return n + 1
}

// Alternative accepted by technique: in-place index marking.
func firstMissingPosMark(in []int) int {
	a := slices.Clone(in)
	n := len(a)
	for i := range a {
		if a[i] <= 0 || a[i] > n {
			a[i] = n + 1
		}
	}
	for i := range a {
		v := max(a[i], -a[i])
		if v <= n {
			a[v-1] = -max(a[v-1], -a[v-1])
		}
	}
	for i := range a {
		if a[i] > 0 {
			return i + 1
		}
	}
	return n + 1
}

func firstMissingPosBrute(a []int) int {
	for v := 1; ; v++ {
		if !slices.Contains(a, v) {
			return v
		}
	}
}

func TestCyclicFirstMissingPositive(t *testing.T) {
	r := newRNG()
	check := func(a []int) {
		want := firstMissingPosBrute(a)
		if got := firstMissingPosLiteral(a); got != want {
			t.Fatalf("a=%v got %d want %d", a, got, want)
		}
		if got := firstMissingPosMark(a); got != want {
			t.Fatalf("mark a=%v got %d want %d", a, got, want)
		}
	}
	check(nil)
	check([]int{1})
	check([]int{2})
	check([]int{-1, 0})
	check([]int{3, 4, -1, 1})
	check([]int{7, 8, 9, 11, 12})
	check([]int{1, 1, 2, 2})
	for range iters {
		check(randInts(r, r.IntN(12), -3, 12))
	}
}

// ---------- iv-merge ----------

type iv struct{ s, e int }

func mergeLiteral(in []iv) []iv {
	a := slices.Clone(in)
	slices.SortFunc(a, func(x, y iv) int { return x.s - y.s })
	var out []iv
	for _, cur := range a {
		if len(out) > 0 && cur.s <= out[len(out)-1].e {
			out[len(out)-1].e = max(out[len(out)-1].e, cur.e)
		} else {
			out = append(out, cur)
		}
	}
	return out
}

// Brute: repeatedly merge any two intersecting closed intervals until fixpoint.
func mergeBrute(in []iv) []iv {
	a := slices.Clone(in)
	for changed := true; changed; {
		changed = false
	loop:
		for i := range a {
			for j := i + 1; j < len(a); j++ {
				if a[i].s <= a[j].e && a[j].s <= a[i].e {
					a[i] = iv{min(a[i].s, a[j].s), max(a[i].e, a[j].e)}
					a = slices.Delete(a, j, j+1)
					changed = true
					break loop
				}
			}
		}
	}
	slices.SortFunc(a, func(x, y iv) int { return x.s - y.s })
	return a
}

func randIntervals(r *rand.Rand, n, maxT int, minLen int) []iv {
	var out []iv
	for range n {
		s := r.IntN(maxT)
		out = append(out, iv{s, s + minLen + r.IntN(5)})
	}
	return out
}

func TestIVMerge(t *testing.T) {
	r := newRNG()
	check := func(a []iv) {
		if got, want := mergeLiteral(a), mergeBrute(a); !slices.Equal(got, want) {
			t.Fatalf("a=%v got %v want %v", a, got, want)
		}
	}
	check(nil)
	check([]iv{{1, 3}})
	check([]iv{{1, 3}, {2, 6}, {8, 10}, {15, 18}})
	check([]iv{{1, 4}, {4, 5}})
	check([]iv{{1, 10}, {2, 3}})
	check([]iv{{2, 2}, {2, 2}})
	for range iters {
		check(randIntervals(r, r.IntN(8), 15, 0))
	}
}

// ---------- iv-meeting-rooms ----------

func roomsSweepLiteral(a []iv) int {
	type ev struct{ t, d int }
	var evs []ev
	for _, x := range a {
		evs = append(evs, ev{x.s, +1}, ev{x.e, -1})
	}
	slices.SortFunc(evs, func(x, y ev) int {
		if x.t != y.t {
			return x.t - y.t
		}
		return x.d - y.d // -1 (end) before +1 (start)
	})
	cur, best := 0, 0
	for _, e := range evs {
		cur += e.d
		best = max(best, cur)
	}
	return best
}

func roomsHeapLiteral(in []iv) int {
	a := slices.Clone(in)
	slices.SortFunc(a, func(x, y iv) int { return x.s - y.s })
	var ends []int // min-heap via sorted slice (small n)
	for _, x := range a {
		if len(ends) > 0 && ends[0] <= x.s {
			ends = ends[1:]
		}
		i, _ := slices.BinarySearch(ends, x.e)
		ends = slices.Insert(ends, i, x.e)
	}
	return len(ends)
}

// Brute: half-open [s,e); max over integer times of active meetings.
func roomsBrute(a []iv) int {
	best := 0
	for t := 0; t <= 30; t++ {
		c := 0
		for _, x := range a {
			if x.s <= t && t < x.e {
				c++
			}
		}
		best = max(best, c)
	}
	return best
}

func TestIVMeetingRooms(t *testing.T) {
	r := newRNG()
	check := func(a []iv) {
		want := roomsBrute(a)
		if got := roomsSweepLiteral(a); got != want {
			t.Fatalf("sweep a=%v got %d want %d", a, got, want)
		}
		if got := roomsHeapLiteral(a); got != want {
			t.Fatalf("heap a=%v got %d want %d", a, got, want)
		}
	}
	check(nil)
	check([]iv{{0, 30}, {5, 10}, {15, 20}})
	check([]iv{{7, 10}, {2, 4}})
	check([]iv{{1, 5}, {5, 10}})
	check([]iv{{1, 5}, {1, 5}, {1, 5}})
	// Prompt: [start, end) with start < end.
	for range iters {
		check(randIntervals(r, r.IntN(8), 15, 1))
	}
}

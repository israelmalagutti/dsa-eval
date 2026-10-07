package basics

import (
	"math/rand/v2"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// ---------- sort-count-inversions ----------

func countInvLiteral(a []int) ([]int, int64) {
	if len(a) <= 1 {
		return slices.Clone(a), 0
	}
	m := len(a) / 2
	left, cl := countInvLiteral(a[:m])
	right, cr := countInvLiteral(a[m:])
	cnt := cl + cr
	out := make([]int, 0, len(a))
	i, j := 0, 0
	for i < len(left) && j < len(right) {
		if left[i] <= right[j] {
			out = append(out, left[i])
			i++
		} else {
			out = append(out, right[j])
			j++
			cnt += int64(len(left) - i)
		}
	}
	out = append(out, left[i:]...)
	out = append(out, right[j:]...)
	return out, cnt
}

func countInvBrute(a []int) int64 {
	var c int64
	for i := range a {
		for j := i + 1; j < len(a); j++ {
			if a[i] > a[j] {
				c++
			}
		}
	}
	return c
}

func TestSortCountInversions(t *testing.T) {
	r := newRNG()
	check := func(a []int) {
		s, got := countInvLiteral(a)
		if want := countInvBrute(a); got != want || !slices.Equal(s, sortedCopy(a)) {
			t.Fatalf("a=%v got %d want %d", a, got, want)
		}
	}
	check(nil)
	check([]int{1})
	check([]int{3, 3, 3})
	check([]int{5, 4, 3, 2, 1})
	check([]int{2, 4, 1, 3, 5})
	for range iters {
		check(randInts(r, r.IntN(14), -4, 4))
	}
}

// ---------- sort-kth-largest ----------

func kthLargestLiteral(in []int, k int, r *rand.Rand) int {
	a := slices.Clone(in)
	target, lo, hi := len(a)-k, 0, len(a)-1
	for {
		v := a[lo+r.IntN(hi-lo+1)]
		lt, i, gt := lo, lo, hi
		for i <= gt {
			if a[i] < v {
				a[lt], a[i] = a[i], a[lt]
				lt++
				i++
			} else if a[i] > v {
				a[i], a[gt] = a[gt], a[i]
				gt--
			} else {
				i++
			}
		}
		if lt <= target && target <= gt {
			return v
		}
		if target < lt {
			hi = lt - 1
		} else {
			lo = gt + 1
		}
	}
}

func TestSortKthLargest(t *testing.T) {
	r := newRNG()
	check := func(a []int, k int) {
		s := sortedCopy(a)
		want := s[len(s)-k]
		if got := kthLargestLiteral(a, k, r); got != want {
			t.Fatalf("a=%v k=%d got %d want %d", a, k, got, want)
		}
	}
	check([]int{7}, 1)
	check([]int{3, 2, 1, 5, 6, 4}, 2)
	check([]int{3, 2, 3, 1, 2, 4, 5, 5, 6}, 4)
	check([]int{2, 2, 2, 2}, 3)
	for range iters {
		n := 1 + r.IntN(12)
		check(randInts(r, n, -5, 5), 1+r.IntN(n))
	}
}

// ---------- sort-top-k-frequent ----------

func topKLiteral(a []int, k int) []int {
	n := len(a)
	cnt := map[int]int{}
	for _, x := range a {
		cnt[x]++
	}
	buckets := make([][]int, n+1)
	for x, f := range cnt {
		buckets[f] = append(buckets[f], x)
	}
	out := []int{}
	for f := n; f >= 1; f-- {
		for _, x := range buckets[f] {
			out = append(out, x)
			if len(out) == k {
				return out
			}
		}
	}
	return out
}

// topKBrute: distinct values sorted by frequency desc; ok=false if the answer is not
// unique (tie straddling position k), which the prompt rules out.
func topKBrute(a []int, k int) (top []int, ok bool) {
	cnt := map[int]int{}
	for _, x := range a {
		cnt[x]++
	}
	var d []int
	for x := range cnt {
		d = append(d, x)
	}
	sort.Slice(d, func(i, j int) bool { return cnt[d[i]] > cnt[d[j]] })
	if k < len(d) && cnt[d[k-1]] == cnt[d[k]] {
		return nil, false
	}
	return sortedCopy(d[:k]), true
}

func TestSortTopKFrequent(t *testing.T) {
	r := newRNG()
	check := func(a []int, k int) {
		want, ok := topKBrute(a, k)
		if !ok {
			t.Fatalf("bad test input a=%v k=%d", a, k)
		}
		if got := sortedCopy(topKLiteral(a, k)); !slices.Equal(got, want) {
			t.Fatalf("a=%v k=%d got %v want %v", a, k, got, want)
		}
	}
	check([]int{1}, 1)
	check([]int{1, 1, 1, 2, 2, 3}, 2)
	check([]int{4, 4, 4}, 1)
	check([]int{1, 2, 3}, 3)
	for n := 0; n < iters; {
		a := randInts(r, 1+r.IntN(14), -4, 4)
		distinct := len(slices.Compact(sortedCopy(a)))
		k := 1 + r.IntN(distinct)
		if _, ok := topKBrute(a, k); ok {
			check(a, k)
			n++
		}
	}
}

// ---------- sort-largest-number ----------

func largestNumberLiteral(nums []int) string {
	s := make([]string, len(nums))
	for i, x := range nums {
		s[i] = strconv.Itoa(x)
	}
	slices.SortFunc(s, func(a, b string) int {
		ab, ba := a+b, b+a
		switch {
		case ab > ba:
			return -1
		case ab < ba:
			return 1
		}
		return 0
	})
	res := strings.Join(s, "")
	if strings.HasPrefix(res, "0") {
		return "0"
	}
	return res
}

func permute(a []string, k int, f func()) {
	if k == len(a) {
		f()
		return
	}
	for i := k; i < len(a); i++ {
		a[k], a[i] = a[i], a[k]
		permute(a, k+1, f)
		a[k], a[i] = a[i], a[k]
	}
}

// Brute: try all permutations; equal-length digit strings compare lexicographically.
func largestNumberBrute(nums []int) string {
	s := make([]string, len(nums))
	for i, x := range nums {
		s[i] = strconv.Itoa(x)
	}
	best := ""
	permute(s, 0, func() {
		if c := strings.Join(s, ""); c > best {
			best = c
		}
	})
	if t := strings.TrimLeft(best, "0"); t == "" {
		return "0"
	}
	return best
}

func TestSortLargestNumber(t *testing.T) {
	r := newRNG()
	check := func(a []int) {
		if got, want := largestNumberLiteral(a), largestNumberBrute(a); got != want {
			t.Fatalf("a=%v got %q want %q", a, got, want)
		}
	}
	check([]int{0})
	check([]int{0, 0})
	check([]int{10, 2})
	check([]int{3, 30, 34, 5, 9})
	check([]int{121, 12})
	check([]int{432, 43243})
	pool := []int{0, 1, 2, 3, 9, 10, 12, 21, 30, 34, 90, 99, 100, 121, 212, 1000}
	for range iters {
		n := 1 + r.IntN(6)
		a := make([]int, n)
		for i := range a {
			if r.IntN(2) == 0 {
				a[i] = pool[r.IntN(len(pool))]
			} else {
				a[i] = r.IntN(400)
			}
		}
		check(a)
	}
}

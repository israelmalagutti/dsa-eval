package basics

import (
	"math"
	"slices"
	"testing"
)

// ---------- bs-first-occurrence ----------

func firstOccLiteral(a []int, t int) int {
	lo, hi := 0, len(a)
	for lo < hi {
		mid := (lo + hi) / 2
		if a[mid] < t {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	if lo < len(a) && a[lo] == t {
		return lo
	}
	return -1
}

func firstOccBrute(a []int, t int) int {
	for i, x := range a {
		if x == t {
			return i
		}
	}
	return -1
}

func TestBSFirstOccurrence(t *testing.T) {
	r := newRNG()
	check := func(a []int, tg int) {
		if got, want := firstOccLiteral(a, tg), firstOccBrute(a, tg); got != want {
			t.Fatalf("a=%v t=%d got %d want %d", a, tg, got, want)
		}
	}
	for _, c := range []struct {
		a []int
		t int
	}{{nil, 1}, {[]int{5}, 5}, {[]int{5}, 4}, {[]int{5}, 6}, {[]int{2, 2, 2, 2}, 2}, {[]int{-3, -3, 0, 0, 7}, 0}} {
		check(c.a, c.t)
	}
	for range iters {
		a := sortedCopy(randInts(r, r.IntN(12), -5, 5))
		check(a, r.IntN(13)-6)
	}
}

// ---------- bs-ship-capacity ----------

func shipFeasible(w []int, c, D int) bool {
	days, load := 1, 0
	for _, x := range w {
		if load+x > c {
			days++
			load = 0
		}
		load += x
	}
	return days <= D
}

func shipLiteral(w []int, D int) int {
	lo, hi := slices.Max(w), 0
	for _, x := range w {
		hi += x
	}
	for lo < hi {
		mid := (lo + hi) / 2
		if shipFeasible(w, mid, D) {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	return lo
}

// Brute force: enumerate every way to cut w into at most D contiguous runs,
// minimize the max run sum.
func shipBrute(w []int, D int) int {
	n := len(w)
	best := math.MaxInt
	// cut mask: bit i set => cut between i and i+1
	for mask := 0; mask < 1<<(n-1); mask++ {
		runs, cur, mx := 1, 0, 0
		for i, x := range w {
			cur += x
			if i == n-1 || mask>>i&1 == 1 {
				mx = max(mx, cur)
				cur = 0
				if i != n-1 {
					runs++
				}
			}
		}
		if runs <= D {
			best = min(best, mx)
		}
	}
	return best
}

func TestBSShipCapacity(t *testing.T) {
	r := newRNG()
	check := func(w []int, D int) {
		if got, want := shipLiteral(w, D), shipBrute(w, D); got != want {
			t.Fatalf("w=%v D=%d got %d want %d", w, D, got, want)
		}
	}
	check([]int{7}, 1)
	check([]int{7}, 3)
	check([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 5) // LC example -> 15
	check([]int{3, 3, 3, 3}, 4)
	check([]int{3, 3, 3, 3}, 1)
	for range iters {
		n := 1 + r.IntN(9)
		check(randInts(r, n, 1, 10), 1+r.IntN(n+1))
	}
}

// ---------- bs-rotated ----------

func rotatedLiteral(a []int, t int) int {
	lo, hi := 0, len(a)-1
	for lo <= hi {
		mid := (lo + hi) / 2
		if a[mid] == t {
			return mid
		}
		if a[lo] <= a[mid] {
			if a[lo] <= t && t < a[mid] {
				hi = mid - 1
			} else {
				lo = mid + 1
			}
		} else {
			if a[mid] < t && t <= a[hi] {
				lo = mid + 1
			} else {
				hi = mid - 1
			}
		}
	}
	return -1
}

func TestBSRotated(t *testing.T) {
	r := newRNG()
	check := func(a []int, tg int) {
		if got, want := rotatedLiteral(a, tg), firstOccBrute(a, tg); got != want {
			t.Fatalf("a=%v t=%d got %d want %d", a, tg, got, want)
		}
	}
	check(nil, 3)
	check([]int{4, 5, 6, 7, 0, 1, 2}, 0)
	check([]int{4, 5, 6, 7, 0, 1, 2}, 3)
	check([]int{1}, 1)
	check([]int{1}, 0)
	check([]int{3, 1}, 1)
	for range iters {
		n := 1 + r.IntN(10)
		a := sortedCopy(randDistinct(r, n, -10, 10))
		k := r.IntN(n)
		a = append(a[k:], a[:k]...)
		check(a, r.IntN(23)-11)
	}
}

// ---------- bs-peak ----------

func peakLiteral(a []int) int {
	lo, hi := 0, len(a)-1
	for lo < hi {
		mid := (lo + hi) / 2
		if a[mid] < a[mid+1] {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo
}

func isPeak(a []int, i int) bool {
	if i < 0 || i >= len(a) {
		return false
	}
	return (i == 0 || a[i] > a[i-1]) && (i == len(a)-1 || a[i] > a[i+1])
}

func TestBSPeak(t *testing.T) {
	r := newRNG()
	check := func(a []int) {
		if i := peakLiteral(a); !isPeak(a, i) {
			t.Fatalf("a=%v returned %d, not a peak", a, i)
		}
	}
	check([]int{1})
	check([]int{1, 2})
	check([]int{2, 1})
	check([]int{1, 2, 3, 4})
	check([]int{4, 3, 2, 1})
	check([]int{1, 3, 1, 3, 1})
	for range iters {
		n := 1 + r.IntN(12)
		a := []int{r.IntN(9) - 4}
		for len(a) < n {
			x := r.IntN(9) - 4
			if x != a[len(a)-1] { // adjacent elements differ
				a = append(a, x)
			}
		}
		check(a)
	}
}

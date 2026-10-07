package advanced

import (
	"math"
	"slices"
	"testing"
)

// ---------- gd-non-overlapping ----------
// Patched prompt: intervals that only touch, e.g. [1,2] and [2,3], do NOT overlap.

func eraseOverlapLiteral(iv [][2]int) int {
	iv = slices.Clone(iv)
	slices.SortFunc(iv, func(a, b [2]int) int { return a[1] - b[1] })
	kept, lastEnd := 0, math.MinInt
	for _, x := range iv {
		if x[0] >= lastEnd {
			kept++
			lastEnd = x[1]
		}
	}
	return len(iv) - kept
}

// brute: max subset size that is pairwise non-overlapping (half-open semantics).
func eraseOverlapBrute(iv [][2]int) int {
	n, best := len(iv), 0
	for mask := 0; mask < 1<<n; mask++ {
		ok, cnt := true, 0
		for i := 0; i < n && ok; i++ {
			if mask>>i&1 == 0 {
				continue
			}
			cnt++
			for j := i + 1; j < n; j++ {
				if mask>>j&1 == 1 && iv[i][0] < iv[j][1] && iv[j][0] < iv[i][1] {
					ok = false
					break
				}
			}
		}
		if ok && cnt > best {
			best = cnt
		}
	}
	return n - best
}

func TestNonOverlapping(t *testing.T) {
	r := newRng(10)
	check := func(iv [][2]int) {
		if g, w := eraseOverlapLiteral(iv), eraseOverlapBrute(iv); g != w {
			t.Fatalf("iv=%v got=%d want=%d", iv, g, w)
		}
	}
	check(nil)
	check([][2]int{{1, 2}})
	check([][2]int{{1, 2}, {2, 3}, {3, 4}, {1, 3}})
	check([][2]int{{1, 2}, {1, 2}, {1, 2}})
	check([][2]int{{-5, -1}, {-3, 0}})
	for range 5000 {
		n := r.IntN(9)
		iv := make([][2]int, n)
		for i := range iv {
			a := r.IntN(10) - 3
			iv[i] = [2]int{a, a + 1 + r.IntN(5)}
		}
		check(iv)
	}
}

// ---------- gd-jump-game ----------

func canJumpLiteral(nums []int) bool {
	far := 0
	for i := range nums {
		if i > far {
			return false
		}
		far = max(far, i+nums[i])
	}
	return true
}

// Alternative in patched technique: backward greedy, tracking the leftmost index
// known to reach the end.
func canJumpBackward(nums []int) bool {
	last := len(nums) - 1
	for i := len(nums) - 2; i >= 0; i-- {
		if i+nums[i] >= last {
			last = i
		}
	}
	return last == 0
}

func canJumpBrute(nums []int) bool {
	n := len(nums)
	reach := make([]bool, n)
	reach[0] = true
	for i := range n {
		if !reach[i] {
			continue
		}
		for j := i + 1; j <= i+nums[i] && j < n; j++ {
			reach[j] = true
		}
	}
	return reach[n-1]
}

func TestJumpGame(t *testing.T) {
	r := newRng(11)
	check := func(a []int) {
		w := canJumpBrute(a)
		if g := canJumpLiteral(a); g != w {
			t.Fatalf("a=%v got=%v want=%v", a, g, w)
		}
		if g := canJumpBackward(a); g != w {
			t.Fatalf("backward: a=%v got=%v want=%v", a, g, w)
		}
	}
	check([]int{0})
	check([]int{2, 3, 1, 1, 4})
	check([]int{3, 2, 1, 0, 4})
	check([]int{0, 1})
	for range 5000 {
		n := 1 + r.IntN(10)
		a := make([]int, n)
		for i := range a {
			a[i] = r.IntN(4)
		}
		check(a)
	}
}

// ---------- gd-gas-station ----------

func gasLiteral(gas, cost []int) int {
	sg, sc := 0, 0
	for i := range gas {
		sg += gas[i]
		sc += cost[i]
	}
	if sg < sc {
		return -1
	}
	tank, start := 0, 0
	for i := range gas {
		tank += gas[i] - cost[i]
		if tank < 0 {
			start = i + 1
			tank = 0
		}
	}
	return start
}

// brute: simulate from every start; return all valid starts.
func gasBrute(gas, cost []int) []int {
	n := len(gas)
	var ok []int
	for s := range n {
		tank, good := 0, true
		for k := range n {
			i := (s + k) % n
			tank += gas[i] - cost[i]
			if tank < 0 {
				good = false
				break
			}
		}
		if good {
			ok = append(ok, s)
		}
	}
	return ok
}

func TestGasStation(t *testing.T) {
	r := newRng(12)
	tested := 0
	check := func(gas, cost []int) {
		valid := gasBrute(gas, cost)
		if len(valid) > 1 {
			return // prompt guarantees uniqueness; skip inputs outside the constraint
		}
		tested++
		want := -1
		if len(valid) == 1 {
			want = valid[0]
		}
		if g := gasLiteral(gas, cost); g != want {
			t.Fatalf("gas=%v cost=%v got=%d want=%d", gas, cost, g, want)
		}
	}
	check([]int{1, 2, 3, 4, 5}, []int{3, 4, 5, 1, 2})
	check([]int{2, 3, 4}, []int{3, 4, 3})
	check([]int{5}, []int{4})
	check([]int{4}, []int{5})
	for range 20000 {
		n := 1 + r.IntN(7)
		gas, cost := make([]int, n), make([]int, n)
		for i := range n {
			gas[i], cost[i] = r.IntN(6), r.IntN(6)
		}
		check(gas, cost)
	}
	if tested < 5000 {
		t.Fatalf("too few valid cases: %d", tested)
	}
}

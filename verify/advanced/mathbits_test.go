package advanced

import (
	"math"
	"math/big"
	"math/bits"
	"math/rand/v2"
	"strings"
	"testing"
)

// ---------- mb-single-number-iii ----------

func singleNumberIIILiteral(a []int) (int, int) {
	x := 0
	for _, v := range a {
		x ^= v
	}
	bit := x & -x
	p := 0
	for _, v := range a {
		if v&bit != 0 {
			p ^= v
		}
	}
	return p, x ^ p
}

func TestSingleNumberIII(t *testing.T) {
	r := newRng(40)
	check := func(a []int) {
		cnt := map[int]int{}
		for _, v := range a {
			cnt[v]++
		}
		var want []int
		for v, c := range cnt {
			if c == 1 {
				want = append(want, v)
			}
		}
		g1, g2 := singleNumberIIILiteral(a)
		if !((g1 == want[0] && g2 == want[1]) || (g1 == want[1] && g2 == want[0])) {
			t.Fatalf("a=%v got=%d,%d want=%v", a, g1, g2, want)
		}
	}
	check([]int{1, 2})
	check([]int{0, 1})
	check([]int{-1, 0})
	check([]int{1, 2, 1, 3, 2, 5})
	check([]int{math.MinInt, math.MaxInt, 7, 7})
	for range 5000 {
		k := r.IntN(6)
		vals := map[int]bool{}
		for len(vals) < k+2 {
			vals[r.IntN(41)-20] = true
		}
		var a []int
		i := 0
		for v := range vals {
			a = append(a, v)
			if i >= 2 {
				a = append(a, v)
			}
			i++
		}
		r.Shuffle(len(a), func(i, j int) { a[i], a[j] = a[j], a[i] })
		check(a)
	}
}

// ---------- mb-counting-bits ----------

func TestCountingBits(t *testing.T) {
	for _, n := range []int{0, 1, 2, 3, 1000} {
		a, b := make([]int, n+1), make([]int, n+1)
		for i := 1; i <= n; i++ {
			a[i] = a[i&(i-1)] + 1
			b[i] = b[i>>1] + (i & 1)
		}
		for i := 0; i <= n; i++ {
			w := bits.OnesCount(uint(i))
			if a[i] != w || b[i] != w {
				t.Fatalf("i=%d got %d/%d want %d", i, a[i], b[i], w)
			}
		}
	}
}

// ---------- mb-gcd-strings ----------

func gcdInt(a, b int) int {
	if b == 0 {
		return a
	}
	return gcdInt(b, a%b)
}

func gcdStringsLiteral(s1, s2 string) string {
	if s1+s2 != s2+s1 {
		return ""
	}
	return s1[:gcdInt(len(s1), len(s2))]
}

func gcdStringsBrute(s1, s2 string) string {
	divides := func(x, s string) bool {
		return len(x) > 0 && len(s)%len(x) == 0 && strings.Repeat(x, len(s)/len(x)) == s
	}
	for k := min(len(s1), len(s2)); k >= 1; k-- {
		x := s1[:k]
		if divides(x, s1) && divides(x, s2) {
			return x
		}
	}
	return ""
}

func TestGCDStrings(t *testing.T) {
	r := newRng(41)
	check := func(a, b string) {
		if g, w := gcdStringsLiteral(a, b), gcdStringsBrute(a, b); g != w {
			t.Fatalf("a=%q b=%q got=%q want=%q", a, b, g, w)
		}
	}
	check("ABCABC", "ABC")
	check("ABABAB", "ABAB")
	check("LEET", "CODE")
	check("A", "A")
	check("AAAA", "AA")
	for range 10000 {
		if r.IntN(2) == 0 {
			u := randStr(r, 1+r.IntN(3), "ab")
			check(strings.Repeat(u, 1+r.IntN(5)), strings.Repeat(u, 1+r.IntN(5)))
		} else {
			check(randStr(r, 1+r.IntN(8), "ab"), randStr(r, 1+r.IntN(8), "ab"))
		}
	}
}

// ---------- mb-count-primes ----------

// Literal transcription of the PATCHED key_idea ("If n < 2, return 0. Otherwise ...").
// The original text (no guard) indexed is[0], is[1] out of range for n=0 and n=1.
func countPrimesLiteral(n int) int {
	if n < 2 {
		return 0
	}
	is := make([]bool, n)
	for i := range is {
		is[i] = true
	}
	is[0], is[1] = false, false
	for i := 2; i*i < n; i++ {
		if is[i] {
			for j := i * i; j < n; j += i {
				is[j] = false
			}
		}
	}
	c := 0
	for _, v := range is {
		if v {
			c++
		}
	}
	return c
}

func countPrimesBrute(n int) int {
	c := 0
	for k := 2; k < n; k++ {
		p := true
		for d := 2; d*d <= k; d++ {
			if k%d == 0 {
				p = false
				break
			}
		}
		if p {
			c++
		}
	}
	return c
}

func TestCountPrimes(t *testing.T) {
	for n := 0; n <= 3000; n++ { // includes n = 0, 1, 2
		if g, w := countPrimesLiteral(n), countPrimesBrute(n); g != w {
			t.Fatalf("n=%d got=%d want=%d", n, g, w)
		}
	}
	if g := countPrimesLiteral(5_000_000); g != 348513 { // pi(5e6 - 1) = 348513
		t.Fatalf("n=5e6 got=%d", g)
	}
}

// ---------- mb-pow ----------

func powLiteral(x float64, n32 int32) float64 {
	n := int64(n32)
	if n < 0 {
		x = 1 / x
		n = -n
	}
	res := 1.0
	for n != 0 {
		if n&1 == 1 {
			res *= x
		}
		x *= x
		n >>= 1
	}
	return res
}

func TestPow(t *testing.T) {
	r := newRng(42)
	check := func(x float64, n int32) {
		g, w := powLiteral(x, n), math.Pow(x, float64(n))
		if g == w || (math.IsNaN(g) && math.IsNaN(w)) {
			return
		}
		// repeated squaring accumulates ~2*log2|n| rounding errors; 1e-12 relative is ample
		if math.IsInf(g, 0) || math.IsInf(w, 0) || math.Abs(g-w) > 1e-12*math.Abs(w) {
			// tolerate pure under/overflow-boundary disagreements only when both are tiny
			if !(math.Abs(g) < 1e-300 && math.Abs(w) < 1e-300) {
				t.Fatalf("x=%v n=%d got=%v want=%v", x, n, g, w)
			}
		}
	}
	for _, c := range []struct {
		x float64
		n int32
	}{{2, 10}, {2.1, 3}, {2, -2}, {1, math.MinInt32}, {-1, math.MinInt32}, {-1, math.MaxInt32},
		{2, math.MinInt32}, {0.5, math.MinInt32}, {5, 0}, {0, 0}, {0, 3}, {0, -3}, {-2, -3}, {1.0000001, math.MaxInt32}} {
		check(c.x, c.n)
	}
	for range 20000 {
		x := (r.Float64()*4 - 2)
		var n int32
		if r.IntN(2) == 0 {
			n = int32(r.IntN(81) - 40)
		} else {
			n = int32(r.Uint32())
		}
		check(x, n)
	}
}

// ---------- mb-super-pow ----------

func superPowLiteral(a int, b []int) int {
	const m = 1337
	a %= m                         // patched: reduce a first
	powmod := func(x, e int) int { // fast exponentiation, reduce after every multiplication
		r := 1
		for e > 0 {
			if e&1 == 1 {
				r = r * x % m
			}
			x = x * x % m
			e >>= 1
		}
		return r
	}
	res := 1
	for _, d := range b {
		res = powmod(res, 10) * powmod(a, d) % m
	}
	return res
}

func TestSuperPow(t *testing.T) {
	r := newRng(43)
	check := func(a int, b []int) {
		var sb strings.Builder
		for _, d := range b {
			sb.WriteByte(byte('0' + d))
		}
		e, _ := new(big.Int).SetString(sb.String(), 10)
		w := int(new(big.Int).Exp(big.NewInt(int64(a)), e, big.NewInt(1337)).Int64())
		if g := superPowLiteral(a, b); g != w {
			t.Fatalf("a=%d b=%v got=%d want=%d", a, b, g, w)
		}
	}
	check(2, []int{3})
	check(2, []int{1, 0})
	check(1, []int{4, 3, 3, 8, 5, 2})
	check(2147483647, []int{2, 0, 0})
	check(1337, []int{1})
	check(0, []int{5})
	for range 5000 {
		var a int
		if r.IntN(2) == 0 {
			a = 1 + r.IntN(3000)
		} else {
			a = 1 + r.IntN(math.MaxInt32)
		}
		b := make([]int, 1+r.IntN(30))
		for i := range b {
			b[i] = r.IntN(10)
		}
		b[0] = 1 + r.IntN(9) // positive, no leading zero
		check(a, b)
	}
}

// ---------- mb-random-node ----------

type listNode struct {
	val  int
	next *listNode
}

func randomNodeLiteral(head *listNode, rng *rand.Rand) int {
	pick := 0 // rand(i) == rng.IntN(i), uniform in [0, i)
	i := 0
	for n := head; n != nil; n = n.next {
		i++
		if rng.IntN(i) == 0 {
			pick = n.val
		}
	}
	return pick
}

func TestRandomNode(t *testing.T) {
	rng := newRng(44)
	for n := 1; n <= 8; n++ {
		var head *listNode
		for v := n - 1; v >= 0; v-- {
			head = &listNode{val: v, next: head}
		}
		const trials = 80000
		cnt := make([]int, n)
		for range trials {
			cnt[randomNodeLiteral(head, rng)]++
		}
		if n == 1 {
			if cnt[0] != trials {
				t.Fatalf("n=1 must always return the only node")
			}
			continue
		}
		// Pearson chi-square vs uniform, df = n-1 <= 7. The p=0.001 critical value
		// for df=7 is 24.32 (smaller df have smaller criticals: df=1 -> 10.83), so
		// using a per-df table keeps the false-alarm rate at 0.1% per n; with a fixed
		// seed the result is deterministic anyway. A biased sampler (e.g. always
		// replacing with prob 1/2) gives chi2 in the thousands at this trial count.
		crit := []float64{0, 10.83, 13.82, 16.27, 18.47, 20.52, 22.46, 24.32}
		exp := float64(trials) / float64(n)
		chi2 := 0.0
		for _, c := range cnt {
			d := float64(c) - exp
			chi2 += d * d / exp
		}
		if chi2 > crit[n-1] {
			t.Fatalf("n=%d chi2=%.2f > %.2f counts=%v", n, chi2, crit[n-1], cnt)
		}
	}
}

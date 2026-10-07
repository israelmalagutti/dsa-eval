package structures

import (
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// ===== ms-daily-temps =====

// Literal (patched): ans = n zeros; stack of indices, initially empty. For i in 0..n-1:
// while the stack is not empty and t[i] > t[top]: j = pop, ans[j] = i - j. Then push i.
func dailyTempsLiteral(t []int) []int {
	ans := make([]int, len(t))
	var st []int
	for i := range t {
		for len(st) > 0 && t[i] > t[st[len(st)-1]] {
			j := st[len(st)-1]
			st = st[:len(st)-1]
			ans[j] = i - j
		}
		st = append(st, i)
	}
	return ans
}

func dailyTempsBrute(t []int) []int {
	ans := make([]int, len(t))
	for i := range t {
		for j := i + 1; j < len(t); j++ {
			if t[j] > t[i] {
				ans[i] = j - i
				break
			}
		}
	}
	return ans
}

func TestDailyTemps(t *testing.T) {
	r := newRNG(1)
	cases := [][]int{{}, {50}, {30, 30, 30}, {73, 74, 75, 71, 69, 72, 76, 73}, {5, 4, 3, 2, 1}, {1, 2, 3, 4}, {-5, -10, -3}}
	for range 5000 {
		cases = append(cases, randInts(r, r.IntN(12), -3, 3))
	}
	for _, c := range cases {
		if got, want := dailyTempsLiteral(c), dailyTempsBrute(c); !slices.Equal(got, want) {
			t.Fatalf("%v: got %v want %v", c, got, want)
		}
	}
}

// ===== ms-histogram =====

// Literal (patched): append a sentinel 0; stack empty; best = 0. For i: while the stack
// is not empty and h[i] < h[top]: height = h[pop]; width = i if the stack is now empty,
// else i - top - 1 (top after the pop); best = max(best, height*width). Then push i.
func histogramLiteral(heights []int) int {
	h := append(slices.Clone(heights), 0)
	var st []int
	best := 0
	for i := range h {
		for len(st) > 0 && h[i] < h[st[len(st)-1]] {
			height := h[st[len(st)-1]]
			st = st[:len(st)-1]
			var width int
			if len(st) == 0 {
				width = i
			} else {
				width = i - st[len(st)-1] - 1
			}
			best = max(best, height*width)
		}
		st = append(st, i)
	}
	return best
}

func histogramBrute(h []int) int {
	best := 0
	for i := range h {
		mn := h[i]
		for j := i; j < len(h); j++ {
			mn = min(mn, h[j])
			best = max(best, mn*(j-i+1))
		}
	}
	return best
}

func TestHistogram(t *testing.T) {
	r := newRNG(2)
	cases := [][]int{{}, {0}, {7}, {2, 1, 5, 6, 2, 3}, {2, 2, 2, 2}, {1, 2, 3, 4, 5}, {5, 4, 3, 2, 1}, {0, 0, 3, 0}}
	for range 5000 {
		cases = append(cases, randInts(r, r.IntN(12), 0, 5))
	}
	for _, c := range cases {
		if got, want := histogramLiteral(c), histogramBrute(c); got != want {
			t.Fatalf("%v: got %d want %d", c, got, want)
		}
	}
}

// ===== dq-window-max =====

// Literal (patched): deque empty. For i: while the deque is not empty and a[back] <= a[i],
// pop the back; push i; if front <= i-k pop the front; if i >= k-1 emit a[front].
func windowMaxLiteral(a []int, k int) []int {
	var dq, out []int
	for i := range a {
		for len(dq) > 0 && a[dq[len(dq)-1]] <= a[i] {
			dq = dq[:len(dq)-1]
		}
		dq = append(dq, i)
		if dq[0] <= i-k {
			dq = dq[1:]
		}
		if i >= k-1 {
			out = append(out, a[dq[0]])
		}
	}
	return out
}

func windowMaxBrute(a []int, k int) []int {
	var out []int
	for i := 0; i+k <= len(a); i++ {
		out = append(out, slices.Max(a[i:i+k]))
	}
	return out
}

func TestWindowMax(t *testing.T) {
	r := newRNG(3)
	type tc struct {
		a []int
		k int
	}
	cases := []tc{{[]int{5}, 1}, {[]int{1, 3, -1, -3, 5, 3, 6, 7}, 3}, {[]int{2, 2, 2}, 2}, {[]int{4, 3, 2, 1}, 4}, {[]int{1, 2, 3}, 1}}
	for range 5000 {
		n := 1 + r.IntN(12)
		cases = append(cases, tc{randInts(r, n, -3, 3), 1 + r.IntN(n)})
	}
	for _, c := range cases {
		if got, want := windowMaxLiteral(c.a, c.k), windowMaxBrute(c.a, c.k); !slices.Equal(got, want) {
			t.Fatalf("%v k=%d: got %v want %v", c.a, c.k, got, want)
		}
	}
}

// ===== st-decode-string =====

// Literal (patched): cur = empty builder, num = 0, stack empty. Digit: num = num*10 + d.
// '[': push (cur, num), cur = new empty builder, num = 0. ']': pop (prev, k); append cur's
// contents to prev k times; cur = prev. Letter: append to cur. Return cur's contents.
func decodeLiteral(s string) string {
	type frame struct {
		b *strings.Builder
		k int
	}
	var st []frame
	cur, num := &strings.Builder{}, 0
	for _, c := range s {
		switch {
		case c >= '0' && c <= '9':
			num = num*10 + int(c-'0')
		case c == '[':
			st = append(st, frame{cur, num})
			cur, num = &strings.Builder{}, 0
		case c == ']':
			f := st[len(st)-1]
			st = st[:len(st)-1]
			inner := cur.String()
			for range f.k {
				f.b.WriteString(inner)
			}
			cur = f.b
		default:
			cur.WriteRune(c)
		}
	}
	return cur.String()
}

// Brute: recursive-descent parser, independent of the stack formulation.
func decodeBrute(s string) string {
	var parse func(i int) (string, int)
	parse = func(i int) (string, int) {
		var sb strings.Builder
		for i < len(s) && s[i] != ']' {
			if s[i] >= '0' && s[i] <= '9' {
				j := i
				for s[j] >= '0' && s[j] <= '9' {
					j++
				}
				k, _ := strconv.Atoi(s[i:j])
				inner, end := parse(j + 1) // skip '['
				for range k {
					sb.WriteString(inner)
				}
				i = end + 1 // skip ']'
			} else {
				sb.WriteByte(s[i])
				i++
			}
		}
		return sb.String(), i
	}
	out, _ := parse(0)
	return out
}

// genEncoded produces a valid encoded string; depth-bounded, k in [1,12] (exercises multi-digit).
func genEncoded(r *rand.Rand, depth int) string {
	var sb strings.Builder
	parts := r.IntN(4)
	for range parts {
		if depth > 0 && r.IntN(2) == 0 {
			sb.WriteString(strconv.Itoa(1 + r.IntN(12)))
			sb.WriteString("[" + genEncoded(r, depth-1) + "]")
		} else {
			sb.WriteByte(byte('a' + r.IntN(3)))
		}
	}
	return sb.String()
}

func TestDecodeString(t *testing.T) {
	r := newRNG(4)
	cases := []string{"", "abc", "3[a2[c]]", "2[abc]3[cd]ef", "10[a]", "1[]", "2[]x", "100[ab]", "a2[b3[c]]d"}
	for range 5000 {
		cases = append(cases, genEncoded(r, 3))
	}
	for _, c := range cases {
		if got, want := decodeLiteral(c), decodeBrute(c); got != want {
			t.Fatalf("%q: got %q want %q", c, got, want)
		}
	}
}

// ===== st-basic-calculator =====

// Literal (patched): res = 0, sign = 1, num = 0, stack empty. Space: skip. Digit:
// num = num*10 + digit. '+'/'-': res += sign*num, num = 0, sign = +1/-1. '(': push res,
// push sign, then res = 0, sign = 1. ')': res += sign*num, num = 0, then
// res = pop_res + pop_sign*res. At the end return res + sign*num.
// (The original text never reset num and never flushed the last number:
// "1" -> 0, "1+2" -> 1, "(1)+2" -> 2.)
func calcLiteral(s string) int {
	res, sign, num := 0, 1, 0
	var st []int
	for _, c := range s {
		switch {
		case c == ' ':
		case c >= '0' && c <= '9':
			num = num*10 + int(c-'0')
		case c == '+' || c == '-':
			res += sign * num
			num = 0
			if c == '+' {
				sign = 1
			} else {
				sign = -1
			}
		case c == '(':
			st = append(st, res, sign)
			res, sign = 0, 1
		case c == ')':
			res += sign * num
			num = 0
			popSign, popRes := st[len(st)-1], st[len(st)-2]
			st = st[:len(st)-2]
			res = popRes + popSign*res
		}
	}
	return res + sign*num
}

// genExpr returns an expression string and its value computed during generation
// (the brute force: value is known by construction). Unary '-' appears only where
// the patched prompt allows it.
func genExpr(r *rand.Rand, depth int) (string, int) {
	term := func() (string, int) {
		if depth > 0 && r.IntN(3) == 0 {
			s, v := genExpr(r, depth-1)
			return "(" + s + ")", v
		}
		v := r.IntN(120)
		return strconv.Itoa(v), v
	}
	sp := func() string { return strings.Repeat(" ", r.IntN(2)) }
	s, v := term()
	if r.IntN(4) == 0 { // unary minus: only at the start of an expression (incl. right after '(')
		s, v = "-"+sp()+s, -v
	}
	s = sp() + s
	for range r.IntN(4) {
		ts, tv := term()
		if r.IntN(2) == 0 {
			s += sp() + "+" + sp() + ts
			v += tv
		} else {
			s += sp() + "-" + sp() + ts
			v -= tv
		}
	}
	return s + sp(), v
}

func TestBasicCalculator(t *testing.T) {
	r := newRNG(5)
	type tc struct {
		s string
		v int
	}
	cases := []tc{{"0", 0}, {"1", 1}, {"1+2", 3}, {"(1)+2", 3}, {"1 + 1", 2}, {" 2-1 + 2 ", 3},
		{"(1+(4+5+2)-3)+(6+8)", 23}, {"1 - (2 + 3) + 4", 0}, {"10 - (20 - (30 - 40))", -20},
		{"((((7))))", 7}, {"123", 123}, {"-2+1", -1}, {"-(2+3)", -5}, {"1 - (-2)", 3}, {"-(-(-4))", -4}}
	for range 8000 {
		s, v := genExpr(r, 3)
		cases = append(cases, tc{s, v})
	}
	for _, c := range cases {
		if got := calcLiteral(c.s); got != c.v {
			t.Fatalf("%q: got %d want %d", c.s, got, c.v)
		}
	}
}

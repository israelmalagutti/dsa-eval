package advanced

import (
	"strings"
	"testing"
)

// ---------- str-strstr-kmp ----------
// Patched prompt: needle is non-empty, return -1 if absent. Tests use m >= 1.

func kmpLiteral(h, p string) int {
	m := len(p)
	lps := make([]int, m)
	for i, k := 1, 0; i < m; i++ {
		for k > 0 && p[i] != p[k] {
			k = lps[k-1]
		}
		if p[i] == p[k] {
			k++
		}
		lps[i] = k
	}
	j := 0
	for i := 0; i < len(h); i++ { // text pointer never moves backward
		for j > 0 && h[i] != p[j] {
			j = lps[j-1]
		}
		if h[i] == p[j] {
			j++
		}
		if j == m {
			return i - m + 1
		}
	}
	return -1
}

func lpsBrute(p string) []int {
	out := make([]int, len(p))
	for i := range p {
		s := p[:i+1]
		for k := i; k > 0; k-- {
			if s[:k] == s[len(s)-k:] {
				out[i] = k
				break
			}
		}
	}
	return out
}

func TestKMP(t *testing.T) {
	r := newRng(30)
	check := func(h, p string) {
		if g, w := kmpLiteral(h, p), strings.Index(h, p); g != w {
			t.Fatalf("h=%q p=%q got=%d want=%d", h, p, g, w)
		}
	}
	check("", "a")
	check("a", "a")
	check("sadbutsad", "sad")
	check("leetcode", "leeto")
	check("aaaaab", "aab")
	check("abababca", "ababca")
	for range 20000 {
		check(randStr(r, r.IntN(15), "ab"), randStr(r, 1+r.IntN(5), "ab"))
	}
	// also check the lps definition directly (via the same loop as above)
	for range 2000 {
		p := randStr(r, 1+r.IntN(10), "ab")
		lps := make([]int, len(p))
		for i, k := 1, 0; i < len(p); i++ {
			for k > 0 && p[i] != p[k] {
				k = lps[k-1]
			}
			if p[i] == p[k] {
				k++
			}
			lps[i] = k
		}
		w := lpsBrute(p)
		for i := range w {
			if lps[i] != w[i] {
				t.Fatalf("lps(%q) got=%v want=%v", p, lps, w)
			}
		}
	}
}

// ---------- str-longest-dup-substring ----------

func longestDupLiteral(s string) string {
	const B, M = 131, 1_000_000_007
	n := len(s)
	check := func(L int) int { // returns start index or -1
		pw := 1
		for range L {
			pw = pw * B % M
		}
		seen := map[int][]int{}
		h := 0
		for i := range n {
			// h = (h*B - (s[i-L]*P if i >= L else 0) + s[i]) mod M, normalized
			sub := 0
			if i >= L {
				sub = int(s[i-L]) * pw % M
			}
			h = (h*B - sub + int(s[i])) % M
			if h < 0 {
				h += M
			}
			if i >= L-1 {
				st := i - L + 1
				for _, o := range seen[h] { // verify on collision
					if s[o:o+L] == s[st:st+L] {
						return st
					}
				}
				seen[h] = append(seen[h], st)
			}
		}
		return -1
	}
	lo, hi, best, bestAt := 1, n-1, 0, 0
	for lo <= hi {
		mid := (lo + hi) / 2
		if at := check(mid); at >= 0 {
			best, bestAt, lo = mid, at, mid+1
		} else {
			hi = mid - 1
		}
	}
	return s[bestAt : bestAt+best]
}

func longestDupBruteLen(s string) int {
	for L := len(s) - 1; L >= 1; L-- {
		seen := map[string]bool{}
		for i := 0; i+L <= len(s); i++ {
			if seen[s[i:i+L]] {
				return L
			}
			seen[s[i:i+L]] = true
		}
	}
	return 0
}

func TestLongestDupSubstring(t *testing.T) {
	r := newRng(31)
	check := func(s string) {
		g := longestDupLiteral(s)
		w := longestDupBruteLen(s)
		occ := 0
		if g != "" {
			for i := 0; i+len(g) <= len(s); i++ {
				if s[i:i+len(g)] == g {
					occ++
				}
			}
		}
		if len(g) != w || (g != "" && occ < 2) {
			t.Fatalf("s=%q got=%q (occ %d) want len %d", s, g, occ, w)
		}
	}
	check("")
	check("a")
	check("abcd")
	check("banana")
	check("aaaa")
	for range 5000 {
		check(randStr(r, r.IntN(16), "abc"[:1+r.IntN(3)]))
	}
}

// ---------- str-shortest-palindrome ----------
// Patched prompt restricts s to lowercase letters, so '#' is a safe separator.
// (Unpatched, s="#" gave t="###", last lps=2>len(s), and s[2:] was out of range.)

func shortestPalLiteral(s string) string {
	rev := func(x string) string {
		b := []byte(x)
		for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
			b[i], b[j] = b[j], b[i]
		}
		return string(b)
	}
	t := s + "#" + rev(s)
	lps := make([]int, len(t))
	for i, k := 1, 0; i < len(t); i++ {
		for k > 0 && t[i] != t[k] {
			k = lps[k-1]
		}
		if t[i] == t[k] {
			k++
		}
		lps[i] = k
	}
	k := lps[len(t)-1]
	return rev(s[k:]) + s
}

// Z alternative from the patched key_idea: k = largest len(t)-j over j > len(s)
// with j + z[j] == len(t), or 0.
func shortestPalZ(s string) string {
	b := []byte(s)
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	rs := string(b)
	t := s + "#" + rs
	n := len(t)
	z := make([]int, n)
	for i, l, r := 1, 0, 0; i < n; i++ {
		if i < r {
			z[i] = min(r-i, z[i-l])
		}
		for i+z[i] < n && t[z[i]] == t[i+z[i]] {
			z[i]++
		}
		if i+z[i] > r {
			l, r = i, i+z[i]
		}
	}
	k := 0
	for j := len(s) + 1; j < n; j++ {
		if j+z[j] == n && n-j > k {
			k = n - j
		}
	}
	return rs[:len(s)-k] + s
}

func isPal(s string) bool {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		if s[i] != s[j] {
			return false
		}
	}
	return true
}

func shortestPalBrute(s string) string {
	for k := len(s); k >= 0; k-- {
		if isPal(s[:k]) {
			suf := []byte(s[k:])
			for i, j := 0, len(suf)-1; i < j; i, j = i+1, j-1 {
				suf[i], suf[j] = suf[j], suf[i]
			}
			return string(suf) + s
		}
	}
	return s
}

func TestShortestPalindrome(t *testing.T) {
	r := newRng(32)
	check := func(s string) {
		w := shortestPalBrute(s)
		if g := shortestPalLiteral(s); g != w {
			t.Fatalf("s=%q got=%q want=%q", s, g, w)
		}
		if g := shortestPalZ(s); g != w {
			t.Fatalf("Z: s=%q got=%q want=%q", s, g, w)
		}
	}
	for _, s := range []string{"", "a", "aa", "ab", "aacecaaa", "abcd", "abab"} {
		check(s)
	}
	for range 10000 {
		check(randStr(r, r.IntN(12), "ab"[:1+r.IntN(2)]))
	}
}

// ---------- str-longest-palindrome ----------

func manacherLiteral(s string) string {
	tb := []byte{'^'}
	for i := 0; i < len(s); i++ {
		tb = append(tb, '#', s[i])
	}
	tb = append(tb, '#', '$')
	T := string(tb)
	p := make([]int, len(T))
	C, R := 0, 0
	for i := 1; i < len(T)-1; i++ {
		if i < R {
			p[i] = min(R-i, p[2*C-i])
		}
		for T[i+1+p[i]] == T[i-1-p[i]] {
			p[i]++
		}
		if i+p[i] > R {
			C, R = i, i+p[i]
		}
	}
	best, center := 0, 0
	for i := 1; i < len(T)-1; i++ {
		if p[i] > best {
			best, center = p[i], i
		}
	}
	start := (center - best) / 2
	return s[start : start+best]
}

func TestLongestPalindrome(t *testing.T) {
	r := newRng(33)
	check := func(s string) {
		g := manacherLiteral(s)
		w := 0
		for i := range len(s) {
			for j := i + 1; j <= len(s); j++ {
				if j-i > w && isPal(s[i:j]) {
					w = j - i
				}
			}
		}
		if len(g) != w || !isPal(g) || !strings.Contains(s, g) {
			t.Fatalf("s=%q got=%q want len %d", s, g, w)
		}
	}
	for _, s := range []string{"", "a", "ab", "aa", "babad", "cbbd", "aaaa", "abacdfgdcaba"} {
		check(s)
	}
	for range 10000 {
		check(randStr(r, r.IntN(16), "abc"[:1+r.IntN(3)]))
	}
}

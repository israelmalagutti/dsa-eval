package basics

import (
	"math/rand/v2"
	"slices"
)

const iters = 5000

func newRNG() *rand.Rand { return rand.New(rand.NewPCG(42, 1337)) }

// randInts returns a slice of length n with values in [lo, hi].
func randInts(r *rand.Rand, n, lo, hi int) []int {
	a := make([]int, n)
	for i := range a {
		a[i] = lo + r.IntN(hi-lo+1)
	}
	return a
}

// randDistinct returns n distinct values from [lo, hi] (requires hi-lo+1 >= n).
func randDistinct(r *rand.Rand, n, lo, hi int) []int {
	p := r.Perm(hi - lo + 1)[:n]
	for i := range p {
		p[i] += lo
	}
	return p
}

func sortedCopy(a []int) []int {
	b := slices.Clone(a)
	slices.Sort(b)
	return b
}

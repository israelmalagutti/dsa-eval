package structures

import (
	"container/heap"
	"math/rand/v2"
	"slices"
	"testing"
)

// generic heap over container/heap
type hp[T any] struct {
	a    []T
	less func(x, y T) bool
}

func (h *hp[T]) Len() int           { return len(h.a) }
func (h *hp[T]) Less(i, j int) bool { return h.less(h.a[i], h.a[j]) }
func (h *hp[T]) Swap(i, j int)      { h.a[i], h.a[j] = h.a[j], h.a[i] }
func (h *hp[T]) Push(x any)         { h.a = append(h.a, x.(T)) }
func (h *hp[T]) Pop() any {
	x := h.a[len(h.a)-1]
	h.a = h.a[:len(h.a)-1]
	return x
}
func (h *hp[T]) top() T { return h.a[0] }

// ===== hp-k-closest =====

type pt struct{ x, y int }

func dist(p pt) int { return p.x*p.x + p.y*p.y }

// Literal (patched): d = x*x + y*y; push (d, p) into a max-heap; if size > k pop the max.
func kClosestLiteral(pts []pt, k int) []pt {
	type item struct {
		d int
		p pt
	}
	h := &hp[item]{less: func(a, b item) bool { return a.d > b.d }}
	for _, p := range pts {
		heap.Push(h, item{dist(p), p})
		if h.Len() > k {
			heap.Pop(h)
		}
	}
	var out []pt
	for _, it := range h.a {
		out = append(out, it.p)
	}
	return out
}

func TestKClosest(t *testing.T) {
	r := newRNG(20)
	for it := range 5000 {
		n := 1 + r.IntN(15)
		pts := make([]pt, n)
		for i := range pts {
			pts[i] = pt{r.IntN(9) - 4, r.IntN(9) - 4}
		}
		if it == 0 {
			pts = []pt{{1, 3}, {-2, 2}}
		}
		k := 1 + r.IntN(n)
		got := kClosestLiteral(pts, k)
		// Brute: sort by distance. Ties make the exact set non-unique, so compare
		// distance multisets and check every returned point is from the input.
		sorted := slices.Clone(pts)
		slices.SortFunc(sorted, func(a, b pt) int { return dist(a) - dist(b) })
		var wd, gd []int
		for _, p := range sorted[:k] {
			wd = append(wd, dist(p))
		}
		cnt := map[pt]int{}
		for _, p := range pts {
			cnt[p]++
		}
		for _, p := range got {
			gd = append(gd, dist(p))
			cnt[p]--
			if cnt[p] < 0 {
				t.Fatalf("point %v not in input (or used too often)", p)
			}
		}
		slices.Sort(gd)
		if !slices.Equal(gd, wd) {
			t.Fatalf("%v k=%d: got %v want dists %v", pts, k, got, wd)
		}
	}
}

// ===== hp-merge-k-lists =====

// Literal (patched): min-heap of (val, idx, node) by val then idx; push non-null heads;
// pop, append to tail, push node.next if it exists.
func mergeKLiteral(lists []*ListNode) *ListNode {
	type item struct {
		val, idx int
		node     *ListNode
	}
	h := &hp[item]{less: func(a, b item) bool {
		if a.val != b.val {
			return a.val < b.val
		}
		return a.idx < b.idx
	}}
	for i, l := range lists {
		if l != nil {
			heap.Push(h, item{l.Val, i, l})
		}
	}
	d := &ListNode{}
	t := d
	for h.Len() > 0 {
		it := heap.Pop(h).(item)
		t.Next = it.node
		t = t.Next
		if it.node.Next != nil {
			heap.Push(h, item{it.node.Next.Val, it.idx, it.node.Next})
		}
	}
	t.Next = nil
	return d.Next
}

func TestMergeKLists(t *testing.T) {
	r := newRNG(21)
	for it := range 5000 {
		k := r.IntN(6)
		var lists []*ListNode
		var all []int
		for range k {
			a := randInts(r, r.IntN(6), -4, 4)
			slices.Sort(a)
			all = append(all, a...)
			lists = append(lists, fromSlice(a))
		}
		if it == 0 {
			lists, all = nil, nil
		}
		slices.Sort(all)
		if got := toSlice(mergeKLiteral(lists)); !slices.Equal(got, all) {
			t.Fatalf("got %v want %v", got, all)
		}
	}
}

// ===== hp-median-stream =====

// Literal (patched): push x into low; pop low's top into high; if size(high) > size(low)
// pop high's top into low. Median: equal sizes -> (low.top+high.top)/2, else low.top.
type medianLiteral struct {
	low, high *hp[int]
}

func newMedianLiteral() *medianLiteral {
	return &medianLiteral{
		low:  &hp[int]{less: func(a, b int) bool { return a > b }},
		high: &hp[int]{less: func(a, b int) bool { return a < b }},
	}
}

func (m *medianLiteral) addNum(x int) {
	heap.Push(m.low, x)
	heap.Push(m.high, heap.Pop(m.low))
	if m.high.Len() > m.low.Len() {
		heap.Push(m.low, heap.Pop(m.high))
	}
}

func (m *medianLiteral) findMedian() float64 {
	if m.low.Len() == m.high.Len() {
		return (float64(m.low.top()) + float64(m.high.top())) / 2
	}
	return float64(m.low.top())
}

func TestMedianStream(t *testing.T) {
	r := newRNG(22)
	for range 2000 {
		m := newMedianLiteral()
		var model []int
		ops := 1 + r.IntN(30)
		for range ops {
			x := r.IntN(21) - 10
			m.addNum(x)
			model = append(model, x)
			s := slices.Clone(model)
			slices.Sort(s)
			n := len(s)
			var want float64
			if n%2 == 1 {
				want = float64(s[n/2])
			} else {
				want = (float64(s[n/2-1]) + float64(s[n/2])) / 2
			}
			if got := m.findMedian(); got != want {
				t.Fatalf("stream %v: got %v want %v", model, got, want)
			}
		}
	}
}

// Patched key_idea alternative: quickselect on squared distance around a random pivot
// until the pivot lands at index k-1; return the first k points.
func kClosestQuickselect(pts []pt, k int, r *rand.Rand) []pt {
	a := slices.Clone(pts)
	lo, hi := 0, len(a)-1
	for lo < hi {
		pi := lo + r.IntN(hi-lo+1)
		a[pi], a[hi] = a[hi], a[pi]
		store := lo
		for i := lo; i < hi; i++ { // Lomuto partition
			if dist(a[i]) < dist(a[hi]) {
				a[i], a[store] = a[store], a[i]
				store++
			}
		}
		a[store], a[hi] = a[hi], a[store]
		switch {
		case store == k-1:
			return a[:k]
		case store < k-1:
			lo = store + 1
		default:
			hi = store - 1
		}
	}
	return a[:k]
}

func TestKClosestQuickselect(t *testing.T) {
	r := newRNG(23)
	for range 5000 {
		n := 1 + r.IntN(15)
		pts := make([]pt, n)
		for i := range pts {
			pts[i] = pt{r.IntN(9) - 4, r.IntN(9) - 4}
		}
		k := 1 + r.IntN(n)
		var gd, wd []int
		for _, p := range kClosestQuickselect(pts, k, r) {
			gd = append(gd, dist(p))
		}
		for _, p := range kClosestLiteral(pts, k) {
			wd = append(wd, dist(p))
		}
		slices.Sort(gd)
		slices.Sort(wd)
		if !slices.Equal(gd, wd) {
			t.Fatalf("%v k=%d: got dists %v want %v", pts, k, gd, wd)
		}
	}
}

// Alternative listed in the patched technique: divide-and-conquer pairwise merging.
func mergeKPairwise(lists []*ListNode) *ListNode {
	if len(lists) == 0 {
		return nil
	}
	for len(lists) > 1 {
		var next []*ListNode
		for i := 0; i < len(lists); i += 2 {
			if i+1 < len(lists) {
				next = append(next, mergeTwo(lists[i], lists[i+1]))
			} else {
				next = append(next, lists[i])
			}
		}
		lists = next
	}
	return lists[0]
}

func TestMergeKPairwise(t *testing.T) {
	r := newRNG(24)
	for range 5000 {
		var lists []*ListNode
		var all []int
		for range r.IntN(7) {
			a := randInts(r, r.IntN(6), -4, 4)
			slices.Sort(a)
			all = append(all, a...)
			lists = append(lists, fromSlice(a))
		}
		slices.Sort(all)
		if got := toSlice(mergeKPairwise(lists)); !slices.Equal(got, all) {
			t.Fatalf("got %v want %v", got, all)
		}
	}
}

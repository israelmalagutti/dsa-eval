package structures

import (
	"slices"
	"testing"
)

// ===== ll-reverse-k-group =====

// Literal (patched): dummy.next = head, groupPrev = dummy. Loop: kth = groupPrev moved k
// steps; null -> stop. nextGroup = kth.next, oldHead = groupPrev.next. prev = nextGroup,
// cur = oldHead; while cur != nextGroup: reverse one link. groupPrev.next = kth; groupPrev = oldHead.
func reverseKGroupLiteral(head *ListNode, k int) *ListNode {
	dummy := &ListNode{Next: head}
	groupPrev := dummy
	for {
		// check that k nodes exist
		kth := groupPrev
		for i := 0; i < k && kth != nil; i++ {
			kth = kth.Next
		}
		if kth == nil {
			break
		}
		nextGroup := kth.Next
		oldHead := groupPrev.Next
		// reverse with prev/cur/next
		var prev *ListNode = nextGroup
		cur := oldHead
		for cur != nextGroup {
			next := cur.Next
			cur.Next = prev
			prev = cur
			cur = next
		}
		groupPrev.Next = prev // new head (= kth)
		oldHead.Next = nextGroup
		groupPrev = oldHead
	}
	return dummy.Next
}

func reverseKGroupBrute(a []int, k int) []int {
	out := slices.Clone(a)
	for i := 0; i+k <= len(out); i += k {
		slices.Reverse(out[i : i+k])
	}
	return out
}

func TestReverseKGroup(t *testing.T) {
	r := newRNG(10)
	type tc struct {
		a []int
		k int
	}
	cases := []tc{{nil, 1}, {nil, 3}, {[]int{1}, 1}, {[]int{1}, 2}, {[]int{1, 2, 3, 4, 5}, 2}, {[]int{1, 2, 3, 4, 5}, 3}, {[]int{1, 2, 3, 4, 5}, 5}, {[]int{1, 2, 3, 4, 5}, 6}, {[]int{1, 2, 3, 4, 5, 6}, 3}}
	for range 5000 {
		cases = append(cases, tc{randInts(r, r.IntN(13), -5, 5), 1 + r.IntN(7)})
	}
	for _, c := range cases {
		// also confirm nodes are re-linked, not re-valued: collect node identities
		h := fromSlice(c.a)
		var orig []*ListNode
		for p := h; p != nil; p = p.Next {
			orig = append(orig, p)
		}
		res := reverseKGroupLiteral(h, c.k)
		if got, want := toSlice(res), reverseKGroupBrute(c.a, c.k); !slices.Equal(got, want) {
			t.Fatalf("%v k=%d: got %v want %v", c.a, c.k, got, want)
		}
		seen := map[*ListNode]bool{}
		for p := res; p != nil; p = p.Next {
			seen[p] = true
		}
		if len(seen) != len(orig) {
			t.Fatalf("node set changed")
		}
	}
}

// ===== ll-cycle-start =====

// Literal (patched): slow = fast = head; while fast and fast.next: step 1/2, break on meet;
// no meet -> null. slow = head; step both by 1 until equal; return slow.
func cycleStartLiteral(head *ListNode) *ListNode {
	slow, fast := head, head
	for {
		if fast == nil || fast.Next == nil {
			return nil
		}
		slow, fast = slow.Next, fast.Next.Next
		if slow == fast {
			break
		}
	}
	slow = head
	for slow != fast {
		slow, fast = slow.Next, fast.Next
	}
	return slow
}

func cycleStartBrute(head *ListNode) *ListNode {
	seen := map[*ListNode]bool{}
	for p := head; p != nil; p = p.Next {
		if seen[p] {
			return p
		}
		seen[p] = true
	}
	return nil
}

func TestCycleStart(t *testing.T) {
	r := newRNG(11)
	for n := 0; n <= 40; n++ {
		for pos := -1; pos < n; pos++ { // exhaustive small cases
			nodes := make([]*ListNode, n)
			for i := range nodes {
				nodes[i] = &ListNode{Val: r.IntN(3)} // duplicate values on purpose
				if i > 0 {
					nodes[i-1].Next = nodes[i]
				}
			}
			var head *ListNode
			if n > 0 {
				head = nodes[0]
				if pos >= 0 {
					nodes[n-1].Next = nodes[pos]
				}
			}
			if got, want := cycleStartLiteral(head), cycleStartBrute(head); got != want {
				t.Fatalf("n=%d pos=%d: wrong node", n, pos)
			}
		}
	}
}

// ===== ll-palindrome =====

// Literal (patched): slow = fast = head; while fast and fast.next: advance. Reverse from
// slow; p = head, q = reversed head; while q: mismatch -> false; advance. Optionally restore.
func isPalindromeLiteral(head *ListNode) bool {
	slow, fast := head, head
	for fast != nil && fast.Next != nil {
		slow, fast = slow.Next, fast.Next.Next
	}
	rev := reverseList(slow)
	ok := true
	for p, q := head, rev; q != nil; p, q = p.Next, q.Next {
		if p.Val != q.Val {
			ok = false
			break
		}
	}
	reverseList(rev) // restore
	return ok
}

func reverseList(h *ListNode) *ListNode {
	var prev *ListNode
	for h != nil {
		h.Next, prev, h = prev, h, h.Next
	}
	return prev
}

func TestPalindrome(t *testing.T) {
	r := newRNG(12)
	cases := [][]int{{}, {1}, {1, 1}, {1, 2}, {1, 2, 1}, {1, 2, 2, 1}, {1, 2, 3, 1}, {-1, 0, -1}}
	for range 5000 {
		a := randInts(r, r.IntN(10), 0, 2)
		if r.IntN(2) == 0 { // force palindromes often
			b := slices.Clone(a)
			slices.Reverse(b)
			if r.IntN(2) == 0 && len(a) > 0 {
				a = a[:len(a)-1]
			}
			a = append(a, b...)
		}
		cases = append(cases, a)
	}
	for _, c := range cases {
		rev := slices.Clone(c)
		slices.Reverse(rev)
		want := slices.Equal(c, rev)
		h := fromSlice(c)
		if got := isPalindromeLiteral(h); got != want {
			t.Fatalf("%v: got %v want %v", c, got, want)
		}
		if !slices.Equal(toSlice(h), c) {
			t.Fatalf("%v: list not restored", c)
		}
	}
}

// ===== ll-sort-list =====

// Literal (patched): split at the FIRST middle (slow = head, fast = head.next), sort halves
// recursively, merge with a dummy head taking the smaller val (left on ties).
// The original text did not say which middle; see TestSortListSecondMiddlePitfall.
func sortListLiteral(head *ListNode) *ListNode {
	if head == nil || head.Next == nil {
		return head
	}
	slow, fast := head, head.Next
	for fast != nil && fast.Next != nil {
		slow, fast = slow.Next, fast.Next.Next
	}
	right := slow.Next
	slow.Next = nil
	return mergeTwo(sortListLiteral(head), sortListLiteral(right))
}

func mergeTwo(a, b *ListNode) *ListNode {
	d := &ListNode{}
	t := d
	for a != nil && b != nil {
		if a.Val <= b.Val {
			t.Next, a = a, a.Next
		} else {
			t.Next, b = b, b.Next
		}
		t = t.Next
	}
	if a != nil {
		t.Next = a
	} else {
		t.Next = b
	}
	return d.Next
}

func TestSortList(t *testing.T) {
	r := newRNG(13)
	cases := [][]int{{}, {1}, {2, 1}, {1, 1, 1}, {4, 2, 1, 3}, {-1, 5, 3, 4, 0}, {5, 4, 3, 2, 1}}
	for range 5000 {
		cases = append(cases, randInts(r, r.IntN(20), -5, 5))
	}
	for _, c := range cases {
		want := slices.Clone(c)
		slices.Sort(want)
		if got := toSlice(sortListLiteral(fromSlice(c))); !slices.Equal(got, want) {
			t.Fatalf("%v: got %v want %v", c, got, want)
		}
	}
}

// Documents the ambiguity: with slow=fast=head and split after slow, a 2-node list
// gives left = both nodes, right = nil -> infinite recursion.
func TestSortListSecondMiddlePitfall(t *testing.T) {
	h := fromSlice([]int{2, 1})
	slow, fast := h, h
	for fast != nil && fast.Next != nil {
		slow, fast = slow.Next, fast.Next.Next
	}
	if slow.Next != nil {
		t.Fatalf("expected the second-middle split to leave right half empty for n=2")
	}
}

// Alternative listed in the patched technique: bottom-up merge sort, O(1) extra space.
func sortListBottomUp(head *ListNode) *ListNode {
	n := 0
	for p := head; p != nil; p = p.Next {
		n++
	}
	dummy := &ListNode{Next: head}
	for size := 1; size < n; size *= 2 {
		tail, cur := dummy, dummy.Next
		for cur != nil {
			left := cur
			right := splitAfter(left, size)
			cur = splitAfter(right, size)
			tail.Next = mergeTwo(left, right)
			for tail.Next != nil {
				tail = tail.Next
			}
		}
	}
	return dummy.Next
}

// splitAfter cuts the list after size nodes and returns the rest.
func splitAfter(h *ListNode, size int) *ListNode {
	for i := 1; h != nil && i < size; i++ {
		h = h.Next
	}
	if h == nil {
		return nil
	}
	rest := h.Next
	h.Next = nil
	return rest
}

func TestSortListBottomUp(t *testing.T) {
	r := newRNG(14)
	for range 5000 {
		c := randInts(r, r.IntN(20), -5, 5)
		want := slices.Clone(c)
		slices.Sort(want)
		if got := toSlice(sortListBottomUp(fromSlice(c))); !slices.Equal(got, want) {
			t.Fatalf("%v: got %v want %v", c, got, want)
		}
	}
}

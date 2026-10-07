package main

import (
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestBank(t *testing.T) {
	bank, err := loadBank()
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, q := range bank {
		if seen[q.ID] {
			t.Errorf("duplicate id %s", q.ID)
		}
		seen[q.ID] = true
		for name, v := range map[string]string{"category": q.Category, "topic": q.Topic, "prompt": q.Prompt,
			"technique": q.Technique, "signal": q.Signal, "complexity": q.Complexity, "key_idea": q.KeyIdea} {
			if v == "" {
				t.Errorf("%s: empty %s", q.ID, name)
			}
		}
	}
}

func TestExamsNeverRepeat(t *testing.T) {
	bank, _ := loadBank()
	rng := rand.New(rand.NewPCG(1, 2))
	for _, n := range []int{1, 6, 10, len(bank)} {
		var history []Snapshot
		seen := map[string]bool{}
		for range 30 {
			plan := pickQuestions(bank, history, n, rng)
			if len(plan) != n {
				t.Fatalf("n=%d: got %d questions", n, len(plan))
			}
			if len(slices.Compact(slices.Sorted(slices.Values(plan)))) != n {
				t.Fatalf("n=%d: duplicate question in plan", n)
			}
			key := strings.Join(plan, ",")
			if seen[key] {
				t.Fatalf("n=%d: repeated exam %s", n, key)
			}
			seen[key] = true
			history = append(history, Snapshot{Plan: plan})
		}
	}
}

func TestTierFor(t *testing.T) {
	cases := []struct{ grade, sec, want int }{{0, 10, 0}, {1, 10, 1}, {2, 180, 3}, {2, 181, 2}}
	for _, c := range cases {
		if got := tierFor(c.grade, c.sec); got != c.want {
			t.Errorf("tierFor(%d,%d)=%d want %d", c.grade, c.sec, got, c.want)
		}
	}
}

func TestNewID(t *testing.T) {
	now := time.Now()
	plan := []string{"a", "b"}
	id := newID(now, plan, nil)
	if len(id) != 6 || strings.Trim(id, "0123456789ABCDEF") != "" {
		t.Fatalf("bad id %q", id)
	}
	if again := newID(now, plan, []Snapshot{{ID: id}}); again == id {
		t.Fatalf("collision not re-salted: %q", again)
	}
}

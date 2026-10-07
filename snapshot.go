package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const fluentSeconds = 180

type Answer struct {
	Technique  string `json:"technique"`
	Signal     string `json:"signal"`
	Complexity string `json:"complexity"`
	KeyIdea    string `json:"key_idea"`
}

type Entry struct {
	QuestionID  string    `json:"question_id"`
	Category    string    `json:"category"`
	Topic       string    `json:"topic"`
	Prompt      string    `json:"prompt"`
	Answer      Answer    `json:"answer"`
	Skipped     bool      `json:"skipped"`
	StartedAt   time.Time `json:"started_at"`
	EndedAt     time.Time `json:"ended_at"`
	DurationSec int       `json:"duration_sec"`
	// Grades are 0 (wrong technique), 1 (technique + signal),
	// 2 (also complexity + key idea). Tiers add 3 (Fluent) for grade 2 within fluentSeconds.
	// Nil means not graded yet.
	SelfGrade  *int   `json:"self_grade"`
	SelfTier   *int   `json:"self_tier"`
	AIGrade    *int   `json:"ai_grade"`
	AITier     *int   `json:"ai_tier"`
	AIFeedback string `json:"ai_feedback"`
}

// Tier prefers the AI grade when one exists; ok is false for ungraded entries.
func (e Entry) Tier() (tier int, ok bool) {
	if e.AITier != nil {
		return *e.AITier, true
	}
	if e.SelfTier != nil {
		return *e.SelfTier, true
	}
	return 0, false
}

func tierFor(grade, durationSec int) int {
	if grade == 2 && durationSec <= fluentSeconds {
		return 3
	}
	return grade
}

type Snapshot struct {
	Version     int         `json:"version"`
	ID          string      `json:"id"`
	Mode        string      `json:"mode"`
	StartedAt   time.Time   `json:"started_at"`
	SubmittedAt *time.Time  `json:"submitted_at"` // answers locked, review phase begins
	EndedAt     *time.Time  `json:"ended_at"`
	ResumedAt   []time.Time `json:"resumed_at,omitempty"`
	Plan        []string    `json:"plan"`
	Entries     []Entry     `json:"entries"` // one per plan item, same order
}

func (s Snapshot) Finished() bool { return s.EndedAt != nil }

func (s Snapshot) Name() string { return "EXAM-" + s.ID }

func save(dir string, s Snapshot) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, s.ID+".json"), append(data, '\n'), 0o644)
}

// loadAll returns every snapshot in dir, oldest first.
func loadAll(dir string) ([]Snapshot, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, err
	}
	var out []Snapshot
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		var s Snapshot
		if err := json.Unmarshal(data, &s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].StartedAt.Before(out[j].StartedAt) })
	return out, nil
}

// topicTiers averages, per topic, the latest attempt of each question.
func topicTiers(history []Snapshot) map[string]float64 {
	type attempt struct {
		topic string
		at    time.Time
		tier  int
	}
	latest := map[string]attempt{}
	for _, s := range history {
		for _, e := range s.Entries {
			tier, graded := e.Tier()
			if !graded {
				continue
			}
			if a, ok := latest[e.QuestionID]; !ok || e.EndedAt.After(a.at) {
				latest[e.QuestionID] = attempt{e.Topic, e.EndedAt, tier}
			}
		}
	}
	sum := map[string]int{}
	count := map[string]int{}
	for _, a := range latest {
		sum[a.topic] += a.tier
		count[a.topic]++
	}
	out := map[string]float64{}
	for t, c := range count {
		out[t] = float64(sum[t]) / float64(c)
	}
	return out
}

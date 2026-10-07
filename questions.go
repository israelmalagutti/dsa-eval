package main

import (
	_ "embed"
	"encoding/json"
	"math/rand/v2"
	"slices"
	"strings"
)

//go:embed questions.json
var questionsJSON []byte

type Question struct {
	ID         string `json:"id"`
	Category   string `json:"category"`
	Topic      string `json:"topic"`
	Prompt     string `json:"prompt"`
	Technique  string `json:"technique"`
	Signal     string `json:"signal"`
	Complexity string `json:"complexity"`
	KeyIdea    string `json:"key_idea"`
}

func loadBank() ([]Question, error) {
	var bank []Question
	err := json.Unmarshal(questionsJSON, &bank)
	return bank, err
}

// pickQuestions draws n random questions, favoring topics that are
// unassessed or rated low, and never returns an exam identical to a past one.
func pickQuestions(bank []Question, history []Snapshot, n int, rng *rand.Rand) []string {
	n = min(n, len(bank))
	tiers := topicTiers(history)
	asked := map[string]bool{}
	prevSets := map[string]bool{}
	prevOrders := map[string]bool{}
	for _, s := range history {
		for _, e := range s.Entries {
			if !e.StartedAt.IsZero() {
				asked[e.QuestionID] = true
			}
		}
		prevOrders[strings.Join(s.Plan, ",")] = true
		sorted := slices.Sorted(slices.Values(s.Plan))
		prevSets[strings.Join(sorted, ",")] = true
	}

	var plan []string
	for range 500 {
		plan = sample(bank, tiers, asked, n, rng)
		sorted := slices.Sorted(slices.Values(plan))
		if !prevSets[strings.Join(sorted, ",")] {
			return plan
		}
	}
	// Every question set has been used (e.g. n == bank size): at least vary the order.
	for prevOrders[strings.Join(plan, ",")] {
		rng.Shuffle(len(plan), func(i, j int) { plan[i], plan[j] = plan[j], plan[i] })
	}
	return plan
}

func sample(bank []Question, tiers map[string]float64, asked map[string]bool, n int, rng *rand.Rand) []string {
	picked := make([]bool, len(bank))
	perTopic := map[string]int{}
	plan := make([]string, 0, n)
	for range n {
		weights := make([]float64, len(bank))
		total := 0.0
		for i, q := range bank {
			if picked[i] {
				continue
			}
			w := 5.0 // unassessed topic
			if t, ok := tiers[q.Topic]; ok {
				w = 4 - t // Unknown 4 ... Fluent 1
			}
			if !asked[q.ID] {
				w *= 2
			}
			k := float64(1 + perTopic[q.Topic])
			w /= k * k // spread the exam across topics
			weights[i] = w
			total += w
		}
		r := rng.Float64() * total
		i := 0
		for ; i < len(bank)-1; i++ {
			if picked[i] {
				continue
			}
			r -= weights[i]
			if r < 0 {
				break
			}
		}
		for picked[i] { // float rounding fallthrough
			i--
		}
		picked[i] = true
		perTopic[bank[i].Topic]++
		plan = append(plan, bank[i].ID)
	}
	return plan
}

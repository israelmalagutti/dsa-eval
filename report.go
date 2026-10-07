package main

import (
	"fmt"
	"math"
	"os"
	"slices"
	"sort"
	"text/tabwriter"
)

var tierNames = []string{"Unknown", "Recognize", "Can implement", "Fluent"}

func tierName(avg float64) string {
	return tierNames[int(math.Round(avg))]
}

func report(bank []Question, history []Snapshot) {
	tiers := topicTiers(history)

	unfinished, answered := 0, 0
	for _, s := range history {
		if !s.Finished() {
			unfinished++
		}
		for _, e := range s.Entries {
			if _, ok := e.Tier(); ok {
				answered++
			}
		}
	}
	fmt.Printf("Sessions: %d (%d unfinished)   Graded answers: %d\n\n", len(history), unfinished, answered)

	var categories []string
	topicsByCat := map[string][]string{}
	for _, q := range bank {
		if !slices.Contains(categories, q.Category) {
			categories = append(categories, q.Category)
		}
		if !slices.Contains(topicsByCat[q.Category], q.Topic) {
			topicsByCat[q.Category] = append(topicsByCat[q.Category], q.Topic)
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "CATEGORY\tCOVERAGE\tTIER\t")
	for _, c := range categories {
		sum, n := 0.0, 0
		for _, t := range topicsByCat[c] {
			if v, ok := tiers[t]; ok {
				sum += v
				n++
			}
		}
		tier := "-"
		if n > 0 {
			tier = fmt.Sprintf("%.1f %s", sum/float64(n), tierName(sum/float64(n)))
		}
		fmt.Fprintf(w, "%s\t%d/%d\t%s\t\n", c, n, len(topicsByCat[c]), tier)
	}
	w.Flush()

	fmt.Println("\nTOPICS")
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, c := range categories {
		fmt.Fprintf(w, "%s\t\t\n", c)
		for _, t := range topicsByCat[c] {
			tier := "-"
			if v, ok := tiers[t]; ok {
				tier = fmt.Sprintf("%.1f %s", v, tierName(v))
			}
			fmt.Fprintf(w, "  %s\t%s\t\n", t, tier)
		}
	}
	w.Flush()

	var assessed []string
	for t := range tiers {
		assessed = append(assessed, t)
	}
	sort.Slice(assessed, func(i, j int) bool {
		if tiers[assessed[i]] != tiers[assessed[j]] {
			return tiers[assessed[i]] < tiers[assessed[j]]
		}
		return assessed[i] < assessed[j]
	})
	if len(assessed) > 5 {
		assessed = assessed[:5]
	}
	fmt.Println("\nWEAKEST")
	for _, t := range assessed {
		fmt.Printf("  %.1f  %s\n", tiers[t], t)
	}
}

func list(history []Snapshot) {
	if len(history) == 0 {
		fmt.Println("No exams yet. Start one with: dsa-eval -quick")
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "EXAM\tSTARTED\tMODE\tSTATUS\t")
	for _, s := range slices.Backward(history) {
		done, status := 0, "answering"
		for _, e := range s.Entries {
			if s.SubmittedAt != nil && e.SelfGrade != nil || s.SubmittedAt == nil && (e.Answer.Technique != "" || e.Skipped) {
				done++
			}
		}
		switch {
		case s.Finished():
			status = "finished"
		case s.SubmittedAt != nil:
			status = fmt.Sprintf("review %d/%d graded", done, len(s.Plan))
		default:
			status = fmt.Sprintf("answering %d/%d", done, len(s.Plan))
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t\n", s.Name(), s.StartedAt.Format(dateFormat), s.Mode, status)
	}
	w.Flush()
}

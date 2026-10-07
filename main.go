// Command dsa-eval runs a pattern-recognition DSA exam and snapshots the session.
//
//	dsa-eval [run] [-quick | -n N]   start a new exam (default 10 questions)
//	dsa-eval resume [ID]             continue an unfinished exam (latest if no ID)
//	dsa-eval list                    all exams, newest first
//	dsa-eval report                  per-category and per-topic tiers
//	dsa-eval refs                    references map: topic dependency tree, CLRS sections, progress
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

const dateFormat = "02 / 01 / 2006 - 15:04:05"

// defaultDir is $XDG_DATA_HOME/dsa-eval/snapshots, falling back to ~/.local/share.
func defaultDir() string {
	data := os.Getenv("XDG_DATA_HOME")
	if data == "" {
		home, _ := os.UserHomeDir()
		data = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(data, "dsa-eval", "snapshots")
}

// newID returns the first 6 hex digits of a SHA-256 over the start time and plan,
// re-salted on the rare collision with an existing exam.
func newID(now time.Time, plan []string, history []Snapshot) string {
	for salt := 0; ; salt++ {
		sum := sha256.Sum256(fmt.Appendf(nil, "%d|%s|%d", now.UnixNano(), strings.Join(plan, ","), salt))
		id := strings.ToUpper(hex.EncodeToString(sum[:3]))
		if !slices.ContainsFunc(history, func(h Snapshot) bool { return h.ID == id }) {
			return id
		}
	}
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "dsa-eval:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	cmd := "run"
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		cmd, args = args[0], args[1:]
	}
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	dir := fs.String("dir", defaultDir(), "snapshot directory")
	quick := fs.Bool("quick", false, "quick exam: 6 questions (~15 min)")
	n := fs.Int("n", 10, "number of questions (10 is ~30 min)")
	fs.Parse(args)

	bank, err := loadBank()
	if err != nil {
		return err
	}
	history, err := loadAll(*dir)
	if err != nil {
		return err
	}

	switch cmd {
	case "run":
		mode := "standard"
		if *quick {
			mode, *n = "quick", 6
		} else if *n != 10 {
			mode = "custom"
		}
		if *n < 1 {
			return errors.New("-n must be at least 1")
		}
		now := time.Now()
		rng := rand.New(rand.NewPCG(uint64(now.UnixNano()), 0))
		plan := pickQuestions(bank, history, *n, rng)
		s := Snapshot{
			Version:   1,
			ID:        newID(now, plan, history),
			Mode:      mode,
			StartedAt: now,
			Plan:      plan,
			Entries:   []Entry{},
		}
		return exam(bank, *dir, s)
	case "resume":
		var s *Snapshot
		want := strings.TrimPrefix(strings.ToUpper(fs.Arg(0)), "EXAM-")
		for i := len(history) - 1; i >= 0; i-- {
			h := history[i]
			if (fs.NArg() == 0 && !h.Finished()) || strings.ToUpper(h.ID) == want {
				s = &history[i]
				break
			}
		}
		if s == nil {
			return errors.New("no unfinished exam found")
		}
		if s.Finished() {
			return fmt.Errorf("%s is already finished", s.Name())
		}
		s.ResumedAt = append(s.ResumedAt, time.Now())
		return exam(bank, *dir, *s)
	case "list":
		list(history)
		return nil
	case "report":
		report(bank, history)
		return nil
	case "refs":
		return browseRefs(history)
	}
	return fmt.Errorf("unknown command %q", cmd)
}

func exam(bank []Question, dir string, s Snapshot) error {
	byID := map[string]Question{}
	for _, q := range bank {
		byID[q.ID] = q
	}
	qs := make([]Question, len(s.Plan))
	for i, id := range s.Plan {
		q, ok := byID[id]
		if !ok {
			return fmt.Errorf("question %q no longer exists in the bank", id)
		}
		qs[i] = q
	}

	final, err := tea.NewProgram(newModel(dir, s, qs)).Run()
	if err != nil {
		return err
	}
	m := final.(*model)
	if m.err != nil {
		return m.err
	}
	if !m.s.Finished() {
		fmt.Printf("Saved %s. Resume with: dsa-eval resume %s\n", m.s.Name(), m.s.ID)
		return nil
	}
	sum := 0
	for _, e := range m.s.Entries {
		sum += *e.SelfTier
	}
	fmt.Printf("%s done at %s. Average self tier: %.1f. Snapshot: %s/%s.json\n",
		m.s.Name(), m.s.EndedAt.Format(dateFormat), float64(sum)/float64(len(m.s.Entries)), dir, m.s.ID)
	return nil
}

package main

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

func press(m *model, keys ...tea.KeyPressMsg) {
	for _, k := range keys {
		m.Update(k)
	}
}

func typeText(m *model, s string) {
	for _, r := range s {
		press(m, tea.KeyPressMsg{Code: r, Text: string(r)})
	}
}

var (
	shiftRight = tea.KeyPressMsg{Code: tea.KeyRight, Mod: tea.ModShift}
	shiftLeft  = tea.KeyPressMsg{Code: tea.KeyLeft, Mod: tea.ModShift}
	enter      = tea.KeyPressMsg{Code: tea.KeyEnter}
	ctrlS      = tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl}
)

func TestExamFlow(t *testing.T) {
	bank, _ := loadBank()
	qs := bank[:3]
	s := Snapshot{Version: 1, ID: "test", StartedAt: time.Now(), Plan: []string{qs[0].ID, qs[1].ID, qs[2].ID}}
	dir := t.TempDir()
	m := newModel(dir, s, qs)
	m.Init()
	m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	typeText(m, "bfs")
	press(m, enter)
	typeText(m, "unweighted")
	press(m, shiftRight)
	if m.cur != 1 {
		t.Fatalf("shift+right: cur=%d", m.cur)
	}
	if a := m.s.Entries[0].Answer; a.Technique != "bfs" || a.Signal != "unweighted" {
		t.Fatalf("answer not stored: %+v", a)
	}

	// Time accumulates across visits: 100s now, 100s more after coming back.
	press(m, shiftLeft)
	if m.fields[0].Value() != "bfs" {
		t.Fatalf("answer not restored: %q", m.fields[0].Value())
	}
	m.viewStart = m.viewStart.Add(-100 * time.Second)
	press(m, shiftRight, shiftLeft)
	m.viewStart = m.viewStart.Add(-100 * time.Second)
	press(m, shiftRight)
	if d := m.s.Entries[0].DurationSec; d < 200 {
		t.Fatalf("duration not accumulated: %d", d)
	}

	typeText(m, ":s")
	press(m, enter)
	if !m.s.Entries[1].Skipped || m.cur != 2 {
		t.Fatalf(":s: skipped=%v cur=%d", m.s.Entries[1].Skipped, m.cur)
	}

	press(m, ctrlS)
	if m.reviewing() {
		t.Fatal("submitted without confirming unanswered questions")
	}
	press(m, ctrlS)
	if !m.reviewing() || m.cur != 0 {
		t.Fatalf("submit: reviewing=%v cur=%d", m.reviewing(), m.cur)
	}
	if g := m.s.Entries[2].SelfGrade; g == nil || *g != 0 {
		t.Fatal("unanswered question not auto-graded 0")
	}

	press(m, tea.KeyPressMsg{Code: '2', Text: "2"})
	if e := m.s.Entries[0]; *e.SelfGrade != 2 || *e.SelfTier != 2 {
		t.Fatalf("grade 2 over 3 min should be tier 2, got grade=%d tier=%d", *e.SelfGrade, *e.SelfTier)
	}
	press(m, enter)

	saved, err := loadAll(dir)
	if err != nil || len(saved) != 1 || !saved[0].Finished() {
		t.Fatalf("snapshot not finished on disk: %v %+v", err, saved)
	}
}

package main

import (
	"slices"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// a, b roots; c <- a; d <- a, b; e <- c, d; f <- a, e (longest chain a-c-e-f).
const refsFixture = `{
  "book": {"title": "Test Book"},
  "nodes": [
    {"id": "f", "title": "Node F", "prereqs": ["a", "e"], "summary": "", "topics": []},
    {"id": "a", "title": "Node A", "prereqs": [], "summary": "", "topics": []},
    {"id": "d", "title": "Node D", "prereqs": ["a", "b"], "summary": "", "topics": []},
    {"id": "b", "title": "Node B", "prereqs": [], "summary": "", "topics": []},
    {"id": "c", "title": "Node C", "prereqs": ["a"], "summary": "Sum C.", "topics": [
      {"name": "Hashing", "bank_topic": "Hash map lookup", "summary": "Hash it.", "clrs": [
        {"ref": "11.2", "title": "Hash tables", "book_page": 253, "pdf_page": 274, "note": "Chaining."},
        {"ref": "11.3", "title": "Hash functions", "book_page": 262, "pdf_page": 283, "note": ""}
      ], "external": [{"title": "Notes", "url": "https://example.com"}]},
      {"name": "Reading", "bank_topic": "", "summary": "", "clrs": [
        {"ref": "11.4", "title": "Open addressing", "book_page": 269, "pdf_page": 290, "note": ""}
      ], "external": []}
    ]},
    {"id": "e", "title": "Node E", "prereqs": ["c", "d"], "summary": "", "topics": []}
  ]
}`

func fixtureRefs(t *testing.T) Refs {
	t.Helper()
	r, err := parseRefs([]byte(refsFixture))
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func ids(r Refs, nodes []int) []string {
	var out []string
	for _, i := range nodes {
		out = append(out, r.Nodes[i].ID)
	}
	return out
}

func TestLayers(t *testing.T) {
	r := fixtureRefs(t)
	var got []string
	for _, l := range r.layers {
		got = append(got, strings.Join(ids(r, l), ","))
	}
	if want := []string{"a,b", "c,d", "e", "f"}; !slices.Equal(got, want) {
		t.Fatalf("layers %v want %v", got, want)
	}

	for name, js := range map[string]string{
		"cycle":   `{"nodes": [{"id": "x", "prereqs": ["y"]}, {"id": "y", "prereqs": ["x"]}]}`,
		"unknown": `{"nodes": [{"id": "x", "prereqs": ["nope"]}]}`,
		"dup":     `{"nodes": [{"id": "x"}, {"id": "x"}]}`,
	} {
		if _, err := parseRefs([]byte(js)); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
}

func TestEmbeddedRefs(t *testing.T) {
	r, err := loadRefs() // also checks that prereqs exist and the graph is acyclic
	if err != nil {
		t.Fatal(err)
	}
	bank, _ := loadBank()
	for _, n := range r.Nodes {
		for _, tp := range n.Topics {
			if tp.BankTopic != "" && !slices.ContainsFunc(bank, func(q Question) bool { return q.Topic == tp.BankTopic }) {
				t.Errorf("%s: bank_topic %q is not a topic in questions.json", n.ID, tp.BankTopic)
			}
		}
	}
}

func TestNodeFor(t *testing.T) {
	r := fixtureRefs(t)
	if i := r.nodeFor("Hash map lookup"); i < 0 || r.Nodes[i].ID != "c" {
		t.Fatalf("nodeFor: %d", i)
	}
	if i := r.nodeFor("Nope"); i != -1 {
		t.Fatalf("nodeFor unknown topic: %d", i)
	}
}

func TestPDFURL(t *testing.T) {
	got := pdfURL("/home/u/My Books/Intro, 3rd Ed.pdf", 274)
	if want := "file:///home/u/My%20Books/Intro%2C%203rd%20Ed.pdf#page=274"; got != want {
		t.Fatalf("pdfURL=%s want %s", got, want)
	}
}

func pressRefs(m *refsModel, keys ...tea.KeyPressMsg) {
	for _, k := range keys {
		m.Update(k)
	}
}

var (
	left     = tea.KeyPressMsg{Code: tea.KeyLeft}
	right    = tea.KeyPressMsg{Code: tea.KeyRight}
	up       = tea.KeyPressMsg{Code: tea.KeyUp}
	down     = tea.KeyPressMsg{Code: tea.KeyDown}
	esc      = tea.KeyPressMsg{Code: tea.KeyEscape}
	tab      = tea.KeyPressMsg{Code: tea.KeyTab}
	shiftTab = tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
)

func selected(m *refsModel) string { return m.refs.Nodes[m.sel].ID }

func TestRefsMapNavigation(t *testing.T) {
	m := newRefsModel(fixtureRefs(t), map[string]float64{"Hash map lookup": 1.5})
	m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	steps := []struct {
		key  tea.KeyPressMsg
		want string
	}{
		{left, "a"}, {right, "b"}, {right, "b"}, {down, "d"}, {down, "e"}, {down, "f"}, {down, "f"},
		{up, "e"}, {up, "c"}, {up, "a"}, {up, "a"},
	}
	for i, s := range steps {
		pressRefs(m, s.key)
		if got := selected(m); got != s.want {
			t.Fatalf("step %d (%s): selected %s want %s", i, s.key, got, s.want)
		}
	}
	pressRefs(m, down)
	if v := m.View().Content; !strings.Contains(v, "Node C") || !strings.Contains(v, "1/1 assessed · 1.5") {
		t.Fatalf("map view missing box or progress:\n%s", v)
	}

	pressRefs(m, enter)
	if !m.detail {
		t.Fatal("enter did not open details")
	}
	v := m.View().Content
	for _, want := range []string{"Node C", "Requires:", "Node A", "Unlocks:", "Node E", "tier 1.5 Can implement", "▸ CLRS 11.2 Hash tables — p.253", "https://example.com"} {
		if !strings.Contains(v, want) {
			t.Fatalf("detail view missing %q:\n%s", want, v)
		}
	}
	pressRefs(m, tab, tab)
	if m.refCursor != 2 || !strings.Contains(m.View().Content, "▸ CLRS 11.4") {
		t.Fatalf("tab: cursor %d", m.refCursor)
	}
	pressRefs(m, tab)
	if m.refCursor != 0 {
		t.Fatalf("tab wraps around: cursor %d", m.refCursor)
	}
	pressRefs(m, shiftTab)
	if m.refCursor != 2 {
		t.Fatalf("tab wraps around: cursor %d", m.refCursor)
	}
	pressRefs(m, esc)
	if m.detail || m.done {
		t.Fatal("esc should return to the map")
	}

	// Narrow terminal: one box per row, so a layer wraps and down walks within it.
	m.Update(tea.WindowSizeMsg{Width: 30, Height: 20})
	pressRefs(m, up, up, up, up, up)
	if got := selected(m); got != "a" {
		t.Fatalf("narrow: top is %s", got)
	}
	pressRefs(m, down)
	if got := selected(m); got != "b" {
		t.Fatalf("narrow: down from a went to %s, want b", got)
	}
	for _, line := range strings.Split(m.View().Content, "\n") {
		if lipgloss.Width(line) > 30 {
			t.Fatalf("line wider than terminal: %q", line)
		}
	}
	if h := strings.Count(m.View().Content, "\n") + 1; h > 20 {
		t.Fatalf("view taller than terminal: %d lines", h)
	}
}

func TestReviewOpensRefs(t *testing.T) {
	saved := referencesJSON
	referencesJSON = []byte(refsFixture)
	t.Cleanup(func() { referencesJSON = saved })

	qs := []Question{
		{ID: "q1", Topic: "Other topic", Prompt: "P1"},
		{ID: "q2", Topic: "Hash map lookup", Prompt: "P2"},
	}
	s := Snapshot{Version: 1, ID: "test", StartedAt: time.Now(), Plan: []string{"q1", "q2"}}
	m := newModel(t.TempDir(), s, qs)
	m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	press(m, shiftRight)
	typeText(m, "r")
	if m.refs != nil {
		t.Fatal("r opened references while answering")
	}
	press(m, ctrlS, ctrlS, shiftRight)
	if !m.reviewing() || m.cur != 1 {
		t.Fatalf("not reviewing question 2: reviewing=%v cur=%d", m.reviewing(), m.cur)
	}

	press(m, tea.KeyPressMsg{Code: 'r', Text: "r"})
	if m.refs == nil || !m.refs.detail || selected(m.refs) != "c" {
		t.Fatalf("r did not open the detail of node c: %+v", m.refs)
	}
	if v := m.View().Content; !strings.Contains(v, "Hash tables") {
		t.Fatalf("exam view does not show references:\n%s", v)
	}
	press(m, esc)
	if m.refs != nil || m.cur != 1 || !m.reviewing() {
		t.Fatalf("esc did not return to the review: refs=%v cur=%d", m.refs != nil, m.cur)
	}
	if v := m.View().Content; !strings.Contains(v, "r references") {
		t.Fatalf("review footer missing r:\n%s", v)
	}

	// Via the map: backspace leaves the detail, esc on the map goes back to the review.
	press(m, tea.KeyPressMsg{Code: 'r', Text: "r"}, tea.KeyPressMsg{Code: tea.KeyBackspace})
	if m.refs == nil || m.refs.detail {
		t.Fatal("backspace should show the map")
	}
	press(m, enter, esc)
	if m.refs == nil || m.refs.detail {
		t.Fatal("esc from a detail opened on the map should return to the map")
	}
	press(m, esc)
	if m.refs != nil || m.cur != 1 {
		t.Fatal("esc on the map did not return to the review")
	}
}

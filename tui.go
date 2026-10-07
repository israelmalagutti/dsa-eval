package main

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const helpText = `Technique   The algorithm or pattern that solves it (e.g. "sliding window").
Signal      The clue in the problem statement that points to the technique, e.g.
              "sorted array" + "O(log n)"            -> binary search
              "longest contiguous substring with X"  -> sliding window
              "subarray sum = k, negatives allowed"  -> prefix sum + hashmap
              "next greater element"                 -> monotonic stack
              "shortest path, unweighted"            -> BFS
Complexity  Time and space, e.g. "O(n) time, O(1) space".
Key idea    2-4 lines of the invariant or core loop, pseudocode or any language.

Type these in a one-line field and press enter:
  :h  toggle this help    :s  skip question    :q  save and quit (resume later)

Esc closes this help.`

var (
	dimStyle    = lipgloss.NewStyle().Faint(true)
	boldStyle   = lipgloss.NewStyle().Bold(true)
	accentStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	warnStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
)

var fieldLabels = [4]string{"Technique", "Signal", "Complexity", "Key idea"}

const keyIdeaField = 3

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

type model struct {
	dir string
	s   Snapshot
	qs  []Question // aligned with s.Plan

	cur       int
	fields    [4]textarea.Model
	focus     int
	spent     []time.Duration // on-screen time per question, accumulated across visits
	viewStart time.Time

	width, height int
	help          bool
	confirmSubmit bool
	status        string
	err           error

	refs *refsModel // references screen opened from review, nil when closed
}

func newModel(dir string, s Snapshot, qs []Question) *model {
	// Pad snapshots that predate one-entry-per-question.
	for i := len(s.Entries); i < len(s.Plan); i++ {
		q := qs[i]
		s.Entries = append(s.Entries, Entry{QuestionID: q.ID, Category: q.Category, Topic: q.Topic, Prompt: q.Prompt})
	}
	m := &model{dir: dir, s: s, qs: qs, width: 80, height: 24, spent: make([]time.Duration, len(s.Plan))}
	for i, e := range s.Entries {
		m.spent[i] = time.Duration(e.DurationSec) * time.Second
	}
	for i := range m.fields {
		ta := textarea.New()
		ta.Prompt = ""
		ta.ShowLineNumbers = false
		ta.CharLimit = 0
		if i == keyIdeaField {
			ta.SetHeight(6)
		} else {
			ta.SetHeight(1)
			ta.KeyMap.InsertNewline.SetEnabled(false)
		}
		m.fields[i] = ta
	}
	if s.SubmittedAt == nil {
		m.enter(0)
	}
	return m
}

func (m *model) reviewing() bool { return m.s.SubmittedAt != nil }

func (m *model) Init() tea.Cmd {
	if m.reviewing() {
		return tick()
	}
	return tea.Batch(tick(), m.setFocus(0))
}

// enter shows question i and starts its clock.
func (m *model) enter(i int) {
	m.cur = i
	m.confirmSubmit = false
	if m.reviewing() {
		return
	}
	now := time.Now()
	m.viewStart = now
	e := &m.s.Entries[i]
	if e.StartedAt.IsZero() {
		e.StartedAt = now
	}
	a := e.Answer
	for f, v := range []string{a.Technique, a.Signal, a.Complexity, a.KeyIdea} {
		m.fields[f].SetValue(v)
	}
}

// leave stores the current answer and stops its clock.
func (m *model) leave() {
	if m.reviewing() {
		return
	}
	now := time.Now()
	e := &m.s.Entries[m.cur]
	m.spent[m.cur] += now.Sub(m.viewStart)
	m.viewStart = now
	e.EndedAt = now
	e.DurationSec = int(m.spent[m.cur].Seconds())
	e.Answer = Answer{
		Technique:  strings.TrimSpace(m.fields[0].Value()),
		Signal:     strings.TrimSpace(m.fields[1].Value()),
		Complexity: strings.TrimSpace(m.fields[2].Value()),
		KeyIdea:    strings.TrimSpace(m.fields[keyIdeaField].Value()),
	}
	if e.Answer.Technique != "" {
		e.Skipped = false
	}
}

func (m *model) save() {
	m.leave()
	if err := save(m.dir, m.s); err != nil {
		m.err = err
	}
}

func (m *model) goTo(i int) tea.Cmd {
	if i < 0 || i >= len(m.s.Plan) || i == m.cur {
		return nil
	}
	m.save()
	m.enter(i)
	m.status = ""
	if m.reviewing() {
		return nil
	}
	return m.setFocus(0)
}

func (m *model) setFocus(f int) tea.Cmd {
	m.focus = f
	for i := range m.fields {
		m.fields[i].Blur()
	}
	return m.fields[f].Focus()
}

func (m *model) quit() (tea.Model, tea.Cmd) {
	m.save()
	return m, tea.Quit
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.WindowSizeMsg, tickMsg:
	default:
		if m.refs != nil {
			return m.updateRefs(msg)
		}
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if m.refs != nil {
			m.refs.Update(msg)
		}
		for i := range m.fields {
			m.fields[i].SetWidth(max(20, msg.Width-14))
		}
		m.fields[keyIdeaField].SetWidth(max(20, msg.Width-2))
		return m, nil
	case tickMsg:
		return m, tick()
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m.quit()
		case "shift+left":
			return m, m.goTo(m.cur - 1)
		case "shift+right":
			return m, m.goTo(m.cur + 1)
		}
		if m.reviewing() {
			return m.updateReview(msg)
		}
		return m.updateAnswer(msg)
	}
	if m.reviewing() {
		return m, nil
	}
	var cmd tea.Cmd
	m.fields[m.focus], cmd = m.fields[m.focus].Update(msg)
	return m, cmd
}

func (m *model) updateAnswer(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.help = false
		return m, nil
	case "tab":
		return m, m.setFocus((m.focus + 1) % len(m.fields))
	case "shift+tab":
		return m, m.setFocus((m.focus + len(m.fields) - 1) % len(m.fields))
	case "ctrl+s":
		return m.submit()
	case "enter":
		if m.focus == keyIdeaField {
			break
		}
		switch strings.TrimSpace(m.fields[m.focus].Value()) {
		case ":h":
			m.fields[m.focus].SetValue("")
			m.help = !m.help
			return m, nil
		case ":q":
			m.fields[m.focus].SetValue("")
			return m.quit()
		case ":s":
			m.fields[m.focus].SetValue("")
			m.s.Entries[m.cur].Skipped = true
			if m.cur == len(m.s.Plan)-1 {
				m.save()
				m.status = "Skipped. That was the last question: ctrl+s to submit."
				return m, nil
			}
			return m, m.goTo(m.cur + 1)
		}
		return m, m.setFocus(m.focus + 1)
	}
	m.confirmSubmit = false
	var cmd tea.Cmd
	m.fields[m.focus], cmd = m.fields[m.focus].Update(msg)
	return m, cmd
}

func (m *model) submit() (tea.Model, tea.Cmd) {
	m.leave()
	unanswered := 0
	for _, e := range m.s.Entries {
		if e.Answer.Technique == "" {
			unanswered++
		}
	}
	if unanswered > 0 && !m.confirmSubmit {
		m.confirmSubmit = true
		m.status = fmt.Sprintf("%d unanswered (counted as skipped). Press ctrl+s again to submit.", unanswered)
		return m, nil
	}
	now := time.Now()
	m.s.SubmittedAt = &now
	zero := 0
	for i := range m.s.Entries {
		e := &m.s.Entries[i]
		if e.Answer.Technique == "" {
			e.Skipped = true
			e.SelfGrade, e.SelfTier = &zero, &zero
		}
	}
	m.save()
	m.help = false
	m.enter(0)
	m.status = "Answers locked. Compare with the reference and grade each one."
	return m, nil
}

func (m *model) updateReview(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch k := msg.String(); k {
	case "q":
		return m.quit()
	case "r":
		m.openRefs()
	case "0", "1", "2":
		e := &m.s.Entries[m.cur]
		if e.Skipped {
			return m, nil
		}
		g := int(k[0] - '0')
		t := tierFor(g, e.DurationSec)
		e.SelfGrade, e.SelfTier = &g, &t
		m.save()
		if next := m.nextUngraded(); next >= 0 {
			m.enter(next)
		} else {
			m.status = "All graded. Press enter to finish."
		}
	case "enter", "ctrl+s":
		if next := m.nextUngraded(); next >= 0 {
			m.enter(next)
			m.status = "Grade every question before finishing."
			return m, nil
		}
		now := time.Now()
		m.s.EndedAt = &now
		return m.quit()
	}
	return m, nil
}

func (m *model) nextUngraded() int {
	n := len(m.s.Entries)
	for k := 1; k <= n; k++ {
		if i := (m.cur + k) % n; m.s.Entries[i].SelfGrade == nil {
			return i
		}
	}
	return -1
}

func (m *model) View() tea.View {
	if m.refs != nil {
		return m.refs.View()
	}
	var b strings.Builder
	w := max(40, m.width)
	wrap := lipgloss.NewStyle().Width(w - 2)

	phase := "Answering"
	if m.reviewing() {
		phase = "Review"
	}
	fmt.Fprintf(&b, "%s\n%s\n", boldStyle.Render(m.s.Name()+" · "+strings.ToUpper(m.s.StartedAt.Format("02 / Jan / 2006"))),
		accentStyle.Render(fmt.Sprintf("%s · Question %d/%d", phase, m.cur+1, len(m.s.Plan))))
	b.WriteString(m.progress() + "\n\n")

	q := m.qs[m.cur]
	e := m.s.Entries[m.cur]
	if m.help {
		b.WriteString(helpText + "\n")
	} else if !m.reviewing() {
		elapsed := m.spent[m.cur] + time.Since(m.viewStart)
		b.WriteString(wrap.Render(q.Prompt) + "\n\n")
		for i := range keyIdeaField {
			fmt.Fprintf(&b, "%-12s%s\n", m.label(i), m.fields[i].View())
		}
		fmt.Fprintf(&b, "%s\n%s\n\n", m.label(keyIdeaField), m.fields[keyIdeaField].View())
		b.WriteString(dimStyle.Render("time on this question: "+clock(elapsed)) + "\n")
	} else {
		fmt.Fprintf(&b, "%s\n%s\n\n", dimStyle.Render(fmt.Sprintf("%s › %s · your time %s", q.Category, q.Topic, clock(m.spent[m.cur]))), wrap.Render(q.Prompt))
		yours := e.Answer
		if e.Skipped {
			yours = Answer{Technique: "(skipped)"}
		}
		b.WriteString(boldStyle.Render("YOUR ANSWER") + "\n" + answerBlock(yours, wrap) + "\n")
		b.WriteString(boldStyle.Render("REFERENCE") + "\n" +
			answerBlock(Answer{q.Technique, q.Signal, q.Complexity, q.KeyIdea}, wrap) + "\n")
		grade := "not graded"
		if e.SelfGrade != nil {
			grade = fmt.Sprintf("%d → tier %d %s", *e.SelfGrade, *e.SelfTier, tierNames[*e.SelfTier])
		}
		fmt.Fprintf(&b, "Grade: %s\n%s\n", accentStyle.Render(grade),
			dimStyle.Render("0 wrong technique · 1 technique + signal · 2 also complexity + key idea"))
	}

	if m.status != "" {
		b.WriteString("\n" + warnStyle.Render(m.status) + "\n")
	}
	if m.err != nil {
		b.WriteString("\n" + warnStyle.Render("save failed: "+m.err.Error()) + "\n")
	}
	keys := "shift+←/→ question · tab field · enter next field · ctrl+s submit · :h help · :s skip · :q quit"
	if m.reviewing() {
		keys = "shift+←/→ question · 0/1/2 grade · r references · enter finish · q quit (resume later)"
	}
	b.WriteString("\n" + dimStyle.Render(keys))

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}

func (m *model) label(i int) string {
	if i == m.focus {
		return boldStyle.Render(fieldLabels[i])
	}
	return dimStyle.Render(fieldLabels[i])
}

// progress renders one marker per question: ● answered/graded, ○ open, – skipped.
func (m *model) progress() string {
	var parts []string
	for i, e := range m.s.Entries {
		mark := "○"
		switch {
		case e.Skipped:
			mark = "–"
		case m.reviewing() && e.SelfGrade != nil, !m.reviewing() && e.Answer.Technique != "":
			mark = "●"
		}
		if i == m.cur {
			mark = accentStyle.Render("[" + mark + "]")
		} else {
			mark = " " + mark + " "
		}
		parts = append(parts, mark)
	}
	return strings.Join(parts, "")
}

func answerBlock(a Answer, wrap lipgloss.Style) string {
	var b strings.Builder
	for i, v := range []string{a.Technique, a.Signal, a.Complexity, a.KeyIdea} {
		if v == "" {
			v = "-"
		}
		fmt.Fprintf(&b, "%s\n", wrap.Render(fmt.Sprintf("  %-11s %s", fieldLabels[i]+":", strings.ReplaceAll(v, "\n", "\n              "))))
	}
	return b.String()
}

func clock(d time.Duration) string {
	s := int(d.Seconds())
	return fmt.Sprintf("%02d:%02d", s/60, s%60)
}

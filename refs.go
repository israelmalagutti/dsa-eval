package main

import (
	"cmp"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

//go:embed references.json
var referencesJSON []byte

type Refs struct {
	Book struct {
		Title string `json:"title"`
	} `json:"book"`
	Nodes []RefNode `json:"nodes"`

	index      map[string]int // node id -> position in Nodes
	dependents [][]int        // nodes that list Nodes[i] as a prereq
	layers     [][]int        // node positions per layer, roots first
}

type RefNode struct {
	ID      string     `json:"id"`
	Title   string     `json:"title"`
	Prereqs []string   `json:"prereqs"`
	Summary string     `json:"summary"`
	Topics  []RefTopic `json:"topics"`
}

type RefTopic struct {
	Name      string    `json:"name"`
	BankTopic string    `json:"bank_topic"` // matches Question.Topic, or "" for reading-only topics
	Summary   string    `json:"summary"`
	CLRS      []CLRSRef `json:"clrs"`
	External  []Link    `json:"external"`
}

type CLRSRef struct {
	Ref      string `json:"ref"`
	Title    string `json:"title"`
	BookPage int    `json:"book_page"`
	PDFPage  int    `json:"pdf_page"` // 1-based page in the local PDF
	Note     string `json:"note"`
}

type Link struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

func loadRefs() (Refs, error) {
	return parseRefs(referencesJSON)
}

func parseRefs(data []byte) (Refs, error) {
	var r Refs
	if err := json.Unmarshal(data, &r); err != nil {
		return r, err
	}
	if len(r.Nodes) == 0 {
		return r, errors.New("no nodes")
	}
	r.index = map[string]int{}
	for i, n := range r.Nodes {
		if _, dup := r.index[n.ID]; dup {
			return r, fmt.Errorf("duplicate node id %q", n.ID)
		}
		r.index[n.ID] = i
	}
	r.dependents = make([][]int, len(r.Nodes))
	for i, n := range r.Nodes {
		for _, p := range n.Prereqs {
			if j, ok := r.index[p]; ok {
				r.dependents[j] = append(r.dependents[j], i)
			}
		}
	}
	var err error
	r.layers, err = layerNodes(r.Nodes, r.index)
	return r, err
}

// layerNodes puts each node on the layer given by its longest prereq chain from a root.
// It fails on unknown prereqs and cycles.
func layerNodes(nodes []RefNode, index map[string]int) ([][]int, error) {
	depth := make([]int, len(nodes))
	state := make([]int, len(nodes)) // 0 unvisited, 1 on the DFS stack, 2 done
	var visit func(i int) error
	visit = func(i int) error {
		switch state[i] {
		case 1:
			return fmt.Errorf("prereq cycle through %q", nodes[i].ID)
		case 2:
			return nil
		}
		state[i] = 1
		for _, p := range nodes[i].Prereqs {
			j, ok := index[p]
			if !ok {
				return fmt.Errorf("%s: unknown prereq %q", nodes[i].ID, p)
			}
			if err := visit(j); err != nil {
				return err
			}
			depth[i] = max(depth[i], depth[j]+1)
		}
		state[i] = 2
		return nil
	}
	var layers [][]int
	for i := range nodes {
		if err := visit(i); err != nil {
			return nil, err
		}
		for len(layers) <= depth[i] {
			layers = append(layers, nil)
		}
	}
	for i, d := range depth {
		layers[d] = append(layers[d], i)
	}
	// Order each layer by where its prereqs sit, so related boxes stay close.
	pos := make([]float64, len(nodes))
	for _, l := range layers {
		key := func(i int) float64 {
			sum := 0.0
			for _, p := range nodes[i].Prereqs {
				sum += pos[index[p]]
			}
			return sum / float64(max(1, len(nodes[i].Prereqs)))
		}
		slices.SortStableFunc(l, func(a, b int) int { return cmp.Compare(key(a), key(b)) })
		for c, i := range l {
			pos[i] = (float64(c) + 0.5) / float64(len(l))
		}
	}
	return layers, nil
}

// nodeFor returns the node with a topic for the given bank topic, or -1.
func (r Refs) nodeFor(bankTopic string) int {
	for i, n := range r.Nodes {
		if slices.ContainsFunc(n.Topics, func(t RefTopic) bool { return t.BankTopic == bankTopic }) {
			return i
		}
	}
	return -1
}

func (r Refs) titles(nodes []int) string {
	if len(nodes) == 0 {
		return dimStyle.Render("nothing")
	}
	var out []string
	for _, i := range nodes {
		out = append(out, r.Nodes[i].Title)
	}
	return strings.Join(out, ", ")
}

func (r Refs) prereqs(i int) []int {
	var out []int
	for _, p := range r.Nodes[i].Prereqs {
		out = append(out, r.index[p])
	}
	return out
}

const (
	boxWidth = 26
	boxGap   = 2
)

var (
	tierColors     = []string{"1", "3", "2", "10"} // Unknown red ... Fluent bright green
	selectedColor  = lipgloss.Color("6")
	requiresColor  = lipgloss.Color("5")
	unlocksColor   = lipgloss.Color("4")
	boxBorderColor = lipgloss.Color("8")
)

func tierStyle(avg float64) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(tierColors[int(avg+0.5)]))
}

type openedMsg string

type refsModel struct {
	refs  Refs
	tiers map[string]float64 // from topicTiers

	sel         int // selected node
	detail      bool
	vp          viewport.Model
	refCursor   int  // selected CLRS ref within the node, in reading order
	embedded    bool // opened from an exam review: esc and q go back to it
	startDetail bool // the detail view the embedded screen opened on
	done        bool // embedded screen asks to close

	width, height int
	status        string
}

func newRefsModel(refs Refs, tiers map[string]float64) *refsModel {
	m := &refsModel{refs: refs, tiers: tiers, sel: refs.layers[0][0], vp: viewport.New()}
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	return m
}

func browseRefs(history []Snapshot) error {
	refs, err := loadRefs()
	if err != nil {
		return fmt.Errorf("references.json: %w", err)
	}
	_, err = tea.NewProgram(newRefsModel(refs, topicTiers(history))).Run()
	return err
}

func (m *refsModel) Init() tea.Cmd { return nil }

func (m *refsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.vp.SetWidth(msg.Width)
		m.vp.SetHeight(max(1, msg.Height-3))
		if m.detail {
			content, _ := m.detailContent()
			m.vp.SetContent(content)
		}
		return m, nil
	case openedMsg:
		m.status = string(msg)
		return m, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.embedded {
				m.done = true
				return m, nil
			}
			return m, tea.Quit
		}
		if m.detail {
			return m.updateDetail(msg)
		}
		return m.updateMap(msg)
	}
	return m, nil
}

func (m *refsModel) updateMap(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	m.status = ""
	switch msg.String() {
	case "esc":
		m.done = m.embedded
	case "left":
		m.moveCol(-1)
	case "right":
		m.moveCol(1)
	case "up":
		m.moveRow(-1)
	case "down":
		m.moveRow(1)
	case "enter":
		m.openDetail(m.sel, false)
	}
	return m, nil
}

func (m *refsModel) updateDetail(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		if m.startDetail {
			m.done = true
			return m, nil
		}
		m.detail = false
	case "backspace":
		m.detail, m.startDetail = false, false
	case "tab":
		m.moveRef(1)
	case "shift+tab":
		m.moveRef(-1)
	case "o":
		refs := m.nodeRefs()
		if len(refs) == 0 {
			m.status = "This node has no CLRS refs."
			return m, nil
		}
		return m, openRef(refs[m.refCursor])
	default:
		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(msg)
		return m, cmd
	}
	return m, nil
}

// openDetail shows node i; start marks the view an embedded screen opened on.
func (m *refsModel) openDetail(i int, start bool) {
	m.sel, m.detail, m.startDetail = i, true, start
	m.refCursor, m.status = 0, ""
	content, _ := m.detailContent()
	m.vp.SetContent(content)
	m.vp.GotoTop()
}

func (m *refsModel) nodeRefs() []CLRSRef {
	var out []CLRSRef
	for _, t := range m.refs.Nodes[m.sel].Topics {
		out = append(out, t.CLRS...)
	}
	return out
}

func (m *refsModel) moveRef(d int) {
	n := len(m.nodeRefs())
	if n == 0 {
		return
	}
	m.refCursor = (m.refCursor + d + n) % n
	content, line := m.detailContent()
	m.vp.SetContent(content)
	m.vp.EnsureVisible(line, 0, 0)
}

func (m *refsModel) locate(i int) (layer, col int) {
	for l, nodes := range m.refs.layers {
		if c := slices.Index(nodes, i); c >= 0 {
			return l, c
		}
	}
	return 0, 0
}

func (m *refsModel) moveCol(d int) {
	l, c := m.locate(m.sel)
	if nodes := m.refs.layers[l]; c+d >= 0 && c+d < len(nodes) {
		m.sel = nodes[c+d]
	}
}

// mapRow is one screen row of boxes; a layer wider than the terminal spans several.
type mapRow []int

func (m *refsModel) boxWidth() int { return min(boxWidth, max(12, m.width)) }

func (m *refsModel) rows() []mapRow {
	bw := m.boxWidth()
	per := max(1, (m.width+boxGap)/(bw+boxGap))
	var rows []mapRow
	for _, nodes := range m.refs.layers {
		for s := 0; s < len(nodes); s += per {
			rows = append(rows, nodes[s:min(s+per, len(nodes))])
		}
	}
	return rows
}

// boxCenter is the x of the k-th box's center in a centered row of n boxes.
func (m *refsModel) boxCenter(k, n int) int {
	bw := m.boxWidth()
	left := (m.width - (n*bw + (n-1)*boxGap)) / 2
	return left + k*(bw+boxGap) + bw/2
}

// moveRow selects the box nearest horizontally in the row above (d=-1) or below (d=1).
func (m *refsModel) moveRow(d int) {
	rows := m.rows()
	for r, row := range rows {
		k := slices.Index(row, m.sel)
		if k < 0 {
			continue
		}
		if r+d < 0 || r+d >= len(rows) {
			return
		}
		x, next := m.boxCenter(k, len(row)), rows[r+d]
		best := 0
		for j := range next {
			if abs(m.boxCenter(j, len(next))-x) < abs(m.boxCenter(best, len(next))-x) {
				best = j
			}
		}
		m.sel = next[best]
		return
	}
}

func abs(x int) int { return max(x, -x) }

// nodeProgress is "assessed/total · avg tier" over the node's bank topics.
func (m *refsModel) nodeProgress(n RefNode) string {
	total, assessed, sum := 0, 0, 0.0
	for _, t := range n.Topics {
		if t.BankTopic == "" {
			continue
		}
		total++
		if v, ok := m.tiers[t.BankTopic]; ok {
			assessed++
			sum += v
		}
	}
	if total == 0 {
		return dimStyle.Render("reading only")
	}
	s := fmt.Sprintf("%d/%d assessed", assessed, total)
	if assessed == 0 {
		return dimStyle.Render(s)
	}
	avg := sum / float64(assessed)
	return tierStyle(avg).Render(fmt.Sprintf("%s · %.1f", s, avg))
}

func (m *refsModel) boxStyle(i int) lipgloss.Style {
	st := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(boxBorderColor).Padding(0, 1).Width(m.boxWidth())
	switch {
	case i == m.sel:
		st = st.Border(lipgloss.ThickBorder()).BorderForeground(selectedColor)
	case slices.Contains(m.refs.Nodes[m.sel].Prereqs, m.refs.Nodes[i].ID):
		st = st.BorderForeground(requiresColor)
	case slices.Contains(m.refs.Nodes[i].Prereqs, m.refs.Nodes[m.sel].ID):
		st = st.BorderForeground(unlocksColor)
	}
	return st
}

func (m *refsModel) mapView() string {
	w := max(12, m.width)
	clip := lipgloss.NewStyle().MaxWidth(w)
	header := clip.Render(boldStyle.Render("References") + dimStyle.Render(" · "+m.refs.Book.Title))

	var body []string
	selTop, selBottom := 0, 0
	for _, row := range m.rows() {
		h := 0
		for _, i := range row {
			n := m.refs.Nodes[i]
			h = max(h, lipgloss.Height(m.boxStyle(i).Render(boldStyle.Render(n.Title)+"\n"+m.nodeProgress(n))))
		}
		var boxes []string
		for k, i := range row {
			if k > 0 {
				boxes = append(boxes, strings.Repeat(" ", boxGap))
			}
			n := m.refs.Nodes[i]
			boxes = append(boxes, m.boxStyle(i).Height(h).Render(boldStyle.Render(n.Title)+"\n"+m.nodeProgress(n)))
		}
		if slices.Contains(row, m.sel) {
			selTop, selBottom = len(body), len(body)+h
		}
		line := lipgloss.PlaceHorizontal(w, lipgloss.Center, lipgloss.JoinHorizontal(lipgloss.Top, boxes...))
		body = append(body, strings.Split(line, "\n")...)
		body = append(body, "")
	}

	keys := "←/→/↑/↓ move · enter details · q quit"
	if m.embedded {
		keys = "←/→/↑/↓ move · enter details · esc back to review"
	}
	footer := []string{
		clip.Render(boldStyle.Render("Requires: ") + m.refs.titles(m.refs.prereqs(m.sel))),
		clip.Render(boldStyle.Render("Unlocks: ") + m.refs.titles(m.refs.dependents[m.sel])),
		clip.Render(lipgloss.NewStyle().Foreground(selectedColor).Render("■ selected") + "  " +
			lipgloss.NewStyle().Foreground(requiresColor).Render("■ requires") + "  " +
			lipgloss.NewStyle().Foreground(unlocksColor).Render("■ unlocks")),
	}
	if m.status != "" {
		footer = append(footer, clip.Render(warnStyle.Render(m.status)))
	}
	footer = append(footer, clip.Render(dimStyle.Render(keys)))

	// Scroll just enough to keep the selected row on screen.
	avail := max(1, m.height-2-len(footer))
	off := min(max(0, selBottom-avail), selTop)
	body = body[off:min(len(body), off+avail)]
	return header + "\n\n" + strings.Join(body, "\n") + "\n" + strings.Join(footer, "\n")
}

// detailContent renders the selected node; line is where the selected CLRS ref starts.
func (m *refsModel) detailContent() (content string, line int) {
	r := m.refs
	n := r.Nodes[m.sel]
	w := max(20, m.width-2)
	wrap := lipgloss.NewStyle().Width(w)
	note := lipgloss.NewStyle().Width(w).PaddingLeft(4).Faint(true)

	var b strings.Builder
	fmt.Fprintf(&b, "%s\n%s\n\n", accentStyle.Render(n.Title), wrap.Render(n.Summary))
	fmt.Fprintf(&b, "%s\n", wrap.Render(boldStyle.Render("Requires: ")+r.titles(r.prereqs(m.sel))))
	fmt.Fprintf(&b, "%s\n", wrap.Render(boldStyle.Render("Unlocks: ")+r.titles(r.dependents[m.sel])))
	k := 0
	for _, t := range n.Topics {
		head := boldStyle.Render(t.Name)
		if v, ok := m.tiers[t.BankTopic]; ok && t.BankTopic != "" {
			head += "  " + tierStyle(v).Render(fmt.Sprintf("tier %.1f %s", v, tierName(v)))
		} else if t.BankTopic != "" {
			head += "  " + dimStyle.Render("not assessed")
		}
		fmt.Fprintf(&b, "\n%s\n", wrap.Render(head))
		if t.Summary != "" {
			fmt.Fprintf(&b, "%s\n", wrap.Render(t.Summary))
		}
		for _, c := range t.CLRS {
			text := fmt.Sprintf("  CLRS %s %s — p.%d", c.Ref, c.Title, c.BookPage)
			if k == m.refCursor {
				line = strings.Count(b.String(), "\n")
				text = accentStyle.Render(fmt.Sprintf("▸ CLRS %s %s — p.%d", c.Ref, c.Title, c.BookPage))
			}
			fmt.Fprintf(&b, "%s\n", wrap.Render(text))
			if c.Note != "" {
				fmt.Fprintf(&b, "%s\n", note.Render(c.Note))
			}
			k++
		}
		for _, e := range t.External {
			fmt.Fprintf(&b, "%s\n", wrap.Render("  "+e.Title+" "+dimStyle.Render(e.URL)))
		}
	}
	return b.String(), line
}

func (m *refsModel) View() tea.View {
	var s string
	if !m.detail {
		s = m.mapView()
	} else {
		keys := "↑/↓ scroll · tab/shift+tab CLRS ref · o open in PDF · esc map · q quit"
		if m.startDetail {
			keys = "↑/↓ scroll · tab/shift+tab CLRS ref · o open in PDF · backspace map · esc back to review"
		} else if m.embedded {
			keys = "↑/↓ scroll · tab/shift+tab CLRS ref · o open in PDF · esc map"
		}
		clip := lipgloss.NewStyle().MaxWidth(max(12, m.width))
		status := ""
		if m.status != "" {
			status = clip.Render(warnStyle.Render(m.status))
		}
		s = m.vp.View() + "\n" + status + "\n" + clip.Render(dimStyle.Render(keys))
	}
	v := tea.NewView(s)
	v.AltScreen = true
	return v
}

// openRefs shows the references screen on the node that covers the current question's topic.
func (m *model) openRefs() {
	topic := m.qs[m.cur].Topic
	refs, err := loadRefs()
	if err != nil {
		m.status = "references.json: " + err.Error()
		return
	}
	i := refs.nodeFor(topic)
	if i < 0 {
		m.status = "No reference node covers " + topic + "."
		return
	}
	history, err := loadAll(m.dir)
	if err != nil {
		m.status = err.Error()
		return
	}
	r := newRefsModel(refs, topicTiers(history))
	r.embedded = true
	r.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
	r.openDetail(i, true)
	m.refs = r
}

// updateRefs delegates to the open references screen; ctrl+c still saves and quits.
func (m *model) updateRefs(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok && k.String() == "ctrl+c" {
		return m.quit()
	}
	_, cmd := m.refs.Update(msg)
	if m.refs.done {
		m.refs = nil
	}
	return m, cmd
}

// pdfURL is the file URL of path at a page; browsers honor the #page fragment.
func pdfURL(path string, page int) string {
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	segs := strings.Split(filepath.ToSlash(path), "/")
	for i, s := range segs {
		segs[i] = url.PathEscape(s)
	}
	return fmt.Sprintf("file://%s#page=%d", strings.Join(segs, "/"), page)
}

var browsers = []string{"chrome", "chromium", "firefox", "brave", "edge", "vivaldi", "opera", "librewolf"}

// viewerCmd opens path at page: $CLRS_VIEWER with the file URL, else the default PDF app
// with its page flag (or the URL for browsers), else xdg-open on the plain path.
func viewerCmd(path string, page int) *exec.Cmd {
	u := pdfURL(path, page)
	if v := strings.Fields(os.Getenv("CLRS_VIEWER")); len(v) > 0 {
		return exec.Command(v[0], append(v[1:], u)...)
	}
	out, _ := exec.Command("xdg-mime", "query", "default", "application/pdf").Output()
	app := strings.TrimSpace(string(out))
	id, bin, n := strings.ToLower(app), desktopExec(app), strconv.Itoa(page)
	switch {
	case bin == "":
	case strings.Contains(id, "zathura"):
		return exec.Command(bin, "--page="+n, path)
	case strings.Contains(id, "okular"):
		return exec.Command(bin, "-p", n, path)
	case strings.Contains(id, "evince"):
		return exec.Command(bin, "-i", n, path)
	case slices.ContainsFunc(browsers, func(b string) bool { return strings.Contains(id, b) }):
		return exec.Command(bin, u)
	}
	return exec.Command("xdg-open", path)
}

// desktopExec returns the binary from a .desktop file's Exec line, or "".
func desktopExec(app string) string {
	if app == "" {
		return ""
	}
	home, _ := os.UserHomeDir()
	dataHome := cmp.Or(os.Getenv("XDG_DATA_HOME"), filepath.Join(home, ".local", "share"))
	dataDirs := cmp.Or(os.Getenv("XDG_DATA_DIRS"), "/usr/local/share:/usr/share")
	for _, d := range append([]string{dataHome}, filepath.SplitList(dataDirs)...) {
		data, err := os.ReadFile(filepath.Join(d, "applications", app))
		if err != nil {
			continue
		}
		for line := range strings.Lines(string(data)) {
			if f := strings.Fields(strings.TrimPrefix(line, "Exec=")); strings.HasPrefix(line, "Exec=") && len(f) > 0 {
				return strings.Trim(f[0], `"`)
			}
		}
	}
	return ""
}

// openRef starts the PDF viewer in the background and reports what it opened.
func openRef(ref CLRSRef) tea.Cmd {
	return func() tea.Msg {
		path := os.Getenv("CLRS_PDF")
		if path == "" {
			return openedMsg("set CLRS_PDF to the path of your copy of CLRS (3rd edition) to open references")
		}
		if _, err := os.Stat(path); err != nil {
			return openedMsg("CLRS PDF not found: " + path + " (set CLRS_PDF)")
		}
		cmd := viewerCmd(path, ref.PDFPage)
		if err := cmd.Start(); err != nil {
			return openedMsg("could not open the PDF: " + err.Error())
		}
		go cmd.Wait()
		return openedMsg(fmt.Sprintf("Opened CLRS %s (PDF page %d) with %s.", ref.Ref, ref.PDFPage, filepath.Base(cmd.Path)))
	}
}

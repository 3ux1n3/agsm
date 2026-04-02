package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/3ux1n3/agsm/internal/config"
	"github.com/3ux1n3/agsm/internal/registry"
	"github.com/3ux1n3/agsm/internal/session"
)

type mode int

const (
	modeList mode = iota
	modeSearch
	modeRename
	modeDeleteConfirm
	modeNewSession
)

type refreshMsg struct {
	items []session.Session
	err   error
}

type actionDoneMsg struct {
	err error
}

type app struct {
	cfg         config.Config
	registry    *registry.Registry
	styles      styles
	keys        keyMap
	mode        mode
	items       []session.Session
	selected    int
	width       int
	height      int
	searchInput textinput.Model
	renameInput textinput.Model
	dirInput    textinput.Model
	nameInput   textinput.Model
	newField    int
	status      string
	err         error
	searchQuery string
	listOffset  int
	lastRefresh time.Time
}

func NewApp(cfg config.Config, registry *registry.Registry) *app {
	searchInput := newSearchInput()
	renameInput := newRenameInput()
	dirInput := newDirectoryInput()
	nameInput := newOptionalNameInput()

	return &app{
		cfg:         cfg,
		registry:    registry,
		styles:      defaultStyles(),
		keys:        defaultKeys(),
		mode:        modeList,
		searchInput: searchInput,
		renameInput: renameInput,
		dirInput:    dirInput,
		nameInput:   nameInput,
		status:      "Press / to search, Enter to resume, q to quit.",
	}
}

func (a *app) Run() error {
	program := tea.NewProgram(a, tea.WithAltScreen())
	_, err := program.Run()
	return err
}

func (a *app) Init() tea.Cmd {
	return a.refreshCmd()
}

func (a *app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
	case refreshMsg:
		a.err = msg.err
		if msg.err == nil {
			a.items = msg.items
			a.clampSelection()
			a.status = fmt.Sprintf("%d sessions loaded", len(a.items))
			a.lastRefresh = time.Now()
		}
	case actionDoneMsg:
		a.err = msg.err
		if msg.err == nil {
			a.status = "Action completed"
			return a, a.refreshCmd()
		}
	}

	switch a.mode {
	case modeSearch:
		return a.updateSearch(msg)
	case modeRename:
		return a.updateRename(msg)
	case modeDeleteConfirm:
		return a.updateDelete(msg)
	case modeNewSession:
		return a.updateNewSession(msg)
	default:
		return a.updateList(msg)
	}
}

func (a *app) View() string {
	width := max(1, a.width)
	height := max(1, a.height)
	base := a.renderBaseView()
	if !a.hasOverlay() {
		return lipgloss.Place(width, height, lipgloss.Left, lipgloss.Top, base)
	}

	dimmed := a.styles.backdrop.Render(base)
	modal := a.renderOverlay()
	return placeOverlay(width, height, dimmed, modal)
}

func (a *app) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, a.keys.Quit):
			return a, tea.Quit
		case key.Matches(keyMsg, a.keys.Up):
			a.move(-1)
		case key.Matches(keyMsg, a.keys.Down):
			a.move(1)
		case key.Matches(keyMsg, a.keys.Esc):
			if a.searchQuery != "" {
				a.searchQuery = ""
				return a, a.refreshCmd()
			}
		case key.Matches(keyMsg, a.keys.Search):
			a.mode = modeSearch
			a.searchInput.SetValue(a.searchQuery)
			a.searchInput.Focus()
			a.err = nil
		case key.Matches(keyMsg, a.keys.Refresh):
			return a, a.refreshCmd()
		case key.Matches(keyMsg, a.keys.Rename):
			if current, ok := a.current(); ok {
				a.mode = modeRename
				a.renameInput.SetValue(current.DisplayName())
				a.renameInput.Focus()
				a.err = nil
			}
		case key.Matches(keyMsg, a.keys.Delete):
			if len(a.items) > 0 {
				a.mode = modeDeleteConfirm
				a.err = nil
			}
		case key.Matches(keyMsg, a.keys.New):
			a.mode = modeNewSession
			a.newField = 0
			a.dirInput.SetValue("")
			a.nameInput.SetValue("")
			a.focusNewField()
			a.err = nil
		case key.Matches(keyMsg, a.keys.Enter):
			if current, ok := a.current(); ok {
				adapter := a.registry.AdapterFor(current.Agent)
				if adapter == nil {
					a.err = fmt.Errorf("no adapter for %s", current.Agent)
					return a, nil
				}
				cmd := adapter.ResumeCommand(current)
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				return a, tea.ExecProcess(cmd, func(err error) tea.Msg { return actionDoneMsg{err: err} })
			}
		}
	}
	return a, nil
}

func (a *app) updateSearch(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, a.keys.Up):
			a.move(-1)
			return a, nil
		case key.Matches(keyMsg, a.keys.Down):
			a.move(1)
			return a, nil
		case key.Matches(keyMsg, a.keys.Esc):
			a.searchQuery = ""
			a.closeOverlay()
			return a, a.refreshCmd()
		case key.Matches(keyMsg, a.keys.Enter):
			a.closeOverlay()
			return a, nil
		}
	}

	var cmd tea.Cmd
	a.searchInput, cmd = a.searchInput.Update(msg)
	a.searchQuery = a.searchInput.Value()
	a.items = a.registry.Filter(a.searchQuery)
	a.clampSelection()
	return a, cmd
}

func (a *app) updateRename(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, a.keys.Esc):
			a.closeOverlay()
			return a, nil
		case key.Matches(keyMsg, a.keys.Enter):
			if current, ok := a.current(); ok {
				a.err = a.registry.Rename(current, a.renameInput.Value())
				if a.err == nil {
					a.items = a.registry.Filter(a.searchQuery)
					a.status = "Session renamed"
				}
			}
			a.closeOverlay()
			return a, nil
		}
	}

	var cmd tea.Cmd
	a.renameInput, cmd = a.renameInput.Update(msg)
	return a, cmd
}

func (a *app) updateDelete(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc", "n":
			a.closeOverlay()
			return a, nil
		case "y":
			if current, ok := a.current(); ok {
				if err := a.registry.Delete(current); err != nil {
					a.err = err
				} else {
					a.items = a.registry.Filter(a.searchQuery)
					a.status = "Session deleted"
				}
			}
			a.closeOverlay()
			return a, nil
		}
	}
	return a, nil
}

func (a *app) updateNewSession(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, a.keys.Esc):
			a.closeOverlay()
			return a, nil
		case key.Matches(keyMsg, a.keys.Tab):
			a.newField = (a.newField + 1) % 2
			a.focusNewField()
			return a, nil
		case key.Matches(keyMsg, a.keys.Enter):
			dir := strings.TrimSpace(a.dirInput.Value())
			if dir == "" {
				a.err = fmt.Errorf("directory is required")
				return a, nil
			}
			resolved, err := filepath.Abs(dir)
			if err != nil {
				a.err = err
				return a, nil
			}
			info, err := os.Stat(resolved)
			if err != nil || !info.IsDir() {
				a.err = fmt.Errorf("directory does not exist: %s", resolved)
				return a, nil
			}

			adapter := a.registry.AdapterFor("opencode")
			if adapter == nil {
				a.err = fmt.Errorf("opencode adapter is not configured")
				return a, nil
			}

			cmd := adapter.NewCommand(resolved)
			if sessionName := strings.TrimSpace(a.nameInput.Value()); sessionName != "" {
				a.status = "Launching new OpenCode session: " + sessionName
			} else {
				a.status = "Launching new OpenCode session"
			}
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			a.closeOverlay()
			return a, tea.ExecProcess(cmd, func(err error) tea.Msg { return actionDoneMsg{err: err} })
		}
	}

	if a.newField == 0 {
		var cmd tea.Cmd
		a.dirInput, cmd = a.dirInput.Update(msg)
		return a, cmd
	}

	var cmd tea.Cmd
	a.nameInput, cmd = a.nameInput.Update(msg)
	return a, cmd
}

func (a *app) renderBaseView() string {
	frameWidth := max(20, a.width)
	frameHeight := max(8, a.height)
	contentWidth := max(1, frameWidth-2)
	availableHeight := max(1, frameHeight-2)
	now := time.Now()

	titleLeft := a.styles.title.Render(" AGSM") + "  " + a.styles.subtitle.Render("agent session manager")
	titleRight := a.styles.headerMeta.Render(strings.Join(a.headerMetaParts(now), "  •  ") + " ")
	titleBar := a.styles.titleBar.Width(contentWidth).Render(joinEdge(contentWidth, titleLeft, titleRight))

	statsLine := strings.Join([]string{
		statBlock(a.styles, "agent", "opencode"),
		statBlock(a.styles, "shown", fmt.Sprintf("%d", len(a.items))),
		statBlock(a.styles, "selected", a.selectedSummary()),
		statBlock(a.styles, "refresh", formatRefreshTime(a.lastRefresh, now)),
	}, "  ")
	if a.searchQuery != "" {
		statsLine = joinEdge(contentWidth, " "+statsLine, a.styles.filterPill.Render("/ "+a.searchQuery+" "))
	} else {
		statsLine = padRight(" "+statsLine, contentWidth)
	}
	statBar := a.styles.statBar.Width(contentWidth).Render(statsLine)

	footer := a.renderFooter(contentWidth)
	chromeHeight := lipgloss.Height(titleBar) + lipgloss.Height(statBar) + lipgloss.Height(footer)
	listHeight := max(6, availableHeight-chromeHeight)
	a.ensureListOffset(max(1, listHeight-1))

	listContent := renderListWindow(contentWidth, listHeight, a.styles, a.items, a.selected, a.listOffset)
	view := lipgloss.JoinVertical(lipgloss.Left, titleBar, statBar, listContent, footer)
	inner := fillVertical(view, availableHeight)
	framed := a.styles.frame.Width(contentWidth).Height(availableHeight).Render(inner)
	return a.styles.app.Render(fillVertical(framed, frameHeight))
}

func (a *app) renderOverlay() string {
	switch a.mode {
	case modeSearch:
		return a.renderSearchModal()
	case modeRename:
		return a.renderRenameModal()
	case modeDeleteConfirm:
		return a.renderDeleteModal()
	case modeNewSession:
		return a.renderNewSessionModal()
	default:
		return ""
	}
}

func (a *app) renderRenameModal() string {
	current, _ := a.current()
	body := []string{
		a.styles.modalTitle.Render("Rename Session"),
		a.styles.muted.Render("Update the AGSM custom name for this session."),
		"",
		a.styles.footerMeta.Render("Current: " + current.DisplayName()),
		a.renameInput.View(),
		"",
		a.styles.footerMeta.Render("Enter to save • Esc to cancel"),
	}
	return a.styles.modal.Width(56).Render(strings.Join(body, "\n"))
}

func (a *app) renderDeleteModal() string {
	current, _ := a.current()
	body := []string{
		a.styles.modalDanger.Render("Delete Session"),
		a.styles.muted.Render("This removes the session from OpenCode storage."),
		"",
		a.styles.footerMeta.Render("Session: " + current.DisplayName()),
		a.styles.footerMeta.Render("Folder:  " + shortenHome(current.ProjectDir)),
		"",
		a.styles.footerMeta.Render("Press y to confirm • n or Esc to cancel"),
	}
	return a.styles.modal.Width(60).Render(strings.Join(body, "\n"))
}

func (a *app) renderNewSessionModal() string {
	fields := []string{
		a.styles.modalTitle.Render("New OpenCode Session"),
		a.styles.muted.Render("Choose a working directory and optionally label the launch."),
		"",
		a.dirInput.View(),
		a.nameInput.View(),
	}
	if a.err != nil {
		fields = append(fields, "", a.styles.errorText.Render(a.err.Error()))
	}
	fields = append(fields, "", a.styles.footerMeta.Render("Tab to switch fields • Enter to launch • Esc to cancel"))
	return a.styles.modal.Width(68).Render(strings.Join(fields, "\n"))
}

func (a *app) renderSearchModal() string {
	modalWidth := min(max(64, a.width-12), 96)
	modalHeight := min(max(12, a.height-8), 20)
	resultsHeight := max(4, modalHeight-7)
	a.ensureListOffset(resultsHeight)
	results := renderFlatListWindow(modalWidth-4, resultsHeight, a.styles, a.items, a.selected, a.listOffset)
	body := []string{
		a.styles.modalTitle.Render("Jump To Session"),
		a.styles.muted.Render("Filter by title, folder, or agent."),
		"",
		a.searchInput.View(),
		"",
		results,
		"",
		a.styles.footerMeta.Render("Up/Down move • Enter keep filtered list • Esc clear search"),
	}
	return a.styles.modal.Width(modalWidth).Render(strings.Join(body, "\n"))
}

func (a *app) renderFooter(width int) string {
	actions := truncate("[Enter] resume  [/] search  [Ctrl+N] new  [Ctrl+R] rename  [Ctrl+D] delete  [Ctrl+L] refresh  [q] quit", width)

	metaText := a.status
	if a.err != nil {
		metaText = a.err.Error()
	} else if current, ok := a.current(); ok {
		metaText = shortenHome(current.ProjectDir) + "  •  " + current.ID
	}
	meta := a.styles.footerMeta.Render(truncate(metaText, width))
	if a.err != nil {
		meta = a.styles.errorText.Render(truncate(metaText, width))
	}
	if a.searchQuery != "" {
		meta = meta + "\n" + a.styles.footerMeta.Render("Esc clears the current filter")
	}
	return a.styles.footer.Width(width).Render(a.styles.footerMeta.Render(actions) + "\n" + meta)
}

func joinEdge(width int, left, right string) string {
	if lipgloss.Width(left)+lipgloss.Width(right) >= width {
		return padRight(left, width)
	}
	gap := max(1, width-lipgloss.Width(left)-lipgloss.Width(right))
	return left + strings.Repeat(" ", gap) + right
}

func (a *app) refreshCmd() tea.Cmd {
	return func() tea.Msg {
		items, err := a.registry.Refresh()
		if err == nil && strings.TrimSpace(a.searchQuery) != "" {
			items = a.registry.Filter(a.searchQuery)
		}
		return refreshMsg{items: items, err: err}
	}
}

func (a *app) current() (session.Session, bool) {
	if len(a.items) == 0 || a.selected < 0 || a.selected >= len(a.items) {
		return session.Session{}, false
	}
	return a.items[a.selected], true
}

func (a *app) move(delta int) {
	if len(a.items) == 0 {
		return
	}
	a.selected += delta
	if a.selected < 0 {
		a.selected = 0
	}
	if a.selected >= len(a.items) {
		a.selected = len(a.items) - 1
	}
	a.ensureListOffset(0)
}

func (a *app) clampSelection() {
	if len(a.items) == 0 {
		a.selected = 0
		a.listOffset = 0
		return
	}
	if a.selected >= len(a.items) {
		a.selected = len(a.items) - 1
	}
	if a.selected < 0 {
		a.selected = 0
	}
	a.ensureListOffset(0)
}

func (a *app) focusNewField() {
	if a.newField == 0 {
		a.dirInput.Focus()
		a.nameInput.Blur()
		return
	}
	a.dirInput.Blur()
	a.nameInput.Focus()
}

func (a *app) hasOverlay() bool {
	return a.mode == modeSearch || a.mode == modeRename || a.mode == modeDeleteConfirm || a.mode == modeNewSession
}

func (a *app) closeOverlay() {
	a.mode = modeList
	a.searchInput.Blur()
	a.renameInput.Blur()
	a.dirInput.Blur()
	a.nameInput.Blur()
}

func placeOverlay(width, height int, base, modal string) string {
	baseView := fillVertical(lipgloss.Place(width, height, lipgloss.Left, lipgloss.Top, base), height)
	baseLines := strings.Split(baseView, "\n")
	modalLines := strings.Split(modal, "\n")
	if len(modalLines) > height {
		modalLines = modalLines[:height]
	}
	top := max(0, (height-len(modalLines))/2)
	for i, line := range modalLines {
		if top+i >= len(baseLines) {
			break
		}
		baseLines[top+i] = lipgloss.Place(width, 1, lipgloss.Center, lipgloss.Center, line)
	}
	return strings.Join(baseLines, "\n")
}

func (a *app) ensureListOffset(listHeight int) {
	if len(a.items) == 0 {
		a.listOffset = 0
		return
	}
	visibleRows := listHeight
	if visibleRows <= 0 {
		visibleRows = max(6, a.height-8)
	}
	visibleRows = max(1, visibleRows)
	if a.selected < a.listOffset {
		a.listOffset = a.selected
	}
	if a.selected >= a.listOffset+visibleRows {
		a.listOffset = a.selected - visibleRows + 1
	}
	maxOffset := max(0, len(a.items)-visibleRows)
	if a.listOffset > maxOffset {
		a.listOffset = maxOffset
	}
	if a.listOffset < 0 {
		a.listOffset = 0
	}
}

func (a *app) headerMetaParts(now time.Time) []string {
	parts := []string{fmt.Sprintf("%d sessions", len(a.items))}
	if a.searchQuery != "" {
		parts = append(parts, "filtered")
	}
	if refreshed := formatRefreshTime(a.lastRefresh, now); refreshed != "" {
		parts = append(parts, refreshed)
	}
	return parts
}

func (a *app) selectedSummary() string {
	if len(a.items) == 0 {
		return "0/0"
	}
	return fmt.Sprintf("%d/%d", a.selected+1, len(a.items))
}

func statBlock(st styles, label, value string) string {
	if strings.TrimSpace(value) == "" {
		value = "-"
	}
	return st.statLabel.Render(strings.ToUpper(label)) + " " + st.statValue.Render(value)
}

func padRight(value string, width int) string {
	gap := max(0, width-lipgloss.Width(value))
	return value + strings.Repeat(" ", gap)
}

package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/3ux1n3/agsm/internal/session"
)

func renderListWindow(width, height int, st styles, items []session.Session, selected, offset int) string {
	return renderCompactListWindow(width, height, st, items, selected, offset, true)
}

func renderFlatListWindow(width, height int, st styles, items []session.Session, selected, offset int) string {
	return renderCompactListWindow(width, height, st, items, selected, offset, false)
}

func renderCompactListWindow(width, height int, st styles, items []session.Session, selected, offset int, withHeader bool) string {
	contentWidth := max(40, width)
	headerHeight := 0
	lines := make([]string, 0, max(1, height))
	if withHeader {
		lines = append(lines, renderColumnHeader(contentWidth, st))
		headerHeight = 1
	}

	if len(items) == 0 {
		lines = append(lines, st.muted.Render("No sessions found."))
		return fillVertical(strings.Join(lines, "\n"), height)
	}

	visibleRows := len(items)
	if height > 0 {
		visibleRows = max(1, height-headerHeight)
	}
	if offset < 0 {
		offset = 0
	}
	if offset > len(items) {
		offset = len(items)
	}
	end := min(len(items), offset+visibleRows)

	now := time.Now()
	for i := offset; i < end; i++ {
		lines = append(lines, renderCompactRow(contentWidth, st, items[i], i == selected, now))
	}

	return fillVertical(strings.Join(lines, "\n"), height)
}

func renderColumnHeader(width int, st styles) string {
	indicatorWidth := 2
	agentWidth := 10
	updatedWidth := 8
	titleWidth := max(16, width*45/100)
	projectWidth := max(10, width-indicatorWidth-agentWidth-titleWidth-updatedWidth-3)

	indicator := lipgloss.NewStyle().Width(indicatorWidth).Render(" ")
	agent := lipgloss.NewStyle().Width(agentWidth).Render("AGENT")
	title := lipgloss.NewStyle().Width(titleWidth).Render("SESSION")
	project := lipgloss.NewStyle().Width(projectWidth).Render("PROJECT")
	updated := lipgloss.NewStyle().Width(updatedWidth).Align(lipgloss.Right).Render("UPDATED")

	row := indicator + agent + " " + title + " " + project + " " + updated
	return st.listHeader.Width(width).Render(row)
}

func renderCompactRow(width int, st styles, item session.Session, selected bool, now time.Time) string {
	indicatorWidth := 2
	agentWidth := 10
	updatedWidth := 8
	titleWidth := max(16, width*45/100)
	projectWidth := max(10, width-indicatorWidth-agentWidth-titleWidth-updatedWidth-3)

	indicator := st.rowMeta.Render("·")
	if selected {
		indicator = st.rowAccent.Render("▌")
	}
	indicatorCol := lipgloss.NewStyle().Width(indicatorWidth).Render(indicator)
	agentText := strings.ToUpper(truncate(item.Agent, 8))
	agentCol := renderAgentBadge(st, item.Agent, agentText)
	titleCol := lipgloss.NewStyle().Width(titleWidth).Render(st.rowTitle.Render(truncate(item.DisplayName(), titleWidth)))
	projectCol := lipgloss.NewStyle().Width(projectWidth).Render(st.rowMeta.Render(truncate(projectName(item.ProjectDir), projectWidth)))
	updatedCol := lipgloss.NewStyle().Width(updatedWidth).Align(lipgloss.Right).Render(st.timeText.Render(formatSessionTime(item.LastActive, now)))

	row := indicatorCol + agentCol + " " + titleCol + " " + projectCol + " " + updatedCol
	if selected {
		return st.selectedRow.Width(width).Render(row)
	}
	return st.row.Width(width).Render(row)
}

func renderAgentBadge(st styles, agent, text string) string {
	switch strings.ToLower(strings.TrimSpace(agent)) {
	case "claude":
		return st.agentClaude.Render(text)
	case "opencode":
		return st.agentOpen.Render(text)
	default:
		return st.agentBadge.Render(text)
	}
}

func truncate(v string, width int) string {
	if lipgloss.Width(v) <= width {
		return v
	}
	if width <= 1 {
		return ""
	}

	targetWidth := width - 1
	currentWidth := 0
	var out strings.Builder
	for _, r := range v {
		runeWidth := lipgloss.Width(string(r))
		if currentWidth+runeWidth > targetWidth {
			break
		}
		out.WriteRune(r)
		currentWidth += runeWidth
	}
	return out.String() + "…"
}

func projectName(path string) string {
	if path == "" {
		return "-"
	}
	return filepath.Base(path)
}

func shortenHome(path string) string {
	if path == "" {
		return "-"
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}
	if strings.HasPrefix(path, home) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}

func formatRefreshTime(ts, now time.Time) string {
	if ts.IsZero() {
		return ""
	}
	delta := now.Sub(ts)
	if delta < time.Minute {
		return "just now"
	}
	if delta < time.Hour {
		return fmt.Sprintf("%dm ago", int(delta.Minutes()))
	}
	if sameDay(ts, now) {
		return ts.Format("15:04")
	}
	return formatSessionTime(ts, now)
}

func formatSessionTime(ts, now time.Time) string {
	if ts.IsZero() {
		return "unknown"
	}
	if sameDay(ts, now) {
		return ts.Format("15:04")
	}
	if sameDay(ts, now.AddDate(0, 0, -1)) {
		return "yday"
	}
	if now.Sub(ts) < 7*24*time.Hour {
		return strings.ToLower(ts.Format("Mon"))
	}
	if ts.Year() == now.Year() {
		return ts.Format("Jan 2")
	}
	return ts.Format("06-01-02")
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

func fillVertical(content string, height int) string {
	if height <= 0 {
		return content
	}
	lines := strings.Split(content, "\n")
	for len(lines) < height {
		lines = append(lines, "")
	}
	if len(lines) > height {
		lines = lines[:height]
	}
	return strings.Join(lines, "\n")
}

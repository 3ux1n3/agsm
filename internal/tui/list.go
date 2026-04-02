package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/3ux1n3/agsm/internal/session"
)

func renderList(width int, st styles, items []session.Session, selected int) string {
	return renderListWindow(width, 0, st, items, selected, 0)
}

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
	updatedWidth := 10
	titleWidth := max(16, width/3)
	folderWidth := max(14, width-indicatorWidth-agentWidth-titleWidth-updatedWidth-3)

	indicator := lipgloss.NewStyle().Width(indicatorWidth).Render(" ")
	agent := lipgloss.NewStyle().Width(agentWidth).Render("AGENT")
	title := lipgloss.NewStyle().Width(titleWidth).Render("SESSION")
	folder := lipgloss.NewStyle().Width(folderWidth).Render("FOLDER")
	updated := lipgloss.NewStyle().Width(updatedWidth).Align(lipgloss.Right).Render("UPDATED")

	row := indicator + agent + " " + title + " " + folder + " " + updated
	return st.listHeader.Width(width).Render(row)
}

func renderCompactRow(width int, st styles, item session.Session, selected bool, now time.Time) string {
	indicatorWidth := 2
	agentWidth := 10
	updatedWidth := 10
	titleWidth := max(16, width/3)
	folderWidth := max(14, width-indicatorWidth-agentWidth-titleWidth-updatedWidth-3)

	indicator := " "
	if selected {
		indicator = ">"
	}
	indicatorCol := lipgloss.NewStyle().Width(indicatorWidth).Render(indicator)
	agentText := strings.ToUpper(truncate(item.Agent, 8))
	agentCol := lipgloss.NewStyle().Width(agentWidth).Render(st.agentBadge.Render(agentText))
	titleCol := lipgloss.NewStyle().Width(titleWidth).Render(st.rowTitle.Render(truncate(item.DisplayName(), titleWidth)))
	folderCol := lipgloss.NewStyle().Width(folderWidth).Render(st.rowMeta.Render(truncate(shortenHome(item.ProjectDir), folderWidth)))
	updatedCol := lipgloss.NewStyle().Width(updatedWidth).Align(lipgloss.Right).Render(st.timeText.Render(formatSessionTime(item.LastActive, now)))

	row := indicatorCol + agentCol + " " + titleCol + " " + folderCol + " " + updatedCol
	if selected {
		return st.selectedRow.Width(width).Render(row)
	}
	return st.row.Width(width).Render(row)
}

func truncate(v string, width int) string {
	if lipgloss.Width(v) <= width {
		return v
	}
	if width <= 1 {
		return ""
	}
	return v[:max(0, width-1)] + "…"
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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

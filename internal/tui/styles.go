package tui

import "github.com/charmbracelet/lipgloss"

type styles struct {
	app         lipgloss.Style
	frame       lipgloss.Style
	titleBar    lipgloss.Style
	statBar     lipgloss.Style
	title       lipgloss.Style
	subtitle    lipgloss.Style
	headerMeta  lipgloss.Style
	statLabel   lipgloss.Style
	statValue   lipgloss.Style
	listHeader  lipgloss.Style
	row         lipgloss.Style
	rowTitle    lipgloss.Style
	rowMeta     lipgloss.Style
	selectedRow lipgloss.Style
	muted       lipgloss.Style
	errorText   lipgloss.Style
	footer      lipgloss.Style
	footerKey   lipgloss.Style
	footerMeta  lipgloss.Style
	modal       lipgloss.Style
	modalTitle  lipgloss.Style
	modalDanger lipgloss.Style
	backdrop    lipgloss.Style
	agentBadge  lipgloss.Style
	timeText    lipgloss.Style
	filterPill  lipgloss.Style
	status      lipgloss.Style
}

func defaultStyles() styles {
	border := lipgloss.NormalBorder()

	return styles{
		app:         lipgloss.NewStyle(),
		frame:       lipgloss.NewStyle().BorderStyle(border).BorderForeground(adaptive("6", "4")),
		titleBar:    lipgloss.NewStyle().Background(adaptive("24", "12")).Foreground(adaptive("15", "15")).Bold(true),
		statBar:     lipgloss.NewStyle().Background(adaptive("236", "254")),
		title:       lipgloss.NewStyle().Bold(true).Foreground(adaptive("15", "0")),
		subtitle:    lipgloss.NewStyle().Foreground(adaptive("7", "8")),
		headerMeta:  lipgloss.NewStyle().Foreground(adaptive("15", "15")),
		statLabel:   lipgloss.NewStyle().Foreground(adaptive("8", "8")),
		statValue:   lipgloss.NewStyle().Bold(true).Foreground(adaptive("10", "2")),
		listHeader:  lipgloss.NewStyle().Background(adaptive("238", "252")).Foreground(adaptive("14", "4")).Bold(true),
		row:         lipgloss.NewStyle(),
		rowTitle:    lipgloss.NewStyle().Foreground(adaptive("15", "0")),
		rowMeta:     lipgloss.NewStyle().Foreground(adaptive("8", "8")),
		selectedRow: lipgloss.NewStyle().Background(adaptive("24", "12")).Foreground(adaptive("15", "15")).Bold(true),
		muted:       lipgloss.NewStyle().Foreground(adaptive("8", "8")),
		errorText:   lipgloss.NewStyle().Foreground(adaptive("9", "1")).Bold(true),
		footer:      lipgloss.NewStyle().BorderTop(true).BorderForeground(adaptive("8", "7")).PaddingTop(0).Foreground(adaptive("7", "8")),
		footerKey:   lipgloss.NewStyle().Background(adaptive("24", "12")).Foreground(adaptive("15", "15")).Bold(true).Padding(0, 1),
		footerMeta:  lipgloss.NewStyle().Foreground(adaptive("8", "8")),
		modal:       lipgloss.NewStyle().BorderStyle(border).BorderForeground(adaptive("12", "4")).Background(adaptive("236", "255")).Padding(1, 2),
		modalTitle:  lipgloss.NewStyle().Bold(true).Foreground(adaptive("15", "0")),
		modalDanger: lipgloss.NewStyle().Bold(true).Foreground(adaptive("9", "1")),
		backdrop:    lipgloss.NewStyle().Foreground(adaptive("8", "8")),
		agentBadge:  lipgloss.NewStyle().Foreground(adaptive("14", "4")).Bold(true),
		timeText:    lipgloss.NewStyle().Foreground(adaptive("8", "8")),
		filterPill:  lipgloss.NewStyle().Foreground(adaptive("15", "15")).Background(adaptive("24", "12")).Padding(0, 1),
		status:      lipgloss.NewStyle().Foreground(adaptive("10", "2")),
	}
}

func adaptive(dark, light string) lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: dark, Light: light}
}

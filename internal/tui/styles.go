package tui

import "github.com/charmbracelet/lipgloss"

type styles struct {
	app          lipgloss.Style
	frame        lipgloss.Style
	titleBar     lipgloss.Style
	statBar      lipgloss.Style
	statBlock    lipgloss.Style
	title        lipgloss.Style
	subtitle     lipgloss.Style
	headerMeta   lipgloss.Style
	statLabel    lipgloss.Style
	statValue    lipgloss.Style
	listHeader   lipgloss.Style
	row          lipgloss.Style
	rowTitle     lipgloss.Style
	rowAccent    lipgloss.Style
	rowMeta      lipgloss.Style
	selectedRow  lipgloss.Style
	muted        lipgloss.Style
	errorText    lipgloss.Style
	statusBar    lipgloss.Style
	statusMode   lipgloss.Style
	statusPath   lipgloss.Style
	statusKey    lipgloss.Style
	statusText   lipgloss.Style
	statusErr    lipgloss.Style
	footerMeta   lipgloss.Style
	modal        lipgloss.Style
	modalTitle   lipgloss.Style
	modalDanger  lipgloss.Style
	modalSection lipgloss.Style
	backdrop     lipgloss.Style
	agentBadge   lipgloss.Style
	agentClaude  lipgloss.Style
	agentOpen    lipgloss.Style
	timeText     lipgloss.Style
	filterPill   lipgloss.Style
	status       lipgloss.Style
	fieldLabel   lipgloss.Style
	focusRing    lipgloss.Style
}

func defaultStyles() styles {
	border := lipgloss.RoundedBorder()

	return styles{
		app:          lipgloss.NewStyle().Padding(0, 1),
		frame:        lipgloss.NewStyle().BorderStyle(border).BorderForeground(adaptive("240", "252")).Padding(0, 1),
		titleBar:     lipgloss.NewStyle().Background(adaptive("17", "12")).Foreground(adaptive("15", "15")).Bold(true),
		statBar:      lipgloss.NewStyle().Background(adaptive("235", "253")),
		statBlock:    lipgloss.NewStyle().Background(adaptive("236", "254")).MarginRight(1),
		title:        lipgloss.NewStyle().Bold(true).Foreground(adaptive("15", "0")),
		subtitle:     lipgloss.NewStyle().Foreground(adaptive("14", "8")),
		headerMeta:   lipgloss.NewStyle().Foreground(adaptive("15", "15")),
		statLabel:    lipgloss.NewStyle().Foreground(adaptive("245", "8")),
		statValue:    lipgloss.NewStyle().Bold(true).Foreground(adaptive("14", "4")),
		listHeader:   lipgloss.NewStyle().Foreground(adaptive("245", "8")).Bold(true),
		row:          lipgloss.NewStyle(),
		rowTitle:     lipgloss.NewStyle().Foreground(adaptive("15", "0")),
		rowAccent:    lipgloss.NewStyle().Foreground(adaptive("6", "4")),
		rowMeta:      lipgloss.NewStyle().Foreground(adaptive("8", "8")),
		selectedRow:  lipgloss.NewStyle().Background(adaptive("17", "12")).Foreground(adaptive("15", "15")).Bold(true),
		muted:        lipgloss.NewStyle().Foreground(adaptive("8", "8")),
		errorText:    lipgloss.NewStyle().Foreground(adaptive("9", "1")).Bold(true),
		statusBar:    lipgloss.NewStyle().Background(adaptive("236", "254")),
		statusMode:   lipgloss.NewStyle().Background(adaptive("17", "12")).Foreground(adaptive("15", "15")).Bold(true).Padding(0, 1),
		statusPath:   lipgloss.NewStyle().Background(adaptive("238", "252")).Foreground(adaptive("252", "238")).Padding(0, 1),
		statusKey:    lipgloss.NewStyle().Foreground(adaptive("14", "4")).Background(adaptive("236", "254")).Bold(true),
		statusText:   lipgloss.NewStyle().Foreground(adaptive("245", "8")).Background(adaptive("236", "254")),
		statusErr:    lipgloss.NewStyle().Foreground(adaptive("9", "1")).Background(adaptive("236", "254")).Bold(true),
		footerMeta:   lipgloss.NewStyle().Foreground(adaptive("8", "8")),
		modal:        lipgloss.NewStyle().BorderStyle(border).BorderForeground(adaptive("240", "252")).Background(adaptive("236", "255")).Padding(1, 2).Width(68),
		modalTitle:   lipgloss.NewStyle().Bold(true).Foreground(adaptive("15", "0")),
		modalDanger:  lipgloss.NewStyle().Bold(true).Foreground(adaptive("9", "1")),
		modalSection: lipgloss.NewStyle().BorderLeft(true).BorderForeground(adaptive("6", "4")).PaddingLeft(1),
		backdrop:     lipgloss.NewStyle().Foreground(adaptive("8", "8")),
		agentBadge:   lipgloss.NewStyle().Foreground(adaptive("252", "238")).Background(adaptive("238", "252")).Bold(true).Width(10),
		agentClaude:  lipgloss.NewStyle().Foreground(adaptive("183", "90")).Background(adaptive("53", "225")).Bold(true).Width(10),
		agentOpen:    lipgloss.NewStyle().Foreground(adaptive("117", "23")).Background(adaptive("23", "195")).Bold(true).Width(10),
		timeText:     lipgloss.NewStyle().Foreground(adaptive("8", "8")),
		filterPill:   lipgloss.NewStyle().Foreground(adaptive("15", "15")).Background(adaptive("24", "12")).Padding(0, 1),
		status:       lipgloss.NewStyle().Foreground(adaptive("10", "2")),
		fieldLabel:   lipgloss.NewStyle().Foreground(adaptive("14", "4")).Bold(true),
		focusRing:    lipgloss.NewStyle().BorderLeft(true).BorderForeground(adaptive("13", "5")).PaddingLeft(1),
	}
}

func adaptive(dark, light string) lipgloss.AdaptiveColor {
	return lipgloss.AdaptiveColor{Dark: dark, Light: light}
}

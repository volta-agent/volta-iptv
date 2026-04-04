package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	primaryColor    = lipgloss.Color("#7C3AED")
	secondaryColor  = lipgloss.Color("#3B82F6")
	accentColor     = lipgloss.Color("#10B981")
	errorColor      = lipgloss.Color("#EF4444")
	warningColor    = lipgloss.Color("#F59E0B")
	textColor       = lipgloss.Color("#E5E7EB")
	mutedColor      = lipgloss.Color("#6B7280")
	backgroundColor = lipgloss.Color("#1F2937")
	surfaceColor    = lipgloss.Color("#374151")

	TitleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Padding(0, 1)

	TabStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Padding(0, 2)

	ActiveTabStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Padding(0, 2).
			Underline(true)

	SearchBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(0, 1).
			Margin(1, 0)

	ListItemStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Padding(0, 1)

	SelectedListItemStyle = lipgloss.NewStyle().
				Foreground(accentColor).
				Background(surfaceColor).
				Bold(true).
				Padding(0, 1)

	HelpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Padding(1, 2)

	StatusStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Background(surfaceColor).
			Padding(0, 1)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Padding(1, 2)

	FavoriteStyle = lipgloss.NewStyle().
			Foreground(warningColor)

	InfoStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Padding(0, 1)

	HeaderStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			MarginBottom(1)

	CategoryStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			MarginRight(1)

	CountryStyle = lipgloss.NewStyle().
			Foreground(textColor).
			MarginRight(1)

	QualityStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)
)

var baseStyle = lipgloss.NewStyle().
	Background(backgroundColor).
	Padding(1, 2)

func getTabNames() []string {
	return []string{
		"Channels",
		"Countries",
		"Languages",
		"Categories",
		"Guides",
	}
}

func renderTabs(activeTab int) string {
	tabNames := getTabNames()
	var renderedTabs []string
	for i, name := range tabNames {
		if i == activeTab {
			renderedTabs = append(renderedTabs, ActiveTabStyle.Render(name))
		} else {
			renderedTabs = append(renderedTabs, TabStyle.Render(name))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

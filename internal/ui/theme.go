package ui

import "charm.land/lipgloss/v2"

type Theme struct {
	Brand      lipgloss.Style
	Title      lipgloss.Style
	Section    lipgloss.Style
	ModalTitle lipgloss.Style
	Key        lipgloss.Style
	Heading    lipgloss.Style
	Name       lipgloss.Style
	Handle     lipgloss.Style
	Dim        lipgloss.Style
	Accent     lipgloss.Style
	Active     lipgloss.Style
	Button     lipgloss.Style
	Warn       lipgloss.Style
	Bad        lipgloss.Style
	Column     lipgloss.Style
	Box        lipgloss.Style
	Dialog     lipgloss.Style
}

func NewTheme() Theme {
	accent := lipgloss.Color("6")
	dialogBackground := lipgloss.Color("235")
	dialogForeground := lipgloss.Color("252")

	return Theme{
		Brand: lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(accent).
			Padding(0, 1).
			Bold(true),
		Title: lipgloss.NewStyle().
			Foreground(accent).
			Bold(true).
			Underline(true),
		Section: lipgloss.NewStyle().
			Foreground(accent).
			Bold(true),
		ModalTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(accent).
			Padding(0, 1).
			Bold(true),
		Key:     lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true),
		Heading: lipgloss.NewStyle().Bold(true),
		Name:    lipgloss.NewStyle().Bold(true),
		Handle:  lipgloss.NewStyle().Foreground(lipgloss.Color("7")),
		Dim:     lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		Accent:  lipgloss.NewStyle().Foreground(accent).Bold(true),
		Active:  lipgloss.NewStyle().Foreground(accent).Bold(true),
		Button: lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(accent).
			Bold(true),
		Warn: lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
		Bad:  lipgloss.NewStyle().Foreground(lipgloss.Color("1")),
		Column: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, true).
			BorderForeground(lipgloss.Color("8")).
			Padding(0, 1),
		Box: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")),
		Dialog: lipgloss.NewStyle().
			Foreground(dialogForeground).
			Background(dialogBackground).
			Border(lipgloss.DoubleBorder()).
			BorderForeground(accent).
			BorderBackground(dialogBackground),
	}
}

package console

import "github.com/charmbracelet/lipgloss"

// Styles groups the Lip Gloss styles used by the console renderer.
type Styles struct {
	Title     lipgloss.Style
	Subtitle  lipgloss.Style
	Section   lipgloss.Style
	Muted     lipgloss.Style
	Success   lipgloss.Style
	Warning   lipgloss.Style
	Danger    lipgloss.Style
	Box       lipgloss.Style
	TableHead lipgloss.Style
	TableCell lipgloss.Style
}

// NewStyles constructs the style set for the given rendering options.
func NewStyles(opts Options) Styles {
	title := lipgloss.NewStyle().Bold(true)
	subtitle := lipgloss.NewStyle().Bold(true)
	section := lipgloss.NewStyle().Bold(true).Underline(true)
	muted := lipgloss.NewStyle().Faint(true)
	success := lipgloss.NewStyle().Bold(true)
	warning := lipgloss.NewStyle().Bold(true)
	danger := lipgloss.NewStyle().Bold(true)
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)
	tableHead := lipgloss.NewStyle().Bold(true)
	tableCell := lipgloss.NewStyle()

	if opts.Color {
		title = title.Foreground(lipgloss.Color("63"))
		subtitle = subtitle.Foreground(lipgloss.Color("69"))
		section = section.Foreground(lipgloss.Color("111"))
		muted = muted.Foreground(lipgloss.Color("241"))
		success = success.Foreground(lipgloss.Color("42"))
		warning = warning.Foreground(lipgloss.Color("214"))
		danger = danger.Foreground(lipgloss.Color("196"))
		box = box.BorderForeground(lipgloss.Color("240"))
		tableHead = tableHead.Foreground(lipgloss.Color("252"))
	}

	return Styles{
		Title:     title,
		Subtitle:  subtitle,
		Section:   section,
		Muted:     muted,
		Success:   success,
		Warning:   warning,
		Danger:    danger,
		Box:       box,
		TableHead: tableHead,
		TableCell: tableCell,
	}
}

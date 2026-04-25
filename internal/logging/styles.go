package logging

import "github.com/charmbracelet/lipgloss"

// Styles controls pretty development log rendering.
type Styles struct {
	Time                lipgloss.Style
	Info                lipgloss.Style
	Warn                lipgloss.Style
	Error               lipgloss.Style
	Event               lipgloss.Style
	ComparisonStarted   lipgloss.Style
	ComparisonCompleted lipgloss.Style
	TaskStarted         lipgloss.Style
	TaskCompleted       lipgloss.Style
	RunStarted          lipgloss.Style
	RunExecuted         lipgloss.Style
	RunScored           lipgloss.Style
	RunFailed           lipgloss.Style
	ReportCreated       lipgloss.Style
	Muted               lipgloss.Style
	RoleBase            lipgloss.Style
	RoleCand            lipgloss.Style
	Success             lipgloss.Style
	Failure             lipgloss.Style
	Metric              lipgloss.Style
}

// NewStyles constructs the style set for pretty development logs.
func NewStyles(color bool) Styles {
	if !color {
		return Styles{
			Time:                lipgloss.NewStyle(),
			Info:                lipgloss.NewStyle(),
			Warn:                lipgloss.NewStyle(),
			Error:               lipgloss.NewStyle(),
			Event:               lipgloss.NewStyle(),
			ComparisonStarted:   lipgloss.NewStyle(),
			ComparisonCompleted: lipgloss.NewStyle(),
			TaskStarted:         lipgloss.NewStyle(),
			TaskCompleted:       lipgloss.NewStyle(),
			RunStarted:          lipgloss.NewStyle(),
			RunExecuted:         lipgloss.NewStyle(),
			RunScored:           lipgloss.NewStyle(),
			RunFailed:           lipgloss.NewStyle(),
			ReportCreated:       lipgloss.NewStyle(),
			Muted:               lipgloss.NewStyle(),
			RoleBase:            lipgloss.NewStyle(),
			RoleCand:            lipgloss.NewStyle(),
			Success:             lipgloss.NewStyle(),
			Failure:             lipgloss.NewStyle(),
			Metric:              lipgloss.NewStyle(),
		}
	}

	return Styles{
		Time:                lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
		Info:                lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true),
		Warn:                lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true),
		Error:               lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true),
		Event:               lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Bold(true),
		ComparisonStarted:   lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true),
		ComparisonCompleted: lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true),
		TaskStarted:         lipgloss.NewStyle().Foreground(lipgloss.Color("75")).Bold(true),
		TaskCompleted:       lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true),
		RunStarted:          lipgloss.NewStyle().Foreground(lipgloss.Color("219")).Bold(true),
		RunExecuted:         lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true),
		RunScored:           lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true),
		RunFailed:           lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true),
		ReportCreated:       lipgloss.NewStyle().Foreground(lipgloss.Color("141")).Bold(true),
		Muted:               lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		RoleBase:            lipgloss.NewStyle().Foreground(lipgloss.Color("81")).Bold(true),
		RoleCand:            lipgloss.NewStyle().Foreground(lipgloss.Color("121")).Bold(true),
		Success:             lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true),
		Failure:             lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true),
		Metric:              lipgloss.NewStyle().Foreground(lipgloss.Color("220")),
	}
}

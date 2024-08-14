package cmd

import "github.com/charmbracelet/lipgloss"

const (
	purple      = lipgloss.Color("#BF40BF")
	lightPurple = lipgloss.Color("#800080")
	white       = lipgloss.Color("#FFFFFF")
	gray        = lipgloss.Color("#808080")
	lightGray   = lipgloss.Color("#D3D3D3")
	red         = lipgloss.Color("#FF0000")
	green       = lipgloss.Color("#00FF00")
)

var (
	HeaderStyle   = lipgloss.NewStyle().Foreground(purple).Bold(true).Underline(true).Align(lipgloss.Left).Margin(0, 0)
	CellStyle     = lipgloss.NewStyle().Padding(0, 0)
	OddRowStyle   = CellStyle.Foreground(lightGray)
	EvenRowStyle  = CellStyle.Foreground(gray)
	UpToDateStyle = CellStyle.Foreground(green)
	OutdatedStyle = CellStyle.Foreground(red)

	PurpleStyle      = lipgloss.NewStyle().Foreground(purple)
	LightPurpleStyle = lipgloss.NewStyle().Foreground(lightPurple)
	WhiteStyle       = lipgloss.NewStyle().Foreground(white)
)

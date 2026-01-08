// Package output provides terminal output formatting using lipgloss.
package output

import (
	"github.com/charmbracelet/lipgloss"
)

// Colors used throughout the application.
var (
	colorPrimary   = lipgloss.Color("39")  // Bright blue
	colorSuccess   = lipgloss.Color("42")  // Green
	colorError     = lipgloss.Color("196") // Red
	colorWarning   = lipgloss.Color("214") // Orange
	colorMuted     = lipgloss.Color("245") // Gray
	colorHighlight = lipgloss.Color("177") // Purple
)

// Styles for different output elements.
var (
	// Header styles for major sections
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			BorderStyle(lipgloss.DoubleBorder()).
			BorderForeground(colorPrimary).
			Padding(0, 1)

	// Step header style
	stepHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colorMuted).
			Padding(0, 1)

	// Success style
	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorSuccess)

	// Error style
	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorError)

	// Muted style for secondary information
	mutedStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	// Label style
	labelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorHighlight)

	// Tool name style
	toolNameStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWarning)

	// Divider style
	dividerStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	// Summary box style
	summaryStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.DoubleBorder()).
			BorderForeground(colorPrimary).
			Padding(0, 1)

	// Queue header style
	queueHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorHighlight).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(colorHighlight).
				Padding(0, 1)
)

// Icons used in output.
const (
	iconSuccess    = "✓"
	iconError      = "✗"
	iconPending    = "○"
	iconInProgress = "●"
	iconTool       = "┌─"
	iconToolEnd    = "└─"
	iconToolLine   = "│"
)

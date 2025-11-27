package tui

import (
	"fmt"
	"strings"

	"github.com/bayhaqi/kv/pkg/keyvault"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Styles
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Padding(0, 1)

	secretNameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FBBF24")).
			Bold(true)

	versionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4"))

	latestBadgeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#10B981")).
				Bold(true)

	lineNumStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Width(4).
			Align(lipgloss.Right)
)

// Model represents the TUI model
type Model struct {
	versions   []keyvault.SecretVersion
	secretName string
	currentIdx int
	viewport   viewport.Model
	ready      bool
	width      int
	height     int
}

// NewModel creates a new TUI model
func NewModel(versions []keyvault.SecretVersion, secretName string) Model {
	return Model{
		versions:   versions,
		secretName: secretName,
		currentIdx: 0,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "ctrl+c":
			return m, tea.Quit
		case "left", "h":
			if m.currentIdx > 0 {
				m.currentIdx--
				m.updateViewportContent()
			}
			return m, nil
		case "right", "l":
			if m.currentIdx < len(m.versions)-1 {
				m.currentIdx++
				m.updateViewportContent()
			}
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		footerHeight := 3 // Footer with secret name, version, and help

		if !m.ready {
			m.viewport = viewport.New(msg.Width-4, msg.Height-footerHeight-2) // -4 for box border, -2 for box padding
			m.ready = true
			m.updateViewportContent()
		} else {
			m.viewport.Width = msg.Width - 4
			m.viewport.Height = msg.Height - footerHeight - 2
			m.updateViewportContent()
		}
		return m, nil
	}

	// Handle viewport updates for scrolling
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// updateViewportContent updates the viewport with the current version details
func (m *Model) updateViewportContent() {
	if len(m.versions) == 0 {
		return
	}

	version := m.versions[m.currentIdx]

	// Just display the secret value with line numbers
	maxWidth := m.viewport.Width - 8 // Account for line numbers and padding
	if maxWidth < 20 {
		maxWidth = 20
	}
	wrappedValue := wrapTextWithLineNumbers(version.Value, maxWidth)

	m.viewport.SetContent(wrappedValue)
	m.viewport.GotoTop()
}

// wrapTextWithLineNumbers wraps text preserving \n and adds line numbers
func wrapTextWithLineNumbers(text string, width int) string {
	if width <= 0 {
		width = 40
	}

	var result strings.Builder
	lines := strings.Split(text, "\n")
	lineNum := 1

	for _, line := range lines {
		if line == "" {
			// Empty line, just add line number
			result.WriteString(lineNumStyle.Render(fmt.Sprintf("%d", lineNum)))
			result.WriteString(" │ \n")
			lineNum++
			continue
		}

		// Wrap long lines
		wrappedLines := wrapLine(line, width)
		for i, wrappedLine := range wrappedLines {
			if i == 0 {
				// First line of wrapped content gets the line number
				result.WriteString(lineNumStyle.Render(fmt.Sprintf("%d", lineNum)))
				result.WriteString(" │ ")
			} else {
				// Continuation lines get empty space
				result.WriteString(lineNumStyle.Render(""))
				result.WriteString(" │ ")
			}
			result.WriteString(wrappedLine)
			result.WriteString("\n")
		}
		lineNum++
	}

	return strings.TrimRight(result.String(), "\n")
}

// wrapLine wraps a single line to the specified width
func wrapLine(line string, width int) []string {
	if len(line) <= width {
		return []string{line}
	}

	var wrapped []string
	remaining := line

	for len(remaining) > 0 {
		if len(remaining) <= width {
			wrapped = append(wrapped, remaining)
			break
		}

		// Find a good breaking point (space) before width
		breakPoint := width
		for i := width; i > width-20 && i > 0; i-- {
			if i < len(remaining) && remaining[i] == ' ' {
				breakPoint = i
				break
			}
		}

		// If no space found, just break at width
		if breakPoint >= len(remaining) {
			breakPoint = width
		}

		wrapped = append(wrapped, remaining[:breakPoint])
		remaining = strings.TrimLeft(remaining[breakPoint:], " ")
	}

	return wrapped
}

// View renders the TUI
func (m Model) View() string {
	if len(m.versions) == 0 {
		return "No versions available.\n"
	}

	if !m.ready {
		return "\n  Initializing..."
	}

	// Build the content box with viewport
	content := boxStyle.
		Width(m.width - 2).
		Height(m.height - 4).
		Render(m.viewport.View())

	// Build footer with secret name and version
	versionName := m.versions[m.currentIdx].Version
	if len(versionName) > 8 {
		versionName = versionName[:8]
	}

	// Check if this is the latest version (index 0)
	latestBadge := ""
	if m.currentIdx == 0 {
		latestBadge = latestBadgeStyle.Render(" [latest]")
	}

	footer := footerStyle.Render(
		fmt.Sprintf("%s • %s (%d/%d)%s",
			secretNameStyle.Render(m.secretName),
			versionStyle.Render(versionName),
			m.currentIdx+1,
			len(m.versions),
			latestBadge,
		),
	)

	// Help text
	help := footerStyle.Render("← → Navigate • ↑↓ Scroll • ESC/Q Quit")

	// Combine all parts
	return fmt.Sprintf("%s\n%s\n%s", content, footer, help)
}

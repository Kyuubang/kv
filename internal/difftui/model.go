package difftui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Styles
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1)

	leftBoxStyle = boxStyle.Copy().
			BorderForeground(lipgloss.Color("#EF4444"))

	rightBoxStyle = boxStyle.Copy().
			BorderForeground(lipgloss.Color("#10B981"))

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1)

	leftTitleStyle = titleStyle.Copy().
			Foreground(lipgloss.Color("#EF4444"))

	rightTitleStyle = titleStyle.Copy().
			Foreground(lipgloss.Color("#10B981"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Padding(0, 1)

	lineNumStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Width(4).
			Align(lipgloss.Right)

	removedLineStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#3D1E1E")).
				Foreground(lipgloss.Color("#FF6B6B"))

	addedLineStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1E3D1E")).
			Foreground(lipgloss.Color("#69DB7C"))

	unchangedLineStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F3F4F6"))
)

type diffLine struct {
	lineNum int
	content string
	isDiff  bool
	isAdded bool
}

// Model represents the diff TUI model
type Model struct {
	oldValue      string
	newValue      string
	secretName    string
	leftViewport  viewport.Model
	rightViewport viewport.Model
	ready         bool
	width         int
	height        int
	confirmed     bool
	cancelled     bool
}

// NewModel creates a new diff TUI model
func NewModel(oldValue, newValue, secretName string) Model {
	return Model{
		oldValue:   oldValue,
		newValue:   newValue,
		secretName: secretName,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "n", "N":
			m.cancelled = true
			return m, tea.Quit
		case "y", "Y", "enter":
			m.confirmed = true
			return m, tea.Quit
		case "up", "k":
			var cmd tea.Cmd
			m.leftViewport, cmd = m.leftViewport.Update(msg)
			m.rightViewport.SetYOffset(m.leftViewport.YOffset)
			return m, cmd
		case "down", "j":
			var cmd tea.Cmd
			m.leftViewport, cmd = m.leftViewport.Update(msg)
			m.rightViewport.SetYOffset(m.leftViewport.YOffset)
			return m, cmd
		case "pgup", "ctrl+b":
			var cmd tea.Cmd
			m.leftViewport, cmd = m.leftViewport.Update(msg)
			m.rightViewport.SetYOffset(m.leftViewport.YOffset)
			return m, cmd
		case "pgdown", "ctrl+f":
			var cmd tea.Cmd
			m.leftViewport, cmd = m.leftViewport.Update(msg)
			m.rightViewport.SetYOffset(m.leftViewport.YOffset)
			return m, cmd
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := 3 // Title
		footerHeight := 3 // Help text

		viewportWidth := (msg.Width / 2) - 4
		viewportHeight := msg.Height - headerHeight - footerHeight

		if !m.ready {
			m.leftViewport = viewport.New(viewportWidth, viewportHeight)
			m.rightViewport = viewport.New(viewportWidth, viewportHeight)
			m.ready = true
			m.updateViewportContent()
		} else {
			m.leftViewport.Width = viewportWidth
			m.leftViewport.Height = viewportHeight
			m.rightViewport.Width = viewportWidth
			m.rightViewport.Height = viewportHeight
			m.updateViewportContent()
		}
		return m, nil
	}

	return m, nil
}

// updateViewportContent updates both viewports with content
func (m *Model) updateViewportContent() {
	maxWidth := m.leftViewport.Width - 10

	oldLines := strings.Split(m.oldValue, "\n")
	newLines := strings.Split(m.newValue, "\n")

	leftDiff, rightDiff := computeDiff(oldLines, newLines)

	oldContent := renderDiffLines(leftDiff, maxWidth, true)
	newContent := renderDiffLines(rightDiff, maxWidth, false)

	m.leftViewport.SetContent(oldContent)
	m.rightViewport.SetContent(newContent)
}

// computeDiff compares two sets of lines and marks differences
func computeDiff(oldLines, newLines []string) ([]diffLine, []diffLine) {
	var leftDiff, rightDiff []diffLine

	maxLen := len(oldLines)
	if len(newLines) > maxLen {
		maxLen = len(newLines)
	}

	for i := 0; i < maxLen; i++ {
		oldLine := ""
		newLine := ""

		if i < len(oldLines) {
			oldLine = oldLines[i]
		}
		if i < len(newLines) {
			newLine = newLines[i]
		}

		isDifferent := oldLine != newLine

		// Left side (old)
		if i < len(oldLines) {
			leftDiff = append(leftDiff, diffLine{
				lineNum: i + 1,
				content: oldLine,
				isDiff:  isDifferent, // Mark as diff if line is different or removed
				isAdded: false,
			})
		} else {
			// Line doesn't exist in old (was added in new)
			leftDiff = append(leftDiff, diffLine{
				lineNum: 0,
				content: "",
				isDiff:  false,
				isAdded: false,
			})
		}

		// Right side (new)
		if i < len(newLines) {
			rightDiff = append(rightDiff, diffLine{
				lineNum: i + 1,
				content: newLine,
				isDiff:  isDifferent && oldLine != "",
				isAdded: oldLine == "",
			})
		} else {
			// Line doesn't exist in new (was removed)
			rightDiff = append(rightDiff, diffLine{
				lineNum: 0,
				content: "",
				isDiff:  false,
				isAdded: false,
			})
		}
	}

	return leftDiff, rightDiff
}

// renderDiffLines renders diff lines with appropriate styling
func renderDiffLines(lines []diffLine, width int, isLeft bool) string {
	var result strings.Builder

	for _, line := range lines {
		if line.lineNum == 0 {
			// Empty line placeholder
			result.WriteString(lineNumStyle.Render("    "))
			result.WriteString(" │ \n")
			continue
		}

		wrappedLines := wrapLine(line.content, width)
		for i, wrappedContent := range wrappedLines {
			// Line number only on first wrapped line
			if i == 0 {
				result.WriteString(lineNumStyle.Render(fmt.Sprintf("%d", line.lineNum)))
				if isLeft && line.isDiff {
					result.WriteString(removedLineStyle.Render(" - "))
				} else if !isLeft && line.isAdded {
					result.WriteString(addedLineStyle.Render(" + "))
				} else if !isLeft && line.isDiff {
					result.WriteString(addedLineStyle.Render(" ~ "))
				} else {
					result.WriteString(" │ ")
				}
			} else {
				// Continuation lines
				result.WriteString(lineNumStyle.Render(""))
				result.WriteString("   ")
			}

			// Apply styling to content
			if line.isDiff && isLeft {
				result.WriteString(removedLineStyle.Render(wrappedContent))
			} else if (line.isDiff || line.isAdded) && !isLeft {
				result.WriteString(addedLineStyle.Render(wrappedContent))
			} else {
				result.WriteString(unchangedLineStyle.Render(wrappedContent))
			}

			result.WriteString("\n")
		}
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

		breakPoint := width
		for i := width; i > width-20 && i > 0; i-- {
			if i < len(remaining) && remaining[i] == ' ' {
				breakPoint = i
				break
			}
		}

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
	if !m.ready {
		return "\n  Initializing..."
	}

	// Calculate box dimensions
	boxWidth := (m.width / 2) - 2
	boxHeight := m.height - 6

	// Left side (old version)
	leftTitle := leftTitleStyle.Render("Previous Version")
	leftBox := leftBoxStyle.
		Width(boxWidth).
		Height(boxHeight).
		Render(m.leftViewport.View())

	// Right side (new version)
	rightTitle := rightTitleStyle.Render("New Version")
	rightBox := rightBoxStyle.
		Width(boxWidth).
		Height(boxHeight).
		Render(m.rightViewport.View())

	// Combine side by side
	leftColumn := lipgloss.JoinVertical(lipgloss.Left, leftTitle, leftBox)
	rightColumn := lipgloss.JoinVertical(lipgloss.Left, rightTitle, rightBox)
	content := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightColumn)

	// Footer
	footer := footerStyle.Render(
		fmt.Sprintf("Secret: %s", m.secretName),
	)
	help := footerStyle.Render("↑↓ Scroll • Y/Enter Confirm • N/ESC Cancel")

	return fmt.Sprintf("%s\n%s\n%s", content, footer, help)
}

// Confirmed returns whether the user confirmed the changes
func (m Model) Confirmed() bool {
	return m.confirmed
}

// Cancelled returns whether the user cancelled
func (m Model) Cancelled() bool {
	return m.cancelled
}

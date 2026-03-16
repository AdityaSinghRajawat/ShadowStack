package ui

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FFAA")).
			MarginBottom(1).
			Underline(true)
	pidStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(12)
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).MarginTop(1)
)

// highlightSQL applies a Monokai color theme to the SQL string
func highlightSQL(query string) string {
	var buf bytes.Buffer
	// Lexer: sql, Formatter: terminal256, Theme: monokai
	err := quick.Highlight(&buf, query, "sql", "terminal256", "monokai")
	if err != nil {
		return query // Fallback to plain text if parsing fails
	}

	// Chroma adds a newline at the end by default, let's trim it
	return strings.TrimSpace(buf.String())
}

func (m Model) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("🕵️‍♂️ ShadowStack Live SQL Traffic"))
	b.WriteString("\n")

	if len(m.queries) == 0 {
		b.WriteString("Listening for database traffic... (Send a query!)\n")
	} else {
		for _, q := range m.queries {
			pid := pidStyle.Render(fmt.Sprintf("[PID: %d]", q.PID))

			// Pass the query through our new syntax highlighter
			highlightedQuery := highlightSQL(q.Query)

			b.WriteString(fmt.Sprintf("%s %s\n", pid, highlightedQuery))
		}
	}

	b.WriteString(helpStyle.Render("\nPress 'q' or 'ctrl+c' to exit"))
	return b.String()
}

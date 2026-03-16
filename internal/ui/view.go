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

	// Custom badge styles for our Factory databases
	postgresStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#336791")).Bold(true).Width(12)
	mysqlStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#E6873C")).Bold(true).Width(12)
	mongoStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#4DB33D")).Bold(true).Width(12)
	redisStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#D82C20")).Bold(true).Width(12)
	defaultDBStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).Bold(true).Width(12)

	// Fixed width for the Process & PID column
	pidStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(16)
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).MarginTop(1)
)

func getDBStyle(dbType string) lipgloss.Style {
	switch dbType {
	case "PostgreSQL":
		return postgresStyle
	case "MySQL":
		return mysqlStyle
	case "MongoDB":
		return mongoStyle
	case "Redis":
		return redisStyle
	default:
		return defaultDBStyle
	}
}
func highlightSQL(query string, dbType string) string {
	var buf bytes.Buffer
	lexer := "sql"
	if dbType == "MongoDB" {
		lexer = "javascript" // MongoDB queries look like JS/JSON
	}

	err := quick.Highlight(&buf, query, lexer, "terminal256", "monokai")
	if err != nil {
		return query
	}
	return strings.TrimSpace(buf.String())
}

func (m Model) View() string {
	// Don't render until Bubble Tea tells us the terminal size
	if m.width == 0 {
		return "Initializing Ghost-Trace UI..."
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render("🕵️‍♂️ ShadowStack Live Database Traffic"))
	b.WriteString("\n")

	if len(m.queries) == 0 {
		b.WriteString(
			lipgloss.NewStyle().
				Italic(true).
				Foreground(lipgloss.Color("#666666")).
				Render("Listening for traffic on all ports... (Send a query!)"),
		)
		b.WriteString("\n")
	} else {
		// Calculate exactly how much space is left for the SQL query
		// DB Col (12) + PID Col (16) + Spacing (2) = 30
		queryMaxWidth := m.width - 30
		if queryMaxWidth < 20 {
			queryMaxWidth = 20
		}

		for _, q := range m.queries {
			// 1. Render Database Badge
			dbIcon := q.DBType
			switch dbIcon {
			case "PostgreSQL":
				dbIcon = "🐘 Postgres"
			case "MySQL":
				dbIcon = "🐬 MySQL"
			case "MongoDB":
				dbIcon = "🍃 Mongo"
			case "Redis":
				dbIcon = "🔴 Redis"
			}
			dbCol := getDBStyle(q.DBType).Render(dbIcon)

			// 2. Render Process & PID
			pidText := fmt.Sprintf("[%s:%d]", q.Comm, q.PID)
			// MaxWidth gracefully truncates weirdly long process names
			pidCol := pidStyle.Copy().MaxWidth(16).Render(pidText)

			// 3. Render Syntax-Highlighted Query
			highlighted := highlightSQL(q.Query, q.DBType)
			// MaxWidth ensures the text cuts off with ... instead of wrapping
			queryCol := lipgloss.NewStyle().MaxWidth(queryMaxWidth).Render(highlighted)

			// 4. Join them into a perfect row!
			row := lipgloss.JoinHorizontal(lipgloss.Left, dbCol, " ", pidCol, " ", queryCol)
			b.WriteString(row + "\n")
		}
	}

	b.WriteString(
		helpStyle.Render(
			fmt.Sprintf("\nPress 'q' or 'ctrl+c' to exit | Window: %dx%d", m.width, m.height),
		),
	)
	return b.String()
}

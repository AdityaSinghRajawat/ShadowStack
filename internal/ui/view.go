package ui

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FFAA")).
			MarginBottom(1)

	// Metrics Dashboard Styles
	metricBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#444444")).
			Padding(0, 1).
			MarginBottom(1)

	metricValueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFAA")).Bold(true)
	metricLabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	postgresStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#336791")).Bold(true).Width(12)
	mysqlStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#E6873C")).Bold(true).Width(12)
	mongoStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#4DB33D")).Bold(true).Width(12)
	redisStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#D82C20")).Bold(true).Width(12)
	defaultDBStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).Bold(true).Width(12)

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
		lexer = "javascript"
	}
	err := quick.Highlight(&buf, query, lexer, "terminal256", "monokai")
	if err != nil {
		return query
	}
	return strings.TrimSpace(buf.String())
}

func (m Model) View() string {
	if m.width == 0 {
		return "Initializing ShadowStack UI..."
	}

	var b strings.Builder

	// 1. Title
	b.WriteString(titleStyle.Render("🕵️‍♂️ ShadowStack Live Profiler"))
	b.WriteString("\n")

	// 2. Metrics Dashboard
	stats := fmt.Sprintf(
		"%s %s  │  %s %s  │  %s %s  │  %s %s  │  🐘 %d  🐬 %d  🍃 %d  🔴 %d",
		metricLabelStyle.Render("TOTAL:"),
		metricValueStyle.Render(fmt.Sprintf("%d", m.totalQueries)),
		metricLabelStyle.Render("RPS:"),
		metricValueStyle.Render(fmt.Sprintf("%.1f", m.rps)),
		metricLabelStyle.Render("SLOWEST:"),
		metricValueStyle.Copy().
			Foreground(lipgloss.Color("#FF4444")).
			Render(m.slowest.Round(time.Millisecond/10).String()),
		metricLabelStyle.Render("FASTEST:"),
		metricValueStyle.Render(m.fastest.Round(time.Millisecond/10).String()),
		m.dbCounts["PostgreSQL"],
		m.dbCounts["MySQL"],
		m.dbCounts["MongoDB"],
		m.dbCounts["Redis"],
	)

	// Make the box stretch to terminal width
	boxWidth := m.width - 2
	if boxWidth > 0 {
		b.WriteString(metricBoxStyle.Width(boxWidth).Render(stats))
		b.WriteString("\n")
	}

	// 3. Query List
	if len(m.queries) == 0 {
		b.WriteString(
			lipgloss.NewStyle().
				Italic(true).
				Foreground(lipgloss.Color("#666666")).
				Render("Listening for traffic on interface..."),
		)
		b.WriteString("\n")
	} else {
		// Calculate space for SQL query: DB (12) + PID (16) + Latency (12) + Spaces (3) = 43
		queryMaxWidth := m.width - 45
		if queryMaxWidth < 20 {
			queryMaxWidth = 20
		}

		for _, q := range m.queries {
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
			pidCol := pidStyle.Copy().MaxWidth(16).Render(fmt.Sprintf("[%s:%d]", q.Comm, q.PID))

			// Render Latency Tag
			latencyStr := "[pending]"
			latColor := "#888888" // Gray for pending
			if q.Latency > 0 {
				latencyStr = fmt.Sprintf("[%s]", q.Latency.Round(time.Millisecond/10).String())
				if q.Latency > 50*time.Millisecond {
					latColor = "#FF4444" // Red if slow
				} else {
					latColor = "#00FFAA" // Green if fast
				}
			}
			latCol := lipgloss.NewStyle().Foreground(lipgloss.Color(latColor)).Width(12).Render(latencyStr)

			queryCol := lipgloss.NewStyle().MaxWidth(queryMaxWidth).Render(highlightSQL(q.Query, q.DBType))

			// Join them all together!
			row := lipgloss.JoinHorizontal(lipgloss.Left, dbCol, " ", pidCol, " ", latCol, " ", queryCol)
			b.WriteString(row + "\n")
		}
	}

	// 4. Footer
	b.WriteString(
		helpStyle.Render(
			fmt.Sprintf("\nPress 'q' or 'ctrl+c' to exit | Window: %dx%d", m.width, m.height),
		),
	)
	return b.String()
}

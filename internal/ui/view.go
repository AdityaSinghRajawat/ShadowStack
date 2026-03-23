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
	metricBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#444444")).
			Padding(0, 1).
			MarginBottom(1)
	metricValueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFAA")).Bold(true)
	metricLabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	postgresStyle    = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#336791")).
				Bold(true).
				Width(12)
	mysqlStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E6873C")).
			Bold(true).
			Width(12)
	mongoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4DB33D")).
			Bold(true).
			Width(12)
	redisStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D82C20")).
			Bold(true).
			Width(12)
	defaultDBStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA")).
			Bold(true).
			Width(12)
	pidStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(16)
	helpStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).MarginTop(1)
	selectedRowStyle = lipgloss.NewStyle().Background(lipgloss.Color("#444444")).Bold(true)
	modalStyle       = lipgloss.NewStyle().
				Border(lipgloss.ThickBorder()).
				BorderForeground(lipgloss.Color("#00FFAA")).
				Padding(1, 2)

	// Filter Styles
	filterLabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFAA")).Bold(true)
	filterInputStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#333333")).
				Padding(0, 1)
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

	// 1. FILTER LOGIC
	var filtered []QueryMsg
	filterLower := strings.ToLower(m.filterText)
	for _, q := range m.queries {
		if filterLower == "" ||
			strings.Contains(strings.ToLower(q.Query), filterLower) ||
			strings.Contains(strings.ToLower(q.DBType), filterLower) {
			filtered = append(filtered, q)
		}
	}

	// Calculate visible screen space
	maxRows := m.height - 10
	if maxRows < 5 {
		maxRows = 5
	}

	displayQueries := filtered
	if len(displayQueries) > maxRows {
		displayQueries = displayQueries[len(displayQueries)-maxRows:]
	}

	// Clamp selectedIndex so it doesn't crash if filter changes
	if !m.isPaused || m.selectedIndex >= len(displayQueries) {
		m.selectedIndex = len(displayQueries) - 1
	}
	if m.selectedIndex < 0 {
		m.selectedIndex = 0
	}

	// 2. HEADER
	pausedIndicator := ""
	if m.isPaused && !m.isModalOpen {
		pausedIndicator = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAA00")).
			Render(" [PAUSED]")
	}
	if m.filterText != "" {
		pausedIndicator += lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF00AA")).
			Render(" [FILTERED]")
	}

	b.WriteString(titleStyle.Render("🕵️‍♂️ ShadowStack Live Profiler" + pausedIndicator))
	b.WriteString("\n")

	stats := fmt.Sprintf(
		"%s %s  │  %s %s  │  %s %s  │  %s %s  │  🐘 %d  🐬 %d  🍃 %d  🔴 %d",
		metricLabelStyle.Render(
			"TOTAL:",
		),
		metricValueStyle.Render(fmt.Sprintf("%d", m.totalQueries)),
		metricLabelStyle.Render("RPS:"),
		metricValueStyle.Render(fmt.Sprintf("%.1f", m.rps)),
		metricLabelStyle.Render(
			"SLOWEST:",
		),
		metricValueStyle.Copy().
			Foreground(lipgloss.Color("#FF4444")).
			Render(m.slowest.Round(time.Millisecond/10).String()),
		metricLabelStyle.Render(
			"FASTEST:",
		),
		metricValueStyle.Render(m.fastest.Round(time.Millisecond/10).String()),
		m.dbCounts["PostgreSQL"],
		m.dbCounts["MySQL"],
		m.dbCounts["MongoDB"],
		m.dbCounts["Redis"],
	)

	boxWidth := m.width - 2
	if boxWidth > 0 {
		b.WriteString(metricBoxStyle.Width(boxWidth).Render(stats))
		b.WriteString("\n")
	}

	// 3. MODAL OVERLAY
	if m.isModalOpen && len(displayQueries) > 0 {
		q := displayQueries[m.selectedIndex]
		modalWidth := m.width - 6
		if modalWidth < 40 {
			modalWidth = 40
		}

		queryText := lipgloss.NewStyle().
			Width(modalWidth - 4).
			Render(highlightSQL(q.Query, q.DBType))
		details := fmt.Sprintf(
			"%s %s\n%s %s\n%s %s\n\n%s\n%s",
			metricLabelStyle.Render("Database:"),
			getDBStyle(q.DBType).Render(q.DBType),
			metricLabelStyle.Render("Latency: "),
			metricValueStyle.Render(q.Latency.String()),
			metricLabelStyle.Render(
				"Process: ",
			),
			pidStyle.Render(fmt.Sprintf("[%s:%d]", q.Comm, q.PID)),
			metricLabelStyle.Render("Full Query:"),
			queryText,
		)

		b.WriteString(modalStyle.Width(modalWidth).Render(details))
		b.WriteString("\n" + helpStyle.Render("\nPress 'Esc' or 'Enter' to close modal"))
		return b.String()
	}

	// 4. MAIN LIST
	if len(displayQueries) == 0 {
		b.WriteString(
			lipgloss.NewStyle().
				Italic(true).
				Foreground(lipgloss.Color("#666666")).
				Render("No queries match filter or waiting for traffic..."),
		)
		b.WriteString("\n")
	} else {
		queryMaxWidth := m.width - 48
		if queryMaxWidth < 20 {
			queryMaxWidth = 20
		}

		for i, q := range displayQueries {
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

			latencyStr := "[pending]"
			latColor := "#888888"
			if q.Latency > 0 {
				latencyStr = fmt.Sprintf("[%s]", q.Latency.Round(time.Millisecond/10).String())
				if q.Latency > 50*time.Millisecond {
					latColor = "#FF4444"
				} else {
					latColor = "#00FFAA"
				}
			}
			latCol := lipgloss.NewStyle().Foreground(lipgloss.Color(latColor)).Width(12).Render(latencyStr)
			queryCol := lipgloss.NewStyle().MaxWidth(queryMaxWidth).Render(highlightSQL(q.Query, q.DBType))

			cursor := "  "
			if m.isPaused && i == m.selectedIndex {
				cursor = "▶ "
			}
			cursorCol := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFAA")).Bold(true).Render(cursor)

			row := lipgloss.JoinHorizontal(lipgloss.Left, cursorCol, dbCol, " ", pidCol, " ", latCol, " ", queryCol)
			if m.isPaused && i == m.selectedIndex {
				row = selectedRowStyle.Width(m.width).Render(row)
			}
			b.WriteString(row + "\n")
		}
	}

	// 5. FOOTER & SEARCH BAR
	if m.isFiltering {
		b.WriteString(fmt.Sprintf("\n%s %s█\n%s",
			filterLabelStyle.Render("🔍 FILTER:"),
			filterInputStyle.Render(m.filterText),
			helpStyle.Render("Press 'Enter' to confirm or 'Esc' to clear"),
		))
	} else if m.isPaused {
		b.WriteString(helpStyle.Render(fmt.Sprintf("\nArrows: Navigate | Enter: Inspect | Esc: Resume | Window: %dx%d", m.width, m.height)))
	} else {
		b.WriteString(helpStyle.Render(fmt.Sprintf("\nPress 'Up' to pause | '/' to filter | 'q' to exit | Window: %dx%d", m.width, m.height)))
	}

	return b.String()
}

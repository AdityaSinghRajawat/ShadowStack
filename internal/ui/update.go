package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	case TickMsg:
		// 1. Calculate RPS (queries in the last 1 second)
		now := time.Now()
		cutoff := now.Add(-time.Second)

		var recent int
		var newQueryTimes []time.Time

		for _, t := range m.queryTimes {
			if t.After(cutoff) {
				recent++
				newQueryTimes = append(newQueryTimes, t)
			}
		}

		m.rps = float64(recent)
		m.queryTimes = newQueryTimes // Shrink the sliding window

		// 2. Schedule the next tick
		return m, tickCmd()

	case QueryMsg:
		// Update Metrics
		m.totalQueries++
		m.dbCounts[msg.DBType]++
		m.queryTimes = append(m.queryTimes, time.Now())

		m.queries = append(m.queries, msg)

		// Reserving space for the new Metrics Header (10 lines)
		maxQueries := m.height - 10
		if maxQueries < 5 {
			maxQueries = 5
		}

		if len(m.queries) > maxQueries {
			m.queries = m.queries[1:]
		}
		return m, nil
	}

	return m, nil
}

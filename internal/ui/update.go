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
		// Calculate RPS (queries in the last 1 second)
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
		m.queryTimes = newQueryTimes

		return m, tickCmd()

	case QueryMsg:
		if msg.IsUpdate {
			// Find the pending query and update its latency
			for i := len(m.queries) - 1; i >= 0; i-- {
				if m.queries[i].Port == msg.Port && m.queries[i].Latency == 0 {
					m.queries[i].Latency = msg.Latency

					// Update global Slowest/Fastest records
					if msg.Latency > m.slowest {
						m.slowest = msg.Latency
					}
					if m.fastest == 0 || msg.Latency < m.fastest {
						m.fastest = msg.Latency
					}
					break
				}
			}
			return m, nil
		}

		// New query: Update metrics and append
		m.totalQueries++
		m.dbCounts[msg.DBType]++
		m.queryTimes = append(m.queryTimes, time.Now())

		m.queries = append(m.queries, msg)

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

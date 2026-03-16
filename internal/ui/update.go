package ui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Handle keystrokes
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	// Handle incoming SQL queries from eBPF
	case QueryMsg:
		m.queries = append(m.queries, msg)
		// Keep the dashboard clean by only showing the last 15 queries
		if len(m.queries) > 15 {
			m.queries = m.queries[1:]
		}
		return m, nil
	}

	return m, nil
}

package ui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Listen for terminal resize events!
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	case QueryMsg:
		m.queries = append(m.queries, msg)

		// Dynamically calculate how many rows fit on the screen
		// Reserving 6 lines for the Header, Footer, and spacing
		maxQueries := m.height - 6
		if maxQueries < 5 {
			maxQueries = 5 // Sane minimum
		}

		// Pop the oldest query if we exceed the screen height
		if len(m.queries) > maxQueries {
			m.queries = m.queries[1:]
		}
		return m, nil
	}

	return m, nil
}

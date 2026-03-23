package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		// 1. FILTER TYPING MODE
		if m.isFiltering {
			switch msg.Type {
			case tea.KeyEnter, tea.KeyEsc:
				m.isFiltering = false // Exit filter typing mode
			case tea.KeyBackspace, tea.KeyDelete:
				if len(m.filterText) > 0 {
					m.filterText = m.filterText[:len(m.filterText)-1]
				}
			case tea.KeyRunes:
				m.filterText += string(msg.Runes)
			case tea.KeySpace:
				m.filterText += " "
			}
			return m, nil
		}

		// 2. NORMAL NAVIGATION CONTROLS
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "/":
			if !m.isModalOpen {
				m.isFiltering = true
				m.isPaused = false // Resume scrolling while searching
			}
		case "up":
			if !m.isModalOpen {
				m.isPaused = true
				if m.selectedIndex > 0 {
					m.selectedIndex--
				}
			}
		case "down":
			if !m.isModalOpen {
				// We don't know the exact filtered length here easily,
				// so we just increment and let View() clamp it.
				m.selectedIndex++
			}
		case "enter":
			if m.isPaused {
				m.isModalOpen = !m.isModalOpen
			}
		case "esc":
			m.isModalOpen = false
			m.isPaused = false
		case "s":
			// THE EXPORT FEATURE
			go func(queriesToSave []QueryMsg) {
				// We do this in a goroutine so it doesn't freeze the UI while writing
				saveToFile(queriesToSave)
			}(m.queries)
		}

	case TickMsg:
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
			for i := len(m.queries) - 1; i >= 0; i-- {
				if m.queries[i].Port == msg.Port && m.queries[i].Latency == 0 {
					m.queries[i].Latency = msg.Latency
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

		m.totalQueries++
		m.dbCounts[msg.DBType]++
		m.queryTimes = append(m.queryTimes, time.Now())

		m.queries = append(m.queries, msg)

		// HUGE ARCHITECTURE UPGRADE:
		// Keep up to 1000 queries in memory so the filter has historical data to search!
		if len(m.queries) > 1000 {
			m.queries = m.queries[1:]
		}

		return m, nil
	}

	return m, nil
}

// saveToFile dumps the current in-memory queries to a JSON file
func saveToFile(queries []QueryMsg) {
	if len(queries) == 0 {
		return
	}

	fileName := fmt.Sprintf("shadowstack_export_%d.json", time.Now().Unix())

	// Convert the queries to pretty-printed JSON
	fileData, err := json.MarshalIndent(queries, "", "  ")
	if err != nil {
		return // Silently fail on error to keep UI safe
	}

	os.WriteFile(fileName, fileData, 0644)
}

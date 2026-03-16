package ui

import tea "github.com/charmbracelet/bubbletea"

type QueryMsg struct {
	PID   uint32
	Comm  string
	Query string
}

type Model struct {
	queries []QueryMsg
}

func New() Model {
	return Model{
		queries: make([]QueryMsg, 0),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

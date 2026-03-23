package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type QueryMsg struct {
	PID    uint32
	Comm   string
	DBType string
	Query  string
}

// TickMsg is used to trigger RPS calculations every second
type TickMsg time.Time

type Model struct {
	queries []QueryMsg
	width   int
	height  int

	// Metrics State
	totalQueries int
	queryTimes   []time.Time
	rps          float64
	dbCounts     map[string]int
}

func New() Model {
	return Model{
		queries:  make([]QueryMsg, 0),
		dbCounts: make(map[string]int),
	}
}

// Init starts the ticking clock for live metrics
func (m Model) Init() tea.Cmd {
	return tickCmd()
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

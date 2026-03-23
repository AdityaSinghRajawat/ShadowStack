package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type QueryMsg struct {
	PID      uint32
	Comm     string
	DBType   string
	Query    string
	Port     uint16
	IsUpdate bool
	Latency  time.Duration
}

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

	// Latency Metrics
	slowest time.Duration
	fastest time.Duration

	// Interactive State
	selectedIndex int
	isPaused      bool
	isModalOpen   bool

	// Filter State
	isFiltering bool
	filterText  string
}

func New() Model {
	return Model{
		queries:       make([]QueryMsg, 0),
		dbCounts:      make(map[string]int),
		selectedIndex: -1,
	}
}

func (m Model) Init() tea.Cmd {
	return tickCmd()
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

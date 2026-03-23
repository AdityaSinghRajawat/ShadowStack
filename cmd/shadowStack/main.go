package main

import (
	"fmt"
	"os"

	"shadowStack/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	eventsChan := make(chan ui.QueryMsg, 100)

	// startEngine is defined differently based on the OS we compile for
	go startEngine(eventsChan)

	p := tea.NewProgram(ui.New(), tea.WithAltScreen())

	// Push events directly into BubbleTea
	go func() {
		for event := range eventsChan {
			p.Send(event)
		}
	}()

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error starting UI: %v\n", err)
		os.Exit(1)
	}
}

package main

import (
	"flag"
	"fmt"
	"os"
	"shadowStack/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	remoteTarget = flag.String("remote", "", "Remote DB address (e.g. cluster0.mongodb.net:27017)")
	localPort    = flag.String("port", "27017", "Local port to listen on for proxy")
)

func main() {
	flag.Parse()
	eventsChan := make(chan ui.QueryMsg, 100)

	if *remoteTarget != "" {
		go startProxyEngine(*localPort, *remoteTarget, eventsChan) //
	} else {
		go startEngine(eventsChan) //
	}

	p := tea.NewProgram(ui.New(), tea.WithAltScreen())
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

package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"clockscale/config"
	"clockscale/ui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}

	// Save current terminal title and set ours.
	// OSC 22 pushes the current title onto a stack; OSC 23 pops it back.
	fmt.Fprint(os.Stdout, "\033[22;0t")  // push current title
	fmt.Fprint(os.Stdout, "\033]0;Clockscale\007") // set new title
	defer fmt.Fprint(os.Stdout, "\033[23;0t") // pop (restore) on exit

	p := tea.NewProgram(
		ui.New(cfg),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error running clockscale: %v\n", err)
		os.Exit(1)
	}
}

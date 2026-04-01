package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"clockscale/config"
	"clockscale/ui"
)

const version = "1.2.0"

func main() {
	var (
		showVersion bool
		showHelp    bool
		configPath  string
	)

	flag.BoolVar(&showVersion, "version", false, "print version and exit")
	flag.BoolVar(&showVersion, "v", false, "print version and exit (shorthand)")
	flag.BoolVar(&showHelp, "help", false, "show usage information")
	flag.BoolVar(&showHelp, "h", false, "show usage information (shorthand)")
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.StringVar(&configPath, "c", "", "path to config file (shorthand)")
	flag.Usage = usage
	flag.Parse()

	if showHelp {
		usage()
		return
	}

	if showVersion {
		fmt.Println("clockscale " + version)
		return
	}

	if configPath != "" {
		config.PathOverride = configPath
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}

	// Save current terminal title and set ours.
	// OSC 22 pushes the current title onto a stack; OSC 23 pops it back.
	fmt.Fprint(os.Stdout, "\033[22;0t")           // push current title
	fmt.Fprint(os.Stdout, "\033]0;Clockscale\007") // set new title
	defer fmt.Fprint(os.Stdout, "\033[23;0t")      // pop (restore) on exit

	p := tea.NewProgram(
		ui.New(cfg),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error running clockscale: %v\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `Clockscale — view multiple timezones in a terminal grid

Usage:
  clockscale [flags]

Flags:
  -c, --config <path>   path to config file (default: ~/.config/clockscale/config.yaml)
  -v, --version         print version and exit
  -h, --help            show this help message

Keybindings:
  ←/→       scroll grid left/right
  a         add a timezone
  d         delete a timezone
  r         reload config file
  ?         toggle help overlay
  q         quit
`)
}

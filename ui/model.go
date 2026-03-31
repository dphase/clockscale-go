package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"clockscale/config"
)

type viewState int

const (
	stateNormal viewState = iota
	stateAdd
	stateDelete
)

type tickMsg time.Time

type Model struct {
	config       *config.Config
	scrollOffset int
	width        int
	height       int
	now          time.Time
	state        viewState
	input        string // timezone name being typed in add mode
	cursor       int    // selected row index in delete mode
	errMsg       string
	showHelp     bool
}

func New(cfg *config.Config) Model {
	return Model{
		config: cfg,
		now:    time.Now(),
	}
}

func (m Model) Init() tea.Cmd {
	return tickCmd()
}

func tickCmd() tea.Cmd {
	return tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

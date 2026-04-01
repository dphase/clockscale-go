package ui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"clockscale/config"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tickMsg:
		m.now = time.Time(msg)
		if cfg, err := config.Load(); err == nil {
			m.config = cfg
		}
		return m, tickCmd()
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case stateNormal:
		return m.handleNormalKey(msg)
	case stateAdd:
		return m.handleAddKey(msg)
	case stateDelete:
		return m.handleDeleteKey(msg)
	}
	return m, nil
}

func (m Model) handleNormalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.errMsg = ""
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "?", "/":
		m.showHelp = !m.showHelp
		return m, nil
	case "left":
		m.showHelp = false
		m.scrollOffset--
	case "right":
		m.showHelp = false
		m.scrollOffset++
	case "a":
		m.showHelp = false
		m.state = stateAdd
		m.input = ""
	case "d":
		m.showHelp = false
		if len(m.config.Timezones) > 0 {
			m.state = stateDelete
			m.cursor = 0
		}
	case "r":
		m.showHelp = false
		if cfg, err := config.Load(); err == nil {
			m.config = cfg
		}
	}
	return m, nil
}

func (m Model) handleAddKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		tzName := strings.TrimSpace(m.input)
		if tzName != "" {
			if _, err := time.LoadLocation(tzName); err != nil {
				m.errMsg = "Unknown timezone: " + tzName
				m.state = stateNormal
				m.input = ""
				return m, nil
			}
			m.config.Timezones = append(m.config.Timezones, config.TimezoneConfig{
				Timezone: tzName,
				Label:    tzName,
			})
			if err := config.Save(m.config); err != nil {
				m.errMsg = "failed to save config: " + err.Error()
			}
		}
		m.state = stateNormal
		m.input = ""
	case "esc":
		m.state = stateNormal
		m.input = ""
	case "backspace", "ctrl+h":
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
	default:
		if len(msg.Runes) > 0 {
			m.input += string(msg.Runes)
		}
	}
	return m, nil
}

func (m Model) handleDeleteKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.cursor >= 0 && m.cursor < len(m.config.Timezones) {
			m.config.Timezones = append(
				m.config.Timezones[:m.cursor],
				m.config.Timezones[m.cursor+1:]...,
			)
			if err := config.Save(m.config); err != nil {
				m.errMsg = "failed to save config: " + err.Error()
			}
		}
		m.state = stateNormal
	case "esc":
		m.state = stateNormal
	case "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down":
		if m.cursor < len(m.config.Timezones)-1 {
			m.cursor++
		}
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

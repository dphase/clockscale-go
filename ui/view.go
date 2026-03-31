package ui

import (
	"fmt"
	"strings"
	"time"

	colorful "github.com/lucasb-eyer/go-colorful"
	"github.com/charmbracelet/lipgloss"
)

const lightnessStep = 0.07

// shiftLightness adjusts the HSL lightness of a hex color by steps * lightnessStep.
// Positive steps = brighter, negative = darker.
func shiftLightness(hex string, steps float64) string {
	c, err := colorful.Hex(hex)
	if err != nil {
		return hex
	}
	h, s, l := c.Hsl()
	l += steps * lightnessStep
	if l > 1 {
		l = 1
	} else if l < 0 {
		l = 0
	}
	return colorful.Hsl(h, s, l).Hex()
}

const (
	cellWidth = 3
	numHours  = 24
)

func (m Model) View() string {
	if m.config == nil {
		return "Loading..."
	}

	cfg := m.config
	colors := cfg.Colors

	// Find the local timezone row index
	localIdx := -1
	for i, tz := range cfg.Timezones {
		if tz.Local {
			localIdx = i
			break
		}
	}

	// Compute label column width from the longest label
	maxLabelLen := 0
	for _, tz := range cfg.Timezones {
		if len(tz.Label) > maxLabelLen {
			maxLabelLen = len(tz.Label)
		}
	}
	labelWidth := maxLabelLen + 2

	// Anchor the window at midnight in the local timezone so column 0 = hour 0.
	utcNow := m.now.UTC()
	currentHour := utcNow.Truncate(time.Hour)

	// Find the local timezone to anchor the window
	var localLoc *time.Location
	for _, tz := range cfg.Timezones {
		if tz.Local {
			if loc, err := time.LoadLocation(tz.Timezone); err == nil {
				localLoc = loc
			}
			break
		}
	}
	if localLoc == nil {
		localLoc = time.UTC
	}
	// Start at today's midnight in the local timezone, shifted by scroll offset
	localNow := m.now.In(localLoc)
	midnight := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, localLoc)
	windowStart := midnight.UTC().Add(time.Duration(m.scrollOffset) * time.Hour)

	var rows []string

	for rowIdx, tz := range cfg.Timezones {
		loc, err := time.LoadLocation(tz.Timezone)
		if err != nil {
			continue
		}

		isLocalRow := rowIdx == localIdx

		// Label style
		labelFg := colors.DefaultTimezoneLabel.Fg
		if isLocalRow {
			labelFg = colors.LocalTimezoneLabel.Fg
		}
		labelStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(labelFg)).
			Width(labelWidth).
			Align(lipgloss.Right)

		// In delete mode, highlight the selected row's label
		if m.state == stateDelete && rowIdx == m.cursor {
			labelStyle = labelStyle.Reverse(true)
		}

		var row strings.Builder
		row.WriteString(labelStyle.Render(tz.Label))

		for col := 0; col < numHours; col++ {
			colTime := windowStart.Add(time.Duration(col) * time.Hour)
			localHour := colTime.In(loc).Hour()
			isCurrentCol := colTime.Equal(currentHour)

			var cellStyle lipgloss.Style
			switch {
			case isCurrentCol && isLocalRow:
				cellStyle = lipgloss.NewStyle().
					Background(lipgloss.Color(colors.CurrentTimeCells.Local.Bg)).
					Foreground(lipgloss.Color(colors.CurrentTimeCells.Local.Fg))
			case isCurrentCol:
				cellStyle = lipgloss.NewStyle().
					Background(lipgloss.Color(colors.CurrentTimeCells.Default.Bg)).
					Foreground(lipgloss.Color(colors.CurrentTimeCells.Default.Fg))
			default:
				bg := colors.DefaultCell.EvenBg
				if col%2 != 0 {
					bg = colors.DefaultCell.OddBg
				}
				fg := colors.DefaultCell.Fg
				if isLocalRow {
					bg = shiftLightness(bg, -1.5)
					fg = shiftLightness(fg, 2)
				}
				cellStyle = lipgloss.NewStyle().
					Background(lipgloss.Color(bg)).
					Foreground(lipgloss.Color(fg))
			}

			row.WriteString(
				cellStyle.
					Width(cellWidth).
					Align(lipgloss.Right).
					Render(fmt.Sprintf("%d", localHour)),
			)
		}

		rows = append(rows, row.String())
	}

	view := strings.Join(rows, "\n")

	// Overlay line — rendered flush against the bottom of the grid
	overlayStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#ffffff")).
		Foreground(lipgloss.Color("#000000"))

	switch m.state {
	case stateAdd:
		view += "\n" + overlayStyle.Render("Timezone: "+m.input+"_")
	case stateDelete:
		view += "\n" + overlayStyle.Render("↑/↓ select row  Enter delete  Esc cancel")
	default:
		if m.showHelp {
			helpText := "[a] add  [d] delete  [r] reload  [←/→] scroll  [q] quit  [?] hide help"
			if m.errMsg != "" {
				helpText = m.errMsg + "  —  " + helpText
			}
			view += "\n" + overlayStyle.Render(helpText)
		} else if m.errMsg != "" {
			view += "\n" + overlayStyle.Render(m.errMsg)
		}
	}

	return view
}

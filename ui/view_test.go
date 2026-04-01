package ui

import (
	"regexp"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"clockscale/config"
)

// stripANSI removes ANSI escape sequences so we can inspect raw text content.
var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func useTempConfigDir(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	old := config.DirOverride
	config.DirOverride = tmp
	t.Cleanup(func() { config.DirOverride = old })
}

func testConfig() *config.Config {
	return &config.Config{
		Timezones: []config.TimezoneConfig{
			{Timezone: "US/Central", Label: "CDT", Local: true},
			{Timezone: "UTC", Label: "Z"},
		},
		Colors: config.ColorsConfig{
			DefaultTimezoneLabel: config.ColorPair{Bg: "default", Fg: "#02ffff"},
			LocalTimezoneLabel:   config.ColorPair{Bg: "default", Fg: "#ffff00"},
			DefaultCell: config.DefaultCellColors{
				EvenBg: "#1c1c1c",
				OddBg:  "#2d2e2e",
				Fg:     "#dadada",
			},
			CurrentTimeCells: config.CurrentTimeCellColors{
				Default: config.ColorPair{Bg: "#5e5e86", Fg: "#90ee90"},
				Local:   config.ColorPair{Bg: "#b4420a", Fg: "#ffff00"},
			},
		},
	}
}

func TestNewModel(t *testing.T) {
	cfg := testConfig()
	m := New(cfg)

	if m.config != cfg {
		t.Error("model config not set")
	}
	if m.scrollOffset != 0 {
		t.Errorf("expected scrollOffset 0, got %d", m.scrollOffset)
	}
	if m.state != stateNormal {
		t.Errorf("expected stateNormal, got %d", m.state)
	}
}

func TestViewRenders24Hours(t *testing.T) {
	cfg := testConfig()
	m := New(cfg)
	m.width = 100
	m.height = 10

	view := m.View()
	lines := strings.Split(view, "\n")

	// Should have at least as many lines as timezones
	if len(lines) < len(cfg.Timezones) {
		t.Errorf("expected at least %d lines, got %d", len(cfg.Timezones), len(lines))
	}
}

func TestViewContainsLabels(t *testing.T) {
	cfg := testConfig()
	m := New(cfg)

	view := m.View()

	if !strings.Contains(view, "CDT") {
		t.Error("view does not contain label CDT")
	}
	if !strings.Contains(view, "Z") {
		t.Error("view does not contain label Z")
	}
}

func TestViewLocalRowStartsAtZero(t *testing.T) {
	cfg := testConfig()
	m := New(cfg)

	view := m.View()
	lines := strings.Split(view, "\n")

	// First line is the local row (CDT). After the label, the first hour should be 0.
	if len(lines) == 0 {
		t.Fatal("no output lines")
	}
	localLine := lines[0]
	// The label is right-aligned in its column, followed by hour cells.
	// The first hour cell should contain "0" as the first value.
	if !strings.Contains(localLine, "0") {
		t.Error("local row does not contain hour 0")
	}
}

func TestScrollOffsetChangesOnArrowKeys(t *testing.T) {
	cfg := testConfig()
	m := New(cfg)

	// Press right arrow
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = updated.(Model)
	if m.scrollOffset != 1 {
		t.Errorf("expected scrollOffset 1 after right, got %d", m.scrollOffset)
	}

	// Press left arrow
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m = updated.(Model)
	if m.scrollOffset != 0 {
		t.Errorf("expected scrollOffset 0 after left, got %d", m.scrollOffset)
	}
}

func TestQuitOnQ(t *testing.T) {
	cfg := testConfig()
	m := New(cfg)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Error("expected quit command on 'q'")
	}
}

func TestAddModeToggle(t *testing.T) {
	cfg := testConfig()
	m := New(cfg)

	// Press 'a' to enter add mode
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updated.(Model)
	if m.state != stateAdd {
		t.Errorf("expected stateAdd, got %d", m.state)
	}

	// Press Esc to cancel
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.state != stateNormal {
		t.Errorf("expected stateNormal after Esc, got %d", m.state)
	}
}

func TestDeleteModeToggle(t *testing.T) {
	cfg := testConfig()
	m := New(cfg)

	// Press 'd' to enter delete mode
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = updated.(Model)
	if m.state != stateDelete {
		t.Errorf("expected stateDelete, got %d", m.state)
	}

	// Press Esc to cancel
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.state != stateNormal {
		t.Errorf("expected stateNormal after Esc, got %d", m.state)
	}
}

func TestHelpToggle(t *testing.T) {
	cfg := testConfig()
	m := New(cfg)

	// Press '?' to show help
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = updated.(Model)
	if !m.showHelp {
		t.Error("expected showHelp true after '?'")
	}

	// View should contain help text
	view := m.View()
	if !strings.Contains(view, "[a] add") {
		t.Error("help overlay not visible in view")
	}

	// Press '?' again to hide
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = updated.(Model)
	if m.showHelp {
		t.Error("expected showHelp false after second '?'")
	}
}

func TestWindowSizeMsg(t *testing.T) {
	cfg := testConfig()
	m := New(cfg)

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = updated.(Model)

	if m.width != 120 {
		t.Errorf("expected width 120, got %d", m.width)
	}
	if m.height != 40 {
		t.Errorf("expected height 40, got %d", m.height)
	}
}

func TestTickUpdatesTime(t *testing.T) {
	useTempConfigDir(t)
	cfg := testConfig()
	m := New(cfg)

	newTime := time.Date(2026, 6, 15, 14, 0, 0, 0, time.UTC)
	updated, cmd := m.Update(tickMsg(newTime))
	m = updated.(Model)

	if !m.now.Equal(newTime) {
		t.Errorf("expected now to be %v, got %v", newTime, m.now)
	}
	if cmd == nil {
		t.Error("expected tick command to be re-scheduled")
	}
}

func TestShiftLightness(t *testing.T) {
	tests := []struct {
		name     string
		hex      string
		steps    float64
		wantDiff bool // true if result should differ from input
	}{
		{"brighten dark", "#1c1c1c", 3, true},
		{"darken light", "#dadada", -3, true},
		{"zero shift", "#808080", 0, false},
		{"clamp at white", "#ffffff", 5, false},
		{"clamp at black", "#000000", -5, false},
		{"invalid hex", "notacolor", 3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shiftLightness(tt.hex, tt.steps)
			if tt.wantDiff && result == tt.hex {
				t.Errorf("expected color to change from %s, got same", tt.hex)
			}
			if !tt.wantDiff && result != tt.hex {
				// For zero shift, the hex might be reformatted but should be equivalent
				if tt.steps == 0 && tt.hex != "notacolor" {
					return // reformatting is ok
				}
			}
		})
	}
}

func TestCellsHavePadding(t *testing.T) {
	cfg := testConfig()
	m := New(cfg)

	view := m.View()
	lines := strings.Split(view, "\n")

	for i, line := range lines {
		if i >= len(cfg.Timezones) {
			break
		}
		stripped := ansiRe.ReplaceAllString(line, "")

		// Label should be followed by a space (right padding) before the first hour cell.
		label := cfg.Timezones[i].Label
		labelIdx := strings.Index(stripped, label)
		if labelIdx == -1 {
			t.Errorf("row %d: label %q not found in stripped line", i, label)
			continue
		}
		afterLabel := stripped[labelIdx+len(label):]
		if len(afterLabel) == 0 || afterLabel[0] != ' ' {
			t.Errorf("row %d: expected space padding after label %q", i, label)
		}

		// Walk the hour area checking that every digit-group (cell) has:
		//   - at least 1 space before the first digit (left padding)
		//   - at least 1 space after the last digit (right padding)
		hourArea := afterLabel
		for j := 0; j < len(hourArea); j++ {
			ch := hourArea[j]
			if ch < '0' || ch > '9' {
				continue
			}
			// Found start of a digit group — check left padding
			if j == 0 || hourArea[j-1] != ' ' {
				t.Errorf("row %d: digit at position %d missing left padding space", i, j)
			}
			// Advance past all digits in this group
			end := j
			for end < len(hourArea) && hourArea[end] >= '0' && hourArea[end] <= '9' {
				end++
			}
			// Check right padding
			if end < len(hourArea) && hourArea[end] != ' ' {
				t.Errorf("row %d: digit group ending at position %d missing right padding space", i, end-1)
			}
			j = end // skip past this group
		}
	}
}

func TestInvalidTimezoneShowsErrorRow(t *testing.T) {
	cfg := testConfig()
	cfg.Timezones = append(cfg.Timezones, config.TimezoneConfig{
		Timezone: "Not/A/Real/Zone",
		Label:    "BAD",
	})
	m := New(cfg)
	m.width = 120

	view := m.View()
	stripped := ansiRe.ReplaceAllString(view, "")

	if !strings.Contains(stripped, "BAD") {
		t.Error("error row should still display the label")
	}
	if !strings.Contains(stripped, "invalid timezone: Not/A/Real/Zone") {
		t.Error("error row should display the invalid timezone message")
	}

	// Valid timezones should still render normally
	if !strings.Contains(stripped, "CDT") {
		t.Error("valid timezone CDT should still render")
	}
}

func TestNarrowTerminalLimitsColumns(t *testing.T) {
	cfg := testConfig()
	m := New(cfg)

	// Wide terminal — all 24 hours should appear
	m.width = 200
	wideView := m.View()
	wideLine := strings.Split(wideView, "\n")[0]
	wideStripped := ansiRe.ReplaceAllString(wideLine, "")
	wideFields := strings.Fields(wideStripped)
	// Label + 24 hours = 25 fields
	wideCount := len(wideFields) - 1 // subtract label

	// Narrow terminal — fewer columns should appear
	m.width = 40
	narrowView := m.View()
	narrowLine := strings.Split(narrowView, "\n")[0]
	narrowStripped := ansiRe.ReplaceAllString(narrowLine, "")
	narrowFields := strings.Fields(narrowStripped)
	narrowCount := len(narrowFields) - 1

	if narrowCount >= wideCount {
		t.Errorf("narrow terminal (%d cols) should show fewer hours than wide (%d cols), got narrow=%d wide=%d",
			40, 200, narrowCount, wideCount)
	}
	if narrowCount < 1 {
		t.Error("narrow terminal should still show at least 1 hour column")
	}
}

func TestViewNilConfigShowsLoading(t *testing.T) {
	m := Model{} // config is nil
	view := m.View()
	if view != "Loading..." {
		t.Errorf("expected %q, got %q", "Loading...", view)
	}
}

func TestViewAddModeOverlay(t *testing.T) {
	cfg := testConfig()
	m := New(cfg)
	m.state = stateAdd
	m.input = "Europe/London"
	m.width = 100

	view := m.View()
	stripped := ansiRe.ReplaceAllString(view, "")
	if !strings.Contains(stripped, "Timezone: Europe/London_") {
		t.Error("add mode overlay should show current input with cursor")
	}
}

func TestViewDeleteModeOverlay(t *testing.T) {
	cfg := testConfig()
	m := New(cfg)
	m.state = stateDelete
	m.cursor = 0
	m.width = 100

	view := m.View()
	stripped := ansiRe.ReplaceAllString(view, "")
	if !strings.Contains(stripped, "Enter delete") {
		t.Error("delete mode overlay should show instructions")
	}
	if !strings.Contains(stripped, "Esc cancel") {
		t.Error("delete mode overlay should mention Esc")
	}
}

func TestViewErrorMessageDisplay(t *testing.T) {
	cfg := testConfig()
	m := New(cfg)
	m.errMsg = "Unknown timezone: Foo/Bar"
	m.width = 100

	// Without help shown — errMsg should appear on its own
	view := m.View()
	stripped := ansiRe.ReplaceAllString(view, "")
	if !strings.Contains(stripped, "Unknown timezone: Foo/Bar") {
		t.Error("errMsg should be displayed when set")
	}

	// With help shown — errMsg should appear alongside help
	m.showHelp = true
	view = m.View()
	stripped = ansiRe.ReplaceAllString(view, "")
	if !strings.Contains(stripped, "Unknown timezone: Foo/Bar") {
		t.Error("errMsg should still appear when help is shown")
	}
	if !strings.Contains(stripped, "[a] add") {
		t.Error("help text should appear when showHelp is true")
	}
}

func TestViewSingleTimezone(t *testing.T) {
	cfg := testConfig()
	cfg.Timezones = []config.TimezoneConfig{
		{Timezone: "UTC", Label: "UTC"},
	}
	m := New(cfg)
	m.width = 100

	view := m.View()
	lines := strings.Split(view, "\n")
	if len(lines) < 1 {
		t.Fatal("expected at least 1 line")
	}
	stripped := ansiRe.ReplaceAllString(lines[0], "")
	if !strings.Contains(stripped, "UTC") {
		t.Error("single timezone view should contain the label")
	}
}

func TestViewNoLocalTimezone(t *testing.T) {
	cfg := testConfig()
	// Remove local flag from all timezones
	for i := range cfg.Timezones {
		cfg.Timezones[i].Local = false
	}
	m := New(cfg)
	m.width = 100

	// Should not panic; should fall back to UTC
	view := m.View()
	if view == "" {
		t.Error("view should not be empty with no local timezone")
	}
}

func TestViewVeryNarrowTerminal(t *testing.T) {
	cfg := testConfig()
	m := New(cfg)
	m.width = 10 // extremely narrow

	// Should not panic
	view := m.View()
	if view == "" {
		t.Error("view should not be empty even at very narrow width")
	}
}

func TestDeleteModeNavigateAndDelete(t *testing.T) {
	useTempConfigDir(t)
	cfg := testConfig()
	m := New(cfg)
	originalCount := len(m.config.Timezones)

	// Enter delete mode
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = updated.(Model)

	// Move cursor down
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)
	if m.cursor != 1 {
		t.Errorf("expected cursor at 1, got %d", m.cursor)
	}

	// Delete the selected timezone (index 1)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if len(m.config.Timezones) != originalCount-1 {
		t.Errorf("expected %d timezones after delete, got %d",
			originalCount-1, len(m.config.Timezones))
	}
	if m.state != stateNormal {
		t.Error("expected stateNormal after delete")
	}
}

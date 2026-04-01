package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"clockscale/config"
)

// --- handleAddKey tests ---

func TestAddKeyTypingAccumulatesInput(t *testing.T) {
	cfg := testConfig()
	m := New(cfg)
	m.state = stateAdd

	for _, r := range "US/Pacific" {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(Model)
	}

	if m.input != "US/Pacific" {
		t.Errorf("expected input %q, got %q", "US/Pacific", m.input)
	}
	if m.state != stateAdd {
		t.Error("should still be in stateAdd while typing")
	}
}

func TestAddKeyBackspaceDeletesCharacter(t *testing.T) {
	cfg := testConfig()
	m := New(cfg)
	m.state = stateAdd
	m.input = "US/Pacifi"

	// Backspace should remove last character
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = updated.(Model)
	if m.input != "US/Pacif" {
		t.Errorf("expected %q after backspace, got %q", "US/Pacif", m.input)
	}

	// Ctrl+H should also work as backspace
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlH})
	m = updated.(Model)
	if m.input != "US/Paci" {
		t.Errorf("expected %q after ctrl+h, got %q", "US/Paci", m.input)
	}
}

func TestAddKeyBackspaceOnEmptyInputIsNoop(t *testing.T) {
	cfg := testConfig()
	m := New(cfg)
	m.state = stateAdd
	m.input = ""

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = updated.(Model)
	if m.input != "" {
		t.Errorf("expected empty input after backspace on empty, got %q", m.input)
	}
	if m.state != stateAdd {
		t.Error("should remain in stateAdd")
	}
}

func TestAddKeyEnterWithValidTimezone(t *testing.T) {
	useTempConfigDir(t)
	cfg := testConfig()
	m := New(cfg)
	m.state = stateAdd
	m.input = "US/Pacific"
	originalCount := len(m.config.Timezones)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	if m.state != stateNormal {
		t.Error("expected stateNormal after enter")
	}
	if m.input != "" {
		t.Errorf("expected input cleared, got %q", m.input)
	}
	if len(m.config.Timezones) != originalCount+1 {
		t.Errorf("expected %d timezones, got %d", originalCount+1, len(m.config.Timezones))
	}
	added := m.config.Timezones[len(m.config.Timezones)-1]
	if added.Timezone != "US/Pacific" || added.Label != "US/Pacific" {
		t.Errorf("unexpected added timezone: %+v", added)
	}
	if m.errMsg != "" {
		t.Errorf("unexpected error: %s", m.errMsg)
	}
}

func TestAddKeyEnterWithInvalidTimezone(t *testing.T) {
	cfg := testConfig()
	m := New(cfg)
	m.state = stateAdd
	m.input = "Not/Real/Zone"
	originalCount := len(m.config.Timezones)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	if m.state != stateNormal {
		t.Error("expected stateNormal after invalid timezone")
	}
	if m.errMsg == "" {
		t.Error("expected error message for invalid timezone")
	}
	if len(m.config.Timezones) != originalCount {
		t.Error("timezone list should not change for invalid timezone")
	}
}

func TestAddKeyEnterWithEmptyInput(t *testing.T) {
	cfg := testConfig()
	m := New(cfg)
	m.state = stateAdd
	m.input = ""
	originalCount := len(m.config.Timezones)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	if m.state != stateNormal {
		t.Error("expected stateNormal after empty enter")
	}
	if len(m.config.Timezones) != originalCount {
		t.Error("timezone list should not change for empty input")
	}
}

func TestAddKeyEnterTrimsWhitespace(t *testing.T) {
	useTempConfigDir(t)
	cfg := testConfig()
	m := New(cfg)
	m.state = stateAdd
	m.input = "  US/Pacific  "

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	added := m.config.Timezones[len(m.config.Timezones)-1]
	if added.Timezone != "US/Pacific" {
		t.Errorf("expected trimmed timezone %q, got %q", "US/Pacific", added.Timezone)
	}
}

func TestAddKeyEnterWhitespaceOnlyTreatedAsEmpty(t *testing.T) {
	cfg := testConfig()
	m := New(cfg)
	m.state = stateAdd
	m.input = "   "
	originalCount := len(m.config.Timezones)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	if len(m.config.Timezones) != originalCount {
		t.Error("whitespace-only input should not add a timezone")
	}
}

func TestAddKeyEscCancels(t *testing.T) {
	cfg := testConfig()
	m := New(cfg)
	m.state = stateAdd
	m.input = "US/Pacific"

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	if m.state != stateNormal {
		t.Error("expected stateNormal after esc")
	}
	if m.input != "" {
		t.Errorf("expected input cleared after esc, got %q", m.input)
	}
}

// --- handleDeleteKey edge case tests ---

func TestDeleteKeyCursorUpAtZeroIsNoop(t *testing.T) {
	cfg := testConfig()
	m := New(cfg)
	m.state = stateDelete
	m.cursor = 0

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(Model)
	if m.cursor != 0 {
		t.Errorf("cursor should stay at 0, got %d", m.cursor)
	}
}

func TestDeleteKeyCursorDownAtEndIsNoop(t *testing.T) {
	cfg := testConfig()
	m := New(cfg)
	m.state = stateDelete
	m.cursor = len(cfg.Timezones) - 1

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)
	if m.cursor != len(cfg.Timezones)-1 {
		t.Errorf("cursor should stay at %d, got %d", len(cfg.Timezones)-1, m.cursor)
	}
}

func TestDeleteKeyQuitInDeleteMode(t *testing.T) {
	cfg := testConfig()
	m := New(cfg)
	m.state = stateDelete

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Error("expected quit command on 'q' in delete mode")
	}
}

func TestDeleteKeyCtrlCInDeleteMode(t *testing.T) {
	cfg := testConfig()
	m := New(cfg)
	m.state = stateDelete

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("expected quit command on ctrl+c in delete mode")
	}
}

func TestDeleteKeyDeleteFirstItem(t *testing.T) {
	useTempConfigDir(t)
	cfg := testConfig()
	m := New(cfg)
	m.state = stateDelete
	m.cursor = 0
	second := m.config.Timezones[1].Timezone

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	if len(m.config.Timezones) != 1 {
		t.Fatalf("expected 1 timezone, got %d", len(m.config.Timezones))
	}
	if m.config.Timezones[0].Timezone != second {
		t.Errorf("remaining timezone should be %q, got %q", second, m.config.Timezones[0].Timezone)
	}
}

func TestDeleteModeCannotEnterWithEmptyTimezones(t *testing.T) {
	cfg := testConfig()
	cfg.Timezones = nil
	m := New(cfg)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = updated.(Model)

	if m.state != stateNormal {
		t.Error("should not enter delete mode with no timezones")
	}
}

func TestDeleteKeyEscCancelsWithoutDeleting(t *testing.T) {
	cfg := testConfig()
	m := New(cfg)
	m.state = stateDelete
	m.cursor = 0
	originalCount := len(m.config.Timezones)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	if m.state != stateNormal {
		t.Error("expected stateNormal after esc")
	}
	if len(m.config.Timezones) != originalCount {
		t.Error("esc should not delete any timezone")
	}
}

// --- handleNormalKey additional tests ---

func TestNormalKeyReloadConfig(t *testing.T) {
	useTempConfigDir(t)
	cfg := testConfig()
	m := New(cfg)

	// Save config first so 'r' can reload it
	if err := config.Save(cfg); err != nil {
		t.Fatalf("save: %v", err)
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = updated.(Model)

	if m.state != stateNormal {
		t.Error("should stay in stateNormal after reload")
	}
}

func TestNormalKeyCtrlCQuits(t *testing.T) {
	cfg := testConfig()
	m := New(cfg)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("expected quit command on ctrl+c")
	}
}

func TestNormalKeySlashTogglesHelp(t *testing.T) {
	cfg := testConfig()
	m := New(cfg)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(Model)
	if !m.showHelp {
		t.Error("expected showHelp true after '/'")
	}
}

func TestNormalKeyClearsErrMsg(t *testing.T) {
	cfg := testConfig()
	m := New(cfg)
	m.errMsg = "some error"

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = updated.(Model)
	if m.errMsg != "" {
		t.Errorf("expected errMsg cleared, got %q", m.errMsg)
	}
}

func TestNavigationHidesHelp(t *testing.T) {
	cfg := testConfig()
	m := New(cfg)

	tests := []struct {
		name string
		msg  tea.KeyMsg
	}{
		{"left", tea.KeyMsg{Type: tea.KeyLeft}},
		{"right", tea.KeyMsg{Type: tea.KeyRight}},
		{"a", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}},
		{"d", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}},
		{"r", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m2 := m
			m2.showHelp = true
			updated, _ := m2.Update(tt.msg)
			m2 = updated.(Model)
			if m2.showHelp {
				t.Errorf("key %q should hide help", tt.name)
			}
		})
	}
}

// --- Init test ---

func TestInitReturnsTickCommand(t *testing.T) {
	cfg := testConfig()
	m := New(cfg)

	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() should return a non-nil command (tick)")
	}
}

# Clockscale — Specification

## Overview

Clockscale is a terminal UI application that displays the current time across multiple timezones in a scrollable 24-hour grid.

---

## Display

### Layout

- One row per timezone
- Each row shows 24 columns, one per hour (0–23)
- Timezone label displayed on the left of each row
- All rows scroll horizontally together, except for the timezone label, it remains static/fixed to left side of row

### Columns

- Each cell displays the local hour (0–23) for that timezone at the corresponding UTC offset
- Column width is fixed to accommodate 2-digit hour values

### Highlighting

- **Current hour column** — the column representing "now" is highlighted with a purple/blue background across all rows
- **Local time cell** — the cell in the user's local timezone row is highlighted with an orange background
- All other cells use the default dark background
- All text in cells is right aligned

### Colors

| Element | Color |
|---|---|
| Timezone label | Default background, Cyan (#02ffff) foreground |
| Local Timezone label | Default background, Yellow (#ffff00) foreground |
| Current hour column | Purple/blue background (#5e5e86), light green (#90ee90) foreground |
| Local timezone current cell | Red-orange (#b4420a) background, yellow (#ffff00) foreground |
| Default cell | Horizontally alternating (#1c1c1c (even column) and #2d2e2e (odd column)) dark background, light (#dadada) foreground |

---

## Timezones

- Configurable list of timezones to display
- Each timezone has a short, configurable display label (e.g. `IDT`, `Z`, `CDT`, `PST`, `ARG`)
- Default set of two timezones: Local time and Zulu/GMT/UTC
- One timezone is designated as the **local timezone** (used for orange highlight)

---

## Interaction

| Key | Action |
|---|---|
| `q` / `Ctrl+C` | Quit |
| `←` / `→` | Scroll grid left/right |
| a | Add timezone (modify config file in place, use timezone name as default label) |
| d | Remove timezone (modify config file in place) |
| r | Reload config file

---

## Refresh

- Display updates automatically every minute (or on the minute boundary)
- Current time highlight column moves as time advances
- Config file is automatically reloaded on each update

---

## Configuration

- Timezones and basic configuration stored in a config file (e.g. `~/.config/clockscale/config.json`)
- Supports adding/removing timezones via config or interactive commands

### Sample Configuration File
```json
{
  timezones: [
    { timezone: "Israel", label: "IDT" },
    { timezone: "Zulu", label: "Z" },
    { timezone: "US/Central", label: "Local", local: true },
    { timezone: "US/Pacific", label: "PDT" },
    { timezone: "US/Eastern", label: "EDT" },
  ]
  colors: {
    defaultTimezoneLabel: { bg: "default", fg: "#02ffff" }
    localTimezoneLabel: { bg: "default", fg: "#ffff00" }
    defaultCell: {
      evenBg: "#1c1c1c",
      oddBg: "#2d2e2e",
      fg: "#dadada"
    },
    currentTimeCells: {
      default: { bg: "#5e5e86", fg: "#90ee90" },
      local: { bg: "#b4420a", fg: "#ffff00" }
    }
  }
}
```

---

## Technical

- Language: Go
- TUI library: Bubble Tea
- Target platforms: macOS, Linux
- Minimum terminal width: 80 columns

---

## Out of Scope (v1)

- Mouse support
- 12-hour (AM/PM) mode
- DST change warnings
- Multiple themes

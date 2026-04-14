# waybar-tui

A terminal UI for managing Waybar themes, built with Go and Bubble Tea.

Inspired by [OldJobobo Wayflipper](https://github.com/OldJobobo/wayflipper/).

Browse, install, apply and delete Waybar themes from a keyboard-driven interface that automatically adapts to your terminal color scheme. Designed to work seamlessly with [Omarchy](https://omarchy.org/).

<img width="1909" height="1012" alt="image" src="https://github.com/user-attachments/assets/ff240820-f294-43ba-9ce2-f6806fd85bb6" />

## Features

- Browse installed themes with a live preview of `config.jsonc` and `style.css`
- Install themes directly from any GitHub repository URL
- Auto-detects theme folders within repos regardless of their structure
- Apply themes instantly — restarts Waybar automatically
- Delete themes with a confirmation step
- Automatic backup of your existing config before the first change
- Colors follow your terminal theme (works with any Omarchy theme or terminal palette)
- Opens as a floating centered window on Omarchy / Hyprland

## Requirements

- Go 1.21+ (to build from source)
- git
- Waybar
- Omarchy (optional — provides `omarchy-restart-waybar` and the floating window behavior)

## Installation

### Option 1 — Binary release (no Go required)

Download `waybar-tui-linux-x86_64` from the [latest release](https://github.com/oscardaaz/waybar-tui/releases), then:

```bash
chmod +x waybar-tui-linux-x86_64
./waybar-tui-linux-x86_64
```

### Option 2 — From source with installer (recommended for Omarchy)

```bash
git clone https://github.com/oscardaaz/waybar-tui.git
cd waybar-tui
./install.sh
```

The installer builds the binary, places it in `~/.local/bin/`, and automatically adds the Hyprland window rules for the floating window behavior.

### Option 3 — From source manually

```bash
git clone https://github.com/oscardaaz/waybar-tui.git
cd waybar-tui
go build -o .waybar-tui .
./waybar-tui
```

If you are on Omarchy, add these lines to `~/.config/hypr/hyprland.conf` to get the floating centered window:

```
windowrule = tag +floating-window, match:class org.omarchy.waybar-tui
windowrule = opacity 1.0 1.0, match:class org.omarchy.waybar-tui
```

## Theme structure

Themes live in `~/.config/waybar/themes/` and must contain at minimum:

```
~/.config/waybar/themes/
└── my-theme/
    ├── config.jsonc
    └── style.css
```

When a theme is applied, the `config.jsonc` and `style.css` files in `~/.config/waybar/` become symlinks pointing to the active theme folder.

## Keybindings

| Key         | Action                        |
|-------------|-------------------------------|
| `up / down` | Navigate theme list           |
| `enter`     | Apply selected theme          |
| `i`         | Install theme from GitHub     |
| `d`         | Delete selected theme         |
| `tab`       | Switch preview tab            |
| `r`         | Refresh theme list            |
| `q`         | Quit                          |
| `esc`       | Cancel / close dialog         |

## Installing a theme from GitHub

Press `i` and paste the GitHub repository URL. waybar-tui clones the repo, scans for valid theme folders (any folder containing both `config.jsonc` and `style.css`), and lets you pick one if multiple are found. You can filter the list by typing.

Example repos:

- `https://github.com/HANCORE-linux/waybar-themes`

## License

MIT

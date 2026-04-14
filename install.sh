#!/bin/bash

set -e

BINARY_NAME=".waybar-tui"
SCRIPT_NAME="waybar-tui"
INSTALL_DIR="$HOME/.local/bin"
HYPR_CONF="$HOME/.config/hypr/hyprland.conf"
WINDOW_RULE="windowrule = tag +floating-window, match:class org.omarchy.waybar-tui"
OPACITY_RULE="windowrule = opacity 1.0 1.0, match:class org.omarchy.waybar-tui"

echo "Building waybar-tui..."
go build -ldflags="-s -w" -o "$BINARY_NAME" .

echo "Installing to $INSTALL_DIR..."
mkdir -p "$INSTALL_DIR"
cp "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
cp "$SCRIPT_NAME" "$INSTALL_DIR/$SCRIPT_NAME"
chmod +x "$INSTALL_DIR/$BINARY_NAME"
chmod +x "$INSTALL_DIR/$SCRIPT_NAME"

# Add Hyprland window rules if on Omarchy/Hyprland
if [[ -f "$HYPR_CONF" ]]; then
    if ! grep -q "org.omarchy.waybar-tui" "$HYPR_CONF"; then
        echo "" >> "$HYPR_CONF"
        echo "# waybar-tui" >> "$HYPR_CONF"
        echo "$WINDOW_RULE" >> "$HYPR_CONF"
        echo "$OPACITY_RULE" >> "$HYPR_CONF"
        echo "Hyprland window rules added."
    else
        echo "Hyprland window rules already present, skipping."
    fi
fi

echo "Done. Run 'waybar-tui' to launch."

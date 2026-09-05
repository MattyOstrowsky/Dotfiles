# Windows Terminal — Catppuccin Mocha & Nord color schemes

Windows Terminal lives on the Windows side, so the WSL installer cannot (and
should not) touch it. This is the one color source you set by hand. Paste one
of the schemes below into `settings.json` and point your WSL profile at it;
`bat`, `glow`, `atuin`, and the fish prompt inherit the terminal's 16-color
palette, so everything visually matches once this is done.

## settings.json location

```
C:\Users\<you>\AppData\Local\Packages\Microsoft.WindowsTerminal_8wekyb3d8bbwe\LocalState\settings.json
```

- The `8wekyb3d8bbwe` suffix is constant for the stable MSIX build.
- For Windows Terminal **Preview** the folder is
  `Microsoft.WindowsTerminalPreview_8wekyb3d8bbwe`.
- Pick whichever exists on your machine.
- The file is **UTF-8 with BOM**. Preserve the BOM when saving — an editor
  that strips it can make Windows Terminal fail to parse the file.

## Color schemes

Paste one (or both) objects into the top-level `"schemes"` array:

### Catppuccin Mocha

```json
{
    "name": "Catppuccin Mocha",
    "background": "#1E1E2E",
    "foreground": "#CDD6F4",
    "cursorColor": "#F5E0DC",
    "selectionBackground": "#585B70",
    "black": "#45475A",
    "brightBlack": "#585B70",
    "red": "#F38BA8",
    "brightRed": "#F38BA8",
    "green": "#A6E3A1",
    "brightGreen": "#A6E3A1",
    "yellow": "#F9E2AF",
    "brightYellow": "#F9E2AF",
    "blue": "#89B4FA",
    "brightBlue": "#89B4FA",
    "purple": "#CBA6F7",
    "brightPurple": "#CBA6F7",
    "cyan": "#94E2D5",
    "brightCyan": "#94E2D5",
    "white": "#BAC2DE",
    "brightWhite": "#A6ADC8"
}
```

### Nord

```json
{
    "name": "Nord",
    "background": "#2E3440",
    "foreground": "#D8DEE9",
    "cursorColor": "#D8DEE9",
    "selectionBackground": "#434C5E",
    "black": "#3B4252",
    "brightBlack": "#4C566A",
    "red": "#BF616A",
    "brightRed": "#BF616A",
    "green": "#A3BE8C",
    "brightGreen": "#A3BE8C",
    "yellow": "#EBCB8B",
    "brightYellow": "#EBCB8B",
    "blue": "#81A1C1",
    "brightBlue": "#81A1C1",
    "purple": "#B48EAD",
    "brightPurple": "#B48EAD",
    "cyan": "#88C0D0",
    "brightCyan": "#8FBCBB",
    "white": "#E5E9F0",
    "brightWhite": "#ECEFF4"
}
```

These are the canonical published schemes (catppuccin/windows-terminal and
nordtheme/windows-terminal). Their hex values match the exact colors already
used by the stowed `bat`, `btop`, `k9s`, `starship`, and `lazygit` configs, so
the terminal background/foreground and the tool accent colors align.

## Step-by-step

1. Open Windows Terminal → **Settings** with `Ctrl`+`,` (or the down-arrow
   menu → Settings).
2. Bottom-left → **Open JSON file** (opens `settings.json` in your default
   editor).
3. Paste **one** (or both) of the scheme objects above into the `"schemes": [ … ]`
   array. Add a comma between objects if you paste both.
4. Find your WSL profile under `"profiles" → "list"` — the one with
   `"source": "Windows.Terminal.Wsl"` or your distro `guid` — and add or set:
   ```json
   "colorScheme": "Catppuccin Mocha"
   ```
   (or `"Nord"`).
5. While you're on that profile, confirm the Nerd Font face is set so glyphs
   render:
   ```json
   "font": { "face": "<Nerd Font name>" }
   ```
6. Save. The change applies live — no restart needed.
7. To switch themes later: change only the `"colorScheme"` value on the WSL
   profile and save.

## Reminder

Keep the file UTF-8 with BOM. If your editor offers "UTF-8" vs "UTF-8 with
BOM", pick the BOM variant (Windows Terminal expects it).
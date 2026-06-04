#!/bin/sh
# =============================================================================
# ly startup.sh — One Dark-themed TTY color palette
# =============================================================================
# Sets the 16 ANSI terminal colors to Nord palette before ly renders.
# This ensures the TTY background and box styles match the Nord theme.
#
# ANSI escape sequence format: ESC]4;N;#RRGGBB\a
# where N = color index (0-15)
# =============================================================================

# Only set colors if we're on a Linux console (not in terminal emulator)
[ "$TERM" = "linux" ] || exit 0

# Nord palette — 16 ANSI colors
# Regular (0-7)
printf '\033]4;0;#282c34\a'   # nord0  — black
printf '\033]4;1;#e06c75\a'   # nord11 — red
printf '\033]4;2;#98c379\a'   # nord14 — green
printf '\033]4;3;#e5c07b\a'   # nord13 — yellow
printf '\033]4;4;#61afef\a'   # nord9  — blue
printf '\033]4;5;#c678dd\a'   # nord15 — magenta
printf '\033]4;6;#56b6c2\a'   # nord8  — cyan
printf '\033]4;7;#abb2bf\a'   # nord4  — white

# Bright (8-15)
printf '\033]4;8;#5c6370\a'   # nord3  — bright black
printf '\033]4;9;#e06c75\a'   # nord11 — bright red
printf '\033]4;10;#98c379\a'  # nord14 — bright green
printf '\033]4;11;#e5c07b\a'  # nord13 — bright yellow
printf '\033]4;12;#61afef\a'  # nord9  — bright blue
printf '\033]4;13;#c678dd\a'  # nord15 — bright magenta
printf '\033]4;14;#56b6c2\a'  # one dark cyan
printf '\033]4;15;#abb2bf\a'  # one dark bright white

# Default foreground/background
printf '\033]10;#abb2bf\a'    # nord4 — default fg
printf '\033]11;#282c34\a'    # nord0 — default bg

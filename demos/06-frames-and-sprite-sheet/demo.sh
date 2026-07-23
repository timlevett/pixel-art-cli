#!/bin/sh
# Frames demo: a 3-frame slime squish-bounce cycle, using frame_ghost to
# check alignment against the previous frame while drawing each new one,
# exported as a single tiled sprite sheet.
set -eu
SOCK="${1:?usage: demo.sh <socket>}"
OUT_SHEET="${2:?usage: demo.sh <socket> <sheet.png>}"
GHOST_LOG="${3:?usage: demo.sh <socket> <sheet.png> <ghost.log>}"

pxcli() { /tmp/pxcli-main --socket "$SOCK" "$@"; }

pxcli start --headless --size 16x16 --socket "$SOCK"

# Frame 0: resting
pxcli fill_rect 5 8 6 5 "#4ad66d"
pxcli fill_rect 4 9 1 3 "#4ad66d"
pxcli fill_rect 11 9 1 3 "#4ad66d"
pxcli set_pixel 6 10 "#0a3d0a"
pxcli set_pixel 9 10 "#0a3d0a"

# Frame 1: squished down (wider, shorter) mid-bounce
pxcli frame add
pxcli frame select 1
pxcli fill_rect 3 10 10 3 "#4ad66d"
pxcli fill_rect 2 11 1 2 "#4ad66d"
pxcli fill_rect 13 11 1 2 "#4ad66d"
pxcli set_pixel 5 11 "#0a3d0a"
pxcli set_pixel 10 11 "#0a3d0a"
> "$GHOST_LOG"
echo "frame 1 ghosted against frame 0 (onion-skin check):" >> "$GHOST_LOG"
pxcli frame ghost 0 0.4 >> "$GHOST_LOG"

# Frame 2: stretched up mid-jump
pxcli frame add
pxcli frame select 2
pxcli fill_rect 6 5 4 8 "#4ad66d"
pxcli fill_rect 5 6 1 4 "#4ad66d"
pxcli fill_rect 10 6 1 4 "#4ad66d"
pxcli set_pixel 7 7 "#0a3d0a"
pxcli set_pixel 9 7 "#0a3d0a"
echo "" >> "$GHOST_LOG"
echo "frame 2 ghosted against frame 1 (onion-skin check):" >> "$GHOST_LOG"
pxcli frame ghost 1 0.4 >> "$GHOST_LOG"

pxcli export_sheet "$OUT_SHEET" --cols 3
pxcli stop

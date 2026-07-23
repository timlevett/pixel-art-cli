#!/bin/sh
# Layers demo: background, shadow, and character sprite kept on
# independent, non-destructive layers, composited only on export.
set -eu
SOCK="${1:?usage: demo.sh <socket>}"
OUT="${2:?usage: demo.sh <socket> <out.png>}"

pxcli() { /tmp/pxcli-main --socket "$SOCK" "$@"; }

pxcli start --headless --size 16x16 --socket "$SOCK"

# base = sky/ground background
pxcli fill_rect 0 0 16 11 "#8fd0ff"
pxcli fill_rect 0 11 16 5 "#5ab552"

# shadow layer, drawn beneath the sprite
pxcli layer add shadow
pxcli layer select shadow
pxcli fill_rect 5 12 6 2 "#2d6b2a"

# sprite layer: a little slime character
pxcli layer add sprite
pxcli layer select sprite
pxcli fill_rect 5 8 6 5 "#4ad66d"
pxcli fill_rect 4 9 1 3 "#4ad66d"
pxcli fill_rect 11 9 1 3 "#4ad66d"
pxcli set_pixel 6 10 "#0a3d0a"
pxcli set_pixel 9 10 "#0a3d0a"

pxcli export "$OUT"
pxcli stop

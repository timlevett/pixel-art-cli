#!/bin/sh
# Regions/symmetry demo: draw only the left half of a symmetric potion
# bottle by hand, then copy+mirror it onto the right half instead of
# drawing every pixel twice.
set -eu
SOCK="${1:?usage: demo.sh <socket>}"
OUT="${2:?usage: demo.sh <socket> <out.png>}"

pxcli() { /tmp/pxcli-main --socket "$SOCK" "$@"; }

pxcli start --headless --size 16x16 --socket "$SOCK"

# Left half only (x 0-7), right half (x 8-15) is a mirror copy below.
pxcli fill_rect 6 1 2 3 "#8fd6ff"
pxcli fill_rect 5 4 3 2 "#3a9bd6"
pxcli fill_rect 3 6 5 8 "#3a9bd6"
pxcli fill_rect 4 7 4 5 "#e05a5a"
pxcli set_pixel 4 8 "#ffb3b3"

pxcli copy 0 0 8 16 left
pxcli paste 8 0 left
pxcli mirror 8 0 8 16 horizontal

pxcli export "$OUT"
pxcli stop

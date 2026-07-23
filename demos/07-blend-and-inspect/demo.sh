#!/bin/sh
# Utility demo: blend computes anti-aliasing edge colors by hand instead
# of eyeballing hex math, and inspect dumps a text-grid readout of the
# canvas -- a fast, image-tool-free sanity check for an agent.
set -eu
SOCK="${1:?usage: demo.sh <socket>}"
OUT_PNG="${2:?usage: demo.sh <socket> <out.png>}"
OUT_TXT="${3:?usage: demo.sh <socket> <out.png> <inspect.txt>}"

pxcli() { /tmp/pxcli-main --socket "$SOCK" "$@"; }

pxcli start --headless --size 8x8 --socket "$SOCK"

BG="#1a2f66"
FG="#ffb347"
pxcli clear "$BG"

# Solid AA-free circle first...
pxcli circle 3 3 3 "$FG" fill

# ...then hand-blend the four corner pixels of the disk toward the
# background at intermediate ratios, computed with `blend` rather than
# guessed, to soften the outline.
EDGE1=$(pxcli blend "$FG" "$BG" 0.35 | awk '{print $2}')
EDGE2=$(pxcli blend "$FG" "$BG" 0.65 | awk '{print $2}')
pxcli set_pixel 0 1 "$EDGE2"
pxcli set_pixel 1 0 "$EDGE2"
pxcli set_pixel 5 0 "$EDGE2"
pxcli set_pixel 6 1 "$EDGE2"
pxcli set_pixel 0 4 "$EDGE1"
pxcli set_pixel 1 6 "$EDGE1"
pxcli set_pixel 6 4 "$EDGE1"
pxcli set_pixel 5 6 "$EDGE1"

pxcli export "$OUT_PNG"
pxcli inspect > "$OUT_TXT"

pxcli stop

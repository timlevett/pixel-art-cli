#!/bin/sh
# Palette demo: define one named "fire" palette, reference it by slot
# instead of repeating hex everywhere, then reuse it across many pixels
# to draw a torch/flame icon with consistent, easy-to-adjust colors.
set -eu
SOCK="${1:?usage: demo.sh <socket>}"
OUT="${2:?usage: demo.sh <socket> <out.png>}"

pxcli() { /tmp/pxcli-main --socket "$SOCK" "$@"; }

pxcli start --headless --size 16x16 --socket "$SOCK"
pxcli palette add fire "#3a2a1a" "#7a3b12" "#c1531a" "#f0891f" "#ffc23c" "#fff2b0"
pxcli palette add wood "#4a2f1a" "#6b4423"

pxcli palette use fire
pxcli fill_rect 6 10 4 5 wood:0
pxcli fill_rect 7 10 2 5 wood:1

pxcli fill_rect 5 6 6 5 p:0
pxcli fill_rect 6 5 4 5 p:1
pxcli fill_rect 6 3 4 5 p:2
pxcli fill_rect 7 2 2 4 p:3
pxcli fill_rect 7 1 2 3 p:4
pxcli set_pixel 7 0 p:5
pxcli set_pixel 8 0 p:5

pxcli export "$OUT"
pxcli stop

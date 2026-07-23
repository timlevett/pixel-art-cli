#!/bin/sh
# Reference underlay demo (requires the pxcli build from PR #8 /
# feature/pxcli-reference-underlay -- not yet merged to main): import a
# rough user-supplied sketch as a dimmed, non-drawable underlay, trace a
# clean mushroom sprite over it, and compare export (no underlay) against
# export_debug (underlay included) to show alignment.
set -eu
SOCK="${1:?usage: demo.sh <socket>}"
REF="${2:?usage: demo.sh <socket> <reference.png>}"
OUT_PLAIN="${3:?usage: demo.sh <socket> <reference.png> <plain.png>}"
OUT_DEBUG="${4:?usage: demo.sh <socket> <reference.png> <plain.png> <debug.png>}"

pxcli() { /tmp/pxcli-underlay --socket "$SOCK" "$@"; }

pxcli start --headless --size 16x16 --socket "$SOCK"
pxcli import_reference "$REF" --opacity 0.5

pxcli circle 7 3 3 "#e8b7d6" fill
pxcli set_pixel 6 2 "#ffffff"
pxcli fill_rect 6 6 3 6 "#f2e6c8"
pxcli fill_rect 4 12 7 2 "#c9a876"

pxcli export "$OUT_PLAIN"
pxcli export_debug "$OUT_DEBUG"
pxcli stop

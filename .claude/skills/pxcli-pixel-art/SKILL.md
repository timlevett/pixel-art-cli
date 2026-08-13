---
name: pxcli-pixel-art
description: Use when the user asks to draw, create, edit, or iterate on pixel art / sprites using the pxcli tool (e.g. "draw a pixel art X", "make a sprite of Y", "add Z to this pixel art").
---

# pxcli Pixel Art

## Overview

`pxcli` is a daemon-backed pixel canvas CLI. Commands are sent to a running daemon over a Unix socket, so the daemon must be started before drawing and stopped when done. If a local build isn't on `PATH`, invoke it as `./pxcli` from the repo root (see README's "Build from source").

## Workflow

Every drawing session is a single bracketed sequence: **start → draw → export → stop**.

1. `pxcli start --headless` once, at the very start (default canvas is 32x32; pass `--size WxH` if the user specifies dimensions). If it errors because a daemon is already running, proceed anyway — not fatal. If it errors with a stale pid/socket left behind by an unclean shutdown, `pxcli start` cleans that up automatically and starts fresh — only a live process still listening on the socket causes `daemon_already_running`.
2. Draw with `set_pixel`, `fill_rect`, `line`, `clear`, `circle`, `ellipse`, `dither_fill`, `copy`/`paste`/`move`/`mirror` (run `pxcli <cmd> --help` for exact args if unsure). Use `undo`/`redo` to fix mistakes instead of redrawing from scratch. For anything beyond a handful of pixels, prefer `pxcli script <file>` (below) over one CLI call per pixel — it's a single connection instead of one process spawn per command, and the whole batch is one undo step.
3. `pxcli export <file>.png`, then **Read the PNG** to see the actual result. `pxcli inspect [x y w h]` is a faster, image-tool-free alternative for a quick sanity check mid-drawing — it prints the canvas (or a region) as a text grid of colors. Exporting does not stop the daemon — if it needs more work, keep drawing and export again within the same running daemon.
4. `pxcli stop` once the art is finished and the exported PNG has been confirmed. Always the last step — never leave the daemon running between sessions.

## Command quick reference

Lifecycle: `start, stop`

Drawing: `set_pixel, fill_rect, line, clear, circle, ellipse, dither_fill`
- `circle`/`ellipse` draw an outline by default; add a trailing `fill` argument for a solid disk/region.
- `dither_fill <x> <y> <w> <h> <color1> <color2> [pattern]` alternates two colors (`checkerboard` default, `horizontal`, `vertical`) to approximate shading/gradients — pixel art has no true per-pixel gradient fill.

Batch execution: `script <file>` — run many drawing commands (including `paste`/`move`/`mirror`) from a file (or stdin) over one connection instead of one process per command. One undo step for the whole batch; stops and rolls back on the first bad line, reporting the failing line number. Use this for anything sprite-sized or bigger.

Regions: `copy <x> <y> <w> <h> [clipboard], paste <x> <y> [clipboard], move <x> <y> <w> <h> <dx> <dy>, mirror <x> <y> <w> <h> <horizontal|vertical>`
- Use `copy`+`paste`+`mirror` for symmetric sprite parts (e.g. draw one eye/arm, mirror it) instead of redrawing by hand.
- `copy` is read-only (clipboard only) and not part of undo history or `script` batches; `paste`/`move`/`mirror` are.

Palettes: `palette add <name> <color...>, palette list [name], palette use <name>` — define a named, ordered color set once, then reference it anywhere a `<color>` argument is expected instead of repeating hex: `<name>:<index>` (e.g. `fire:2`) or `p:<index>` for the palette most recently selected with `palette use`. Use this for any sprite with more than 2-3 colors so the palette stays consistent across many draw calls.

Utility: `get_pixel, export, undo, redo, blend <color1> <color2> <ratio>, inspect [x y w h]`
- `blend` linearly interpolates two colors (0=first, 1=second) — use it to compute an anti-aliasing edge color instead of eyeballing hex math.

Layers: `layer add <name>, layer list, layer select <name>, layer visible <name> <true|false>`
- Every drawing command, `get_pixel`, `inspect`, `undo`, `redo` act on whichever layer is active (each has its own independent undo history). `base` always exists and can't be re-added.
- `export` flattens every *visible* layer into the output PNG (non-destructive — layers stay independently editable). The windowed renderer only ever shows `base`; other layers only become visible via `export`.
- Use layers to keep a background, sprite, and effects separable instead of drawing everything onto one flat canvas.

Frames: `frame add, frame list, frame select <index>, frame ghost <index> [opacity], export_sheet <file>.png [--cols N]`
- Each frame is its own independent layer stack (so a frame can itself have multiple layers), addressed by 0-based index instead of name. Frame `0` always exists and can't be re-added.
- `frame select <index>` retargets drawing, `layer *`, `get_pixel`, `inspect`, `undo`, `redo`, and `export` to that frame.
- `frame ghost <index> [opacity]` (default opacity `0.35`) is onion-skinning: it prints a text grid (same format as `inspect`) of the active frame with another frame's content dimmed underneath, so frame-to-frame coherence (e.g. a walk cycle) can be checked without eyeballing separate exports.
- `export_sheet <file>.png [--cols N]` tiles every frame's own flattened composite into one sprite-sheet PNG (default: all frames in one row).
- Use frames for animation cycles (walk/idle/attack): draw frame 0, `frame add` + `frame select` for each subsequent pose, `frame ghost` the previous frame while drawing the next one for consistency, then `export_sheet` once all frames are done.

Reference underlay: `import_reference <path> [--opacity N], export_debug <file>.png`
- `import_reference` imports a local PNG/JPEG (only files the user supplied — never fetch external/copyrighted images) as a non-drawable reference underneath the canvas, dimmed to opacity `N` (default `0.35`), to trace proportions/silhouette against.
- It isn't a layer — drawing commands can never touch it, and it's excluded from `export`/`export_sheet`. Use `export_debug` instead of `export` to check alignment against it.

Run `pxcli <cmd> --help` for exact arguments rather than guessing.

## Color formats

Hex: `#rgb`, `#rrggbb`, `#rrggbbaa`. Named: `black, white, red, green, blue, yellow, orange, purple, cyan, magenta, gray, grey, transparent`. Palette reference: `<name>:<index>` or `p:<index>`. In zsh, `#` starts a comment — always quote hex colors, e.g. `"#ff0000"`.

## Common mistakes

- Out-of-bounds coordinates return `err out_of_bounds` — check canvas size first if unsure.
- Declaring the art done without exporting and reading the PNG back.
- Leaving the daemon running after finishing (always `pxcli stop` at the end).
- Issuing one `set_pixel` call per pixel for a detailed sprite — batch it with `script` instead; it's dramatically faster and cheaper.
- Redrawing symmetric parts (eyes, limbs) pixel-by-pixel instead of `copy`+`mirror`.

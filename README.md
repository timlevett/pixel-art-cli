# pxcli

pxcli is a cli tool meant for AI agents (claude code, opencode, codex) to draw pixel art.

It was created using a ralph loop see `ralph.sh` and an example agent is given in `.opencode/agents/artists.md`

## Install (Homebrew)

macOS and Linux (Linuxbrew):

```bash
brew install vossenwout/pixel-art-cli/pxcli
```

Upgrade:

```bash
brew upgrade pxcli
```

## Build from source

Windowed (GUI):
(don't have windows PC so didn't tests this if you have fixes you can always make PR)

```bash
go build -tags=ebiten ./cmd/pxcli
```

Headless (no GUI deps):

```bash
go build ./cmd/pxcli
```

## Quick start

```bash
pxcli start
pxcli set_pixel 1 1 #ff0000
pxcli export out.png
pxcli stop
```

## CLI API

Lifecycle:

- `pxcli start [--size 32x32] [--scale 10] [--headless] [--socket <path>]`
- `pxcli stop [--socket <path>]`

Drawing:

- `pxcli set_pixel <x> <y> <color>`
- `pxcli fill_rect <x> <y> <w> <h> <color>`
- `pxcli line <x1> <y1> <x2> <y2> <color>`
- `pxcli clear [color]`
- `pxcli circle <cx> <cy> <r> <color> [fill]` — circle outline, or a filled disk with the optional trailing `fill`.
- `pxcli ellipse <cx> <cy> <rx> <ry> <color> [fill]` — ellipse outline, or a filled region with the optional trailing `fill`.
- `pxcli dither_fill <x> <y> <w> <h> <color1> <color2> [pattern]` — fill a rectangle by alternating two colors, approximating a gradient without true per-pixel blending. `pattern` is `checkerboard` (default), `horizontal`, or `vertical`.

Regions:

- `pxcli copy <x> <y> <w> <h> [clipboard]` — capture a rectangle into a named clipboard slot (`default` if omitted). Read-only; not undoable (nothing on the canvas changes).
- `pxcli paste <x> <y> [clipboard]` — stamp a previously copied region with its top-left corner at `(x,y)`.
- `pxcli move <x> <y> <w> <h> <dx> <dy>` — relocate a rectangle by an offset, clearing the source (transparent). Source and destination may overlap safely.
- `pxcli mirror <x> <y> <w> <h> <horizontal|vertical>` — flip a rectangle in place; `horizontal` reverses left-right, `vertical` reverses top-bottom.

Batch execution:

- `pxcli script <file>` — run newline-separated `set_pixel`/`fill_rect`/`line`/`clear`/`circle`/`ellipse`/`dither_fill`/`paste`/`move`/`mirror` commands from a file over a single connection, instead of one process + socket round trip per command. Use `pxcli script` (no file arg, or `-`) to read from stdin.
  - Blank lines and lines starting with `#` are ignored.
  - The whole batch is applied as **one undoable step**: `pxcli undo` reverts the entire script, not individual commands.
  - On a malformed or failing line, execution stops immediately, the canvas is rolled back to its pre-script state (nothing partially applied), and the error reports the offending line number, e.g. `err invalid_args line 3: x must be an integer`.
  - Only mutating commands (`set_pixel`, `fill_rect`, `line`, `clear`, `circle`, `ellipse`, `dither_fill`, `paste`, `move`, `mirror`) are allowed inside a script. `copy` is read-only and not supported inside a script.

```bash
pxcli script art.pxs
```

Palettes:

- `pxcli palette add <name> <color...>` — define (or replace) a named palette from an ordered list of colors (any accepted color format).
- `pxcli palette list [name]` — list palette names, or the colors in a named palette (in slot order).
- `pxcli palette use <name>` — select the active palette for the `p:<index>` shorthand.

Any `<color>` argument on `set_pixel`, `fill_rect`, `line`, `clear`, and inside `script` files also accepts a palette reference instead of a raw color:

- `<name>:<index>` — slot `<index>` of the named palette.
- `p:<index>` — slot `<index>` of the palette most recently selected with `palette use`.

```bash
pxcli palette add fire "#ff0000" "#ffa500" "#ffff00"
pxcli set_pixel 1 1 fire:0       # named reference, no "use" needed
pxcli palette use fire
pxcli fill_rect 2 2 3 3 p:2      # active-palette shorthand
```

Referencing an undefined palette or an out-of-range slot returns `err invalid_color <message>`, same as any other malformed color argument.

Utility:

- `pxcli get_pixel <x> <y>`
- `pxcli export <filename.png>`
- `pxcli undo`
- `pxcli redo`
- `pxcli blend <color1> <color2> <ratio>` — compute a linearly interpolated color (`ratio` 0 = `color1`, 1 = `color2`) for anti-aliasing edge colors by hand. Read-only; both colors accept palette references.
- `pxcli inspect [x y w h]` — dump the canvas (or a sub-region) as a text grid, one row per line, colors space-separated. A fast alternative to `export` + reading the image back, useful when an agent has no image-reading tool. Read-only.

```bash
pxcli blend "#000000" "#ffffff" 0.5   # -> ok #808080ff
pxcli inspect 0 0 4 4
```

Layers:

- `pxcli layer add <name>` — create a new blank, visible layer, same size as the canvas. `base` (the original canvas) always exists and can't be re-added.
- `pxcli layer list` — list layer names in creation order (`base` first).
- `pxcli layer select <name>` — set the active layer. Every drawing command, `get_pixel`, `inspect`, `undo`, and `redo` act on whichever layer is active; each layer has its own independent undo/redo history.
- `pxcli layer visible <name> <true|false>` — include or exclude a layer when `export` flattens all visible layers (`base` first, in creation order, standard alpha "over" compositing) into the output PNG.

Layers are non-destructive: `export` flattens into the output file only — it never merges layers into each other, so `base` and every other layer stay independently editable. **The windowed renderer only ever displays the `base` layer** — other layers are headless-only until you `export` (or select back to `base`) to see them composited.

```bash
pxcli layer add sprite
pxcli layer select sprite
pxcli fill_rect 4 4 8 8 "#ff0000"   # drawn on "sprite", not "base"
pxcli layer select base
pxcli export out.png                # out.png shows base + sprite composited
```

Frames:

- `pxcli frame add` — create a new blank frame (its own independent layer stack, starting with just `base`) and print its 0-based index. Frame `0` always exists and can't be re-added.
- `pxcli frame list` — list frame indices in creation order (`0` first).
- `pxcli frame select <index>` — set the active frame. Every drawing command, `layer *`, `get_pixel`, `inspect`, `undo`, `redo`, and `export` act on whichever frame is active; each frame has its own independent layer stack and undo/redo history.
- `pxcli frame ghost <index> [opacity]` — dump a text grid (same format as `inspect`) of the active frame with another frame's flattened content ghosted underneath at reduced opacity (default `0.35`). Onion-skinning for frame-to-frame coherence, without diffing exports by hand. Read-only.
- `pxcli export_sheet <filename.png> [--cols N]` — tile every frame's flattened content into a single sprite-sheet PNG, `N` frames per row (default: all frames in one row), wrapping to a new row every `N` frames and padding any incomplete final row with transparent pixels.

Frames are non-destructive, same as layers: `export` only ever flattens the *active* frame; `export_sheet` tiles every frame's own flattened composite side by side without merging frames into each other. **The windowed renderer only ever displays frame `0`'s `base` layer** — other frames are headless-only until `export`/`export_sheet` (or selecting back to frame `0`) shows them.

```bash
pxcli frame add                      # -> ok 1
pxcli frame select 1
pxcli fill_rect 0 0 8 8 "#ff0000"    # drawn on frame 1
pxcli frame ghost 0 0.5              # frame 1 on top, frame 0 dimmed underneath
pxcli export_sheet sheet.png --cols 4
```

Common error codes:

- `invalid_command` unknown command
- `invalid_args` wrong argument count or type
- `invalid_color` unsupported color format (including a palette reference that fails to resolve)
- `invalid_palette` palette management error (e.g. `palette use`/`palette list` on an undefined palette)
- `invalid_clipboard` `paste` referenced a clipboard slot that has not been `copy`'d into yet
- `invalid_layer` referenced a layer name that doesn't exist (`layer select`/`layer visible`)
- `invalid_frame` referenced a frame index that doesn't exist (`frame select`/`frame ghost`)
- `out_of_bounds` coordinate outside canvas
- `no_history` undo/redo with empty history
- `io` export file error

## Color formats

Accepted input formats:

- Hex: `#rgb`, `#rrggbb`, `#rrggbbaa`
- Named: `black`, `white`, `red`, `green`, `blue`, `yellow`, `orange`, `purple`, `cyan`, `magenta`, `gray`, `grey`, `transparent`
- Palette reference: `<name>:<index>` or `p:<index>` (active palette) — see [Palettes](#cli-api)

!!! For zsh shells you have to put colors between "" parenthesis. !!!

## Headless vs windowed

- Windowed mode is the default when built with `-tags=ebiten`.
- Headless mode is opt-in; pass `--headless` for CI or a headless container.
- If the binary is built without the `ebiten` tag, starting without `--headless` returns `err renderer_unavailable ...`.
- The headless container/CI environment cannot open a window; build and run windowed mode locally.

## Linux GUI requirements

For the GUI on Linux, you need X11/OpenGL runtime libraries. Examples:

- Debian/Ubuntu: `sudo apt-get install -y libx11-6 libxext6 libxrandr2 libxinerama1 libxcursor1 libxi6 libgl1`
- Fedora: `sudo dnf install libX11 libXext libXrandr libXinerama libXcursor libXi mesa-libGL`

If you are in a headless container, use `--headless`.

## Development

Typical commands:

```bash
go test ./...
go build ./cmd/pxcli
go build -tags=ebiten ./cmd/pxcli
```

If you are developing in a headless container, use `--headless` when running the daemon.

package daemon

import (
	"errors"
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"pxcli/internal/canvas"
	"pxcli/internal/clipboard"
	pxcolor "pxcli/internal/color"
	"pxcli/internal/history"
	"pxcli/internal/palette"
	"pxcli/internal/protocol"
)

// errScriptFailed marks a script batch mutation as failed; the actual
// error details are carried out-of-band via the enclosing closure.
var errScriptFailed = errors.New("script failed")

// Handler maps protocol requests to canvas operations.
type Handler struct {
	history   *history.Manager
	palette   *palette.Store
	clipboard *clipboard.Store
	onStop    func()
}

// NewHandler creates a command handler for the provided history manager.
func NewHandler(history *history.Manager, onStop func()) *Handler {
	return &Handler{history: history, palette: palette.New(), clipboard: clipboard.New(), onStop: onStop}
}

// Handle executes a command and returns a single-line protocol response.
func (h *Handler) Handle(request protocol.Request) string {
	switch request.Command {
	case "set_pixel":
		return h.handleSetPixel(request.Args)
	case "get_pixel":
		return h.handleGetPixel(request.Args)
	case "fill_rect":
		return h.handleFillRect(request.Args)
	case "line":
		return h.handleLine(request.Args)
	case "clear":
		return h.handleClear(request.Args)
	case "circle":
		return h.handleCircle(request.Args)
	case "ellipse":
		return h.handleEllipse(request.Args)
	case "dither_fill":
		return h.handleDitherFill(request.Args)
	case "copy":
		return h.handleCopy(request.Args)
	case "paste":
		return h.handlePaste(request.Args)
	case "move":
		return h.handleMove(request.Args)
	case "mirror":
		return h.handleMirror(request.Args)
	case "export":
		return h.handleExport(request.Args)
	case "undo":
		return h.handleUndo(request.Args)
	case "redo":
		return h.handleRedo(request.Args)
	case "palette_add":
		return h.handlePaletteAdd(request.Args)
	case "palette_list":
		return h.handlePaletteList(request.Args)
	case "palette_use":
		return h.handlePaletteUse(request.Args)
	case "blend":
		return h.handleBlend(request.Args)
	case "inspect":
		return h.handleInspect(request.Args)
	case "stop":
		return h.handleStop(request.Args)
	default:
		return protocol.FormatError("invalid_command", fmt.Sprintf("unknown command %q", request.Command))
	}
}

func (h *Handler) handleSetPixel(args []string) string {
	if err := h.history.Apply(func(c *canvas.Canvas) error {
		return applySetPixel(c, h.palette, args)
	}); err != nil {
		return formatError(err)
	}
	return protocol.FormatOK("")
}

func applySetPixel(c *canvas.Canvas, store *palette.Store, args []string) error {
	if len(args) != 3 {
		return invalidArgCountErr(3, len(args))
	}
	x, err := parseIntArg(args[0], "x")
	if err != nil {
		return err
	}
	y, err := parseIntArg(args[1], "y")
	if err != nil {
		return err
	}
	value, err := resolveColor(store, args[2])
	if err != nil {
		return err
	}
	return c.SetPixel(x, y, value)
}

func (h *Handler) handleGetPixel(args []string) string {
	if len(args) != 2 {
		return invalidArgCount(2, len(args))
	}
	x, err := parseIntArg(args[0], "x")
	if err != nil {
		return formatError(err)
	}
	y, err := parseIntArg(args[1], "y")
	if err != nil {
		return formatError(err)
	}
	value, err := h.history.Canvas().GetPixel(x, y)
	if err != nil {
		return formatError(err)
	}
	return protocol.FormatOK(pxcolor.Format(value))
}

func (h *Handler) handleFillRect(args []string) string {
	if err := h.history.Apply(func(c *canvas.Canvas) error {
		return applyFillRect(c, h.palette, args)
	}); err != nil {
		return formatError(err)
	}
	return protocol.FormatOK("")
}

func applyFillRect(c *canvas.Canvas, store *palette.Store, args []string) error {
	if len(args) != 5 {
		return invalidArgCountErr(5, len(args))
	}
	x, err := parseIntArg(args[0], "x")
	if err != nil {
		return err
	}
	y, err := parseIntArg(args[1], "y")
	if err != nil {
		return err
	}
	w, err := parseIntArg(args[2], "w")
	if err != nil {
		return err
	}
	hgt, err := parseIntArg(args[3], "h")
	if err != nil {
		return err
	}
	value, err := resolveColor(store, args[4])
	if err != nil {
		return err
	}
	return c.FillRect(x, y, w, hgt, value)
}

func (h *Handler) handleLine(args []string) string {
	if err := h.history.Apply(func(c *canvas.Canvas) error {
		return applyLine(c, h.palette, args)
	}); err != nil {
		return formatError(err)
	}
	return protocol.FormatOK("")
}

func applyLine(c *canvas.Canvas, store *palette.Store, args []string) error {
	if len(args) != 5 {
		return invalidArgCountErr(5, len(args))
	}
	x1, err := parseIntArg(args[0], "x1")
	if err != nil {
		return err
	}
	y1, err := parseIntArg(args[1], "y1")
	if err != nil {
		return err
	}
	x2, err := parseIntArg(args[2], "x2")
	if err != nil {
		return err
	}
	y2, err := parseIntArg(args[3], "y2")
	if err != nil {
		return err
	}
	value, err := resolveColor(store, args[4])
	if err != nil {
		return err
	}
	return c.Line(x1, y1, x2, y2, value)
}

func (h *Handler) handleClear(args []string) string {
	if err := h.history.Apply(func(c *canvas.Canvas) error {
		return applyClear(c, h.palette, args)
	}); err != nil {
		return formatError(err)
	}
	return protocol.FormatOK("")
}

func applyClear(c *canvas.Canvas, store *palette.Store, args []string) error {
	if len(args) > 1 {
		return invalidArgCountErr(1, len(args))
	}
	value := canvasTransparent()
	if len(args) == 1 {
		parsed, err := resolveColor(store, args[0])
		if err != nil {
			return err
		}
		value = parsed
	}
	c.Clear(value)
	return nil
}

func (h *Handler) handleCircle(args []string) string {
	if err := h.history.Apply(func(c *canvas.Canvas) error {
		return applyCircle(c, h.palette, args)
	}); err != nil {
		return formatError(err)
	}
	return protocol.FormatOK("")
}

// applyCircle draws a circle. args is <cx> <cy> <r> <color> with an
// optional trailing "fill" literal to draw a solid disk instead of an
// outline.
func applyCircle(c *canvas.Canvas, store *palette.Store, args []string) error {
	filled := false
	if len(args) == 5 && args[4] == "fill" {
		filled = true
		args = args[:4]
	}
	if len(args) != 4 {
		return invalidArgCountErr(4, len(args))
	}
	cx, err := parseIntArg(args[0], "cx")
	if err != nil {
		return err
	}
	cy, err := parseIntArg(args[1], "cy")
	if err != nil {
		return err
	}
	r, err := parseIntArg(args[2], "r")
	if err != nil {
		return err
	}
	value, err := resolveColor(store, args[3])
	if err != nil {
		return err
	}
	return c.Circle(cx, cy, r, value, filled)
}

func (h *Handler) handleEllipse(args []string) string {
	if err := h.history.Apply(func(c *canvas.Canvas) error {
		return applyEllipse(c, h.palette, args)
	}); err != nil {
		return formatError(err)
	}
	return protocol.FormatOK("")
}

// applyEllipse draws an ellipse. args is <cx> <cy> <rx> <ry> <color> with an
// optional trailing "fill" literal to draw a solid region instead of an
// outline.
func applyEllipse(c *canvas.Canvas, store *palette.Store, args []string) error {
	filled := false
	if len(args) == 6 && args[5] == "fill" {
		filled = true
		args = args[:5]
	}
	if len(args) != 5 {
		return invalidArgCountErr(5, len(args))
	}
	cx, err := parseIntArg(args[0], "cx")
	if err != nil {
		return err
	}
	cy, err := parseIntArg(args[1], "cy")
	if err != nil {
		return err
	}
	rx, err := parseIntArg(args[2], "rx")
	if err != nil {
		return err
	}
	ry, err := parseIntArg(args[3], "ry")
	if err != nil {
		return err
	}
	value, err := resolveColor(store, args[4])
	if err != nil {
		return err
	}
	return c.Ellipse(cx, cy, rx, ry, value, filled)
}

func (h *Handler) handleDitherFill(args []string) string {
	if err := h.history.Apply(func(c *canvas.Canvas) error {
		return applyDitherFill(c, h.palette, args)
	}); err != nil {
		return formatError(err)
	}
	return protocol.FormatOK("")
}

// applyDitherFill fills a rectangle with an alternating two-color pattern.
// args is <x> <y> <w> <h> <color1> <color2> with an optional trailing
// pattern name ("checkerboard" (default), "horizontal", "vertical").
func applyDitherFill(c *canvas.Canvas, store *palette.Store, args []string) error {
	pattern := ""
	if len(args) == 7 {
		pattern = args[6]
		args = args[:6]
	}
	if len(args) != 6 {
		return invalidArgCountErr(6, len(args))
	}
	x, err := parseIntArg(args[0], "x")
	if err != nil {
		return err
	}
	y, err := parseIntArg(args[1], "y")
	if err != nil {
		return err
	}
	w, err := parseIntArg(args[2], "w")
	if err != nil {
		return err
	}
	hgt, err := parseIntArg(args[3], "h")
	if err != nil {
		return err
	}
	color1, err := resolveColor(store, args[4])
	if err != nil {
		return err
	}
	color2, err := resolveColor(store, args[5])
	if err != nil {
		return err
	}
	return c.DitherFill(x, y, w, hgt, color1, color2, pattern)
}

// handleCopy captures a rectangle into a named clipboard slot. It reads the
// canvas but does not mutate it, so it is not recorded in undo/redo history.
// args is <x> <y> <w> <h> with an optional trailing clipboard name
// (default clipboard.DefaultName).
func (h *Handler) handleCopy(args []string) string {
	name := clipboard.DefaultName
	if len(args) == 5 {
		name = args[4]
		args = args[:4]
	}
	if len(args) != 4 {
		return invalidArgCount(4, len(args))
	}
	x, err := parseIntArg(args[0], "x")
	if err != nil {
		return formatError(err)
	}
	y, err := parseIntArg(args[1], "y")
	if err != nil {
		return formatError(err)
	}
	w, err := parseIntArg(args[2], "w")
	if err != nil {
		return formatError(err)
	}
	hgt, err := parseIntArg(args[3], "h")
	if err != nil {
		return formatError(err)
	}
	pixels, err := h.history.Canvas().CopyRegion(x, y, w, hgt)
	if err != nil {
		return formatError(err)
	}
	if err := h.clipboard.Set(name, clipboard.Region{Width: w, Height: hgt, Pixels: pixels}); err != nil {
		return formatError(err)
	}
	return protocol.FormatOK("")
}

func (h *Handler) handlePaste(args []string) string {
	if err := h.history.Apply(func(c *canvas.Canvas) error {
		return applyPaste(c, h.clipboard, args)
	}); err != nil {
		return formatError(err)
	}
	return protocol.FormatOK("")
}

// applyPaste stamps a clipboard region with its top-left corner at (x,y).
// args is <x> <y> with an optional trailing clipboard name (default
// clipboard.DefaultName).
func applyPaste(c *canvas.Canvas, store *clipboard.Store, args []string) error {
	name := clipboard.DefaultName
	if len(args) == 3 {
		name = args[2]
		args = args[:2]
	}
	if len(args) != 2 {
		return invalidArgCountErr(2, len(args))
	}
	x, err := parseIntArg(args[0], "x")
	if err != nil {
		return err
	}
	y, err := parseIntArg(args[1], "y")
	if err != nil {
		return err
	}
	region, err := store.Get(name)
	if err != nil {
		return err
	}
	return c.PasteRegion(x, y, region.Width, region.Height, region.Pixels)
}

func (h *Handler) handleMove(args []string) string {
	if err := h.history.Apply(func(c *canvas.Canvas) error {
		return applyMove(c, args)
	}); err != nil {
		return formatError(err)
	}
	return protocol.FormatOK("")
}

// applyMove relocates a rectangle by an offset. args is
// <x> <y> <w> <h> <dx> <dy>.
func applyMove(c *canvas.Canvas, args []string) error {
	if len(args) != 6 {
		return invalidArgCountErr(6, len(args))
	}
	x, err := parseIntArg(args[0], "x")
	if err != nil {
		return err
	}
	y, err := parseIntArg(args[1], "y")
	if err != nil {
		return err
	}
	w, err := parseIntArg(args[2], "w")
	if err != nil {
		return err
	}
	hgt, err := parseIntArg(args[3], "h")
	if err != nil {
		return err
	}
	dx, err := parseIntArg(args[4], "dx")
	if err != nil {
		return err
	}
	dy, err := parseIntArg(args[5], "dy")
	if err != nil {
		return err
	}
	return c.MoveRegion(x, y, w, hgt, dx, dy)
}

func (h *Handler) handleMirror(args []string) string {
	if err := h.history.Apply(func(c *canvas.Canvas) error {
		return applyMirror(c, args)
	}); err != nil {
		return formatError(err)
	}
	return protocol.FormatOK("")
}

// applyMirror flips a rectangle in place. args is
// <x> <y> <w> <h> <horizontal|vertical>.
func applyMirror(c *canvas.Canvas, args []string) error {
	if len(args) != 5 {
		return invalidArgCountErr(5, len(args))
	}
	x, err := parseIntArg(args[0], "x")
	if err != nil {
		return err
	}
	y, err := parseIntArg(args[1], "y")
	if err != nil {
		return err
	}
	w, err := parseIntArg(args[2], "w")
	if err != nil {
		return err
	}
	hgt, err := parseIntArg(args[3], "h")
	if err != nil {
		return err
	}
	return c.MirrorRegion(x, y, w, hgt, args[4])
}

func (h *Handler) handleExport(args []string) string {
	if len(args) != 1 {
		return invalidArgCount(1, len(args))
	}
	if err := h.history.Canvas().ExportPNG(args[0]); err != nil {
		return formatError(err)
	}
	return protocol.FormatOK("")
}

func (h *Handler) handleUndo(args []string) string {
	if len(args) != 0 {
		return invalidArgCount(0, len(args))
	}
	if err := h.history.Undo(); err != nil {
		return formatError(err)
	}
	return protocol.FormatOK("")
}

func (h *Handler) handleRedo(args []string) string {
	if len(args) != 0 {
		return invalidArgCount(0, len(args))
	}
	if err := h.history.Redo(); err != nil {
		return formatError(err)
	}
	return protocol.FormatOK("")
}

// HandleScript executes a batch of newline-separated commands as a single
// undoable step. Blank lines and lines starting with '#' are ignored; only
// canvas-mutating commands (set_pixel, fill_rect, line, clear, circle,
// ellipse, dither_fill, paste, move, mirror) are allowed. "copy" is
// read-only (it only writes to the clipboard, not the canvas) and is not
// supported inside a script.
// Execution stops at the first error, the canvas is rolled back to its
// pre-script state, and the response reports the 1-based line number that
// failed. On success the whole batch is recorded as one undo entry.
func (h *Handler) HandleScript(lines []string) string {
	type scriptCommand struct {
		line    int
		request protocol.Request
	}

	commands := make([]scriptCommand, 0, len(lines))
	for i, raw := range lines {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		request, err := protocol.ParseLine(trimmed)
		if err != nil {
			return protocol.FormatError("invalid_command", fmt.Sprintf("line %d: %s", i+1, err.Error()))
		}
		commands = append(commands, scriptCommand{line: i + 1, request: request})
	}
	if len(commands) == 0 {
		return protocol.FormatOK("")
	}

	type failure struct {
		line    int
		code    string
		message string
	}
	var failed *failure
	err := h.history.Apply(func(c *canvas.Canvas) error {
		pre := c.Snapshot()
		for _, cmd := range commands {
			if err := applyScriptCommand(c, h.palette, h.clipboard, cmd.request); err != nil {
				code, message := errorCodeAndMessage(err)
				failed = &failure{line: cmd.line, code: code, message: message}
				_ = c.Restore(pre)
				return errScriptFailed
			}
		}
		return nil
	})
	if err != nil {
		if failed != nil {
			return protocol.FormatError(failed.code, fmt.Sprintf("line %d: %s", failed.line, failed.message))
		}
		return formatError(err)
	}
	return protocol.FormatOK("")
}

func applyScriptCommand(c *canvas.Canvas, store *palette.Store, clip *clipboard.Store, request protocol.Request) error {
	switch request.Command {
	case "set_pixel":
		return applySetPixel(c, store, request.Args)
	case "fill_rect":
		return applyFillRect(c, store, request.Args)
	case "line":
		return applyLine(c, store, request.Args)
	case "clear":
		return applyClear(c, store, request.Args)
	case "circle":
		return applyCircle(c, store, request.Args)
	case "ellipse":
		return applyEllipse(c, store, request.Args)
	case "dither_fill":
		return applyDitherFill(c, store, request.Args)
	case "paste":
		return applyPaste(c, clip, request.Args)
	case "move":
		return applyMove(c, request.Args)
	case "mirror":
		return applyMirror(c, request.Args)
	default:
		return handlerError{Code: "invalid_command", Message: fmt.Sprintf("unsupported command %q in script", request.Command)}
	}
}

// resolveColor resolves a color argument that is either a palette reference
// ("<name>:<index>", or "p:<index>" for the active palette) or a raw color
// accepted by pxcolor.Parse (hex/named). Palette lookup failures are
// reported with the invalid_color code so callers see a consistent error
// shape regardless of which form they used.
func resolveColor(store *palette.Store, arg string) (color.RGBA, error) {
	if value, matched, err := store.ResolveRef(arg); matched {
		if err != nil {
			var perr palette.Error
			if errors.As(err, &perr) {
				return color.RGBA{}, pxcolor.Error{Code: "invalid_color", Message: perr.Message}
			}
			return color.RGBA{}, err
		}
		return value, nil
	}
	return pxcolor.Parse(arg)
}

func (h *Handler) handlePaletteAdd(args []string) string {
	if len(args) < 2 {
		return formatError(handlerError{Code: "invalid_args", Message: fmt.Sprintf("expected a name and at least one color, got %d args", len(args))})
	}
	name := args[0]
	colors := make([]color.RGBA, 0, len(args)-1)
	for _, hex := range args[1:] {
		value, err := pxcolor.Parse(hex)
		if err != nil {
			return formatError(err)
		}
		colors = append(colors, value)
	}
	if err := h.palette.Add(name, colors); err != nil {
		return formatError(err)
	}
	return protocol.FormatOK("")
}

func (h *Handler) handlePaletteList(args []string) string {
	if len(args) > 1 {
		return invalidArgCount(1, len(args))
	}
	if len(args) == 0 {
		return protocol.FormatOK(strings.Join(h.palette.List(), ","))
	}
	colors, err := h.palette.Colors(args[0])
	if err != nil {
		return formatError(err)
	}
	formatted := make([]string, len(colors))
	for i, value := range colors {
		formatted[i] = pxcolor.Format(value)
	}
	return protocol.FormatOK(strings.Join(formatted, ","))
}

func (h *Handler) handlePaletteUse(args []string) string {
	if len(args) != 1 {
		return invalidArgCount(1, len(args))
	}
	if err := h.palette.Use(args[0]); err != nil {
		return formatError(err)
	}
	return protocol.FormatOK("")
}

// handleBlend computes a linearly interpolated color between two colors
// (each accepting a palette reference or a raw color, like any other color
// argument) at the given ratio (0 = first color, 1 = second color). It is a
// pure query: it does not touch the canvas and is not part of undo history.
func (h *Handler) handleBlend(args []string) string {
	if len(args) != 3 {
		return invalidArgCount(3, len(args))
	}
	c1, err := resolveColor(h.palette, args[0])
	if err != nil {
		return formatError(err)
	}
	c2, err := resolveColor(h.palette, args[1])
	if err != nil {
		return formatError(err)
	}
	ratio, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return formatError(handlerError{Code: "invalid_args", Message: "ratio must be a number between 0 and 1"})
	}
	if ratio < 0 || ratio > 1 {
		return formatError(handlerError{Code: "invalid_args", Message: "ratio must be between 0 and 1"})
	}
	return protocol.FormatOK(pxcolor.Format(pxcolor.Blend(c1, c2, ratio)))
}

// handleInspect dumps the canvas (or a sub-region) as a text grid of
// canonical #rrggbbaa colors: rows separated by ";", cells within a row
// separated by ",". It is a pure query, like get_pixel, and gives an agent
// a fast alternative to export+read-image for sanity-checking a drawing.
// args is either empty (whole canvas) or <x> <y> <w> <h>.
func (h *Handler) handleInspect(args []string) string {
	if len(args) != 0 && len(args) != 4 {
		return formatError(handlerError{Code: "invalid_args", Message: fmt.Sprintf("expected 0 args (whole canvas) or 4 args (x y w h), got %d", len(args))})
	}
	c := h.history.Canvas()
	x, y, w, hgt := 0, 0, c.Width(), c.Height()
	if len(args) == 4 {
		var err error
		if x, err = parseIntArg(args[0], "x"); err != nil {
			return formatError(err)
		}
		if y, err = parseIntArg(args[1], "y"); err != nil {
			return formatError(err)
		}
		if w, err = parseIntArg(args[2], "w"); err != nil {
			return formatError(err)
		}
		if hgt, err = parseIntArg(args[3], "h"); err != nil {
			return formatError(err)
		}
	}
	pixels, err := c.CopyRegion(x, y, w, hgt)
	if err != nil {
		return formatError(err)
	}
	rows := make([]string, hgt)
	for row := 0; row < hgt; row++ {
		cells := make([]string, w)
		for col := 0; col < w; col++ {
			cells[col] = pxcolor.Format(pixels[row*w+col])
		}
		rows[row] = strings.Join(cells, ",")
	}
	return protocol.FormatOK(strings.Join(rows, ";"))
}

func (h *Handler) handleStop(args []string) string {
	if len(args) != 0 {
		return invalidArgCount(0, len(args))
	}
	if h.onStop != nil {
		h.onStop()
	}
	return protocol.FormatOK("")
}

func parseIntArg(value, name string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, handlerError{Code: "invalid_args", Message: fmt.Sprintf("%s must be an integer", name)}
	}
	return parsed, nil
}

func invalidArgCount(expected, got int) string {
	return formatError(invalidArgCountErr(expected, got))
}

func invalidArgCountErr(expected, got int) error {
	return handlerError{Code: "invalid_args", Message: fmt.Sprintf("expected %d args, got %d", expected, got)}
}

func canvasTransparent() color.RGBA {
	return color.RGBA{R: 0, G: 0, B: 0, A: 0}
}

type handlerError struct {
	Code    string
	Message string
}

func (e handlerError) Error() string {
	if e.Message == "" {
		return e.Code
	}
	return e.Code + ": " + e.Message
}

func formatError(err error) string {
	if err == nil {
		return protocol.FormatError("error", "unknown error")
	}
	code, message := errorCodeAndMessage(err)
	return protocol.FormatError(code, message)
}

func errorCodeAndMessage(err error) (string, string) {
	var herr handlerError
	if errors.As(err, &herr) {
		return herr.Code, herr.Message
	}
	var cerr canvas.Error
	if errors.As(err, &cerr) {
		return cerr.Code, cerr.Message
	}
	var colErr pxcolor.Error
	if errors.As(err, &colErr) {
		return colErr.Code, colErr.Message
	}
	var palErr palette.Error
	if errors.As(err, &palErr) {
		return palErr.Code, palErr.Message
	}
	var clipErr clipboard.Error
	if errors.As(err, &clipErr) {
		return clipErr.Code, clipErr.Message
	}
	var histErr history.Error
	if errors.As(err, &histErr) {
		return histErr.Code, histErr.Message
	}
	return "error", err.Error()
}

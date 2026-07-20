package daemon

import (
	"errors"
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"pxcli/internal/canvas"
	pxcolor "pxcli/internal/color"
	"pxcli/internal/history"
	"pxcli/internal/protocol"
)

// errScriptFailed marks a script batch mutation as failed; the actual
// error details are carried out-of-band via the enclosing closure.
var errScriptFailed = errors.New("script failed")

// Handler maps protocol requests to canvas operations.
type Handler struct {
	history *history.Manager
	onStop  func()
}

// NewHandler creates a command handler for the provided history manager.
func NewHandler(history *history.Manager, onStop func()) *Handler {
	return &Handler{history: history, onStop: onStop}
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
	case "export":
		return h.handleExport(request.Args)
	case "undo":
		return h.handleUndo(request.Args)
	case "redo":
		return h.handleRedo(request.Args)
	case "stop":
		return h.handleStop(request.Args)
	default:
		return protocol.FormatError("invalid_command", fmt.Sprintf("unknown command %q", request.Command))
	}
}

func (h *Handler) handleSetPixel(args []string) string {
	if err := h.history.Apply(func(c *canvas.Canvas) error {
		return applySetPixel(c, args)
	}); err != nil {
		return formatError(err)
	}
	return protocol.FormatOK("")
}

func applySetPixel(c *canvas.Canvas, args []string) error {
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
	value, err := pxcolor.Parse(args[2])
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
		return applyFillRect(c, args)
	}); err != nil {
		return formatError(err)
	}
	return protocol.FormatOK("")
}

func applyFillRect(c *canvas.Canvas, args []string) error {
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
	value, err := pxcolor.Parse(args[4])
	if err != nil {
		return err
	}
	return c.FillRect(x, y, w, hgt, value)
}

func (h *Handler) handleLine(args []string) string {
	if err := h.history.Apply(func(c *canvas.Canvas) error {
		return applyLine(c, args)
	}); err != nil {
		return formatError(err)
	}
	return protocol.FormatOK("")
}

func applyLine(c *canvas.Canvas, args []string) error {
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
	value, err := pxcolor.Parse(args[4])
	if err != nil {
		return err
	}
	return c.Line(x1, y1, x2, y2, value)
}

func (h *Handler) handleClear(args []string) string {
	if err := h.history.Apply(func(c *canvas.Canvas) error {
		return applyClear(c, args)
	}); err != nil {
		return formatError(err)
	}
	return protocol.FormatOK("")
}

func applyClear(c *canvas.Canvas, args []string) error {
	if len(args) > 1 {
		return invalidArgCountErr(1, len(args))
	}
	value := canvasTransparent()
	if len(args) == 1 {
		parsed, err := pxcolor.Parse(args[0])
		if err != nil {
			return err
		}
		value = parsed
	}
	c.Clear(value)
	return nil
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
// canvas-mutating commands (set_pixel, fill_rect, line, clear) are allowed.
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
			if err := applyScriptCommand(c, cmd.request); err != nil {
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

func applyScriptCommand(c *canvas.Canvas, request protocol.Request) error {
	switch request.Command {
	case "set_pixel":
		return applySetPixel(c, request.Args)
	case "fill_rect":
		return applyFillRect(c, request.Args)
	case "line":
		return applyLine(c, request.Args)
	case "clear":
		return applyClear(c, request.Args)
	default:
		return handlerError{Code: "invalid_command", Message: fmt.Sprintf("unsupported command %q in script", request.Command)}
	}
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
	var histErr history.Error
	if errors.As(err, &histErr) {
		return histErr.Code, histErr.Message
	}
	return "error", err.Error()
}

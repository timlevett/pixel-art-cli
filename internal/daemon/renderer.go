package daemon

import (
	"context"

	"pxcli/internal/canvas"
	"pxcli/internal/underlay"
)

const rendererUnavailableMessage = "windowed mode requires -tags=ebiten"

// Renderer runs the GUI loop and supports external close requests.
type Renderer interface {
	Run(ctx context.Context) error
	RequestClose()
}

// RenderSource provides snapshot and dirty information for rendering.
type RenderSource interface {
	Dirty() bool
	RenderSnapshot() canvas.RenderSnapshot
	Width() int
	Height() int
}

// underlayRenderSource composites an imported reference image beneath
// grid's own content for the windowed renderer, so the underlay is
// visible while drawing even though it lives outside the layer/frame
// stack grid belongs to. It never mutates grid.
type underlayRenderSource struct {
	grid              *canvas.Canvas
	underlay          *underlay.Store
	lastSeenVersion   int
	haveSeenAnyRender bool
}

func newUnderlayRenderSource(grid *canvas.Canvas, store *underlay.Store) *underlayRenderSource {
	return &underlayRenderSource{grid: grid, underlay: store}
}

func (s *underlayRenderSource) Width() int  { return s.grid.Width() }
func (s *underlayRenderSource) Height() int { return s.grid.Height() }

// Dirty reports grid's own dirty flag, but also stays true the first time
// it's checked and whenever the underlay has been (re)imported since the
// last RenderSnapshot, since importing a reference doesn't touch grid.
func (s *underlayRenderSource) Dirty() bool {
	if !s.haveSeenAnyRender {
		return true
	}
	if s.underlay.Version() != s.lastSeenVersion {
		return true
	}
	return s.grid.Dirty()
}

func (s *underlayRenderSource) RenderSnapshot() canvas.RenderSnapshot {
	s.haveSeenAnyRender = true
	s.lastSeenVersion = s.underlay.Version()
	if !s.underlay.HasImage() {
		return s.grid.RenderSnapshot()
	}
	composed, err := s.underlay.CompositeUnder(s.grid)
	if err != nil {
		// Composition can only fail on a dimension mismatch, which can't
		// happen here since both are sized from the same canvas config;
		// fall back to the uncomposited grid rather than panicking.
		return s.grid.RenderSnapshot()
	}
	snapshot := composed.RenderSnapshot()
	_ = s.grid.RenderSnapshot() // clear grid's own dirty flag too
	return snapshot
}

// RendererOptions holds future renderer configuration.
type RendererOptions struct {
	Headless bool
}

// RendererUnavailableError reports missing GUI support.
func RendererUnavailableError() Error {
	return Error{Code: "renderer_unavailable", Message: rendererUnavailableMessage}
}

// ValidateRenderer ensures requested headless/windowed mode is supported.
func ValidateRenderer(headless bool) error {
	if headless {
		return nil
	}
	if RendererAvailable() {
		return nil
	}
	return RendererUnavailableError()
}

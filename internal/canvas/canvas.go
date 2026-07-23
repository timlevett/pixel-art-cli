package canvas

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"sync"
)

// Error represents a canvas error with a code and message.
type Error struct {
	Code    string
	Message string
}

func (e Error) Error() string {
	if e.Message == "" {
		return e.Code
	}
	return e.Code + ": " + e.Message
}

// Canvas stores pixel data for a fixed-width, fixed-height image.
type Canvas struct {
	mu     sync.RWMutex
	width  int
	height int
	pixels []color.RGBA
	dirty  bool
}

// Snapshot captures a copy of the canvas pixels.
type Snapshot struct {
	width  int
	height int
	pixels []color.RGBA
}

// RenderSnapshot captures a copy of the canvas in RGBA byte form for rendering.
type RenderSnapshot struct {
	Width  int
	Height int
	Pixels []byte
}

// New creates a canvas with the provided dimensions.
func New(width, height int) (*Canvas, error) {
	if width <= 0 || height <= 0 {
		return nil, Error{Code: "invalid_args", Message: "canvas dimensions must be positive"}
	}
	pixels := make([]color.RGBA, width*height)
	return &Canvas{width: width, height: height, pixels: pixels, dirty: true}, nil
}

// Width returns the canvas width in pixels.
func (c *Canvas) Width() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.width
}

// Height returns the canvas height in pixels.
func (c *Canvas) Height() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.height
}

// SetPixel sets a pixel to the provided color.
func (c *Canvas) SetPixel(x, y int, value color.RGBA) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	idx, err := c.index(x, y)
	if err != nil {
		return err
	}
	c.pixels[idx] = value
	c.dirty = true
	return nil
}

// GetPixel returns the color at the provided coordinates.
func (c *Canvas) GetPixel(x, y int) (color.RGBA, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	idx, err := c.index(x, y)
	if err != nil {
		return color.RGBA{}, err
	}
	return c.pixels[idx], nil
}

// Clear fills the entire canvas with the provided color.
func (c *Canvas) Clear(value color.RGBA) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := range c.pixels {
		c.pixels[i] = value
	}
	c.dirty = true
}

// Snapshot returns a copy of the current canvas state.
func (c *Canvas) Snapshot() Snapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	pixels := make([]color.RGBA, len(c.pixels))
	copy(pixels, c.pixels)
	return Snapshot{width: c.width, height: c.height, pixels: pixels}
}

// RenderSnapshot returns a copy of the canvas as RGBA bytes and clears the dirty flag.
func (c *Canvas) RenderSnapshot() RenderSnapshot {
	c.mu.Lock()
	defer c.mu.Unlock()
	pixels := make([]byte, len(c.pixels)*4)
	for i, value := range c.pixels {
		offset := i * 4
		pixels[offset] = value.R
		pixels[offset+1] = value.G
		pixels[offset+2] = value.B
		pixels[offset+3] = value.A
	}
	c.dirty = false
	return RenderSnapshot{Width: c.width, Height: c.height, Pixels: pixels}
}

// Dirty reports whether the canvas has changed since the last render snapshot.
func (c *Canvas) Dirty() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.dirty
}

// Restore replaces the current canvas state with the snapshot.
func (c *Canvas) Restore(snapshot Snapshot) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if snapshot.width != c.width || snapshot.height != c.height {
		return Error{Code: "invalid_args", Message: "snapshot dimensions do not match canvas"}
	}
	if len(snapshot.pixels) != len(c.pixels) {
		return Error{Code: "invalid_args", Message: "snapshot size does not match canvas"}
	}
	copy(c.pixels, snapshot.pixels)
	c.dirty = true
	return nil
}

// FillRect fills a rectangle with the provided color.
func (c *Canvas) FillRect(x, y, w, h int, value color.RGBA) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if w <= 0 || h <= 0 {
		return Error{Code: "invalid_args", Message: "rect width and height must be positive"}
	}
	if x < 0 || y < 0 || x+w > c.width || y+h > c.height {
		return Error{
			Code:    "out_of_bounds",
			Message: fmt.Sprintf("rect (%d,%d) size %dx%d outside canvas", x, y, w, h),
		}
	}

	for row := y; row < y+h; row++ {
		start := row*c.width + x
		for i := 0; i < w; i++ {
			c.pixels[start+i] = value
		}
	}
	c.dirty = true
	return nil
}

// Line draws a line between two points, inclusive of endpoints.
func (c *Canvas) Line(x1, y1, x2, y2 int, value color.RGBA) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, err := c.index(x1, y1); err != nil {
		return err
	}
	if _, err := c.index(x2, y2); err != nil {
		return err
	}

	dx := absInt(x2 - x1)
	dy := absInt(y2 - y1)
	sx := -1
	if x1 < x2 {
		sx = 1
	}
	sy := -1
	if y1 < y2 {
		sy = 1
	}
	errVal := dx - dy

	for {
		c.pixels[y1*c.width+x1] = value
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * errVal
		if e2 > -dy {
			errVal -= dy
			x1 += sx
		}
		if e2 < dx {
			errVal += dx
			y1 += sy
		}
	}
	c.dirty = true
	return nil
}

// Circle draws a circle centered at (cx,cy) with the given radius, either as
// an outline (Bresenham/midpoint circle algorithm) or, when filled is true,
// as a solid disk. The full bounding box (cx-r,cy-r) to (cx+r,cy+r) must lie
// within the canvas.
func (c *Canvas) Circle(cx, cy, r int, value color.RGBA, filled bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if r <= 0 {
		return Error{Code: "invalid_args", Message: "radius must be positive"}
	}
	if cx-r < 0 || cy-r < 0 || cx+r >= c.width || cy+r >= c.height {
		return Error{
			Code:    "out_of_bounds",
			Message: fmt.Sprintf("circle center (%d,%d) radius %d outside canvas", cx, cy, r),
		}
	}

	if filled {
		for dy := -r; dy <= r; dy++ {
			dx := int(math.Sqrt(float64(r*r - dy*dy)))
			c.fillRow(cy+dy, cx-dx, cx+dx, value)
		}
	} else {
		x, y := r, 0
		errVal := 1 - r
		c.setSymmetricPoints8(cx, cy, x, y, value)
		for x > y {
			y++
			if errVal < 0 {
				errVal += 2*y + 1
			} else {
				x--
				errVal += 2*(y-x) + 1
			}
			c.setSymmetricPoints8(cx, cy, x, y, value)
		}
	}
	c.dirty = true
	return nil
}

// Ellipse draws an ellipse centered at (cx,cy) with radii rx,ry, either as an
// outline (midpoint ellipse algorithm) or, when filled is true, as a solid
// region. The full bounding box (cx-rx,cy-ry) to (cx+rx,cy+ry) must lie
// within the canvas.
func (c *Canvas) Ellipse(cx, cy, rx, ry int, value color.RGBA, filled bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if rx <= 0 || ry <= 0 {
		return Error{Code: "invalid_args", Message: "radii must be positive"}
	}
	if cx-rx < 0 || cy-ry < 0 || cx+rx >= c.width || cy+ry >= c.height {
		return Error{
			Code:    "out_of_bounds",
			Message: fmt.Sprintf("ellipse center (%d,%d) radii %dx%d outside canvas", cx, cy, rx, ry),
		}
	}

	if filled {
		for dy := -ry; dy <= ry; dy++ {
			dx := int(float64(rx) * math.Sqrt(1-float64(dy*dy)/float64(ry*ry)))
			c.fillRow(cy+dy, cx-dx, cx+dx, value)
		}
	} else {
		rx2, ry2 := float64(rx*rx), float64(ry*ry)
		x, y := 0, ry
		c.setSymmetricPoints4(cx, cy, x, y, value)

		// Region 1: slope of the ellipse boundary is shallower than -1.
		px, py := 0.0, 2*rx2*float64(y)
		p1 := ry2 - rx2*float64(ry) + 0.25*rx2
		for px < py {
			x++
			px += 2 * ry2
			if p1 < 0 {
				p1 += ry2 + px
			} else {
				y--
				py -= 2 * rx2
				p1 += ry2 + px - py
			}
			c.setSymmetricPoints4(cx, cy, x, y, value)
		}

		// Region 2: slope of the ellipse boundary is steeper than -1.
		p2 := ry2*(float64(x)+0.5)*(float64(x)+0.5) + rx2*float64(y-1)*float64(y-1) - rx2*ry2
		for y > 0 {
			y--
			py -= 2 * rx2
			if p2 > 0 {
				p2 += rx2 - py
			} else {
				x++
				px += 2 * ry2
				p2 += rx2 - py + px
			}
			c.setSymmetricPoints4(cx, cy, x, y, value)
		}
	}
	c.dirty = true
	return nil
}

// DitherFill fills a rectangle by alternating between two colors according
// to the named pattern, approximating shading gradients without true
// per-pixel color blending. Supported patterns: "checkerboard" (default),
// "horizontal", "vertical".
func (c *Canvas) DitherFill(x, y, w, h int, color1, color2 color.RGBA, pattern string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if w <= 0 || h <= 0 {
		return Error{Code: "invalid_args", Message: "rect width and height must be positive"}
	}
	if x < 0 || y < 0 || x+w > c.width || y+h > c.height {
		return Error{
			Code:    "out_of_bounds",
			Message: fmt.Sprintf("rect (%d,%d) size %dx%d outside canvas", x, y, w, h),
		}
	}
	if pattern == "" {
		pattern = "checkerboard"
	}
	switch pattern {
	case "checkerboard", "horizontal", "vertical":
	default:
		return Error{Code: "invalid_args", Message: fmt.Sprintf("unknown dither pattern %q", pattern)}
	}

	for row := y; row < y+h; row++ {
		start := row * c.width
		for col := x; col < x+w; col++ {
			var useFirst bool
			switch pattern {
			case "horizontal":
				useFirst = row%2 == 0
			case "vertical":
				useFirst = col%2 == 0
			default: // checkerboard
				useFirst = (row+col)%2 == 0
			}
			if useFirst {
				c.pixels[start+col] = color1
			} else {
				c.pixels[start+col] = color2
			}
		}
	}
	c.dirty = true
	return nil
}

// fillRow sets pixels [x1,x2] (inclusive) on the given row to value. Callers
// must hold c.mu and have validated that the row and [x1,x2] fall within
// the canvas (true for any chord of a circle/ellipse whose bounding box was
// already bounds-checked).
func (c *Canvas) fillRow(row, x1, x2 int, value color.RGBA) {
	start := row * c.width
	for i := x1; i <= x2; i++ {
		c.pixels[start+i] = value
	}
}

// setSymmetricPoints8 sets the 8 octant-symmetric points around (cx,cy) for
// a circle midpoint at offset (x,y). Callers must hold c.mu and have
// validated the circle's bounding box is within the canvas.
func (c *Canvas) setSymmetricPoints8(cx, cy, x, y int, value color.RGBA) {
	points := [8][2]int{
		{cx + x, cy + y}, {cx - x, cy + y}, {cx + x, cy - y}, {cx - x, cy - y},
		{cx + y, cy + x}, {cx - y, cy + x}, {cx + y, cy - x}, {cx - y, cy - x},
	}
	for _, p := range points {
		c.pixels[p[1]*c.width+p[0]] = value
	}
}

// setSymmetricPoints4 sets the 4 quadrant-symmetric points around (cx,cy)
// for an ellipse midpoint at offset (x,y). Callers must hold c.mu and have
// validated the ellipse's bounding box is within the canvas.
func (c *Canvas) setSymmetricPoints4(cx, cy, x, y int, value color.RGBA) {
	points := [4][2]int{
		{cx + x, cy + y}, {cx - x, cy + y}, {cx + x, cy - y}, {cx - x, cy - y},
	}
	for _, p := range points {
		c.pixels[p[1]*c.width+p[0]] = value
	}
}

// ExportPNG writes the canvas to a PNG file at the provided path.
func (c *Canvas) ExportPNG(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return Error{Code: "io", Message: err.Error()}
	}

	snapshot := c.Snapshot()
	img := image.NewRGBA(image.Rect(0, 0, snapshot.width, snapshot.height))
	for y := 0; y < snapshot.height; y++ {
		row := y * snapshot.width
		for x := 0; x < snapshot.width; x++ {
			img.SetRGBA(x, y, snapshot.pixels[row+x])
		}
	}

	if err := png.Encode(file, img); err != nil {
		_ = file.Close()
		return Error{Code: "io", Message: err.Error()}
	}
	if err := file.Close(); err != nil {
		return Error{Code: "io", Message: err.Error()}
	}
	return nil
}

func (c *Canvas) index(x, y int) (int, error) {
	if x < 0 || x >= c.width || y < 0 || y >= c.height {
		return 0, Error{Code: "out_of_bounds", Message: fmt.Sprintf("pixel (%d,%d) outside canvas", x, y)}
	}
	return y*c.width + x, nil
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

package canvas

import (
	"fmt"
	"image/color"
)

// CopyRegion returns a row-major copy of the pixels within the rectangle
// [x, x+w) by [y, y+h).
func (c *Canvas) CopyRegion(x, y, w, h int) ([]color.RGBA, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if err := c.validateRegionLocked(x, y, w, h); err != nil {
		return nil, err
	}
	pixels := make([]color.RGBA, w*h)
	for row := 0; row < h; row++ {
		srcStart := (y+row)*c.width + x
		copy(pixels[row*w:row*w+w], c.pixels[srcStart:srcStart+w])
	}
	return pixels, nil
}

// PasteRegion writes row-major pixel data (width w, height h) with its
// top-left corner at (x,y).
func (c *Canvas) PasteRegion(x, y, w, h int, pixels []color.RGBA) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.validateRegionLocked(x, y, w, h); err != nil {
		return err
	}
	if len(pixels) != w*h {
		return Error{Code: "invalid_args", Message: fmt.Sprintf("expected %d pixels for a %dx%d region, got %d", w*h, w, h, len(pixels))}
	}
	for row := 0; row < h; row++ {
		dstStart := (y+row)*c.width + x
		copy(c.pixels[dstStart:dstStart+w], pixels[row*w:row*w+w])
	}
	c.dirty = true
	return nil
}

// MoveRegion relocates the rectangle [x, x+w) by [y, y+h) by (dx,dy):
// captures the source pixels, clears the source rectangle to transparent,
// then pastes the captured pixels at the destination. Both the source and
// destination rectangles must lie within the canvas. Capturing before
// clearing means overlapping source/destination rectangles are handled
// correctly.
func (c *Canvas) MoveRegion(x, y, w, h, dx, dy int) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.validateRegionLocked(x, y, w, h); err != nil {
		return err
	}
	destX, destY := x+dx, y+dy
	if err := c.validateRegionLocked(destX, destY, w, h); err != nil {
		return err
	}

	pixels := make([]color.RGBA, w*h)
	for row := 0; row < h; row++ {
		srcStart := (y+row)*c.width + x
		copy(pixels[row*w:row*w+w], c.pixels[srcStart:srcStart+w])
	}

	transparent := color.RGBA{}
	for row := y; row < y+h; row++ {
		start := row*c.width + x
		for i := 0; i < w; i++ {
			c.pixels[start+i] = transparent
		}
	}

	for row := 0; row < h; row++ {
		dstStart := (destY+row)*c.width + destX
		copy(c.pixels[dstStart:dstStart+w], pixels[row*w:row*w+w])
	}
	c.dirty = true
	return nil
}

// MirrorRegion flips the rectangle [x, x+w) by [y, y+h) in place.
// axis "horizontal" reverses column order (left-right flip); "vertical"
// reverses row order (top-bottom flip).
func (c *Canvas) MirrorRegion(x, y, w, h int, axis string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.validateRegionLocked(x, y, w, h); err != nil {
		return err
	}

	switch axis {
	case "horizontal":
		for row := y; row < y+h; row++ {
			start := row * c.width
			for i := 0; i < w/2; i++ {
				left, right := start+x+i, start+x+w-1-i
				c.pixels[left], c.pixels[right] = c.pixels[right], c.pixels[left]
			}
		}
	case "vertical":
		for i := 0; i < h/2; i++ {
			topStart := (y+i) * c.width
			botStart := (y+h-1-i) * c.width
			for col := x; col < x+w; col++ {
				c.pixels[topStart+col], c.pixels[botStart+col] = c.pixels[botStart+col], c.pixels[topStart+col]
			}
		}
	default:
		return Error{Code: "invalid_args", Message: fmt.Sprintf("unknown mirror axis %q (want horizontal or vertical)", axis)}
	}
	c.dirty = true
	return nil
}

// validateRegionLocked checks that a rectangle has positive dimensions and
// lies fully within the canvas. Callers must hold c.mu.
func (c *Canvas) validateRegionLocked(x, y, w, h int) error {
	if w <= 0 || h <= 0 {
		return Error{Code: "invalid_args", Message: "region width and height must be positive"}
	}
	if x < 0 || y < 0 || x+w > c.width || y+h > c.height {
		return Error{
			Code:    "out_of_bounds",
			Message: fmt.Sprintf("region (%d,%d) size %dx%d outside canvas", x, y, w, h),
		}
	}
	return nil
}

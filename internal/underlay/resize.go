package underlay

import (
	"image"
	"image/color"
)

// resizeNearest scales src to width x height using nearest-neighbor
// sampling, returning row-major RGBA pixels. Reference images are
// typically much higher resolution than a pixel canvas, so this is a
// deliberately simple downscale (or upscale) rather than an averaging
// filter — good enough for tracing silhouette/proportions.
func resizeNearest(src image.Image, width, height int) []color.RGBA {
	bounds := src.Bounds()
	srcW, srcH := bounds.Dx(), bounds.Dy()

	out := make([]color.RGBA, width*height)
	for y := 0; y < height; y++ {
		srcY := bounds.Min.Y + y*srcH/height
		for x := 0; x < width; x++ {
			srcX := bounds.Min.X + x*srcW/width
			r, g, b, a := src.At(srcX, srcY).RGBA()
			out[y*width+x] = color.RGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			}
		}
	}
	return out
}

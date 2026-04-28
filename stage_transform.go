package ddddocr

import (
	"image"
	"image/color"
	"math"

	"github.com/disintegration/imaging"
)

const (
	ocrHeight = 64
	detSize   = 416
)

// OCRPreprocess converts an image into OCR model input data.
func OCRPreprocess(img image.Image, pngFix bool) ([]float32, int, int) {
	if pngFix {
		img = compositeOverWhite(img)
	}
	b := img.Bounds()
	srcW, srcH := b.Dx(), b.Dy()
	targetW := int(float64(srcW) * (float64(ocrHeight) / float64(srcH)))
	if targetW < 1 {
		targetW = 1
	}

	resized := imaging.Resize(img, targetW, ocrHeight, imaging.Lanczos)
	data := make([]float32, ocrHeight*targetW)
	for y := 0; y < ocrHeight; y++ {
		for x := 0; x < targetW; x++ {
			gray := color.GrayModel.Convert(resized.At(x, y)).(color.Gray)
			data[y*targetW+x] = float32(gray.Y) / 255.0
		}
	}
	return data, targetW, ocrHeight
}

// DetectionPreprocess converts an image into detection model input data.
func DetectionPreprocess(img image.Image) ([]float32, float64) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	ratio := math.Min(float64(detSize)/float64(h), float64(detSize)/float64(w))
	newW := int(float64(w) * ratio)
	newH := int(float64(h) * ratio)
	if newW < 1 {
		newW = 1
	}
	if newH < 1 {
		newH = 1
	}

	plane := detSize * detSize
	data := make([]float32, 3*plane)
	for i := range data {
		data[i] = 114
	}
	for y := 0; y < newH; y++ {
		for x := 0; x < newW; x++ {
			r, g, b := bilinearRGB(img, x, y, newW, newH)
			idx := y*detSize + x
			data[idx] = float32(b)
			data[plane+idx] = float32(g)
			data[2*plane+idx] = float32(r)
		}
	}
	return data, ratio
}

func bilinearRGB(img image.Image, dx, dy, dstW, dstH int) (uint8, uint8, uint8) {
	bounds := img.Bounds()
	srcW, srcH := bounds.Dx(), bounds.Dy()
	scaleX := float64(srcW) / float64(dstW)
	scaleY := float64(srcH) / float64(dstH)
	fx := (float64(dx)+0.5)*scaleX - 0.5
	fy := (float64(dy)+0.5)*scaleY - 0.5
	x0 := int(math.Floor(fx))
	y0 := int(math.Floor(fy))
	wx := fx - float64(x0)
	wy := fy - float64(y0)
	x1 := clampInt(x0+1, 0, srcW-1)
	y1 := clampInt(y0+1, 0, srcH-1)
	x0 = clampInt(x0, 0, srcW-1)
	y0 = clampInt(y0, 0, srcH-1)

	r00, g00, b00 := rgbAt(img, bounds.Min.X+x0, bounds.Min.Y+y0)
	r01, g01, b01 := rgbAt(img, bounds.Min.X+x1, bounds.Min.Y+y0)
	r10, g10, b10 := rgbAt(img, bounds.Min.X+x0, bounds.Min.Y+y1)
	r11, g11, b11 := rgbAt(img, bounds.Min.X+x1, bounds.Min.Y+y1)

	r := interp2(r00, r01, r10, r11, wx, wy)
	g := interp2(g00, g01, g10, g11, wx, wy)
	bl := interp2(b00, b01, b10, b11, wx, wy)
	return r, g, bl
}

func rgbAt(img image.Image, x, y int) (float64, float64, float64) {
	r, g, b, _ := img.At(x, y).RGBA()
	return float64(uint8(r >> 8)), float64(uint8(g >> 8)), float64(uint8(b >> 8))
}

func interp2(v00, v01, v10, v11, wx, wy float64) uint8 {
	top := v00*(1-wx) + v01*wx
	bottom := v10*(1-wx) + v11*wx
	v := top*(1-wy) + bottom*wy
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

func compositeOverWhite(src image.Image) image.Image {
	b := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			r, g, bl, a := src.At(b.Min.X+x, b.Min.Y+y).RGBA()
			af := float64(a) / 65535.0
			rr := uint8((float64(r>>8)*af + 255*(1-af)) + 0.5)
			gg := uint8((float64(g>>8)*af + 255*(1-af)) + 0.5)
			bb := uint8((float64(bl>>8)*af + 255*(1-af)) + 0.5)
			dst.SetRGBA(x, y, color.RGBA{R: rr, G: gg, B: bb, A: 255})
		}
	}
	return dst
}

func toGrayMatrix(img image.Image) [][]float64 {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	out := make([][]float64, h)
	for y := 0; y < h; y++ {
		row := make([]float64, w)
		for x := 0; x < w; x++ {
			gray := color.GrayModel.Convert(img.At(b.Min.X+x, b.Min.Y+y)).(color.Gray)
			row[x] = float64(gray.Y)
		}
		out[y] = row
	}
	return out
}

func toRGBMatrix(img image.Image) [][][]uint8 {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	out := make([][][]uint8, h)
	for y := 0; y < h; y++ {
		row := make([][]uint8, w)
		for x := 0; x < w; x++ {
			r, g, b, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
			row[x] = []uint8{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8)}
		}
		out[y] = row
	}
	return out
}

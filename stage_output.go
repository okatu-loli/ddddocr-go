package ddddocr

import (
	"image"
	"math"
)

// RunSlideMatchFile locates a slider target from target and background files.
func RunSlideMatchFile(targetPath, backgroundPath string, simple bool) (SlideResult, error) {
	target, err := LoadImageFile(targetPath)
	if err != nil {
		return SlideResult{}, err
	}
	background, err := LoadImageFile(backgroundPath)
	if err != nil {
		return SlideResult{}, err
	}
	return SlideMatch(target, background, simple), nil
}

// RunSlideComparisonFile locates a slider target by comparing two image files.
func RunSlideComparisonFile(targetPath, backgroundPath string) (SlideResult, error) {
	target, err := LoadImageFile(targetPath)
	if err != nil {
		return SlideResult{}, err
	}
	background, err := LoadImageFile(backgroundPath)
	if err != nil {
		return SlideResult{}, err
	}
	return SlideComparison(target, background), nil
}

// SlideMatch locates a slider target by template matching.
func SlideMatch(target, background image.Image, simple bool) SlideResult {
	t := toGrayMatrix(target)
	b := toGrayMatrix(background)
	if !simple {
		t = edgeMap(t)
		b = edgeMap(b)
	}
	x, y, conf := templateMatch(t, b)
	centerX := x + len(t[0])/2
	centerY := y + len(t)/2
	return SlideResult{Target: []int{centerX, centerY}, TargetX: centerX, TargetY: centerY, Confidence: &conf}
}

// SlideComparison locates a slider target by image difference.
func SlideComparison(target, background image.Image) SlideResult {
	t := toRGBMatrix(target)
	b := toRGBMatrix(background)
	h := min(len(t), len(b))
	if h == 0 {
		return SlideResult{Target: []int{0, 0}, TargetX: 0, TargetY: 0}
	}
	w := min(len(t[0]), len(b[0]))
	mask := make([][]bool, h)
	for y := 0; y < h; y++ {
		mask[y] = make([]bool, w)
		for x := 0; x < w; x++ {
			dr := absInt(int(t[y][x][0]) - int(b[y][x][0]))
			dg := absInt(int(t[y][x][1]) - int(b[y][x][1]))
			db := absInt(int(t[y][x][2]) - int(b[y][x][2]))
			gray := int(0.299*float64(dr) + 0.587*float64(dg) + 0.114*float64(db))
			mask[y][x] = gray > 30
		}
	}
	mask = erode(dilate(mask))
	mask = dilate(erode(mask))
	x1, y1, x2, y2, ok := largestComponentBounds(mask)
	if !ok {
		return SlideResult{Target: []int{0, 0}, TargetX: 0, TargetY: 0}
	}
	cx := x1 + (x2-x1+1)/2
	cy := y1 + (y2-y1+1)/2
	return SlideResult{Target: []int{cx, cy}, TargetX: cx, TargetY: cy}
}

func templateMatch(tpl, bg [][]float64) (int, int, float64) {
	th, tw := len(tpl), len(tpl[0])
	bh, bw := len(bg), len(bg[0])
	tMean, tNorm := stats(tpl, 0, 0, tw, th)
	bestX, bestY := 0, 0
	best := math.Inf(-1)
	for y := 0; y <= bh-th; y++ {
		for x := 0; x <= bw-tw; x++ {
			bMean, bNorm := stats(bg, x, y, tw, th)
			if tNorm == 0 || bNorm == 0 {
				continue
			}
			var num float64
			for yy := 0; yy < th; yy++ {
				for xx := 0; xx < tw; xx++ {
					num += (tpl[yy][xx] - tMean) * (bg[y+yy][x+xx] - bMean)
				}
			}
			score := num / (tNorm * bNorm)
			if score > best {
				best = score
				bestX, bestY = x, y
			}
		}
	}
	if math.IsInf(best, -1) {
		best = 0
	}
	return bestX, bestY, best
}

func stats(m [][]float64, x, y, w, h int) (float64, float64) {
	var sum float64
	n := float64(w * h)
	for yy := 0; yy < h; yy++ {
		for xx := 0; xx < w; xx++ {
			sum += m[y+yy][x+xx]
		}
	}
	mean := sum / n
	var sq float64
	for yy := 0; yy < h; yy++ {
		for xx := 0; xx < w; xx++ {
			d := m[y+yy][x+xx] - mean
			sq += d * d
		}
	}
	return mean, math.Sqrt(sq)
}

func edgeMap(src [][]float64) [][]float64 {
	h, w := len(src), len(src[0])
	dst := make([][]float64, h)
	for y := range dst {
		dst[y] = make([]float64, w)
	}
	for y := 1; y < h-1; y++ {
		for x := 1; x < w-1; x++ {
			gx := -src[y-1][x-1] - 2*src[y][x-1] - src[y+1][x-1] + src[y-1][x+1] + 2*src[y][x+1] + src[y+1][x+1]
			gy := -src[y-1][x-1] - 2*src[y-1][x] - src[y-1][x+1] + src[y+1][x-1] + 2*src[y+1][x] + src[y+1][x+1]
			if math.Hypot(gx, gy) >= 100 {
				dst[y][x] = 255
			}
		}
	}
	return dst
}

func largestComponentBounds(mask [][]bool) (int, int, int, int, bool) {
	h, w := len(mask), len(mask[0])
	seen := make([][]bool, h)
	for y := range seen {
		seen[y] = make([]bool, w)
	}
	bestArea := 0
	best := [4]int{}
	dirs := [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	for sy := 0; sy < h; sy++ {
		for sx := 0; sx < w; sx++ {
			if seen[sy][sx] || !mask[sy][sx] {
				continue
			}
			queue := [][2]int{{sx, sy}}
			seen[sy][sx] = true
			x1, y1, x2, y2, area := sx, sy, sx, sy, 0
			for len(queue) > 0 {
				p := queue[0]
				queue = queue[1:]
				x, y := p[0], p[1]
				area++
				if x < x1 {
					x1 = x
				}
				if y < y1 {
					y1 = y
				}
				if x > x2 {
					x2 = x
				}
				if y > y2 {
					y2 = y
				}
				for _, d := range dirs {
					nx, ny := x+d[0], y+d[1]
					if nx >= 0 && nx < w && ny >= 0 && ny < h && !seen[ny][nx] && mask[ny][nx] {
						seen[ny][nx] = true
						queue = append(queue, [2]int{nx, ny})
					}
				}
			}
			if area > bestArea {
				bestArea = area
				best = [4]int{x1, y1, x2, y2}
			}
		}
	}
	return best[0], best[1], best[2], best[3], bestArea > 0
}

func dilate(src [][]bool) [][]bool {
	h, w := len(src), len(src[0])
	dst := make([][]bool, h)
	for y := range dst {
		dst[y] = make([]bool, w)
		for x := 0; x < w; x++ {
			for yy := max(0, y-1); yy <= min(h-1, y+1); yy++ {
				for xx := max(0, x-1); xx <= min(w-1, x+1); xx++ {
					dst[y][x] = dst[y][x] || src[yy][xx]
				}
			}
		}
	}
	return dst
}

func erode(src [][]bool) [][]bool {
	h, w := len(src), len(src[0])
	dst := make([][]bool, h)
	for y := range dst {
		dst[y] = make([]bool, w)
		for x := 0; x < w; x++ {
			ok := true
			for yy := max(0, y-1); yy <= min(h-1, y+1); yy++ {
				for xx := max(0, x-1); xx <= min(w-1, x+1); xx++ {
					ok = ok && src[yy][xx]
				}
			}
			dst[y][x] = ok
		}
	}
	return dst
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

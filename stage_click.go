package ddddocr

import (
	"fmt"
	"image"
	"math"
	"sort"
)

const shapeGridSize = 64

// ClickCaptchaResult contains click points ordered by the prompt sequence.
type ClickCaptchaResult struct {
	Target [][]int `json:"target"`
}

type shapePoint struct {
	x float64
	y float64
}

type shapeMask struct {
	grid  []bool
	count int
}

// RunClickCaptchaFile resolves ordered click points from an image file.
func RunClickCaptchaFile(cfg DetectionConfig) (ClickCaptchaResult, error) {
	img, err := LoadImageFile(cfg.ImagePath)
	if err != nil {
		return ClickCaptchaResult{}, err
	}
	boxes, err := RunDetectionImage(img, cfg)
	if err != nil {
		return ClickCaptchaResult{}, err
	}
	return ResolveClickCaptcha(img, boxes)
}

// ResolveClickCaptcha resolves ordered click points from a decoded image and
// precomputed detection boxes.
func ResolveClickCaptcha(img image.Image, boxes []DetectionBox) (ClickCaptchaResult, error) {
	promptStrip := locatePromptStrip(img)
	prompts := findPromptBoxes(img, boxes, promptStrip)
	if len(prompts) == 0 {
		return ClickCaptchaResult{}, fmt.Errorf("no prompt icons found")
	}

	candidates := findCandidateBoxes(img, boxes, prompts, promptStrip)
	if len(candidates) == 0 {
		return ClickCaptchaResult{}, fmt.Errorf("no clickable candidates found")
	}

	scores := make([][]float64, len(prompts))
	for i, prompt := range prompts {
		promptMask := shapeMaskForBox(img, prompt, 0)
		scores[i] = make([]float64, len(candidates))
		for j, candidate := range candidates {
			scores[i][j] = bestShapeScore(img, promptMask, candidate)
		}
	}

	assignment := bestAssignment(scores)
	if len(assignment) != len(prompts) {
		return ClickCaptchaResult{}, fmt.Errorf("not enough candidates: prompts=%d candidates=%d", len(prompts), len(candidates))
	}

	result := ClickCaptchaResult{
		Target: make([][]int, 0, len(prompts)),
	}
	for _, candidateIndex := range assignment {
		if candidateIndex < 0 {
			continue
		}
		box := candidates[candidateIndex]
		center := boxCenter(box)
		result.Target = append(result.Target, []int{center[0], center[1]})
	}
	return result, nil
}

func locatePromptStrip(img image.Image) image.Rectangle {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	topH := max(1, h*18/100)
	seen := make([]bool, w*topH)
	bestArea := 0
	best := image.Rect(w/4, 0, w*3/4, max(1, h/8))
	dirs := [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}

	for sy := 0; sy < topH; sy++ {
		for sx := 0; sx < w; sx++ {
			idx := sy*w + sx
			if seen[idx] || !isPromptBackground(img, b.Min.X+sx, b.Min.Y+sy) {
				continue
			}

			queue := [][2]int{{sx, sy}}
			seen[idx] = true
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
					if nx < 0 || nx >= w || ny < 0 || ny >= topH {
						continue
					}
					nidx := ny*w + nx
					if !seen[nidx] && isPromptBackground(img, b.Min.X+nx, b.Min.Y+ny) {
						seen[nidx] = true
						queue = append(queue, [2]int{nx, ny})
					}
				}
			}

			if area > bestArea && x2-x1 >= 40 && y2-y1 >= 20 {
				bestArea = area
				best = image.Rect(x1, y1, x2+1, y2+1)
			}
		}
	}
	return best
}

func isPromptBackground(img image.Image, x, y int) bool {
	r, g, b := rgbAt(img, x, y)
	maxc := math.Max(r, math.Max(g, b))
	minc := math.Min(r, math.Min(g, b))
	avg := (r + g + b) / 3
	return avg >= 145 && avg <= 230 && maxc-minc <= 12
}

func findPromptBoxes(img image.Image, boxes []DetectionBox, strip image.Rectangle) []DetectionBox {
	prompts := make([]DetectionBox, 0)
	expanded := image.Rect(strip.Min.X-4, strip.Min.Y-4, strip.Max.X+4, strip.Max.Y+4)
	for _, box := range boxes {
		center := boxCenter(box)
		if !pointInRect(center[0], center[1], expanded) {
			continue
		}
		if boxWidth(box) < 18 || boxHeight(box) < 18 {
			continue
		}
		_, count := inkStats(img, box)
		if count < 40 {
			continue
		}
		prompts = append(prompts, cloneBox(box))
	}
	if len(prompts) == 0 {
		b := img.Bounds()
		for _, box := range boxes {
			center := boxCenter(box)
			if center[1] <= b.Dy()/8 && center[0] >= b.Dx()/4 {
				prompts = append(prompts, cloneBox(box))
			}
		}
	}
	sort.Slice(prompts, func(i, j int) bool {
		return boxCenter(prompts[i])[0] < boxCenter(prompts[j])[0]
	})
	return prompts
}

func findCandidateBoxes(img image.Image, boxes, prompts []DetectionBox, strip image.Rectangle) []DetectionBox {
	b := img.Bounds()
	candidates := make([]DetectionBox, 0)
	for _, box := range boxes {
		if containsBox(prompts, box) {
			continue
		}
		center := boxCenter(box)
		if center[1] <= strip.Max.Y+8 {
			continue
		}
		if boxWidth(box) < 24 || boxHeight(box) < 24 {
			continue
		}
		if center[0] < 0 || center[0] > b.Dx() || center[1] < 0 || center[1] > b.Dy() {
			continue
		}

		ratio, count := inkStats(img, box)
		if count < 80 || ratio < 0.015 || ratio > 0.45 {
			continue
		}
		candidates = append(candidates, cloneBox(box))
	}
	return candidates
}

func bestShapeScore(img image.Image, prompt shapeMask, candidate DetectionBox) float64 {
	best := 0.0
	for angle := -90.0; angle <= 90.0; angle += 5 {
		candidateMask := shapeMaskForBox(img, candidate, angle)
		score := shapeScore(prompt, candidateMask)
		if score > best {
			best = score
		}
	}
	return best
}

func shapeMaskForBox(img image.Image, box DetectionBox, angle float64) shapeMask {
	points := inkPoints(img, box)
	if len(points) == 0 {
		return shapeMask{grid: make([]bool, shapeGridSize*shapeGridSize)}
	}
	return pointsToShapeMask(points, angle)
}

func inkPoints(img image.Image, box DetectionBox) []shapePoint {
	rect := boxImageRect(img, box)
	points := make([]shapePoint, 0, rect.Dx()*rect.Dy()/8)
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			if isInkPixel(img, x, y) {
				points = append(points, shapePoint{x: float64(x - rect.Min.X), y: float64(y - rect.Min.Y)})
			}
		}
	}
	return points
}

func pointsToShapeMask(points []shapePoint, angle float64) shapeMask {
	minX, minY := points[0].x, points[0].y
	maxX, maxY := minX, minY
	for _, p := range points[1:] {
		minX = math.Min(minX, p.x)
		minY = math.Min(minY, p.y)
		maxX = math.Max(maxX, p.x)
		maxY = math.Max(maxY, p.y)
	}
	cx := (minX + maxX) / 2
	cy := (minY + maxY) / 2
	scale := math.Max(maxX-minX, maxY-minY)
	if scale <= 0 {
		scale = 1
	}

	rad := angle * math.Pi / 180
	cosA, sinA := math.Cos(rad), math.Sin(rad)
	rotated := make([]shapePoint, len(points))
	rMinX, rMinY := math.Inf(1), math.Inf(1)
	rMaxX, rMaxY := math.Inf(-1), math.Inf(-1)
	for i, p := range points {
		x := (p.x - cx) / scale
		y := (p.y - cy) / scale
		rx := x*cosA - y*sinA
		ry := x*sinA + y*cosA
		rotated[i] = shapePoint{x: rx, y: ry}
		rMinX = math.Min(rMinX, rx)
		rMinY = math.Min(rMinY, ry)
		rMaxX = math.Max(rMaxX, rx)
		rMaxY = math.Max(rMaxY, ry)
	}

	rcx := (rMinX + rMaxX) / 2
	rcy := (rMinY + rMaxY) / 2
	rscale := math.Max(rMaxX-rMinX, rMaxY-rMinY)
	if rscale <= 0 {
		rscale = 1
	}

	grid := make([]bool, shapeGridSize*shapeGridSize)
	pad := 4.0
	span := float64(shapeGridSize) - 1 - pad*2
	for _, p := range rotated {
		x := int(((p.x-rcx)/rscale+0.5)*span + pad + 0.5)
		y := int(((p.y-rcy)/rscale+0.5)*span + pad + 0.5)
		if x >= 0 && x < shapeGridSize && y >= 0 && y < shapeGridSize {
			grid[y*shapeGridSize+x] = true
		}
	}
	grid = dilateGrid(grid)
	return shapeMask{grid: grid, count: countGrid(grid)}
}

func shapeScore(a, b shapeMask) float64 {
	if a.count == 0 || b.count == 0 {
		return 0
	}
	da := dilateGrid(a.grid)
	db := dilateGrid(b.grid)
	hits := overlapGrid(da, b.grid) + overlapGrid(a.grid, db)
	return float64(hits) / float64(a.count+b.count)
}

func bestAssignment(scores [][]float64) []int {
	if len(scores) == 0 || len(scores[0]) == 0 {
		return nil
	}
	used := make([]bool, len(scores[0]))
	current := make([]int, len(scores))
	best := make([]int, len(scores))
	bestTotal := math.Inf(-1)

	var walk func(int, float64)
	walk = func(row int, total float64) {
		if row == len(scores) {
			if total > bestTotal {
				bestTotal = total
				copy(best, current)
			}
			return
		}
		for col := range scores[row] {
			if used[col] {
				continue
			}
			used[col] = true
			current[row] = col
			walk(row+1, total+scores[row][col])
			used[col] = false
		}
	}
	walk(0, 0)
	return best
}

func inkStats(img image.Image, box DetectionBox) (float64, int) {
	rect := boxImageRect(img, box)
	if rect.Empty() {
		return 0, 0
	}
	count := 0
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			if isInkPixel(img, x, y) {
				count++
			}
		}
	}
	return float64(count) / float64(rect.Dx()*rect.Dy()), count
}

func isInkPixel(img image.Image, x, y int) bool {
	r, g, b := rgbAt(img, x, y)
	maxc := math.Max(r, math.Max(g, b))
	luma := 0.299*r + 0.587*g + 0.114*b
	return luma < 90 && maxc < 120
}

func boxImageRect(img image.Image, box DetectionBox) image.Rectangle {
	b := img.Bounds()
	return image.Rect(
		b.Min.X+clampInt(box[0], 0, b.Dx()),
		b.Min.Y+clampInt(box[1], 0, b.Dy()),
		b.Min.X+clampInt(box[2], 0, b.Dx()),
		b.Min.Y+clampInt(box[3], 0, b.Dy()),
	)
}

func boxCenter(box DetectionBox) []int {
	return []int{(box[0] + box[2]) / 2, (box[1] + box[3]) / 2}
}

func boxWidth(box DetectionBox) int {
	return absInt(box[2] - box[0])
}

func boxHeight(box DetectionBox) int {
	return absInt(box[3] - box[1])
}

func cloneBox(box DetectionBox) DetectionBox {
	out := make(DetectionBox, len(box))
	copy(out, box)
	return out
}

func pointInRect(x, y int, r image.Rectangle) bool {
	return x >= r.Min.X && x < r.Max.X && y >= r.Min.Y && y < r.Max.Y
}

func containsBox(boxes []DetectionBox, target DetectionBox) bool {
	for _, box := range boxes {
		if len(box) != len(target) {
			continue
		}
		same := true
		for i := range box {
			if box[i] != target[i] {
				same = false
				break
			}
		}
		if same {
			return true
		}
	}
	return false
}

func dilateGrid(src []bool) []bool {
	dst := make([]bool, len(src))
	for y := 0; y < shapeGridSize; y++ {
		for x := 0; x < shapeGridSize; x++ {
			if !src[y*shapeGridSize+x] {
				continue
			}
			for yy := max(0, y-1); yy <= min(shapeGridSize-1, y+1); yy++ {
				for xx := max(0, x-1); xx <= min(shapeGridSize-1, x+1); xx++ {
					dst[yy*shapeGridSize+xx] = true
				}
			}
		}
	}
	return dst
}

func countGrid(grid []bool) int {
	count := 0
	for _, v := range grid {
		if v {
			count++
		}
	}
	return count
}

func overlapGrid(a, b []bool) int {
	count := 0
	for i := range a {
		if a[i] && b[i] {
			count++
		}
	}
	return count
}

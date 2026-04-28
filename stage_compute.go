package ddddocr

import (
	"encoding/json"
	"fmt"
	"image"
	"math"
	"os"
	"sort"
	"sync"

	ort "github.com/yalue/onnxruntime_go"
)

// OCRConfig configures OCR model paths, runtime path, and preprocessing.
type OCRConfig struct {
	ImagePath   string
	ModelPath   string
	CharsetPath string
	CharsetName string
	RuntimePath string
	PNGFix      bool
	Probability bool
}

// DetectionConfig configures the object detection model runtime.
type DetectionConfig struct {
	ImagePath   string
	ModelPath   string
	RuntimePath string
}

// OCRProbability is returned by OCR when probability output is enabled.
type OCRProbability struct {
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
}

// DetectionBox is a rectangle in [x1, y1, x2, y2] pixel coordinates.
type DetectionBox []int

// SlideResult contains a slider target point.
type SlideResult struct {
	Target     []int    `json:"target"`
	TargetX    int      `json:"target_x"`
	TargetY    int      `json:"target_y"`
	Confidence *float64 `json:"confidence,omitempty"`
}

var ortOnce sync.Once
var ortErr error

func ensureORT(runtimePath string) error {
	ortOnce.Do(func() {
		ort.SetSharedLibraryPath(runtimePath)
		ortErr = ort.InitializeEnvironment()
	})
	return ortErr
}

// RunOCRFile recognizes text from an image file using the provided config.
func RunOCRFile(cfg OCRConfig) (any, error) {
	img, err := LoadImageFile(cfg.ImagePath)
	if err != nil {
		return nil, err
	}
	return RunOCRImage(img, cfg)
}

// RunOCRImage recognizes text from a decoded image using the provided config.
func RunOCRImage(img image.Image, cfg OCRConfig) (any, error) {
	if err := ensureORT(cfg.RuntimePath); err != nil {
		return nil, fmt.Errorf("initialize onnxruntime: %w", err)
	}
	charset, err := loadCharset(cfg.CharsetPath, cfg.CharsetName)
	if err != nil {
		return nil, err
	}
	inputData, width, height := OCRPreprocess(img, cfg.PNGFix)
	input, err := ort.NewTensor(ort.NewShape(1, 1, int64(height), int64(width)), inputData)
	if err != nil {
		return nil, fmt.Errorf("create OCR input tensor: %w", err)
	}
	defer input.Destroy()

	session, err := ort.NewDynamicAdvancedSession(cfg.ModelPath, []string{"input1"}, []string{"387"}, nil)
	if err != nil {
		return nil, fmt.Errorf("load OCR model: %w", err)
	}
	defer session.Destroy()

	outputs := []ort.Value{nil}
	if err := session.Run([]ort.Value{input}, outputs); err != nil {
		return nil, fmt.Errorf("run OCR model: %w", err)
	}
	defer outputs[0].Destroy()

	out, ok := outputs[0].(*ort.Tensor[float32])
	if !ok {
		return nil, fmt.Errorf("unexpected OCR output type %T", outputs[0])
	}
	text, confidence := decodeCTC(out.GetData(), out.GetShape(), charset)
	if cfg.Probability {
		return OCRProbability{Text: text, Confidence: confidence}, nil
	}
	return text, nil
}

// RunDetectionFile detects target boxes from an image file.
func RunDetectionFile(cfg DetectionConfig) ([]DetectionBox, error) {
	img, err := LoadImageFile(cfg.ImagePath)
	if err != nil {
		return nil, err
	}
	return RunDetectionImage(img, cfg)
}

// RunDetectionImage detects target boxes from a decoded image.
func RunDetectionImage(img image.Image, cfg DetectionConfig) ([]DetectionBox, error) {
	if err := ensureORT(cfg.RuntimePath); err != nil {
		return nil, fmt.Errorf("initialize onnxruntime: %w", err)
	}
	inputData, ratio := DetectionPreprocess(img)
	input, err := ort.NewTensor(ort.NewShape(1, 3, detSize, detSize), inputData)
	if err != nil {
		return nil, fmt.Errorf("create detection input tensor: %w", err)
	}
	defer input.Destroy()

	session, err := ort.NewDynamicAdvancedSession(cfg.ModelPath, []string{"images"}, []string{"output"}, nil)
	if err != nil {
		return nil, fmt.Errorf("load detection model: %w", err)
	}
	defer session.Destroy()

	outputs := []ort.Value{nil}
	if err := session.Run([]ort.Value{input}, outputs); err != nil {
		return nil, fmt.Errorf("run detection model: %w", err)
	}
	defer outputs[0].Destroy()
	out, ok := outputs[0].(*ort.Tensor[float32])
	if !ok {
		return nil, fmt.Errorf("unexpected detection output type %T", outputs[0])
	}
	b := img.Bounds()
	return postprocessDetection(out.GetData(), ratio, b.Dx(), b.Dy()), nil
}

func loadCharset(path, name string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read charset %s: %w", path, err)
	}
	var all map[string][]string
	if err := json.Unmarshal(data, &all); err != nil {
		return nil, fmt.Errorf("parse charset: %w", err)
	}
	charset, ok := all[name]
	if !ok {
		return nil, fmt.Errorf("charset %q not found", name)
	}
	return charset, nil
}

func decodeCTC(data []float32, shape ort.Shape, charset []string) (string, float64) {
	var predicted []int
	var maxes []float64
	if len(shape) == 3 {
		a, b, c := int(shape[0]), int(shape[1]), int(shape[2])
		switch {
		case b == 1:
			predicted = make([]int, a)
			maxes = make([]float64, a)
			for t := 0; t < a; t++ {
				idx, maxv := argmaxProb(data[t*b*c : t*b*c+c])
				predicted[t], maxes[t] = idx, maxv
			}
		case a == 1:
			predicted = make([]int, b)
			maxes = make([]float64, b)
			for t := 0; t < b; t++ {
				idx, maxv := argmaxProb(data[t*c : t*c+c])
				predicted[t], maxes[t] = idx, maxv
			}
		default:
			predicted = make([]int, b)
			maxes = make([]float64, b)
			for t := 0; t < b; t++ {
				idx, maxv := argmaxProb(data[t*c : t*c+c])
				predicted[t], maxes[t] = idx, maxv
			}
		}
	} else if len(shape) == 2 {
		rows, cols := int(shape[0]), int(shape[1])
		predicted = make([]int, rows)
		maxes = make([]float64, rows)
		for r := 0; r < rows; r++ {
			idx, maxv := argmaxProb(data[r*cols : r*cols+cols])
			predicted[r], maxes[r] = idx, maxv
		}
	} else {
		idx, maxv := argmaxProb(data)
		predicted = []int{idx}
		maxes = []float64{maxv}
	}

	out := ""
	prev := -1
	for _, idx := range predicted {
		if idx != prev && idx != 0 && idx >= 0 && idx < len(charset) {
			out += charset[idx]
		}
		prev = idx
	}

	if len(maxes) == 0 {
		return out, 0
	}
	var sum float64
	for _, v := range maxes {
		sum += v
	}
	return out, sum / float64(len(maxes))
}

func argmaxProb(v []float32) (int, float64) {
	if len(v) == 0 {
		return 0, 0
	}
	maxIdx := 0
	maxVal := v[0]
	for i := 1; i < len(v); i++ {
		if v[i] > maxVal {
			maxVal = v[i]
			maxIdx = i
		}
	}
	var denom float64
	for _, x := range v {
		denom += math.Exp(float64(x - maxVal))
	}
	if denom == 0 {
		return maxIdx, 0
	}
	return maxIdx, 1.0 / denom
}

type detCandidate struct {
	box   [4]float64
	score float64
}

func postprocessDetection(raw []float32, ratio float64, imgW, imgH int) []DetectionBox {
	candidates := make([]detCandidate, 0)
	offset := 0
	for _, stride := range []int{8, 16, 32} {
		hsize := detSize / stride
		wsize := detSize / stride
		for gy := 0; gy < hsize; gy++ {
			for gx := 0; gx < wsize; gx++ {
				if offset+5 >= len(raw) {
					return nil
				}
				x := (float64(raw[offset]) + float64(gx)) * float64(stride)
				y := (float64(raw[offset+1]) + float64(gy)) * float64(stride)
				w := math.Exp(float64(raw[offset+2])) * float64(stride)
				h := math.Exp(float64(raw[offset+3])) * float64(stride)
				score := float64(raw[offset+4] * raw[offset+5])
				if score > 0.1 {
					candidates = append(candidates, detCandidate{
						box: [4]float64{
							(x - w/2) / ratio,
							(y - h/2) / ratio,
							(x + w/2) / ratio,
							(y + h/2) / ratio,
						},
						score: score,
					})
				}
				offset += 6
			}
		}
	}
	keep := nms(candidates, 0.45)
	boxes := make([]DetectionBox, 0, len(keep))
	for _, i := range keep {
		b := candidates[i].box
		x1 := clampInt(int(b[0]), 0, imgW)
		y1 := clampInt(int(b[1]), 0, imgH)
		x2 := clampInt(int(b[2]), 0, imgW)
		y2 := clampInt(int(b[3]), 0, imgH)
		boxes = append(boxes, DetectionBox{x1, y1, x2, y2})
	}
	return boxes
}

func nms(c []detCandidate, thr float64) []int {
	order := make([]int, len(c))
	for i := range order {
		order[i] = i
	}
	sort.Slice(order, func(i, j int) bool { return c[order[i]].score > c[order[j]].score })
	keep := make([]int, 0)
	for len(order) > 0 {
		i := order[0]
		keep = append(keep, i)
		next := order[:0]
		for _, j := range order[1:] {
			if iou(c[i].box, c[j].box) <= thr {
				next = append(next, j)
			}
		}
		order = next
	}
	return keep
}

func iou(a, b [4]float64) float64 {
	xx1 := math.Max(a[0], b[0])
	yy1 := math.Max(a[1], b[1])
	xx2 := math.Min(a[2], b[2])
	yy2 := math.Min(a[3], b[3])
	w := math.Max(0, xx2-xx1+1)
	h := math.Max(0, yy2-yy1+1)
	inter := w * h
	areaA := (a[2] - a[0] + 1) * (a[3] - a[1] + 1)
	areaB := (b[2] - b[0] + 1) * (b[3] - b[1] + 1)
	return inter / (areaA + areaB - inter)
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

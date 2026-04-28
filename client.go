package ddddocr

import (
	"image"
	"os"
	"path/filepath"
	"runtime"
)

const (
	oldOCRModelName    = "common_old.onnx"
	betaOCRModelName   = "common.onnx"
	detectionModelName = "common_det.onnx"
	charsetFileName    = "charsets.json"
	defaultRuntimeName = "onnxruntime_arm64.dylib"
)

type ClientConfig struct {
	// AssetsDir contains ONNX models, charsets.json, and optional runtime files.
	// If empty, DefaultAssetsDir is used.
	AssetsDir string
	// RuntimePath points to the ONNX Runtime shared library. If empty, the
	// client uses onnxruntime_arm64.dylib inside AssetsDir.
	RuntimePath string
	// UseBetaModel switches OCR from common_old.onnx/old charset to
	// common.onnx/beta charset.
	UseBetaModel bool
}

// OCROptions controls OCR preprocessing and response shape.
type OCROptions struct {
	// PNGFix composites transparent PNG pixels over white before OCR.
	PNGFix bool
	// Probability returns OCRProbability instead of a plain string.
	Probability bool
}

// Client is the high-level SDK entry point.
type Client struct {
	assetsDir    string
	runtimePath  string
	useBetaModel bool
}

// NewClient creates a reusable ddddocr client.
func NewClient(cfg ClientConfig) *Client {
	assetsDir := cfg.AssetsDir
	if assetsDir == "" {
		assetsDir = DefaultAssetsDir()
	}
	runtimePath := cfg.RuntimePath
	if runtimePath == "" {
		runtimePath = filepath.Join(assetsDir, defaultRuntimeName)
	}
	return &Client{
		assetsDir:    assetsDir,
		runtimePath:  runtimePath,
		useBetaModel: cfg.UseBetaModel,
	}
}

// DefaultAssetsDir returns the default assets directory used by NewClient.
func DefaultAssetsDir() string {
	if _, err := os.Stat("assets"); err == nil {
		return "assets"
	}
	if _, file, _, ok := runtime.Caller(0); ok {
		dir := filepath.Join(filepath.Dir(file), "assets")
		if _, err := os.Stat(dir); err == nil {
			return dir
		}
	}
	return "/tmp/rewrite/assets"
}

// OCRFile recognizes text from an image file.
func (c *Client) OCRFile(path string, opts OCROptions) (any, error) {
	return RunOCRFile(c.ocrConfig(path, opts))
}

// OCRImage recognizes text from an already decoded image.
func (c *Client) OCRImage(img image.Image, opts OCROptions) (any, error) {
	return RunOCRImage(img, c.ocrConfig("", opts))
}

// DetectFile returns detection boxes in [x1, y1, x2, y2] format.
func (c *Client) DetectFile(path string) ([]DetectionBox, error) {
	return RunDetectionFile(c.detectionConfig(path))
}

// DetectImage returns detection boxes from an already decoded image.
func (c *Client) DetectImage(img image.Image) ([]DetectionBox, error) {
	return RunDetectionImage(img, c.detectionConfig(""))
}

// ClickFile returns click points ordered according to the prompt icons.
func (c *Client) ClickFile(path string) (ClickCaptchaResult, error) {
	return RunClickCaptchaFile(c.detectionConfig(path))
}

// ClickImage returns ordered click points from an already decoded image.
func (c *Client) ClickImage(img image.Image) (ClickCaptchaResult, error) {
	boxes, err := c.DetectImage(img)
	if err != nil {
		return ClickCaptchaResult{}, err
	}
	return ResolveClickCaptcha(img, boxes)
}

// SlideMatchFile locates a slider target by template matching.
func (c *Client) SlideMatchFile(targetPath, backgroundPath string, simple bool) (SlideResult, error) {
	return RunSlideMatchFile(targetPath, backgroundPath, simple)
}

// SlideComparisonFile locates a slider target by comparing two images.
func (c *Client) SlideComparisonFile(targetPath, backgroundPath string) (SlideResult, error) {
	return RunSlideComparisonFile(targetPath, backgroundPath)
}

func (c *Client) ocrConfig(path string, opts OCROptions) OCRConfig {
	modelName := oldOCRModelName
	charsetName := "old"
	if c.useBetaModel {
		modelName = betaOCRModelName
		charsetName = "beta"
	}
	return OCRConfig{
		ImagePath:   path,
		ModelPath:   filepath.Join(c.assetsDir, modelName),
		CharsetPath: filepath.Join(c.assetsDir, charsetFileName),
		CharsetName: charsetName,
		RuntimePath: c.runtimePath,
		PNGFix:      opts.PNGFix,
		Probability: opts.Probability,
	}
}

func (c *Client) detectionConfig(path string) DetectionConfig {
	return DetectionConfig{
		ImagePath:   path,
		ModelPath:   filepath.Join(c.assetsDir, detectionModelName),
		RuntimePath: c.runtimePath,
	}
}

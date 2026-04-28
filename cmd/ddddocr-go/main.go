package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	ddddocr "github.com/okatu-loli/ddddocr-go"
)

func main() {
	var (
		mode         = flag.String("mode", "ocr", "ocr, detect, click, slide-match, or slide-comparison")
		imagePath    = flag.String("image", "", "image path for ocr/detect/click")
		targetPath   = flag.String("target", "", "target image path for slide modes")
		bgPath       = flag.String("background", "", "background image path for slide modes")
		assetsDir    = flag.String("assets", ddddocr.DefaultAssetsDir(), "assets directory")
		libPath      = flag.String("onnxruntime", "", "onnxruntime shared library path")
		beta         = flag.Bool("beta", false, "use common.onnx and beta charset")
		pngFix       = flag.Bool("png-fix", false, "composite RGBA input over white before OCR")
		probability  = flag.Bool("probability", false, "include OCR confidence")
		simpleTarget = flag.Bool("simple-target", false, "use direct template matching for slide-match")
	)
	flag.Parse()

	client := ddddocr.NewClient(ddddocr.ClientConfig{
		AssetsDir:    *assetsDir,
		RuntimePath:  *libPath,
		UseBetaModel: *beta,
	})

	var err error
	var out any

	switch *mode {
	case "ocr":
		if *imagePath == "" {
			fatalf("--image is required for OCR")
		}
		out, err = client.OCRFile(*imagePath, ddddocr.OCROptions{
			PNGFix:      *pngFix,
			Probability: *probability,
		})
	case "detect":
		if *imagePath == "" {
			fatalf("--image is required for detection")
		}
		out, err = client.DetectFile(*imagePath)
	case "click":
		if *imagePath == "" {
			fatalf("--image is required for click captcha")
		}
		out, err = client.ClickFile(*imagePath)
	case "slide-match":
		if *targetPath == "" || *bgPath == "" {
			fatalf("--target and --background are required for slide-match")
		}
		out, err = client.SlideMatchFile(*targetPath, *bgPath, *simpleTarget)
	case "slide-comparison":
		if *targetPath == "" || *bgPath == "" {
			fatalf("--target and --background are required for slide-comparison")
		}
		out, err = client.SlideComparisonFile(*targetPath, *bgPath)
	default:
		fatalf("unknown mode %q", *mode)
	}
	if err != nil {
		fatalf("%v", err)
	}

	switch v := out.(type) {
	case string:
		fmt.Println(v)
	default:
		enc := json.NewEncoder(os.Stdout)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(out); err != nil {
			fatalf("encode output: %v", err)
		}
	}
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

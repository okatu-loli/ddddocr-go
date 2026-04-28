# ddddocr-go

Go SDK and CLI for OCR, target detection, click-captcha ordering, and simple slider matching.

Documentation: English | [Chinese](docs/README.zh-CN.md)

## Attribution

The original models and implementation references come from [sml2h3/ddddocr](https://github.com/sml2h3/ddddocr), a MIT-licensed Python SDK for offline CAPTCHA OCR and related CAPTCHA recognition tasks.

This project packages the related capabilities as a Go SDK and CLI. See [NOTICE](NOTICE) for attribution details.

## Documentation

- [Getting Started](#sdk-usage)
- [CLI Usage](#cli-usage)
- [API Overview](#api)
- [Assets](#assets)
- [Release Checklist](#release-checklist)

## Install

```bash
go get github.com/okatu-loli/ddddocr-go
```

CLI:

```bash
go install github.com/okatu-loli/ddddocr-go/cmd/ddddocr-go@latest
```

## Assets

The SDK uses local ONNX models from `assets/`:

- `common_old.onnx`
- `common.onnx`
- `common_det.onnx`
- `charsets.json`
- `onnxruntime_arm64.dylib`

By default, `NewClient` looks for `assets/` in the current directory, then beside the installed module source. For another runtime or asset directory, pass `ClientConfig`.

```go
client := ddddocr.NewClient(ddddocr.ClientConfig{
    AssetsDir:   "/path/to/assets",
    RuntimePath: "/path/to/onnxruntime.dylib",
})
```

The bundled runtime is macOS arm64. On other platforms, provide the matching ONNX Runtime shared library path.

## SDK Usage

Click captcha:

```go
package main

import (
    "fmt"

    ddddocr "github.com/okatu-loli/ddddocr-go"
)

func main() {
    client := ddddocr.NewClient(ddddocr.ClientConfig{})

    result, err := client.ClickFile("testdata/sample.jpg")
    if err != nil {
        panic(err)
    }

    fmt.Println(result.Target)
}
```

Output:

```json
{"target":[[461,219],[216,298],[380,136]]}
```

OCR:

```go
text, err := client.OCRFile("captcha.png", ddddocr.OCROptions{})
```

Detection:

```go
boxes, err := client.DetectFile("captcha.png")
```

Slider:

```go
result, err := client.SlideMatchFile("target.png", "background.png", false)
```

## CLI Usage

```bash
ddddocr-go --mode click --image testdata/sample.jpg
ddddocr-go --mode detect --image testdata/sample.jpg
ddddocr-go --mode ocr --image captcha.png
ddddocr-go --mode slide-match --target target.png --background background.png
```

From the repo:

```bash
go run ./cmd/ddddocr-go --mode click --image testdata/sample.jpg
```

## API

The primary SDK entry point is:

```go
client := ddddocr.NewClient(ddddocr.ClientConfig{})
```

Main methods:

- `OCRFile`, `OCRImage`
- `DetectFile`, `DetectImage`
- `ClickFile`, `ClickImage`
- `SlideMatchFile`
- `SlideComparisonFile`

## Release Checklist

1. Run `go test ./...`.
2. Commit all files.
3. Push to GitHub and tag a release:

```bash
git tag v0.1.0
git push origin main --tags
```

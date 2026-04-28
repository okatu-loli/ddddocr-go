// Package ddddocr provides local OCR, target detection, click-captcha
// ordering, and simple slider-captcha helpers backed by ONNX models.
//
// The recommended entry point is Client:
//
//	client := ddddocr.NewClient(ddddocr.ClientConfig{})
//	result, err := client.ClickFile("captcha.jpg")
//
// Model files and the ONNX Runtime shared library are loaded from the assets
// directory. By default the package first checks ./assets, then the assets
// directory beside the module source. Set ClientConfig.AssetsDir or
// ClientConfig.RuntimePath when deploying with custom paths.
package ddddocr

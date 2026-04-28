# SDK API

SDK 的主要入口是 `Client`。

```go
client := ddddocr.NewClient(ddddocr.ClientConfig{})
```

## ClientConfig

```go
type ClientConfig struct {
	AssetsDir    string
	RuntimePath  string
	UseBetaModel bool
}
```

字段说明：

- `AssetsDir`：模型和字典文件目录。为空时使用 `DefaultAssetsDir()`。
- `RuntimePath`：ONNX Runtime shared library 路径。为空时使用 `AssetsDir/onnxruntime_arm64.dylib`。
- `UseBetaModel`：OCR 是否使用 `common.onnx` 和 `beta` 字符集。默认使用 `common_old.onnx` 和 `old` 字符集。

## OCR

```go
text, err := client.OCRFile("captcha.png", ddddocr.OCROptions{})
```

默认返回 `any`，实际类型通常是 `string`。

开启置信度：

```go
out, err := client.OCRFile("captcha.png", ddddocr.OCROptions{
	Probability: true,
})
prob := out.(ddddocr.OCRProbability)
```

`OCRProbability`：

```go
type OCRProbability struct {
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
}
```

如果输入 PNG 含透明通道，可以开启白底合成：

```go
out, err := client.OCRFile("captcha.png", ddddocr.OCROptions{
	PNGFix: true,
})
```

## 目标检测

```go
boxes, err := client.DetectFile("captcha.jpg")
```

返回：

```go
[]ddddocr.DetectionBox
```

`DetectionBox` 格式是：

```text
[x1, y1, x2, y2]
```

坐标原点是图片左上角。

## 点选验证码

```go
result, err := client.ClickFile("captcha.jpg")
```

返回：

```go
type ClickCaptchaResult struct {
	Target [][]int `json:"target"`
}
```

`Target` 是按提示顺序排列的点击坐标：

```json
{"target":[[461,219],[216,298],[380,136]]}
```

内部流程：

1. 使用检测模型找出提示区图标和背景候选目标。
2. 根据顶部提示区从左到右确定点击顺序。
3. 对提示图标和候选目标做形状匹配。
4. 输出每个匹配目标的中心点。

不需要识别图标语义，只需要坐标和顺序。

## 滑块匹配

模板匹配：

```go
result, err := client.SlideMatchFile("target.png", "background.png", false)
```

图片差分：

```go
result, err := client.SlideComparisonFile("target.png", "background.png")
```

返回：

```go
type SlideResult struct {
	Target     []int    `json:"target"`
	TargetX    int      `json:"target_x"`
	TargetY    int      `json:"target_y"`
	Confidence *float64 `json:"confidence,omitempty"`
}
```

## Image API

如果你已经解码了图片，可以直接使用 `image.Image` API：

```go
img, err := ddddocr.LoadImageFile("captcha.jpg")
if err != nil {
	panic(err)
}

result, err := client.ClickImage(img)
```

对应方法：

- `OCRImage`
- `DetectImage`
- `ClickImage`


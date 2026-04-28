# 命令行使用

安装：

```bash
go install github.com/okatu-loli/ddddocr-go/cmd/ddddocr-go@latest
```

在仓库内直接运行：

```bash
go run ./cmd/ddddocr-go --mode click --image testdata/sample.jpg
```

## 通用参数

```text
--mode          运行模式
--image         OCR、detect、click 模式的图片路径
--assets        assets 目录
--onnxruntime   ONNX Runtime shared library 路径
```

## OCR

```bash
ddddocr-go --mode ocr --image captcha.png
```

使用 beta 模型：

```bash
ddddocr-go --mode ocr --image captcha.png --beta
```

透明 PNG 白底合成：

```bash
ddddocr-go --mode ocr --image captcha.png --png-fix
```

输出置信度：

```bash
ddddocr-go --mode ocr --image captcha.png --probability
```

## 目标检测

```bash
ddddocr-go --mode detect --image captcha.jpg
```

输出示例：

```json
[[425,184,498,255],[177,257,256,339]]
```

每个框格式为 `[x1,y1,x2,y2]`。

## 点选验证码

```bash
ddddocr-go --mode click --image testdata/sample.jpg
```

输出示例：

```json
{"target":[[461,219],[216,298],[380,136]]}
```

`target` 是按点击顺序排列的坐标。

## 滑块

模板匹配：

```bash
ddddocr-go --mode slide-match --target target.png --background background.png
```

简单模板匹配：

```bash
ddddocr-go --mode slide-match --target target.png --background background.png --simple-target
```

差分匹配：

```bash
ddddocr-go --mode slide-comparison --target target.png --background background.png
```

## 自定义 assets

```bash
ddddocr-go \
  --mode click \
  --image captcha.jpg \
  --assets /path/to/assets \
  --onnxruntime /path/to/libonnxruntime.so
```


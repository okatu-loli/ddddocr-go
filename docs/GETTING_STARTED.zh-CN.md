# 快速开始

`ddddocr-go` 提供 Go SDK 和 CLI，用于本地识别验证码相关任务：

- OCR 文本识别
- 目标检测
- 点选验证码坐标和点击顺序识别
- 简单滑块匹配

## 来源说明

本项目的原始模型和代码实现参考来自 [sml2h3/ddddocr](https://github.com/sml2h3/ddddocr)。上游项目是 MIT 协议的 Python 通用验证码识别 SDK，由 `sml2h3` 与 `kerlomz` 共同开发。

`ddddocr-go` 在此基础上整理为 Go SDK 和 CLI，并补充了当前项目中的点选验证码顺序解析封装。详细归因见仓库根目录的 [NOTICE](../NOTICE)。

## 安装

SDK：

```bash
go get github.com/okatu-loli/ddddocr-go
```

CLI：

```bash
go install github.com/okatu-loli/ddddocr-go/cmd/ddddocr-go@latest
```

## 最小 SDK 示例

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

输出示例：

```json
[[461 219] [216 298] [380 136]]
```

`ClickFile` 返回的是按提示顺序排列的点击坐标。坐标原点是整张图片左上角。

## 最小 CLI 示例

```bash
ddddocr-go --mode click --image testdata/sample.jpg
```

输出：

```json
{"target":[[461,219],[216,298],[380,136]]}
```

## 模型文件

SDK 默认读取 `assets/` 下的模型和运行时文件。仓库中已包含 macOS arm64 的运行时：

```text
assets/
  common_old.onnx
  common.onnx
  common_det.onnx
  charsets.json
  onnxruntime_arm64.dylib
```

如果部署到 Linux、Windows 或 Intel macOS，需要提供对应平台的 ONNX Runtime shared library，并通过 `RuntimePath` 指定。

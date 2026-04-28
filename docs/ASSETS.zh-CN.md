# 模型和运行时资产

`ddddocr-go` 使用本地模型文件，不依赖远程服务。

## 来源

原始模型、字符集和部分实现参考来自 [sml2h3/ddddocr](https://github.com/sml2h3/ddddocr)。上游项目使用 MIT 协议。

发布或二次分发本项目时，请保留 `LICENSE` 和 `NOTICE`，以保留对上游项目的归因。

## assets 目录

默认目录结构：

```text
assets/
  charsets.json
  charsets.py
  common.onnx
  common_det.onnx
  common_old.onnx
  onnxruntime_arm64.dylib
```

用途：

- `common_old.onnx`：默认 OCR 模型。
- `common.onnx`：beta OCR 模型。
- `common_det.onnx`：目标检测模型，点选验证码也依赖它。
- `charsets.json`：OCR 字符集。
- `onnxruntime_arm64.dylib`：macOS arm64 的 ONNX Runtime 动态库。

## 默认查找规则

`DefaultAssetsDir()` 按顺序查找：

1. 当前工作目录下的 `assets/`
2. Go module 源码目录旁的 `assets/`
3. 回退到 `/tmp/rewrite/assets`

推荐在服务端部署时显式指定路径：

```go
client := ddddocr.NewClient(ddddocr.ClientConfig{
	AssetsDir:   "/opt/ddddocr/assets",
	RuntimePath: "/opt/ddddocr/runtime/libonnxruntime.so",
})
```

## 跨平台运行时

仓库内置的 `onnxruntime_arm64.dylib` 只适用于 macOS arm64。

其他平台需要自行准备对应 ONNX Runtime shared library：

- Linux：通常是 `libonnxruntime.so`
- macOS Intel：通常是 `libonnxruntime.dylib`
- Windows：通常是 `onnxruntime.dll`

然后通过 `RuntimePath` 或 CLI 的 `--onnxruntime` 指定。

## 发布体积

模型文件较大，仓库体积主要来自 `assets/`。如果后续要减小 Go module 体积，可以考虑：

- 使用 GitHub Release 分发模型文件
- 使用安装脚本下载 assets
- 使用 Git LFS 管理模型

当前实现为了开箱即用，保留了 `assets/`。

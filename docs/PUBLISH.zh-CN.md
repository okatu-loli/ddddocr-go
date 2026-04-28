# 发布流程

## 1. 检查模块路径

当前模块路径：

```go
module github.com/okatu-loli/ddddocr-go
```

确保 GitHub 仓库地址与模块路径一致。

## 2. 运行验证

```bash
go test ./...
go run ./cmd/ddddocr-go --mode click --image testdata/sample.jpg
go run ./examples/click testdata/sample.jpg
```

期望 click 输出：

```json
{"target":[[461,219],[216,298],[380,136]]}
```

## 3. 初始化 Git 仓库

如果当前目录还不是 Git 仓库：

```bash
git init
git add .
git commit -m "Initial SDK release"
```

如果已经是仓库：

```bash
git status
git add .
git commit -m "Prepare SDK package"
```

## 4. 推送到 GitHub

```bash
git branch -M main
git remote add origin git@github.com:okatu-loli/ddddocr-go.git
git push -u origin main
```

如果 remote 已存在，使用：

```bash
git remote set-url origin git@github.com:okatu-loli/ddddocr-go.git
git push -u origin main
```

## 5. 打 tag

Go module 推荐使用语义化版本：

```bash
git tag v0.1.0
git push origin v0.1.0
```

发布后用户可以安装：

```bash
go get github.com/okatu-loli/ddddocr-go@v0.1.0
go install github.com/okatu-loli/ddddocr-go/cmd/ddddocr-go@v0.1.0
```

## 6. 发布前检查清单

- `go test ./...` 通过
- README 中的安装路径正确
- `LICENSE` 已存在并是 MIT
- `NOTICE` 已存在并包含上游 `sml2h3/ddddocr` 归因
- `assets/` 中模型文件齐全
- `testdata/` 中样例图片不包含敏感信息
- GitHub Release 说明中标注当前内置运行时为 macOS arm64

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

## 5. 自动打 tag

推送到默认分支后，GitHub Actions 会自动执行：

- `go test ./...`
- `go build ./...`
- 构建通过后读取现有最高 `vX.Y.Z` tag，并递增 patch 版本
- 创建并推送新的 annotated tag

例如当前最高 tag 是 `v0.1.0`，默认分支构建通过后会自动创建 `v0.1.1`。

发布后用户可以按对应 tag 安装：

```bash
go get github.com/okatu-loli/ddddocr-go@v0.1.1
go install github.com/okatu-loli/ddddocr-go/cmd/ddddocr-go@v0.1.1
```

## 6. 发布前检查清单

- `go test ./...` 通过
- README 中的安装路径正确
- `LICENSE` 已存在并是 MIT
- `NOTICE` 已存在并包含上游 `sml2h3/ddddocr` 归因
- `assets/` 中模型文件齐全
- `testdata/` 中样例图片不包含敏感信息
- GitHub Release 说明中标注当前内置运行时为 macOS arm64

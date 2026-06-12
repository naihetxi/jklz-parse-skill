# Go CLI 编译指南

## 前提条件

- Go 1.21 或更高版本

## 编译步骤

```bash
cd cli
go build -o jklz-parse-go main.go
```

## 跨平台编译

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o build/jklz-parse-linux-amd64 main.go

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o build/jklz-parse-linux-arm64 main.go

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o build/jklz-parse-darwin-amd64 main.go

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o build/jklz-parse-darwin-arm64 main.go

# Windows x64
GOOS=windows GOARCH=amd64 go build -o build/jklz-parse-windows-x64.exe main.go

# Windows x86
GOOS=windows GOARCH=386 go build -o build/jklz-parse-windows-x86.exe main.go
```

## 故障排除

如果遇到编译错误：
1. 检查 Go 版本：`go version`（需要 >= 1.21）
2. 更新依赖：`go mod tidy`
3. 清理缓存：`go clean -modcache`

如无法编译，请使用 Python CLI，功能完全相同。

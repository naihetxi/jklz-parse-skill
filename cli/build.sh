#!/bin/bash
set -e

VERSION=${VERSION:-v1.0.0}
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)
PKG="github.com/jklz/jklz-parse-cli/cmd"

LDFLAGS="-s -w -X ${PKG}.version=${VERSION} -X ${PKG}.commit=${COMMIT} -X ${PKG}.date=${DATE}"

echo "Building jklz-parse-cli ${VERSION}..."
echo "Commit: ${COMMIT}"
echo "Date: ${DATE}"

mkdir -p dist

# 构建当前平台
echo "Building for current platform..."
go build -ldflags "$LDFLAGS" -o dist/jklz-parse .

# 交叉编译全平台（可选）
if [ "$1" = "all" ]; then
    echo "Cross-compiling for all platforms..."

    GOOS=linux   GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/jklz-parse-linux-amd64 .
    GOOS=linux   GOARCH=arm64 go build -ldflags "$LDFLAGS" -o dist/jklz-parse-linux-arm64 .
    GOOS=darwin  GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/jklz-parse-darwin-amd64 .
    GOOS=darwin  GOARCH=arm64 go build -ldflags "$LDFLAGS" -o dist/jklz-parse-darwin-arm64 .
    GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/jklz-parse-windows-x64.exe .
    GOOS=windows GOARCH=arm64 go build -ldflags "$LDFLAGS" -o dist/jklz-parse-windows-arm64.exe .

    echo "✓ Built 6 binaries in dist/"
else
    echo "✓ Built dist/jklz-parse"
    echo ""
    echo "To build all platforms: ./build.sh all"
fi

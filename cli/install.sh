#!/bin/bash
# jklz-parse-cli 安装脚本
set -e

INSTALL_DIR="${HOME}/.local/bin"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY_NAME="jklz-parse"

echo "正在安装 jklz-parse-cli..."

# 检测操作系统和架构
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "${ARCH}" in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo "不支持的架构: ${ARCH}"
        exit 1
        ;;
esac

BINARY_FILE="${BINARY_NAME}-${OS}-${ARCH}"
SOURCE="${SCRIPT_DIR}/build/${BINARY_FILE}"

# 创建安装目录
mkdir -p "${INSTALL_DIR}"

# 复制压缩包内的二进制文件
echo "安装 ${SOURCE}..."
if [ ! -f "${SOURCE}" ]; then
    echo "错误：未找到二进制文件 ${SOURCE}"
    echo "请确认压缩包中包含 cli/build/${BINARY_FILE}"
    exit 1
fi
cp "${SOURCE}" "${INSTALL_DIR}/${BINARY_NAME}"

# 添加执行权限
chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

# 检查 PATH
if [[ ":$PATH:" != *":${INSTALL_DIR}:"* ]]; then
    echo ""
    echo "⚠️  ${INSTALL_DIR} 不在 PATH 中"
    echo "请将以下行添加到 ~/.bashrc 或 ~/.zshrc："
    echo ""
    echo "    export PATH=\"\$HOME/.local/bin:\$PATH\""
    echo ""
fi

echo "✓ 安装完成！"
echo ""
echo "运行以下命令开始使用："
echo "    jklz-parse --help"
echo ""
echo "配置 API Key："
echo "    jklz-parse config --api-key YOUR_API_KEY"

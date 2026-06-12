#!/bin/bash
# jklz-parse-cli 安装脚本
set -e

INSTALL_DIR="${HOME}/.local/bin"
REPO="naihetxi/jklz-parse-skill"
BASE_URL="${JKLZ_INSTALL_BASE_URL:-https://github.com/${REPO}/releases/latest/download}"
RAW_BASE_URL="https://raw.githubusercontent.com/${REPO}/main/cli/build"
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

# 创建安装目录
mkdir -p "${INSTALL_DIR}"

# 下载二进制文件
echo "下载 ${BINARY_FILE}..."
DOWNLOAD_URL="${BASE_URL}/${BINARY_FILE}"

if command -v curl &> /dev/null; then
    if ! curl -fsSL "${DOWNLOAD_URL}" -o "${INSTALL_DIR}/${BINARY_NAME}"; then
        if [ -z "${JKLZ_INSTALL_BASE_URL:-}" ]; then
            FALLBACK_URL="${RAW_BASE_URL}/${BINARY_FILE}"
            echo "Release asset not found, trying repository binary: ${FALLBACK_URL}"
            curl -fsSL "${FALLBACK_URL}" -o "${INSTALL_DIR}/${BINARY_NAME}"
        else
            exit 1
        fi
    fi
elif command -v wget &> /dev/null; then
    if ! wget -qO "${INSTALL_DIR}/${BINARY_NAME}" "${DOWNLOAD_URL}"; then
        if [ -z "${JKLZ_INSTALL_BASE_URL:-}" ]; then
            FALLBACK_URL="${RAW_BASE_URL}/${BINARY_FILE}"
            echo "Release asset not found, trying repository binary: ${FALLBACK_URL}"
            wget -qO "${INSTALL_DIR}/${BINARY_NAME}" "${FALLBACK_URL}"
        else
            exit 1
        fi
    fi
else
    echo "错误：需要 curl 或 wget"
    exit 1
fi

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

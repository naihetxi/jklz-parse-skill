#!/usr/bin/env bash
set -e

REPO="naihetxi/jklz-parse-skill"
BASE_URL="${JKLZ_INSTALL_BASE_URL:-https://github.com/${REPO}/releases/latest/download}"
RAW_BASE_URL="https://raw.githubusercontent.com/${REPO}/main/cli/build"
INSTALL_DIR="${JKLZ_INSTALL_DIR:-${HOME}/.local/bin}"
EXECUTABLE_NAME="jklz-parse"

echo "=========================================="
echo "    jklz-parse CLI install"
echo "=========================================="
echo ""

OS="$(uname -s)"
case "${OS}" in
    Linux*)     OS_LOWER=linux;;
    Darwin*)    OS_LOWER=darwin;;
    *)          echo "ERROR: unsupported operating system: ${OS}"; exit 1;;
esac

ARCH="$(uname -m)"
case "${ARCH}" in
    x86_64)     ARCH_LOWER=amd64;;
    amd64)      ARCH_LOWER=amd64;;
    arm64)      ARCH_LOWER=arm64;;
    aarch64)    ARCH_LOWER=arm64;;
    *)          echo "ERROR: unsupported architecture: ${ARCH}"; exit 1;;
esac

TARGET="jklz-parse-${OS_LOWER}-${ARCH_LOWER}"
DOWNLOAD_URL="${BASE_URL}/${TARGET}"
DEST="${INSTALL_DIR}/${EXECUTABLE_NAME}"

echo "Detected platform: ${OS_LOWER}/${ARCH_LOWER}"
echo "Download URL: ${DOWNLOAD_URL}"

TMP_FILE=$(mktemp)
if ! curl -fsSL -o "${TMP_FILE}" "${DOWNLOAD_URL}"; then
    if [ -z "${JKLZ_INSTALL_BASE_URL:-}" ]; then
        FALLBACK_URL="${RAW_BASE_URL}/${TARGET}"
        echo "Release asset not found, trying repository binary: ${FALLBACK_URL}"
        if ! curl -fsSL -o "${TMP_FILE}" "${FALLBACK_URL}"; then
            echo "ERROR: download failed. Check network access or binary file: ${TARGET}"
            rm -f "${TMP_FILE}"
            exit 1
        fi
    else
        echo "ERROR: download failed. Check network access or repository file: ${TARGET}"
        rm -f "${TMP_FILE}"
        exit 1
    fi
fi

chmod +x "${TMP_FILE}"

if ! "${TMP_FILE}" --help >/dev/null 2>&1; then
    echo "ERROR: downloaded binary is not executable on this platform."
    rm -f "${TMP_FILE}"
    exit 1
fi

mkdir -p "${INSTALL_DIR}"
mv "${TMP_FILE}" "${DEST}"
chmod +x "${DEST}"

echo ""
echo "Install complete: ${DEST}"
if ! command -v jklz-parse >/dev/null 2>&1; then
    echo ""
    echo "NOTE: ${INSTALL_DIR} is not in your PATH yet."
    echo "Add this line to your shell profile, then reopen the terminal:"
    echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
fi
echo ""
echo "Configure API before first use:"
echo "   jklz-parse config --api-key YOUR_API_KEY --base-url http://192.168.42.15:15216"
echo ""
echo "Verify:"
echo "   jklz-parse health"
echo ""

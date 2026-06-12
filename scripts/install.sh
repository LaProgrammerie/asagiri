#!/usr/bin/env bash
# Install asa from GitHub Releases (macOS / Linux).
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/LaProgrammerie/asagiri/main/scripts/install.sh | bash
#   ASAGIRI_VERSION=v1.0.0 bash scripts/install.sh
set -euo pipefail

REPO="LaProgrammerie/asagiri"
BINARY="asa"
INSTALL_DIR="${INSTALL_DIR:-${HOME}/.local/bin}"
VERIFY_CHECKSUM="${VERIFY_CHECKSUM:-1}"

usage() {
  cat <<EOF
Usage: install.sh

Environment:
  ASAGIRI_VERSION   Release tag (default: latest), e.g. v1.0.0
  INSTALL_DIR       Target directory (default: \$HOME/.local/bin)
  VERIFY_CHECKSUM   1 to verify SHA256 via checksums.txt (default: 1)

Examples:
  curl -fsSL https://raw.githubusercontent.com/${REPO}/main/scripts/install.sh | bash
  ASAGIRI_VERSION=v1.0.0 INSTALL_DIR=/usr/local/bin bash install.sh
EOF
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "install.sh: missing required command: $1" >&2
    exit 1
  fi
}

need_cmd curl
need_cmd tar
need_cmd install
if [[ "${VERIFY_CHECKSUM}" == "1" ]]; then
  need_cmd sha256sum
fi

detect_os() {
  case "$(uname -s)" in
    Darwin) echo "darwin" ;;
    Linux) echo "linux" ;;
    *)
      echo "install.sh: unsupported OS: $(uname -s)" >&2
      exit 1
      ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64 | amd64) echo "amd64" ;;
    arm64 | aarch64) echo "arm64" ;;
    *)
      echo "install.sh: unsupported architecture: $(uname -m)" >&2
      exit 1
      ;;
  esac
}

resolve_version() {
  if [[ -n "${ASAGIRI_VERSION:-}" ]]; then
    echo "${ASAGIRI_VERSION}"
    return
  fi
  curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' \
    | head -n1
}

OS="$(detect_os)"
ARCH="$(detect_arch)"
VERSION="$(resolve_version)"

if [[ -z "${VERSION}" ]]; then
  echo "install.sh: could not resolve release version" >&2
  exit 1
fi

BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"
VERSION_STR="${VERSION#v}"
ARCHIVE="${BINARY}_${VERSION_STR}_${OS}_${ARCH}.tar.gz"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT

echo "==> Installing ${BINARY} ${VERSION} (${OS}/${ARCH}) to ${INSTALL_DIR}"

curl -fsSL "${BASE_URL}/checksums.txt" -o "${TMP_DIR}/checksums.txt"
curl -fsSL "${BASE_URL}/${ARCHIVE}" -o "${TMP_DIR}/${ARCHIVE}"

if [[ "${VERIFY_CHECKSUM}" == "1" ]]; then
  (
    cd "${TMP_DIR}"
    grep " ${ARCHIVE}\$" checksums.txt | sha256sum -c -
  )
fi

tar -xzf "${TMP_DIR}/${ARCHIVE}" -C "${TMP_DIR}"
mkdir -p "${INSTALL_DIR}"
install -m 755 "${TMP_DIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"

if ! command -v "${BINARY}" >/dev/null 2>&1; then
  case ":${PATH}:" in
    *":${INSTALL_DIR}:"*) ;;
    *)
      echo "==> Add ${INSTALL_DIR} to your PATH, e.g.:"
      echo "    export PATH=\"${INSTALL_DIR}:\$PATH\""
      ;;
  esac
fi

"${INSTALL_DIR}/${BINARY}" version
echo "==> Done: ${INSTALL_DIR}/${BINARY}"

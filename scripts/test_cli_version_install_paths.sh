#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

if [[ $# -gt 1 ]]; then
  echo "usage: $0 [path-to-gait-binary]" >&2
  exit 2
fi

if [[ $# -eq 1 ]]; then
  if [[ "$1" = /* ]]; then
    BIN_PATH="$1"
  else
    BIN_PATH="$(pwd)/$1"
  fi
else
  BIN_PATH="${REPO_ROOT}/gait"
  go build -o "${BIN_PATH}" ./cmd/gait
fi

if [[ ! -x "${BIN_PATH}" ]]; then
  echo "binary is not executable: ${BIN_PATH}" >&2
  exit 2
fi

if ! command -v brew >/dev/null 2>&1; then
  echo "brew is required for install-path version smoke" >&2
  exit 2
fi

detect_os() {
  case "$(uname -s)" in
    Linux) echo "linux" ;;
    Darwin) echo "darwin" ;;
    *)
      echo "unsupported OS for install-path version smoke: $(uname -s)" >&2
      exit 2
      ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *)
      echo "unsupported architecture for install-path version smoke: $(uname -m)" >&2
      exit 2
      ;;
  esac
}

sha256_file() {
  local path="$1"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$path" | awk '{print $1}'
    return
  fi
  shasum -a 256 "$path" | awk '{print $1}'
}

extract_version() {
  local bin="$1"
  "$bin" --version | awk 'NR==1{print $2}'
}

assert_version() {
  local label="$1"
  local bin="$2"
  local expected="$3"
  local got
  got="$(extract_version "$bin")"
  if [[ "$got" != "$expected" ]]; then
    echo "${label}: expected version ${expected}, got ${got}" >&2
    exit 1
  fi
  echo "${label}: version ${got}"
}

os="$(detect_os)"
arch="$(detect_arch)"
release_version_tag="v9.9.9-test"
release_version="${release_version_tag#v}"

work_dir="$(mktemp -d)"
trap 'rm -rf "${work_dir}"' EXIT

release_dir="${work_dir}/release"
install_dir="${work_dir}/install/bin"
mkdir -p "${release_dir}" "${install_dir}"

source_bin="${work_dir}/gait-source"
cp "${BIN_PATH}" "${source_bin}"
chmod 0755 "${source_bin}"
assert_version "source-build" "${source_bin}" "0.0.0-dev"

release_bin="${work_dir}/gait-release"
go build -ldflags "-X main.version=${release_version}" -o "${release_bin}" ./cmd/gait

asset_arm64="gait_${release_version}_darwin_arm64.tar.gz"
asset_amd64="gait_${release_version}_darwin_amd64.tar.gz"
project_asset_arm64="gait-local_${release_version}_darwin_arm64.tar.gz"
project_asset_amd64="gait-local_${release_version}_darwin_amd64.tar.gz"

tmp_extract="${work_dir}/extract"
mkdir -p "${tmp_extract}/gait" "${tmp_extract}/gait-local"

cp "${release_bin}" "${tmp_extract}/gait/gait"
cp "${release_bin}" "${tmp_extract}/gait-local/gait-local"

tar -czf "${release_dir}/${asset_arm64}" -C "${tmp_extract}/gait" gait
tar -czf "${release_dir}/${asset_amd64}" -C "${tmp_extract}/gait" gait
tar -czf "${release_dir}/${project_asset_arm64}" -C "${tmp_extract}/gait-local" gait-local
tar -czf "${release_dir}/${project_asset_amd64}" -C "${tmp_extract}/gait-local" gait-local

{
  printf '%s  %s\n' "$(sha256_file "${release_dir}/${asset_amd64}")" "${asset_amd64}"
  printf '%s  %s\n' "$(sha256_file "${release_dir}/${asset_arm64}")" "${asset_arm64}"
  printf '%s  %s\n' "$(sha256_file "${release_dir}/${project_asset_amd64}")" "${project_asset_amd64}"
  printf '%s  %s\n' "$(sha256_file "${release_dir}/${project_asset_arm64}")" "${project_asset_arm64}"
} > "${release_dir}/checksums.txt"

echo "==> install.sh release path"
GAIT_RELEASE_BASE_URL="file://${release_dir}" \
  bash "${REPO_ROOT}/scripts/install.sh" \
    --version "${release_version_tag}" \
    --install-dir "${install_dir}"
assert_version "install-script" "${install_dir}/gait" "${release_version}"

echo "==> homebrew formula local path"
tap_name="local/gait-version-smoke-$$"
cleanup_brew_tap() {
  brew uninstall --formula "${tap_name}/gait-local" >/dev/null 2>&1 || true
  brew untap "${tap_name}" >/dev/null 2>&1 || true
}
trap 'cleanup_brew_tap; rm -rf "${work_dir}"' EXIT
brew tap-new "${tap_name}" >/dev/null
tap_repo="$(brew --repo "${tap_name}")"
brew_formula_path="${tap_repo}/Formula/gait-local.rb"

bash "${REPO_ROOT}/scripts/render_homebrew_formula.sh" \
  --version "${release_version_tag}" \
  --repo "davidahmann/gait" \
  --project "gait-local" \
  --checksums "${release_dir}/checksums.txt" \
  --release-base-url "file://${release_dir}" \
  --out "${brew_formula_path}"

if ! grep -q "version \"${release_version}\"" "${brew_formula_path}"; then
  echo "homebrew formula missing explicit version ${release_version}" >&2
  exit 1
fi

HOMEBREW_NO_AUTO_UPDATE=1 brew install "${tap_name}/gait-local"
brew_prefix="$(brew --prefix)"
assert_version "brew-install" "${brew_prefix}/bin/gait-local" "${release_version}"
cleanup_brew_tap

echo "install-path version smoke: pass"

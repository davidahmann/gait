#!/usr/bin/env bash
set -euo pipefail

REPO_DEFAULT="davidahmann/gait"
VERSION_DEFAULT=""
CHECKSUMS_DEFAULT="dist/checksums.txt"
OUT_DEFAULT="dist/gait.rb"
PROJECT_DEFAULT="gait"
LICENSE_DEFAULT="Apache-2.0"
DESC_DEFAULT="Offline-first control plane for production AI agent tool calls"

usage() {
  cat <<'EOF'
Render a Homebrew formula from release checksums.

Usage:
  render_homebrew_formula.sh --version <tag> [--repo <owner/name>] [--checksums <path>] [--out <path>] [--project <name>]

Options:
  --version <tag>      Release tag (required, e.g. v1.7.0)
  --repo <owner/name>  GitHub repository (default: davidahmann/gait)
  --checksums <path>   checksums.txt path (default: dist/checksums.txt)
  --out <path>         Output formula path (default: dist/gait.rb)
  --project <name>     Release archive prefix (default: gait)
  --license <id>       SPDX license id (default: Apache-2.0)
  --desc <text>        Formula description
  -h, --help           Show this help
EOF
}

repo="$REPO_DEFAULT"
version="$VERSION_DEFAULT"
checksums_path="$CHECKSUMS_DEFAULT"
out_path="$OUT_DEFAULT"
project="$PROJECT_DEFAULT"
license_id="$LICENSE_DEFAULT"
formula_desc="$DESC_DEFAULT"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      [[ $# -ge 2 ]] || { echo "error: --version requires a value" >&2; exit 2; }
      version="$2"
      shift 2
      ;;
    --repo)
      [[ $# -ge 2 ]] || { echo "error: --repo requires a value" >&2; exit 2; }
      repo="$2"
      shift 2
      ;;
    --checksums)
      [[ $# -ge 2 ]] || { echo "error: --checksums requires a value" >&2; exit 2; }
      checksums_path="$2"
      shift 2
      ;;
    --out)
      [[ $# -ge 2 ]] || { echo "error: --out requires a value" >&2; exit 2; }
      out_path="$2"
      shift 2
      ;;
    --project)
      [[ $# -ge 2 ]] || { echo "error: --project requires a value" >&2; exit 2; }
      project="$2"
      shift 2
      ;;
    --license)
      [[ $# -ge 2 ]] || { echo "error: --license requires a value" >&2; exit 2; }
      license_id="$2"
      shift 2
      ;;
    --desc)
      [[ $# -ge 2 ]] || { echo "error: --desc requires a value" >&2; exit 2; }
      formula_desc="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "error: unknown option: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

if [[ -z "$version" ]]; then
  echo "error: --version is required" >&2
  exit 2
fi

if [[ ! -f "$checksums_path" ]]; then
  echo "error: checksums file not found: $checksums_path" >&2
  exit 2
fi

asset_amd64="${project}_${version}_darwin_amd64.tar.gz"
asset_arm64="${project}_${version}_darwin_arm64.tar.gz"

checksum_for() {
  local file="$1"
  local sum
  sum="$(awk -v f="$file" '$2 == f {print $1; exit}' "$checksums_path")"
  if [[ -z "$sum" ]]; then
    echo "error: checksum not found for ${file} in ${checksums_path}" >&2
    exit 2
  fi
  if [[ ! "$sum" =~ ^[a-f0-9]{64}$ ]]; then
    echo "error: invalid checksum format for ${file}: ${sum}" >&2
    exit 2
  fi
  printf '%s\n' "$sum"
}

sha_amd64="$(checksum_for "$asset_amd64")"
sha_arm64="$(checksum_for "$asset_arm64")"

homepage="https://github.com/${repo}"
release_base="https://github.com/${repo}/releases/download/${version}"
formula_name="$(basename "$out_path")"
formula_name="${formula_name%.rb}"
formula_class="$(printf '%s' "$formula_name" | awk -F'[-_]' '{for (i=1;i<=NF;i++) printf toupper(substr($i,1,1)) tolower(substr($i,2)); print ""}')"

mkdir -p "$(dirname "$out_path")"
cat >"$out_path" <<EOF
class ${formula_class} < Formula
  desc "${formula_desc}"
  homepage "${homepage}"
  license "${license_id}"

  on_macos do
    if Hardware::CPU.arm?
      url "${release_base}/${asset_arm64}"
      sha256 "${sha_arm64}"
    else
      url "${release_base}/${asset_amd64}"
      sha256 "${sha_amd64}"
    end
  end

  def install
    bin.install "${project}"
  end

  test do
    assert_match "Usage:", shell_output("\#{bin}/${project} --help")
  end
end
EOF

echo "rendered Homebrew formula: ${out_path}"

#!/usr/bin/env bash
set -euo pipefail

mode="lint"
if [[ $# -ge 1 ]]; then
  case "$1" in
    --mode)
      if [[ $# -lt 2 ]]; then
        echo "--mode requires lint or build" >&2
        exit 2
      fi
      mode="$2"
      ;;
    *)
      echo "unknown argument: $1" >&2
      exit 2
      ;;
  esac
fi

case "$mode" in
  lint|build) ;;
  *)
    echo "unsupported mode: $mode (expected lint or build)" >&2
    exit 2
    ;;
esac

if ! command -v npm >/dev/null 2>&1; then
  echo "npm is required for site stack checks" >&2
  exit 1
fi

run_site_checks() {
  local dir="$1"
  if [[ ! -d "$dir" || ! -f "$dir/package.json" ]]; then
    return 0
  fi

  if [[ ! -d "$dir/node_modules" ]]; then
    echo "missing dependencies for $dir (run: cd $dir && npm ci)" >&2
    exit 1
  fi

  echo "==> [$dir] npm run lint"
  (cd "$dir" && npm run lint)

  if [[ "$mode" == "build" ]]; then
    echo "==> [$dir] npm run build"
    (cd "$dir" && npm run build)
  fi
}

run_site_checks "docs-site"
run_site_checks "ui/local"

echo "site stack checks ($mode): pass"

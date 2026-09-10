#!/usr/bin/env bash

set -euo pipefail

browser=false
log_level="Debug"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --browser)
      browser=true
      ;;
    --log-level)
      log_level="$2"
      shift
      ;;
    -h|--help)
      cat <<'EOF'
Usage: ./dev.sh [--browser] [--log-level LEVEL]

  --browser           Also open the app at its Vite dev server URL (default
                       http://localhost:5173) in your default system browser.
                       Wails wires up a websocket bridge in dev mode, so
                       window.go/window.runtime calls still work there too.
  --log-level LEVEL    Trace, Debug, Info, Warning, or Error (default: Debug)
EOF
      exit 0
      ;;
    *)
      printf "Unknown argument: %s\n" "$1" >&2
      exit 1
      ;;
  esac
  shift
done

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$script_dir"

# `wails dev` builds a debug binary with DevTools always enabled (right-click > Inspect
# works even though build.sh's release build has DevTools off) and serves the frontend
# from disk with hot reload instead of the embedded production bundle.
dev_args=(-loglevel "$log_level")
if [[ "$browser" == true ]]; then
  dev_args+=(-browser)
fi

go run github.com/wailsapp/wails/v2/cmd/wails@v2.15.0 dev "${dev_args[@]}"

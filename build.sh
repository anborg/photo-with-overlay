#!/usr/bin/env bash

set -euo pipefail

skip_tests=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --skip-tests)
      skip_tests=true
      ;;
    -h|--help)
      cat <<'EOF'
Usage: ./build.sh [--skip-tests]
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

if [[ "$skip_tests" != true ]]; then
  go test ./...
fi

output_name="PhotoWithOverlay"
go run github.com/wailsapp/wails/v2/cmd/wails@v2.11.0 build \
  -s \
  -skipbindings \
  -trimpath \
  -ldflags "-s -w" \
  -o "$output_name"

output_dir="$script_dir/build/bin"
binary_path="$output_dir/$output_name"
app_bundle_path="$output_dir/$output_name.app"
slug_binary_path="$output_dir/photo-with-overlay"
slug_app_bundle_path="$output_dir/photo-with-overlay.app"

artifact_path=""
size_path=""

if [[ -f "$binary_path" ]]; then
  artifact_path="$binary_path"
  size_path="$binary_path"
elif [[ -d "$app_bundle_path" ]]; then
  artifact_path="$app_bundle_path"
  if [[ -f "$app_bundle_path/Contents/MacOS/$output_name" ]]; then
    size_path="$app_bundle_path/Contents/MacOS/$output_name"
  fi
elif [[ -f "$slug_binary_path" ]]; then
  artifact_path="$slug_binary_path"
  size_path="$slug_binary_path"
elif [[ -d "$slug_app_bundle_path" ]]; then
  artifact_path="$slug_app_bundle_path"
  if [[ -f "$slug_app_bundle_path/Contents/MacOS/$output_name" ]]; then
    size_path="$slug_app_bundle_path/Contents/MacOS/$output_name"
  elif [[ -f "$slug_app_bundle_path/Contents/MacOS/photo-with-overlay" ]]; then
    size_path="$slug_app_bundle_path/Contents/MacOS/photo-with-overlay"
  fi
else
  printf "Build finished, but no artifact was found in %s.\n" "$output_dir" >&2
  exit 1
fi

size_bytes() {
  stat -f '%z' "$1" 2>/dev/null || stat -c '%s' "$1"
}

reported_size_bytes=0
if [[ -n "$size_path" ]]; then
  reported_size_bytes="$(size_bytes "$size_path")"
fi

printf "Built: %s" "$artifact_path"
if [[ "$reported_size_bytes" -gt 0 ]]; then
  printf " (%.1f MB)" "$(awk -v bytes="$reported_size_bytes" 'BEGIN { print bytes / 1048576 }')"
fi
printf "\n"

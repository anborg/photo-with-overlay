#!/usr/bin/env bash

set -euo pipefail

skip_tests=false
skip_compression=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --skip-tests)
      skip_tests=true
      ;;
    --skip-compression)
      skip_compression=true
      ;;
    -h|--help)
      cat <<'EOF'
Usage: ./buildrelease.sh [--skip-tests] [--skip-compression]
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
  -ldflags "-s -w -buildid=" \
  -o "$output_name"

output_dir="$script_dir/build/bin"
binary_path="$output_dir/$output_name"
app_bundle_path="$output_dir/$output_name.app"
slug_binary_path="$output_dir/photo-with-overlay"
slug_app_bundle_path="$output_dir/photo-with-overlay.app"

artifact_path=""
compressible_path=""

if [[ -f "$binary_path" ]]; then
  artifact_path="$binary_path"
  compressible_path="$binary_path"
elif [[ -d "$app_bundle_path" ]]; then
  artifact_path="$app_bundle_path"
  if [[ -f "$app_bundle_path/Contents/MacOS/$output_name" ]]; then
    compressible_path="$app_bundle_path/Contents/MacOS/$output_name"
  fi
elif [[ -f "$slug_binary_path" ]]; then
  artifact_path="$slug_binary_path"
  compressible_path="$slug_binary_path"
elif [[ -d "$slug_app_bundle_path" ]]; then
  artifact_path="$slug_app_bundle_path"
  if [[ -f "$slug_app_bundle_path/Contents/MacOS/$output_name" ]]; then
    compressible_path="$slug_app_bundle_path/Contents/MacOS/$output_name"
  elif [[ -f "$slug_app_bundle_path/Contents/MacOS/photo-with-overlay" ]]; then
    compressible_path="$slug_app_bundle_path/Contents/MacOS/photo-with-overlay"
  fi
else
  printf "Release build finished, but no artifact was found in %s.\n" "$output_dir" >&2
  exit 1
fi

size_bytes() {
  stat -f '%z' "$1" 2>/dev/null || stat -c '%s' "$1"
}

human_mb() {
  awk -v bytes="$1" 'BEGIN { printf "%.2f", bytes / 1048576 }'
}

uncompressed_size=0
compressed=false

if [[ -n "$compressible_path" ]]; then
  uncompressed_size="$(size_bytes "$compressible_path")"
fi

if [[ "$skip_compression" != true ]]; then
  if command -v upx >/dev/null 2>&1; then
    if [[ -n "$compressible_path" ]]; then
      backup_path="$compressible_path.uncompressed"
      cp "$compressible_path" "$backup_path"
      if upx --best --lzma --no-progress "$compressible_path"; then
        compressed=true
      else
        cp "$backup_path" "$compressible_path"
        printf "Warning: UPX compression failed; keeping the stripped binary.\n" >&2
      fi
      rm -f "$backup_path"
    else
      printf "Warning: no standalone executable was found to compress.\n" >&2
    fi
  else
    printf "Warning: UPX was not found; the binary is stripped but not executable-compressed.\n" >&2
  fi
fi

reported_size_path="$artifact_path"
if [[ -f "$compressible_path" ]]; then
  reported_size_path="$compressible_path"
fi

final_size="$(size_bytes "$reported_size_path")"
saved_percent="0.0"

if [[ "$uncompressed_size" -gt 0 && "$final_size" -le "$uncompressed_size" ]]; then
  saved_percent="$(awk -v before="$uncompressed_size" -v after="$final_size" 'BEGIN { printf "%.1f", ((before - after) / before) * 100 }')"
fi

printf "Built release: %s\n" "$artifact_path"
printf "Size: %s MB" "$(human_mb "$final_size")"
if [[ "$compressed" == true ]]; then
  printf " (saved %s%%)" "$saved_percent"
fi
printf "\n"

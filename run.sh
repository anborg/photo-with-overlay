#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
platform="$(uname -s)"

app_candidates=(
  "$script_dir/build/bin/PhotoWithOverlay.app"
  "$script_dir/build/bin/photo-with-overlay.app"
)

exec_candidates=(
  "$script_dir/build/bin/PhotoWithOverlay"
  "$script_dir/build/bin/PhotoWithOverlay.app/Contents/MacOS/PhotoWithOverlay"
  "$script_dir/build/bin/photo-with-overlay"
  "$script_dir/build/bin/photo-with-overlay.app/Contents/MacOS/photo-with-overlay"
  "$script_dir/build/bin/PhotoWithOverlay.exe"
)

if [[ "$platform" == "Darwin" ]]; then
  for app_path in "${app_candidates[@]}"; do
    if [[ -d "$app_path" ]]; then
      exec open "$app_path"
    fi
  done
fi

for executable_path in "${exec_candidates[@]}"; do
  if [[ -f "$executable_path" ]]; then
    exec "$executable_path"
  fi
done

printf "Application not found in '%s/build/bin'. Build it first.\n" "$script_dir" >&2
printf "Use the repo's build flow first, such as 'build.ps1' or 'buildrelease.ps1' on Windows.\n" >&2
exit 1

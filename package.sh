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
Usage: ./package.sh [--skip-tests] [--skip-compression]
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

if [[ "$(uname -s)" != "Darwin" ]]; then
  printf "package.sh creates a macOS DMG and must be run on macOS.\n" >&2
  exit 1
fi

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$script_dir"

build_args=()
if [[ "$skip_tests" == true ]]; then
  build_args+=(--skip-tests)
fi
if [[ "$skip_compression" == true ]]; then
  build_args+=(--skip-compression)
fi

"$script_dir/buildrelease.sh" "${build_args[@]}"

app_bundle_path="$script_dir/build/bin/photo-with-overlay.app"
if [[ ! -d "$app_bundle_path" ]]; then
  alt_app_bundle_path="$script_dir/build/bin/PhotoWithOverlay.app"
  if [[ -d "$alt_app_bundle_path" ]]; then
    app_bundle_path="$alt_app_bundle_path"
  else
    printf "App bundle was not found after the release build.\n" >&2
    exit 1
  fi
fi

dmg_name="PhotoWithOverlay-macOS"
staging_root="$script_dir/build/dmg"
staging_dir="$staging_root/$dmg_name"
dmg_path="$script_dir/build/$dmg_name.dmg"

rm -rf "$staging_root"
mkdir -p "$staging_dir"
cp -R "$app_bundle_path" "$staging_dir/"
ln -s /Applications "$staging_dir/Applications"
rm -f "$dmg_path"

hdiutil create \
  -volname "Photo With Overlay" \
  -srcfolder "$staging_dir" \
  -ov \
  -format UDZO \
  "$dmg_path"

printf "Created DMG: %s\n" "$dmg_path"

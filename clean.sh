#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
build_output="$script_dir/build/bin"

if [[ "$(cd "$script_dir/build" && pwd)/bin" != "$build_output" ]]; then
  printf "Refusing to clean unexpected path: %s\n" "$build_output" >&2
  exit 1
fi

if [[ ! -e "$build_output" ]]; then
  printf "Already clean: build/bin does not exist.\n"
  exit 0
fi

rm -rf "$build_output"
printf "Cleaned: %s\n" "$build_output"

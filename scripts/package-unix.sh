#!/usr/bin/env bash
set -euo pipefail

version="${1:?version is required}"
platform="${2:?platform is required}"
arch="${3:?architecture is required}"

if [[ "$platform" != "macos" && "$platform" != "linux" ]]; then
  echo "unsupported platform: $platform" >&2
  exit 2
fi
if [[ "$arch" != "x64" && "$arch" != "arm64" ]]; then
  echo "unsupported architecture: $arch" >&2
  exit 2
fi

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
build_dir="$repo_root/build/qt-release-${platform}-${arch}"
stage_dir="$build_dir/stage"
dist_dir="$repo_root/dist"

rm -rf "$build_dir" "$stage_dir"
mkdir -p "$build_dir" "$stage_dir" "$dist_dir"

cmake -S "$repo_root" -B "$build_dir" -G Ninja \
  -DCMAKE_BUILD_TYPE=Release
cmake --build "$build_dir" --parallel
ctest --test-dir "$build_dir" --output-on-failure
cmake --install "$build_dir" --config Release --prefix "$stage_dir"

if [[ "$platform" == "macos" ]]; then
  app_path="$stage_dir/HexImg.app"
  [[ -d "$app_path" ]] || { echo "HexImg.app was not installed" >&2; exit 1; }
  ditto -c -k --sequesterRsrc --keepParent \
    "$app_path" "$dist_dir/HexImg-macos-${arch}-${version}.zip"
else
  tar -C "$stage_dir" -czf \
    "$dist_dir/HexImg-linux-${arch}-${version}.tar.gz" .
fi

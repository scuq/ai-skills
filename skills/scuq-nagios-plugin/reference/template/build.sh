#!/usr/bin/env bash
# build.sh - build the plugin as static Linux binaries for amd64 and arm64.
#
# Usage: bash build.sh [VERSION]
#
# Without VERSION, the script uses the output of git describe, or 0.0.0.
# The binaries go to dist/<module name>-linux-<architecture>.
#
# Exit status:
#   0  The tests passed, and all binaries are static.
#   1  A test failed, a build failed, or a binary is not static.

set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

name=$(basename "$(go list -m)")
version=${1:-$(git describe --tags --always 2>/dev/null || echo 0.0.0)}

go vet ./...
go test ./...

mkdir -p dist
for arch in amd64 arm64; do
  out="dist/${name}-linux-${arch}"
  CGO_ENABLED=0 GOOS=linux GOARCH="$arch" \
    go build -trimpath -ldflags "-s -w -X main.version=${version}" -o "$out" .
  # Go records the build settings in the binary. CGO_ENABLED=0 means no
  # libc and no dynamic loader.
  if ! go version -m "$out" | grep -q 'CGO_ENABLED=0'; then
    echo "The binary $out is not static. Set CGO_ENABLED=0 and build again." >&2
    exit 1
  fi
  echo "Built $out ($version)"
done

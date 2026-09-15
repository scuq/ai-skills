#!/usr/bin/env bash
# build-release.sh - test the repository and build the release assets for a tag.
#
# Usage: scripts/build-release.sh TAG
#
# TAG has the format vMAJOR.MINOR.PATCH. It must match the VERSION file.
# The script writes these files to dist/:
#   ai-skillsctl-linux-amd64    static binaries of ai-skillsctl
#   ai-skillsctl-linux-arm64
#   ai-skillsctl-darwin-amd64
#   ai-skillsctl-darwin-arm64
#   ai-skills-bundle.tar.gz     the Git-tracked files of VERSION, skills/, and agents/
#   SHA256SUMS                  the SHA-256 sums of the other files
#
# Exit status:
#   0  The tests passed, and all assets exist.
#   1  The tag is not valid, a test failed, or a build failed.
#   2  TAG is missing.

set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."

if [[ $# -ne 1 ]]; then
  echo "Usage: scripts/build-release.sh TAG" >&2
  exit 2
fi
tag=$1
if [[ ! $tag =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "The tag $tag is not in the format vMAJOR.MINOR.PATCH." >&2
  exit 1
fi
version=${tag#v}
file_version=$(<VERSION)
if [[ $file_version != "$version" ]]; then
  echo "The VERSION file has $file_version, but the tag is $tag. Write $version in VERSION, commit, and tag again." >&2
  exit 1
fi

bash scripts/test.sh

rm -rf dist
mkdir dist
for target in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64; do
  goos=${target%/*}
  goarch=${target#*/}
  CGO_ENABLED=0 GOOS=$goos GOARCH=$goarch \
    go build -trimpath -ldflags "-s -w -X main.version=${version}" \
    -o "dist/ai-skillsctl-${goos}-${goarch}" ./cmd/ai-skillsctl
done
if ! go version -m dist/ai-skillsctl-linux-amd64 | grep -q 'CGO_ENABLED=0'; then
  echo "The binary dist/ai-skillsctl-linux-amd64 is not static. Set CGO_ENABLED=0 and build again." >&2
  exit 1
fi

# A fixed owner, order, and time give the same archive for the same commit.
mtime=$(git log -1 --format=%ct)
git ls-files -z VERSION skills agents |
  tar --null --files-from=- --sort=name --owner=0 --group=0 --numeric-owner \
    --mtime="@${mtime}" -cf - |
  gzip -n >dist/ai-skills-bundle.tar.gz

(cd dist && sha256sum ai-skillsctl-* ai-skills-bundle.tar.gz >SHA256SUMS)
ls -l dist

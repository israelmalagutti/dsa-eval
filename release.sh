#!/usr/bin/env bash
# Builds dsa-eval for every platform and, after confirmation, tags and publishes a GitHub release.
#
#   ./release.sh v0.1.0
#
# Nothing leaves the machine until you answer "y". Answering anything else leaves the
# build in dist/ for inspection.
set -euo pipefail

version="${1:-}"
targets=(linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64)

die() { echo "release: $*" >&2; exit 1; }

[[ "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || die "usage: ./release.sh vMAJOR.MINOR.PATCH"
cd "$(dirname "$0")"

[[ -z "$(git status --porcelain)" ]] || die "working tree has uncommitted changes"
git fetch --quiet --tags origin
[[ "$(git rev-parse HEAD)" == "$(git rev-parse '@{u}')" ]] || die "HEAD is not in sync with $(git rev-parse --abbrev-ref '@{u}'); push or pull first"
! git rev-parse -q --verify "refs/tags/$version" >/dev/null || die "tag $version already exists"

echo "==> vet and test"
go vet ./...
go test -count=1 ./...

echo "==> build $version"
rm -rf dist && mkdir dist
for t in "${targets[@]}"; do
	os="${t%/*}" arch="${t#*/}" ext=""
	[[ "$os" == windows ]] && ext=".exe"
	name="dsa-eval_${version}_${os}_${arch}"
	mkdir "dist/$name"
	GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o "dist/$name/dsa-eval$ext" .
	cp LICENSE README.md "dist/$name/"
	if [[ "$os" == windows ]]; then
		(cd dist && zip -qr "$name.zip" "$name")
	else
		tar -C dist -czf "dist/$name.tar.gz" "$name"
	fi
	rm -rf "dist/$name"
done
(cd dist && sha256sum ./*.tar.gz ./*.zip | sed 's| \./| |' > checksums.txt)

echo
ls -lh dist | tail -n +2
echo
read -r -p "Tag $version at $(git rev-parse --short HEAD) and publish these files to GitHub? [y/N] " answer
[[ "$answer" == y ]] || { echo "Not published. Build left in dist/."; exit 0; }

git tag -a "$version" -m "$version"
git push origin "$version"
gh release create "$version" dist/*.tar.gz dist/*.zip dist/checksums.txt --title "$version" --generate-notes

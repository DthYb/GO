#!/usr/bin/env bash
# build.sh — cross-compilation de gopack pour les 3 plateformes cibles.
#
# Usage : ./build.sh [version]        (ex : ./build.sh 1.0.0)
#
# Produit dans dist/ :
#   gopack-linux-amd64
#   gopack-windows-amd64.exe
#   gopack-linux-arm64
#   SHA256SUMS
set -euo pipefail

VERSION="${1:-dev}"
DIST="dist"

# Matrice de build : "GOOS/GOARCH"
PLATFORMS=(
  "linux/amd64"
  "windows/amd64"
  "linux/arm64"
)

rm -rf "$DIST"
mkdir -p "$DIST"

for platform in "${PLATFORMS[@]}"; do
  GOOS="${platform%/*}"    # partie avant le /
  GOARCH="${platform#*/}"  # partie après le /

  output="$DIST/gopack-$GOOS-$GOARCH"
  # Sous Windows, un exécutable doit porter l'extension .exe
  if [ "$GOOS" = "windows" ]; then
    output="$output.exe"
  fi

  echo "→ build $GOOS/$GOARCH ($output)"
  # CGO_ENABLED=0  : binaire 100% statique, indispensable en cross-compilation
  # -trimpath      : chemins reproductibles (pas de chemins absolus locaux)
  # -s -w          : binaire allégé (sans table de symboles ni infos DWARF)
  # -X main.version: injection de la version dans la variable main.version
  CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build \
    -trimpath \
    -ldflags "-s -w -X main.version=$VERSION" \
    -o "$output" .
done

# Checksums : permettent de vérifier l'intégrité des binaires distribués.
(cd "$DIST" && sha256sum gopack-* > SHA256SUMS)

echo
echo "Binaires générés :"
ls -lh "$DIST"

#!/usr/bin/env bash
# Fetch Dendrite source for local embedding.
# beetrix uses a replace directive in go.mod to point to ./_dendrite,
# which contains a pkg/embed/ package that wraps Dendrite's internal APIs.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
DENDRITE_DIR="$PROJECT_DIR/_dendrite"

if [ -d "$DENDRITE_DIR" ]; then
    echo "Dendrite already present at $DENDRITE_DIR"
    echo "To update: rm -rf _dendrite && $0"
    exit 0
fi

echo "Cloning Dendrite..."
git clone --depth 1 https://github.com/element-hq/dendrite.git "$DENDRITE_DIR"

# Add the embed package
echo "Adding pkg/embed/ wrapper..."
mkdir -p "$DENDRITE_DIR/pkg/embed"

# The embed package files are tracked in the beetrix repo at _dendrite_embed/
if [ -d "$PROJECT_DIR/_dendrite_embed" ]; then
    cp "$PROJECT_DIR/_dendrite_embed/"*.go "$DENDRITE_DIR/pkg/embed/"
else
    echo "Warning: _dendrite_embed/ not found. You may need to create pkg/embed/ files manually."
fi

echo "Done. Dendrite is ready at $DENDRITE_DIR"

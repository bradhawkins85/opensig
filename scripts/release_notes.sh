#!/usr/bin/env bash
set -euo pipefail
tag="${1:-v0.1.0}"
echo "## ${tag}" > RELEASE_NOTES.md
echo "- Initial skeleton release" >> RELEASE_NOTES.md

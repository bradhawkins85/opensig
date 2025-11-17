#!/usr/bin/env bash
set -e
echo "Formatting Go..."; (cd server && go fmt ./... || true)
echo "Typechecking web..."; (cd web && npm run -s lint || true)

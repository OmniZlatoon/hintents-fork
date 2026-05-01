#!/bin/bash
# Copyright 2026 Erst Users
# SPDX-License-Identifier: Apache-2.0

# Test script to verify strict linting configuration
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${REPO_ROOT}"

echo "Verifying strict linting configuration..."

# golangci-lint v2 no longer accepts bare filenames — it requires a package
# path (e.g. ./...). We create a self-contained temp module so the tool can
# resolve the package correctly regardless of the cwd the CI script is invoked
# from (see ci-standardization.yml which runs scripts from /tmp).
TMP_PKG="$(mktemp -d)"
trap 'rm -rf "${TMP_PKG}"' EXIT

# Write a minimal module manifest.
cat > "${TMP_PKG}/go.mod" << 'EOF'
module linttest

go 1.25.0
EOF

# Write a Go file that has a genuine unused-variable lint issue.
cat > "${TMP_PKG}/main.go" << 'EOF'
package main

import "fmt"

func main() {
	var unused = 1
	fmt.Println("Hello")
}
EOF

# Run golangci-lint and expect it to catch the unused variable.
if command -v golangci-lint &> /dev/null; then
    # Capture output to verify it's the expected lint error
    LINT_OUT=$(golangci-lint run --config="${REPO_ROOT}/.golangci.yml" "${TMP_PKG}/..." 2>&1)
    LINT_EXIT=$?
    
    if [ $LINT_EXIT -eq 0 ]; then
        echo "[FAIL] Strict linting failed to catch unused variable"
        exit 1
    elif echo "$LINT_OUT" | grep -q "unused"; then
        echo "[OK] Strict linting caught unused variable"
    else
        echo "[FAIL] golangci-lint failed with unexpected error:"
        echo "$LINT_OUT"
        exit 1
    fi
else
    echo "golangci-lint not available, skipping verification"
fi

echo "Strict linting verification passed"

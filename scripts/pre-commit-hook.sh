#!/bin/sh

# Pre-commit hook to run gofmt/goimports on staged Go files
# and npx prettier on staged frontend files.

# ---------------------------------------------------------------------------
# Go formatting
# ---------------------------------------------------------------------------

# Get staged Go files (excluding generated files)
STAGED_GO_FILES=$(git diff --cached --name-only --diff-filter=ACMR | grep '\.go$' | grep -v 'gen\.go' | grep -v 'oapi-codegen\.gen\.go')

if [ -n "$STAGED_GO_FILES" ]; then
    echo "Running gofmt and goimports on staged Go files..."

    # Run gofmt -s -w on staged files
    gofmt -s -w $STAGED_GO_FILES

    # Run goimports -w on staged files
    # We use go run to ensure it's available without manual global installation
    for file in $STAGED_GO_FILES; do
        go run golang.org/x/tools/cmd/goimports@latest -w "$file"
    done

    # Re-stage the Go changes
    git add $STAGED_GO_FILES
fi

# ---------------------------------------------------------------------------
# Frontend formatting (Prettier)
# ---------------------------------------------------------------------------

STAGED_FRONTEND_FILES=$(git diff --cached --name-only --diff-filter=ACMR | grep -E '\.(ts|svelte|js|css|json)$' | grep '^frontend/')

if [ -n "$STAGED_FRONTEND_FILES" ]; then
    if ! command -v npx > /dev/null 2>&1; then
        echo "Warning: npx is not installed. Skipping Prettier formatting."
    else
        FRONTEND_DIR="$(git rev-parse --show-toplevel)/frontend"

        # Install prettier if not already present
        if ! npx --no-install prettier --version > /dev/null 2>&1; then
            echo "Prettier not found. Installing..."
            (cd "$FRONTEND_DIR" && npm install --save-dev prettier prettier-plugin-svelte)
        fi

        echo "Running prettier on staged frontend files..."
        for file in $STAGED_FRONTEND_FILES; do
            npx --prefix "$FRONTEND_DIR" prettier --write "$(git rev-parse --show-toplevel)/$file"
        done

        # Re-stage the frontend changes
        git add $STAGED_FRONTEND_FILES
    fi
fi

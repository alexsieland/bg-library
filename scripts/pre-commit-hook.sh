#!/bin/sh

# Pre-commit hook to run gofmt, goimports, and Prettier on staged files.
# This hook works correctly regardless of which directory you commit from.

# Determine the repository root to ensure all operations are relative to it
REPO_ROOT=$(git rev-parse --show-toplevel)
cd "$REPO_ROOT" || exit 1

# Get staged Go files (excluding generated files)
STAGED_GO_FILES=$(git diff --cached --name-only --diff-filter=ACMR | grep '\.go$' | grep -v 'gen\.go' | grep -v 'oapi-codegen\.gen\.go')

# Quick check for staged frontend files (used for the simple existence check)
STAGED_FRONTEND_FILES=$(git diff --cached --name-only --diff-filter=ACMR | grep '^frontend/' || true)

if [ -z "$STAGED_GO_FILES" ] && [ -z "$STAGED_FRONTEND_FILES" ]; then
    exit 0
fi

echo "Running gofmt, goimports, and Prettier on staged files..."

# Run gofmt -s -w on staged files
if [ -n "$STAGED_GO_FILES" ]; then
    gofmt -s -w $STAGED_GO_FILES

    # Run goimports -w on staged files
    # We use go run to ensure it's available without manual global installation
    for file in $STAGED_GO_FILES; do
        go run golang.org/x/tools/cmd/goimports@latest -w "$file"
    done
fi

# Run Prettier on staged frontend files (if any)
# We construct a null-separated list and use xargs -0 to handle filenames safely.
if [ -n "$STAGED_FRONTEND_FILES" ]; then
    echo "Installing frontend dependencies..."
    cd "$REPO_ROOT/frontend"
    npm install

    echo "Running Prettier on staged frontend files..."

    # Run Prettier from the frontend directory so it can find plugins in node_modules.
    # Get the list of staged frontend files and remove the 'frontend/' prefix for processing.
    git diff --cached --name-only --diff-filter=ACMR -z \
      | xargs -0 -n1 printf '%s\n' \
      | grep '^frontend/' \
      | sed 's|^frontend/||' \
      | tr '\n' '\0' \
      | xargs -0 -- npx prettier --write --

    cd "$REPO_ROOT"

    # Re-stage any frontend files that Prettier modified.
    git diff --cached --name-only --diff-filter=ACMR -z \
      | xargs -0 -n1 printf '%s\n' \
      | grep '^frontend/' \
      | tr '\n' '\0' \
      | xargs -0 -- git add --
fi

# Re-stage the Go changes
if [ -n "$STAGED_GO_FILES" ]; then
    git add $STAGED_GO_FILES
fi

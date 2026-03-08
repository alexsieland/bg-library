#!/bin/sh

# Pre-commit hook to run gofmt, goimports, and Prettier on staged files.

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
    cd frontend
    npm install
    cd ..

    echo "Running Prettier on staged frontend files..."

    # Run Prettier on only the staged frontend files using a pipeline that is safe for filenames.
    # Steps:
    # 1) git produces a null-delimited list (-z)
    # 2) xargs -0 -n1 printf '%s\n' converts to newline-separated list
    # 3) grep filters to paths under 'frontend/'
    # 4) tr converts newlines back to nulls
    # 5) xargs -0 invokes prettier with the matched files
    git diff --cached --name-only --diff-filter=ACMR -z \
      | xargs -0 -n1 printf '%s\n' \
      | grep '^frontend/' \
      | tr '\n' '\0' \
      | xargs -0 -- npx prettier --write --

    # Re-stage any frontend files that Prettier modified by re-using the same safe pipeline.
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

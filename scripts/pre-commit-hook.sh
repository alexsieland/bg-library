#!/bin/sh

# Pre-commit hook to run gofmt and goimports on staged files.

# Get staged Go files (excluding generated files)
STAGED_GO_FILES=$(git diff --cached --name-only --diff-filter=ACMR | grep '\.go$' | grep -v 'gen\.go' | grep -v 'oapi-codegen\.gen\.go')

if [ -z "$STAGED_GO_FILES" ]; then
    exit 0
fi

echo "Running gofmt and goimports on staged files..."

# Run gofmt -s -w on staged files
gofmt -s -w $STAGED_GO_FILES

# Run goimports -w on staged files
# We use go run to ensure it's available without manual global installation
for file in $STAGED_GO_FILES; do
    go run golang.org/x/tools/cmd/goimports@latest -w "$file"
done

# Re-stage the changes
git add $STAGED_GO_FILES

pre-commit-hook

This repository includes a Git pre-commit hook script at `scripts/pre-commit-hook.sh` which is intended to be installed in `.git/hooks/pre-commit` (for example via a setup script or manually).

What the hook does

- For staged Go files (excluding generated files):
  - Runs `gofmt -s -w` on the files.
  - Runs `goimports -w` via `go run golang.org/x/tools/cmd/goimports@latest` on each file.
  - Re-stages any modified Go files.

- For staged frontend files under `frontend/`:
  - Runs `npx prettier --write` on only the staged frontend files (using a null-safe pipeline to handle filenames).
  - Re-stages any frontend files that were modified by Prettier.

Notes

- The hook uses `npx prettier` so it requires Node and the project's devDependencies (including `prettier`) to be available. If you prefer not to rely on `npx`, you can install Prettier globally or modify the hook to run a project-local script.
- The pipeline that selects staged frontend files is safe for filenames containing spaces or special characters.
- If your CI enforces formatting, make sure Prettier and goimports versions are pinned in the project devDependencies or tooling scripts to avoid unexpected diffs.


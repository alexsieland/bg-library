# Board Game Library Management System

Full-stack app: Go backend (Gin), Svelte + TypeScript frontend (Vite, Tailwind), PostgreSQL. Development is **contract-first**: `swagger/api.yaml` (OpenAPI) is the source of truth.

## Documentation

Read the relevant doc in `docs/` before working in that area:

- `docs/project-overview.md`: architecture and technology stack
- `docs/coding-guidelines.md`: backend and frontend coding standards
- `docs/testing-guidelines.md`: testing strategy and best practices
- `docs/functional-requirements.md`: features and user workflows
- `docs/frontend-styleguide.md`: UI/UX standards and design principles
- `docs/enduser-documentation-styleguide.md`: rules for all librarian-facing manual pages in `docs/manual/`

## Key Principles

1. **Contract-first**: edit `swagger/api.yaml` first, then regenerate.
2. **Type safety**: use generated types everywhere.
3. **Soft deletes only**: never hard delete; preserve the audit trail.
4. **Error visibility**: log server-side, return appropriate status codes.
5. **Transactions** for multi-step operations.
6. **Tests**: behavior-driven naming; test workflows, not implementation.
7. **Accessibility**: WCAG 2.1 AA on the frontend.
8. **Backend owns business logic and sanitization.**
9. Board game titles must be searchable accent-folded (e.g. "barenpark" matches "Bärenpark").

## Commands

Backend (`backend/`):
```bash
make generate          # Regenerate code from specs
make test              # Unit tests
make test-integration  # Integration tests
make run               # Dev server
```

Frontend (`frontend/`):
```bash
make generate    # Regenerate TypeScript types
npm test         # Unit tests
npm run dev      # Dev server
npm run build    # Production build
```

## Backend (Go 1.25+, Gin, oapi-codegen, sqlc)

### Generated files: never edit by hand
- `backend/api/oapi-codegen.gen.go`
- `backend/db/db.gen.go`
- `backend/db/models.gen.go`
- `backend/db/query.sql.go`

After changing `swagger/api.yaml` or `backend/db/query.sql`, run `make generate` in `backend/`.

### Modern Go idioms
- `any` instead of `interface{}`
- `slices` and `maps` packages for collections
- `t.Context()` in tests instead of `context.Background()`

### Error handling
```go
internalError(c, err)             // 500
notFound(c)                       // 404
validationError(c, errorDetails)  // 400
conflict(c, message)              // 409
malformedJson(c)                  // 400
```
Always log errors server-side before responding: `log.Printf("Error context: %v", err)`.

### Transactions
```go
tx, err := s.Database.BeginTx(c.Request.Context(), pgx.TxOptions{})
if err != nil {
    log.Printf("Error creating transaction: %v", err)
    internalError(c, err)
    return
}

defer func() {
    if tx != nil {
        _ = tx.Rollback(c.Request.Context())
    }
}()

// ... perform operations using &tx ...

if err := tx.Commit(c.Request.Context()); err != nil {
    log.Printf("Error committing transaction: %v", err)
    internalError(c, err)
    return
}
tx = nil // Prevent defer rollback after successful commit
```

### Soft deletion
- No `DELETE` statements. Deletable entities have a `deleted` boolean column (default `FALSE`).
- Use `vw_`-prefixed views that filter out deleted rows (e.g. `vw_library_games`).

### Validation
- Validate all inputs in handlers using `utils.go` helpers, e.g. `ValidateStringLength("name", name, 1, 100, errorDetails)`.
- Don't rely solely on database constraints.

### Database queries
- Define all queries in `backend/db/query.sql` using sqlc naming: `-- name: GetGame :one`.
- snake_case SQL identifiers; add indexes for searchable/filterable columns.

### CSV / bulk import
- Accept CSV as a base64-encoded request body (OpenAPI: `text/plain` with `format: byte`).
- Use one transaction for all-or-nothing behavior; return `BulkAddResponse{Imported: recordCount}`.
```go
decodedReader := base64.NewDecoder(base64.StdEncoding, c.Request.Body)
csvReader := csv.NewReader(decodedReader)

for {
    record, err := csvReader.Read()
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Printf("Error reading CSV: %v", err)
        internalError(c, err)
        return
    }
    if len(record) == 0 {
        continue
    }
    // Process record...
}
```

### Layout
```
backend/
├── api/        # Handlers: api.go (Server), *_api.go, *_api_test.go, utils.go; oapi-codegen.gen.go is GENERATED
├── db/         # schema.sql, query.sql, database.go; *.gen.go are GENERATED
└── tests/      # Integration tests
```

## Frontend (Svelte, TypeScript strict, Tailwind)

- **Types**: always use generated types from `frontend/src/generated/library-api.d.ts` (e.g. `components["schemas"]["Game"]`). Never edit that file by hand; regenerate with `make generate` in `frontend/`. Don't create custom types that duplicate generated ones.
- **Components**: small and focused; props for data flow, events/callbacks for communication; PascalCase filenames (`SearchBar.svelte`). Page-level components are `*View.svelte`.
- **API communication**: centralized utilities (see `frontend/src/lib/api-client.ts`), generated types for payloads, consistent error-state handling.
- **Styling**: light and dark themes via `prefers-color-scheme`; muted colors with cool neutrals (primary: cool blues, success: muted greens, error: muted reds/pinks, warning/edit: muted yellows); desktop-first, mobile-usable; WCAG 2.1 AA.
- **Search**: the frontend sends RAW search strings. The backend does all sanitization (accent folding, case-insensitivity). Never pre-process queries on the frontend.

### Layout
```
frontend/
├── src/
│   ├── lib/         # *View.svelte (pages), *.svelte (reusable), *.test.ts
│   ├── generated/   # library-api.d.ts is GENERATED
│   └── main.ts
└── public/config.js # Runtime configuration
```

## Testing

**Naming**: `Should <exhibit behavior> when <thing happens>`.

### Backend
- Unit tests live beside the code (`backend/api/games_api_test.go`), use `testify/assert` and `testify/mock`, mock the `DB` interface, and use `t.Context()`.
```go
func TestFeatureName(t *testing.T) {
    t.Run("Should exhibit behavior when condition occurs", func(t *testing.T) {
    })
}
```
- Integration tests live in `backend/tests/`, use testcontainers-go with real PostgreSQL, and go through the generated API client ONLY (never internal implementation details). Key workflows: checkout lifecycle, duplicate prevention, CRUD, search.

### Frontend (Vitest + Svelte Testing Library)
Mock `fetch` for API calls and use generated types for mock responses.

**Prefer `data-testid` for element selection** over text, roles, or DOM structure. Before asserting on an element: check whether it has a `data-testid`; if not, add one to the component first, then use it in the test. Use descriptive kebab-case names: `check-out-table`, `ptw-table`, `ptw-record-button-{id}`, `admin-view-tabs`, `ptw-empty-state`.
```typescript
// ❌ Fragile: breaks if heading text changes
expect(screen.getByText('Checkout Games')).toBeInTheDocument();

// ✅ Robust: stable, intentional test hook
expect(screen.getByTestId('check-out-table')).toBeInTheDocument();
```

**Opportunistic cleanup**: when editing an existing test, also bring it in line with current standards (`data-testid`, component isolation, naming), even if that isn't the reason for the change.

**Component isolation**: `Foo.test.ts` tests only `Foo.svelte`. If `Foo` renders a child `Bar.svelte`, mock `Bar`; `Bar.test.ts` covers Bar. UI library components (e.g. Flowbite `Modal`, `Dropdown`) may not render content into the accessible DOM in jsdom, so reaching into them is fragile.

- Test in a parent: the child is opened/shown on the right action, the right props/callbacks are wired (verify via the mock), and the parent calls the right API methods after a callback fires.
- Don't test in a parent: a child modal/dropdown's internal buttons or text, or any DOM the child owns.
```typescript
// ✅ Mock the child, test the wiring
vi.mock('./DeleteConfirmationPrompt.svelte', () => ({
  default: vi.fn(), // or a minimal stub
}));

// ❌ Reaches into a child modal's DOM
const confirmButton = screen.getByRole('button', { name: "Yes, I'm sure" });
```

## Common Workflows

**New API endpoint**
1. Edit `swagger/api.yaml`
2. `make generate` in `backend/`
3. Implement the handler in the appropriate `*_api.go`, with validation
4. Unit tests in `*_api_test.go`; integration test in `backend/tests/`
5. `make generate` in `frontend/` to update types
6. Implement the frontend UI and its tests

**New database query**
1. Add it to `backend/db/query.sql`
2. `make generate` in `backend/`
3. Call the generated method from `s.queries` and handle errors in the handler

## Common Pitfalls

1. Passing `nil` instead of `&tx` to functions expecting a transaction pointer
2. Not checking `io.EOF` before other errors when reading CSV
3. Forgetting `tx = nil` after a successful commit (the defer will roll back)
4. Hard deletes instead of soft deletes
5. Manually editing generated files
6. Not logging errors before returning them to the client
7. Custom TypeScript types instead of generated ones
8. Pre-processing search queries on the frontend
9. Forgetting `make generate` after API spec changes
10. Frontend tests that reach into child component DOM

## When in Doubt

1. Check the relevant doc in `docs/`.
2. Follow existing patterns in similar files.
3. Run tests to validate changes.

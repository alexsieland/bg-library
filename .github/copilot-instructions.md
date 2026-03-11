# AI Assistant Instructions for Board Game Library Project

This document provides comprehensive guidelines for AI assistants working on the Board Game Library Management System. Always reference this file along with the detailed documentation in the `docs/` directory.

---

## Project Context

**Project**: Board Game Library Management System  
**Architecture**: Full-stack application with Go backend, Svelte frontend, PostgreSQL database  
**Approach**: Contract-first development using OpenAPI specification as source of truth  
**Repository**: `/home/nightgecko/Development/bg-library`

### Key Documentation References
- **[Project Overview](docs/project-overview.md)**: High-level architecture and technology stack
- **[Coding Guidelines](docs/coding-guidelines.md)**: Backend and frontend coding standards
- **[Testing Guidelines](docs/testing-guidelines.md)**: Testing strategy and best practices
- **[Functional Requirements](docs/functional-requirements.md)**: Feature requirements and user workflows
- **[Frontend Style Guide](docs/frontend-styleguide.md)**: UI/UX standards and design principles

---

## Backend Development (Go)

### Technology Stack
- **Language**: Go 1.25+
- **Framework**: Gin web framework
- **API Generation**: oapi-codegen (generates from `swagger/api.yaml`)
- **Database Access**: sqlc (generates type-safe code from SQL)
- **Database**: PostgreSQL

### Critical Guidelines

#### 1. Contract-First Development
- **ALWAYS** edit `swagger/api.yaml` first when changing the API
- Regenerate code with `make generate` in the backend directory
- Never manually edit generated files:
  - `backend/api/oapi-codegen.gen.go`
  - `backend/db/db.gen.go`
  - `backend/db/models.gen.go`
  - `backend/db/query.sql.go`

#### 2. Modern Go Idioms (1.25+)
- Use `any` instead of `interface{}`
- Use `slices` and `maps` packages for collections
- Use `t.Context()` in tests instead of `context.Background()`

#### 3. Error Handling Pattern
```go
// Internal errors (500)
internalError(c, err)

// Client errors
notFound(c)                           // 404
validationError(c, errorDetails)      // 400
conflict(c, message)                  // 409
malformedJson(c)                      // 400
```
- **ALWAYS** log errors server-side before returning to client
- Use `log.Printf("Error context: %v", err)`

#### 4. Transaction Handling
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

#### 5. Soft Deletion Policy
- **NO HARD DELETES** - Never use `DELETE` statements
- All deletable entities have a `deleted` boolean column (default `FALSE`)
- Use database views prefixed with `vw_` to filter out deleted records
- Example: `vw_library_games` automatically excludes `deleted = TRUE`

#### 6. Validation
- Validate all inputs in handlers using `utils.go` helpers
- Example: `ValidateStringLength("name", name, 1, 100, errorDetails)`
- Don't rely solely on database constraints

#### 7. Database Queries
- Define all queries in `backend/db/query.sql`
- Use sqlc naming convention: `-- name: GetGame :one`
- Use snake_case for SQL identifiers
- Create indexes for searchable/filterable columns

#### 8. CSV/Bulk Import Pattern
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

---

## Frontend Development (Svelte + TypeScript)

### Technology Stack
- **Framework**: Svelte with Vite
- **Language**: TypeScript (strict mode)
- **Styling**: Tailwind CSS (Material Design principles)
- **Type Generation**: openapi-typescript (from `swagger/api.yaml`)

### Critical Guidelines

#### 1. Type Safety
- **ALWAYS** use generated types from `frontend/src/generated/library-api.d.ts`
- **NEVER** manually edit or regenerate this file
- Reference types: `components["schemas"]["Game"]`
- Regenerate with `make generate` in frontend directory

#### 2. Component Structure
- Keep components small and focused
- Use props for data flow
- Use events/callbacks for communication
- File naming: PascalCase for components (e.g., `SearchBar.svelte`)

#### 3. API Communication
- Create centralized API utility functions
- Use the generated TypeScript types for request/response payloads
- Handle error states consistently
- Example location: `frontend/src/lib/api-client.ts`

#### 4. Styling Standards
- **Theme Support**: Light and Dark modes (use `prefers-color-scheme`)
- **Color Palette**: Muted colors, cool neutrals
  - Primary/Action: Cool blues
  - Success/Available: Muted greens
  - Error/Conflict: Muted reds/pinks
  - Warning/Edit: Muted yellows
- **Accessibility**: WCAG 2.1 Level AA compliance
- **Responsiveness**: Desktop-first, mobile-usable

#### 5. Search Sanitization
- Frontend sends RAW search strings
- Backend handles ALL sanitization (accent folding, case-insensitivity)
- Never pre-process search queries on frontend

---

## Testing Strategy

### Backend Testing

#### Unit Tests
```go
func TestFeatureName(t *testing.T) {
    t.Run("Should exhibit behavior when condition occurs", func(t *testing.T) {
        // Use t.Context() for context
        // Use testify/assert and testify/mock
    })
}
```
- Location: Same package as code (e.g., `backend/api/games_api_test.go`)
- Mock the `DB` interface for isolation

#### Integration Tests
- Use testcontainers-go for real PostgreSQL
- Test via generated API client ONLY (contract-first)
- Never reference internal implementation details
- Location: `backend/tests/`
- Key workflows: checkout lifecycle, duplicate prevention, CRUD operations, search

### Frontend Testing
```typescript
describe('ComponentName', () => {
  it('Should exhibit behavior when condition occurs', () => {
    // Use Vitest + Svelte Testing Library
    // Mock fetch for API calls
    // Use generated types for mock responses
  });
});
```

#### Test Element Selection Best Practices
When writing frontend tests, **always prefer `data-testid` attributes** for element selection over text content, roles, or other fragile DOM selectors.

**Important**: Before writing a test assertion that references text, roles, or other DOM structure:
1. **Check if a `data-testid` already exists** on the element you want to test
2. **If no `data-testid` exists**, add one to the component first
3. **Then use the `data-testid`** in your test

**Naming Convention**: Use descriptive, kebab-case names. Examples:
- `check-out-table`, `check-in-table`, `ptw-table` (for table components)
- `ptw-record-button-{id}` (for buttons with dynamic identifiers)
- `admin-view-tabs` (for container elements)
- `ptw-empty-state` (for conditional rendered states)

**Why this matters**: Text content, role attributes, and DOM structure are fragile and break when UI changes. `data-testid` attributes are intentionally added for testing and remain stable across refactors. This keeps tests maintainable and reliable.

**Example:**
```typescript
// ❌ Fragile — breaks if heading text changes
it('Should display checkout heading', () => {
  render(App);
  expect(screen.getByText('Checkout Games')).toBeInTheDocument();
});

// ✅ Robust — stable, intentional test hook
it('Should display checkout heading', () => {
  render(App);
  expect(screen.getByTestId('check-out-table')).toBeInTheDocument();
});
```

#### Test Element Selection Best Practices
When writing frontend tests, **always prefer `data-testid` attributes** for element selection over text content, roles, or other fragile DOM selectors.

**Important**: Before writing a test assertion that references text, roles, or other DOM structure:
1. **Check if a `data-testid` already exists** on the element you want to test
2. **If no `data-testid` exists**, add one to the component first
3. **Then use the `data-testid`** in your test

**Naming Convention**: Use descriptive, kebab-case names. Examples:
- `check-out-table`, `check-in-table`, `ptw-table` (for table components)
- `ptw-record-button-{id}` (for buttons with dynamic identifiers)
- `admin-view-tabs` (for container elements)
- `ptw-empty-state` (for conditional rendered states)

**Why this matters**: Text content, role attributes, and DOM structure are fragile and break when UI changes. `data-testid` attributes are intentionally added for testing and remain stable across refactors. This keeps tests maintainable and reliable.

**Example:**
```typescript
// ❌ Fragile — breaks if heading text changes
it('Should display checkout heading', () => {
  render(App);
  expect(screen.getByText('Checkout Games')).toBeInTheDocument();
});

// ✅ Robust — stable, intentional test hook
it('Should display checkout heading', () => {
  render(App);
  expect(screen.getByTestId('check-out-table')).toBeInTheDocument();
});
```

#### Opportunistic Test Cleanup
- When editing an existing test, also bring that touched test in line with current testing standards (for example: `data-testid` usage, component isolation, and naming convention), even if those fixes are not the primary reason for the change.


#### Component Isolation Rule
Each test file (`Foo.test.ts`) must **only test the behavior of its own component** (`Foo.svelte`). If `Foo.svelte` renders a child component `Bar.svelte`, mock `Bar.svelte` — do not write assertions that depend on `Bar`'s internal DOM structure or behavior. `Bar.svelte` has its own test file for that.

This matters in practice because UI library components (e.g., Flowbite `Modal`, `Dropdown`) may not render their content into the accessible DOM tree in jsdom. If a parent component test reaches into a child's DOM to find a button or text, it will be fragile and environment-dependent.

**What to test in a parent component:**
- That the child component is opened/shown when the correct action occurs (e.g., a state flag is set)
- That the correct props/callbacks are wired up (verify via the mock)
- That the parent calls the right API methods after a callback fires

**What NOT to test in a parent component:**
- The internal buttons or text of a child modal/dropdown
- Any DOM structure owned by a child component

```typescript
// ✅ Correct — mock the child, test the wiring
vi.mock('./DeleteConfirmationPrompt.svelte', () => ({
  default: vi.fn(), // or a minimal stub
}));

// Then verify the parent passed the right props / called the right API

// ❌ Wrong — reaches into a child modal's DOM
const confirmButton = screen.getByRole('button', { name: "Yes, I'm sure" });
```

### Test Naming Convention
**Pattern**: `Should <exhibit behavior> when <thing happens>`

---

## Common Workflows

### Adding a New API Endpoint

1. Edit `swagger/api.yaml` to define the endpoint
2. Run `make generate` in `backend/` directory
3. Implement handler method in appropriate `*_api.go` file
4. Add validation logic
5. Write unit tests in `*_api_test.go`
6. Write integration test in `backend/tests/`
7. Run `make generate` in `frontend/` directory to update types
8. Implement frontend UI consuming the endpoint
9. Write frontend tests

### Adding a Database Query

1. Edit `backend/db/query.sql` with new query
2. Run `make generate` in `backend/` directory
3. Use generated query method from `s.queries`
4. Handle errors appropriately in handler

### Bulk Import Feature

- Accept CSV as base64-encoded request body
- OpenAPI spec: `text/plain` content type with `format: byte`
- Use transaction for all-or-nothing behavior
- Return `BulkAddResponse{Imported: recordCount}`

---

## File Structure Reference

### Backend
```
backend/
├── api/                    # API handlers (handwritten)
│   ├── api.go             # Server struct, dependencies
│   ├── *_api.go           # Handler implementations
│   ├── *_api_test.go      # Unit tests
│   ├── utils.go           # Validation helpers, converters
│   └── oapi-codegen.gen.go # GENERATED - DO NOT EDIT
├── db/                     # Database layer
│   ├── schema.sql         # Table definitions
│   ├── query.sql          # Query definitions
│   ├── database.go        # Connection management
│   └── *.gen.go           # GENERATED - DO NOT EDIT
└── tests/                  # Integration tests
```

### Frontend
```
frontend/
├── src/
│   ├── lib/               # Svelte components
│   │   ├── *View.svelte   # Page-level components
│   │   ├── *.svelte       # Reusable components
│   │   └── *.test.ts      # Component tests
│   ├── generated/
│   │   └── library-api.d.ts # GENERATED - DO NOT EDIT
│   └── main.ts            # App entry point
└── public/
    └── config.js          # Runtime configuration
```

---

## Common Pitfalls to Avoid

1. ❌ Passing `nil` instead of `&tx` to functions expecting a transaction pointer
2. ❌ Not checking `io.EOF` before other errors when reading CSV
3. ❌ Forgetting to set `tx = nil` after successful commit (causes defer to rollback)
4. ❌ Using hard deletes instead of soft deletes
5. ❌ Editing generated files manually
6. ❌ Not logging errors before returning them to client
7. ❌ Creating custom TypeScript types instead of using generated ones
8. ❌ Pre-processing search queries on frontend
9. ❌ Forgetting to run `make generate` after API spec changes
10. ❌ Writing frontend tests that reach into child component DOM — `Foo.test.ts` should only test `Foo.svelte`; mock child components like modals and dropdowns, and test their behavior in their own dedicated test files

---

## Development Commands

### Backend
```bash
make generate    # Regenerate code from specs
make test        # Run unit tests
make test-integration  # Run integration tests
make run         # Start development server
```

### Frontend
```bash
make generate    # Regenerate TypeScript types
npm test         # Run unit tests
npm run dev      # Start development server
npm run build    # Build for production
```

---

## Key Principles

1. **Contract-First**: OpenAPI spec is source of truth
2. **Type Safety**: Use generated types everywhere
3. **Data Preservation**: Soft deletes only, maintain audit trail
4. **Error Visibility**: Log server-side, return appropriate status codes
5. **Transaction Integrity**: Use transactions for multi-step operations
6. **Test Coverage**: Follow behavior-driven naming, test workflows not implementation
7. **Accessibility**: WCAG 2.1 AA compliance for frontend
8. **Separation of Concerns**: Backend handles all business logic and sanitization

---

## When in Doubt

1. Check the relevant doc in `docs/` directory
2. Follow existing patterns in similar files
3. Run tests to validate changes
4. Check for errors with `get_errors` tool after editing files
5. Commit to type safety and contract compliance
6. **You always have read access to all files in the project and its dependencies.** If you believe you cannot read a file, do not attempt to work around it with grep or assumptions — stop and explicitly request read access before continuing.

---

**Last Updated**: March 9, 2026


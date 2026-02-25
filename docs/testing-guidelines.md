# Testing Guidelines

This document outlines the testing strategy, tools, and best practices for the board game library project.

## 1. General Philosophy

The project follows a behavior-driven naming convention for unit tests to ensure they are descriptive and easy to understand.

### Naming Convention
All unit tests should follow the pattern:
`Should <exhibit behavior> when <thing happens>`

**Example (Go):**
```go
func TestAddGame(t *testing.T) {
    t.Run("Should return 201 Created when a valid game is provided", func(t *testing.T) {
        // ...
    })
}
```

**Example (TypeScript/Vitest):**
```typescript
describe('GameCard', () => {
  it('Should display the "Checked Out" badge when the game is not available', () => {
    // ...
  });
});
```

---

## 2. Backend Testing (Go)

### Unit Tests
- **Tools**: Standard `testing` package, `testify/assert` for assertions, and `testify/mock` for mocking dependencies.
- **Scope**: Focus on individual handlers, utilities, and business logic in isolation.
- **Mocking**: Use the `DB` interface (defined in `api/api.go`) to mock database interactions.
- **Location**: Place unit tests in the same package as the code being tested (e.g., `backend/api/games_api_test.go`).
- **Idioms**: Use `t.Run` for subtests and `t.Context()` when a context is required.

### Integration Tests
- **Tools**: [testcontainers-go](https://golang.testcontainers.org/) to spin up a real PostgreSQL instance.
- **Strategy**: Test real user workflows by interacting exclusively with the API via the generated Go client.
- **Contract-First**: Tests must never reference internal implementation details. They should only rely on the OpenAPI specification (via the generated client).
- **Location**: `backend/tests/`.
- **Workflows Covered**:
    - Full checkout/check-in lifecycle.
    - Duplicate checkout prevention (conflict handling).
    - Patron/Game lifecycle (CRUD + soft-deletion).
    - Search and discovery logic (including accent folding).

---

## 3. Frontend Testing (Svelte)

*Note: Frontend tests are currently being established.*

### Unit & Component Tests
- **Tools**: [Vitest](https://vitest.dev/) and [Svelte Testing Library](https://testing-library.com/docs/svelte-testing-library/intro).
- **Focus**: 
    - Logic within Svelte components.
    - Individual utility functions.
    - API fetch wrappers (mocking global `fetch`).
- **Best Practice**: Use the generated TypeScript types from `api.d.ts` to mock API responses accurately.

### End-to-End (E2E) Tests
- **Tools**: [Playwright](https://playwright.dev/).
- **Focus**: Testing the entire application from the user's perspective in a real browser.
- **Scenarios**:
    - Navigating through the library catalog.
    - Searching for a game and performing a checkout.
    - Managing patrons.
- **Isolation**: E2E tests should ideally run against a clean backend/database environment (potentially using the same Docker-based setup as backend integration tests).

---

## 4. Continuous Integration (CI)

All tests are automatically executed on every push and pull request via GitHub Actions.

- **Workflow**: `.github/workflows/go-tests.yml`
- **Steps**:
    1. Run Go Unit Tests (`make test`).
    2. Run Go Integration Tests (`make test-integration`).
    3. (Future) Run Frontend Unit Tests.
    4. (Future) Run Playwright E2E Tests.

Tests must pass successfully before merging into the `main` branch.

# Junie Project Instructions

This project contains comprehensive documentation that defines the architecture, coding standards, requirements, and testing strategies. Always refer to these documents when assisting with development.

## Project Documentation References

- **[Project Overview](../docs/project-overview.md)**: High-level overview of the board game library system, its goals, and its technology stack.
- **[Functional Requirements](../docs/functional-requirements.md)**: Detailed description of user roles, core workflows (Checkout/Check-in), and future feature goals.
- **[Coding Guidelines](../docs/coding-guidelines.md)**: Standards for backend (Go/Gin/sqlc/oapi-codegen), frontend (Svelte/TypeScript), and persistence layers (PostgreSQL).
- **[Testing Guidelines](../docs/testing-guidelines.md)**: Standards for unit and integration testing, including naming conventions and tools like `testcontainers-go`, `Vitest`, and `Playwright`.
- **[Frontend Style Guide](../docs/frontend-styleguide.md)**: Standards for the visual identity and UI of the application, including Material Design, color palette, accessibility, and responsiveness.

## Key Principles

1.  **Contract-First**: The `swagger/api.yaml` is the single source of truth for the API.
2.  **Soft Deletion**: Hard deletes are strictly forbidden. All data must be preserved using a `deleted` flag.
3.  **Accent-Folded Search**: Board game titles must be searchable regardless of accents (e.g., "barenpark" matches "Bärenpark").
4.  **Modern Idioms**: Use modern Go (1.25+) idioms and patterns as described in the `Coding Guidelines`.

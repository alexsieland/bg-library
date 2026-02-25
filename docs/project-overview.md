# Project Overview

The **Board Game Library Management System** is a full-stack application designed to manage a library of board games, including tracking patrons and game checkout/check-in transactions. 

The project follows a contract-first development approach, using an OpenAPI specification as the source of truth for communication between the backend and frontend.

## Core Features

- **Game Management**: Create, read, update, and soft-delete board games.
- **Patron Management**: Create, read, update, and soft-delete library patrons.
- **Checkout Workflow**: Track which patron has borrowed which game, ensuring that a game cannot be checked out by multiple patrons simultaneously.
- **Search & Filtering**: Search for games by title (including accent-folded search) and patrons by name. Filter games by their current checkout status.
- **Data Integrity**: Enforce a mandatory soft-deletion policy to preserve historical transaction data.

## Technology Stack

### Backend
- **Language**: Go 1.25+
- **Framework**: [Gin](https://gin-gonic.com/) for routing and middleware.
- **API Generation**: [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) to generate server boilerplate and models from the OpenAPI spec.
- **Database Access**: [sqlc](https://sqlc.dev/) for type-safe database queries.
- **Validation**: Custom validation logic in Go.

### Frontend
- **Framework**: [Svelte](https://svelte.dev/) (with Vite) for a reactive UI.
- **Language**: TypeScript for static type safety.
- **Type Generation**: [openapi-typescript](https://github.com/hey-api/openapi-typescript) to generate types directly from the API specification.

### Persistence Layer
- **Database**: PostgreSQL
- **Architecture**: Uses database views (prefixed with `vw_`) to provide a consistent view of active data while maintaining hidden soft-deleted records.

## System Architecture

The system is built with a clear separation of concerns:

1.  **API Layer (`swagger/`)**: The OpenAPI 3.0 specification (`api.yaml`) defines the contract for all system interactions.
2.  **Backend (`backend/`)**: A Go-based REST API that interacts with a PostgreSQL database. It handles business logic, state transitions (like checkouts), and enforces data persistence rules.
3.  **Frontend (`frontend/`)**: A modern web interface that allows library administrators to manage the collection and handle patron transactions.
4.  **Database (`PostgreSQL`)**: Stores all relational data, with a schema optimized for both performance (using indexes) and data preservation (using soft deletes and views).

For detailed technical standards and implementation details, please refer to the [Coding Guidelines](coding-guidelines.md).

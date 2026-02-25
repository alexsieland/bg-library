# Functional Requirements

This document describes the functional requirements for the Board Game Library Management System, detailing what the application must do to satisfy the needs of its users.

## 1. User Roles

The system serves three primary types of users, each with a distinct focus and set of responsibilities:

### 1.1 Patrons
Patrons are users who borrow games from the library to play, typically at conventions or festivals.
*   **Focus**: Browsing the library and borrowing games.
*   **Usage Pattern**: Games are typically checked out for approximately one hour or more, but never exceeding 24 hours.

### 1.2 Librarians
Librarians are staff members who facilitate the loaning of games to patrons.
*   **Focus**: Efficiency and speed in processing checkouts and check-ins.
*   **Key Capability**: Must be able to quickly add new patrons to the system "on the fly" during the checkout process if they are not already registered.

### 1.3 Library Admins
Admins manage the library's catalog and user base.
*   **Focus**: Maintenance of the game collection and patron database.
*   **Responsibilities**: Managing the library catalog (adding/removing games) and maintaining the patron list.

---

## 2. Core Functional Requirements

Based on the API specification and existing workflows, the following core features are required:

### 2.1 Game Management (Admin Console)
*   **Add Game**: Ability to add new board games to the library catalog.
*   **Update Game**: Ability to edit existing game details (e.g., title).
*   **Remove Game**: Ability to soft-delete games from the library. Deleted games are removed from the public catalog but preserved for historical metrics.

### 2.2 Patron Management (Admin/Librarian)
*   **Add Patron**: Ability to register new patrons. Librarians can do this quickly during checkout.
*   **Update Patron**: Ability to edit patron information (Admin Console).
*   **Remove Patron**: Ability to soft-delete patrons. Deleted patrons can no longer borrow games but their history is preserved (Admin Console).

### 2.3 Checkout Workflow (Librarian)
*   **Checkout**: Process a loan by associating a specific game with a specific patron.
    *   The system must prevent a game from being checked out by more than one person at a time (Conflict Handling).
*   **Check-in**: Mark a checked-out game as returned, making it available for the next patron.
*   **Status Tracking**: Real-time visibility into which games are available and which are currently checked out (and to whom).

### 2.4 Discovery and Search
*   **Game Search**: Users can search for games by title.
    *   **Accent Folding**: Searches must be resilient to accents (e.g., searching for "barenpark" should find "Bärenpark").
*   **Patron Search**: Librarians can search for patrons by name to facilitate quick checkouts.
*   **Filtering**: Ability to filter the game list by status (e.g., show only checked-out games).

---

## 3. Administrative Console

While the core checkout functions are used by Librarians on the floor, the following management tasks are restricted to an Administrative Console:
*   Adding new games to the system.
*   Updating or deleting existing games.
*   Updating or deleting patron records.

---

## 4. Future Goals and Improvements

The following features have been identified as high-priority enhancements for future development:

### 4.1 Game Ownership Tracking
*   **Requirement**: Ability to track who owns a specific copy of a game.
*   **Default**: Games are "Library Owned" by default.
*   **Use Case**: Managing private collections lent to conventions.

### 4.2 Play-to-Win (P2W) Management
*   **Requirement**: Distinguish "Play-to-Win" games from standard library games.
*   **Raffle Integration**: Ability to add entries to a raffle for a game after a patron has played/returned it.
*   **Winner Processing**: Once won, P2W games should be removed (deleted) from the library catalog as they are no longer part of the collection.

### 4.3 Game Metrics and Reporting
*   **Requirement**: A dedicated dashboard for library analytics.
*   **Key Metrics**:
    *   **Popularity**: Number of times a game was checked out (aggregating games with the same name).
    *   **Usage Duration**: Average checkout time for games.
    *   **Engagement**: Number of entrants for Play-to-Win raffles.
*   **Data Integrity**: Metrics must include data from deleted games and patrons to ensure historical accuracy.

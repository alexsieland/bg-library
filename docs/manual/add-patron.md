# Adding a Patron

This guide covers how to register a new patron in the system. Patrons can be added at any time by a librarian or admin, including during an active checkout if the patron is not yet registered.

---

## Overview

Patrons must be registered in the system before a game can be checked out to them. Registration only requires a name. The process is quick and can be completed without leaving the checkout flow.

---

## Step-by-Step

### 1. Open the Patron Registration Form

<!-- TODO: screenshot — Patrons_Page.png -->
![Patrons page](img/Patrons_Page.png)

Navigate to the **Patrons** section of the application (accessible from the Admin Console or from the inline "Add Patron" option during checkout).

---

### 2. Enter the Patron's Details

<!-- TODO: screenshot — Patron_Modal.png -->
![Add patron modal](img/Patron_Modal.png)

Fill in the required fields:

| Field | Required | Notes |
|---|---|---|
| Name | ✅ Yes | The patron's full name or preferred display name |

---

### 3. Save the Patron

Click **Add Patron** (or **Save**) to register the patron.

The new patron will immediately appear in the patron registry and be searchable during future checkouts.

---

## Adding a Patron During Checkout

If a patron attempts to borrow a game but is not yet in the system, the librarian can register them on the fly:

1. In the **Loan Modal**, begin typing the patron's name.
2. If no match is found, select the **Add New Patron** option (or click the quick-add button).
3. Complete the patron details in the inline form or modal.
4. The new patron is saved and automatically selected in the Loan Modal — proceed with the checkout as normal.

---

## Notes

- Patron names do not need to be unique; the system allows multiple patrons with the same name.
- Patrons cannot be permanently deleted — records are soft-deleted, preserving checkout history for metrics and audit purposes.
- Updating or removing patron records is restricted to the **Admin Console**.

---

*See also: [Renting a Game](rent-manual.md) · [Returning a Game](return-manual.md)*


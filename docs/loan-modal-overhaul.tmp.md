# Loan Modal Overhaul — Implementation Plan

> ⚠️ **TEMPORARY FILE** — Delete this file before merging the PR.

---

## Background & Motivation

The current `LoanModal` allows `Enter` on the patron name input to trigger `handleLoan()`, which will silently create a new patron if no exact match is found. With barcode scanner support, a librarian could accidentally focus the patron name field and have a raw barcode string (e.g. `9780307455925`) registered as a new patron name. This is the core problem being solved.

The fix involves three sequential stages:

1. Build a standalone `AddPatronModal` component
2. Remove implicit patron creation from `LoanModal`; route new patron creation through `AddPatronModal`
3. Wire the newly created patron back into the loan workflow seamlessly

---

## Stage 1 — `AddPatronModal` Component

### Goal
A reusable, standalone modal for creating a new patron. It must work both inside the loan workflow and independently (e.g. Admin Console, future PDF417 barcode onboarding).

### Component Interface
```typescript
export let open: boolean = false;
export let initialName: string = '';                        // Pre-populate name field
export let onPatronCreated: (patron: Patron) => void = () => {}; // Callback with created Patron
export let onCancel: () => void = () => {};
```

### Behaviour
- Opens with `patronName` pre-populated from `initialName` (may be empty)
- If `isBarcodeEnabled()` is true, shows a barcode input field (matching the style of `BarcodeInput.svelte`) that scans a patron barcode and populates `patronName` via `apiClient.getPatronByBarcode` (for future PDF417 use) — or simply stores the raw barcode value against the patron record if the API supports it
- **Enter key is suppressed on all inputs** unless the Submit button itself is focused — inputs use `onkeydown` handlers that explicitly ignore `Enter`
- Submit is only triggered by:
  - Clicking the "Add Patron" button
  - Pressing `Enter` or `Space` while the "Add Patron" button is focused
- On success: calls `onPatronCreated(newPatron)` and closes
- On cancel: calls `onCancel()` and closes, making no API calls
- Validates that `patronName.trim()` is non-empty before allowing submit
- Follows the same Tailwind/flowbite styling as `LoanModal`

### Key Safety Decision: Enter Key Suppression
HID barcode scanners emit all barcode characters as rapid keystrokes and then send an `Enter` keystroke as their terminator signal. If the patron name input were to submit on `Enter`, a scanner burst that lands in that field would both populate the field with a raw barcode string **and** immediately trigger the add patron workflow before the librarian has any opportunity to intervene — the field would never even be "fully filled" from the librarian's perspective.

The `handleKeydown` on the patron name input should therefore explicitly **not** act on `Enter`. Instead it should do nothing (or close a typeahead dropdown if one is present in future). Submit is only reachable via explicit button interaction. This prevents the barcode-as-patron-name problem at the source and keeps the modal safe in any scanning environment.

### Files to Create
- `frontend/src/lib/AddPatronModal.svelte`

### Tests (added last, after visual confirmation)
- Should render with pre-populated name when `initialName` is provided
- Should not submit when Enter is pressed in the name input
- Should call `onPatronCreated` with the new patron on successful submit
- Should call `onCancel` when the cancel button is clicked
- Should show the barcode input only when `isBarcodeEnabled()` is true
- Should disable the submit button when name is empty

---

## Stage 2 — Remove Implicit Patron Creation from `LoanModal`

### Goal
The patron name input in `LoanModal` becomes a **search-only** field. It can no longer create patrons. A "New Patron" button is added as the explicit path for patron creation.

### Workflow Change Options

#### Option A — "New Patron" button opens `AddPatronModal` inline (recommended)
A small "New Patron" button appears next to the patron name input (or below the search results). Clicking it opens `AddPatronModal` with `initialName` pre-populated from whatever is currently in the patron name field. On success, the new patron is selected automatically (Stage 3).

**Pros:**
- Clear, explicit intent — no accidental creation
- Smooth flow: type a name, see no results, click "New Patron" — name carries over
- `AddPatronModal` handles all creation logic independently
- Librarian stays in context (loan modal stays open behind the add patron modal)

**Cons:**
- Two modals stacking (loan modal behind, add patron modal in front) — needs to be visually clear
- One extra click vs. the current implicit creation

#### Option B — Inline confirmation banner
When no results are found after a search, show a banner: _"No patron named 'X' found. [Add as new patron]"_. Clicking the link/button triggers creation directly without a secondary modal.

**Pros:**
- Fewer clicks than Option A
- Stays in a single modal

**Cons:**
- Mixes creation and search UI in one place — can feel cluttered
- The inline add still risks accidental triggering if the librarian is not paying attention
- Doesn't reuse `AddPatronModal`, so the barcode onboarding path needs to be separately maintained

#### Option C — Separate "Add Patron" tab or page navigation
Route the librarian to the Admin Console patron creation flow.

**Pros:**
- Clean separation of concerns

**Cons:**
- Breaks the checkout flow entirely — librarian loses context of which game they were loaning
- Not suitable for fast-paced convention use

### Recommendation
**Option A** — secondary modal stacking. It is explicit, reuses `AddPatronModal`, and keeps the loan flow intact. The stacking is handled naturally by flowbite's `Modal` z-index layering.

**Decision (March 6, 2026): Option A has been selected for implementation.**

### Behaviour Changes to `LoanModal`
- `handleLoan()` no longer calls `apiClient.addPatron()` under any circumstance
- If `patronName` does not match an existing patron from the search results, the Loan button is **disabled**
- A "New Patron" button appears when the search field has content but no match is selected
- The "New Patron" button passes the current `patronName` value to `AddPatronModal` as `initialName`
- A patron is considered "selected" when:
  - The librarian clicks a name from the dropdown search results, OR
  - A patron is returned via `onPatronCreated` from `AddPatronModal` (Stage 3)
- Selected patron state is stored as `selectedPatron: Patron | null`
- The patron name input's `handleKeydown` no longer calls `handleLoan()` on `Enter` when no patron is selected
  - If a patron **is** selected, `Enter` on the patron name input should trigger `handleLoan()` directly — preserving the original fast-path behaviour for librarians who type a name, pick a result, and confirm without reaching for the mouse
  - If no patron is selected, `Enter` is suppressed — it does not create a patron, does not move focus, and does nothing

### Files to Modify
- `frontend/src/lib/LoanModal.svelte`

### Tests (added last, after visual confirmation)
- Should disable the Loan button when no patron is selected
- Should not create a patron when Enter is pressed in the name input
- Should show "New Patron" button when search has text but no match is selected
- Should open `AddPatronModal` with pre-populated name when "New Patron" is clicked
- Should enable the Loan button after a patron is selected from the dropdown
- Should clear selected patron when the name input is modified after selection

---

## Stage 3 — Wire Newly Created Patron Back into Loan Workflow

### Goal
When `AddPatronModal` calls `onPatronCreated(patron)`, the loan modal should immediately select that patron — populating the name field and enabling the Loan button — without requiring the librarian to search again.

### Behaviour
- `LoanModal` passes an `onPatronCreated` handler to `AddPatronModal`:
  ```typescript
  function handleNewPatronCreated(patron: Patron) {
    selectedPatron = patron;
    patronName = patron.name;
    patrons = [];           // clear search results
    addPatronModalOpen = false;
  }
  ```
- After `handleNewPatronCreated` runs:
  - The patron name input is populated with the new patron's name
  - `selectedPatron` is set, enabling the Loan button
  - Focus should move to the Loan button so the librarian can confirm with a single keystroke
- The librarian **does not** need to search for the newly added patron
- The librarian **does** need to confirm the loan explicitly (Enter on Loan button or click)

### Files to Modify
- `frontend/src/lib/LoanModal.svelte` (add `addPatronModalOpen`, `handleNewPatronCreated`, `selectedPatron` state)

### Files to Use
- `frontend/src/lib/AddPatronModal.svelte` (created in Stage 1)

### Tests (added last, after visual confirmation)
- Should select the new patron and populate the name field after `AddPatronModal` succeeds
- Should enable the Loan button immediately after patron creation without requiring a search
- Should not refetch patrons after patron creation — uses the returned `Patron` object directly
- Should complete a full loan after patron creation when the Loan button is clicked

---

## Implementation Order Summary

| Stage | Description | New Files | Modified Files |
|-------|-------------|-----------|----------------|
| 1 | Build `AddPatronModal` | `AddPatronModal.svelte` | — |
| 2 | Remove implicit creation from `LoanModal` | — | `LoanModal.svelte` |
| 3 | Wire new patron back into loan flow | — | `LoanModal.svelte` |

Tests are written **after visual confirmation** of each stage, not before.

---

## Cross-Cutting Concerns

### `AddPatronModal` Outside the Loan Workflow
The modal is designed to be drop-in reusable. Future uses:
- Admin Console patron management (standalone "Add Patron" button)
- PDF417 barcode scan to auto-populate a new patron's name during onboarding

The `initialName` prop and `onPatronCreated` callback make it composable in any context without coupling to `LoanModal`.

### Enter Key Policy (applies to all stages)
HID barcode scanners use `Enter` as their terminator — the final keystroke after every scan burst. Any text input that acts on `Enter` is therefore a potential vector for a scanner accidentally triggering a workflow mid-scan, before the librarian has reviewed the field contents. No text input in any modal should trigger patron creation or loan initiation on `Enter`. All confirmations require explicit button activation (click, or `Enter`/`Space` while the button itself is focused).

**A11y note**: Suppressing `Enter` on `<input>` elements does not affect the native behaviour of `<button>` elements. A button that receives focus via keyboard navigation (e.g. `Tab`) will still respond to `Enter` and `Space` as expected per WCAG 2.1 Success Criterion 2.1.1. The `onkeydown` suppression is scoped only to the input handlers — the button's own keypress handling is left entirely to the browser's default behaviour, preserving full keyboard operability for users who navigate without a mouse.

### Barcode Field in `AddPatronModal`
The barcode field in `AddPatronModal` serves a different purpose from the barcode fields in `CheckInTable`/`CheckOutTable`. It is for **patron identity** (scanning a convention badge or ID), not game lookup. It should not use `barcodeScanner` (the global listener) — it should be an explicit input field only, since the modal already has focus management.

---

> ⚠️ **Reminder**: Delete `docs/loan-modal-overhaul.tmp.md` before opening the PR.







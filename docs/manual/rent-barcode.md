# Renting a Game (Barcode-Assisted Process)

This guide covers how to check out a game to a patron using a USB or Bluetooth HID barcode scanner for faster processing. Barcode scanning must be enabled in the deployment configuration (`BARCODE_ENABLED = true`).

---

## Overview

The barcode-assisted rent workflow lets a librarian scan a game's barcode label to instantly locate the game, then optionally scan the patron's ID card to identify them — eliminating manual searching entirely. This is the recommended workflow in high-traffic environments such as conventions.

---

## Prerequisites

- Barcode scanning is enabled (`BARCODE_ENABLED = true` in `config.js`).
- Game copies have barcode labels affixed.
- Patrons have barcode-enabled ID cards (optional but recommended).

---

## Step-by-Step

### 1. Navigate to the Checkout Page

<!-- TODO: screenshot — Rent_Page.png (barcode field visible) -->
![Checkout page with barcode input](img/Rent_Page.png)

Open the application and ensure you are on the **Check Out** tab. When barcode support is enabled, a **BARCODE** input field is visible in the toolbar.

---

### 2. Scan the Game Barcode

Point the scanner at the barcode label on the game box and scan it.

**If no element on the page is focused**, the global listener will automatically detect the scan burst, focus the barcode input field, and trigger the game lookup — no clicking required.

**If the BARCODE field is not focused**, click into it first, then scan.

The system will look up the game by barcode and highlight it in the table (or open the Loan Modal directly if the game is uniquely identified and available).

---

### 3. Open the Loan Modal

<!-- TODO: screenshot — Rent_Modal.png -->
![Loan modal](img/Rent_Modal.png)

If the game lookup succeeds and the game is available, the **Loan Modal** opens automatically. If the game is already checked out, a conflict message is displayed instead.

---

### 4. Identify the Patron

In the **Patron Name** field inside the Loan Modal, either:

- **Type** the patron's name to search manually, or
- **Scan the patron's barcode ID card** — because the patron name field is focused while the modal is open, the scanner input flows directly into the field.

> **Note:** When the Loan Modal is open, the global barcode listener is suppressed. Any scan while the patron name field has focus will type into that field, not trigger a game lookup. This is intentional.

Select the correct patron from the suggestions.

---

### 5. Confirm the Checkout

Review the game title and patron name, then click **Check Out** to confirm the loan.

The game's status will update to **Checked Out** in the table, and the barcode field will be cleared and ready for the next scan.

---

## Barcode Field Behavior Summary

| Situation | What happens |
|---|---|
| No element focused, game barcode scanned | Global listener detects burst → focuses barcode field → triggers game lookup |
| BARCODE field manually focused, game barcode scanned | Scan types into field → `Enter` triggers game lookup |
| Patron name field focused (modal open), patron barcode scanned | Characters type into patron name field → `Enter` submits the name search |
| Search bar focused, barcode scanned | Characters type into search bar; no game lookup triggered |

---

## Notes

- The global listener uses an 80 ms inter-keystroke timing threshold to distinguish scanner bursts from human typing. Adjust `SCAN_THRESHOLD_MS` if your scanner model behaves differently.
- If the scanner misfires or the timing heuristic fails, click into the **BARCODE** field and scan again — the explicit field always works as a reliable fallback.
- A game can only be checked out to one patron at a time. Scanning an already-checked-out game will display a conflict message.

---

*See also: [Renting Manually](rent-manual.md) · [Returning with a Barcode Scanner](return-barcode.md) · [Adding a Patron](add-patron.md)*


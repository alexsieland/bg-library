# Barcode Scanner Input Handling

This document researches and evaluates approaches for integrating a USB/Bluetooth HID barcode scanner into the frontend checkout workflow, with no dependency on external native applications or browser extensions.

---

## Feature Flag

All barcode-related UI and logic is controlled by the `BARCODE_ENABLED` runtime configuration flag, exposed via `isBarcodeEnabled()` from `frontend/src/lib/config.ts`.

**Any component, listener, action, or UI element related to barcode scanning must be gated behind this flag.** If `isBarcodeEnabled()` returns `false`:

- The global keyboard listener must not be registered.
- Barcode scan input fields must not be rendered.
- Barcode-triggered API calls (`getGameByBarcode`, `getPatronByBarcode`) must not be invoked from the barcode path.
- No barcode-related UI affordances (buttons, labels, hints) should be visible to the user.

```svelte
<!-- Example: conditionally mounting barcode functionality -->
{#if isBarcodeEnabled()}
  <!-- scanner action / scan field / conflict resolution UI -->
{/if}
```

The flag is set in `public/config.js` at deployment time, allowing barcode support to be enabled or disabled per environment without a code change or rebuild.

---

## Background: How HID Barcode Scanners Work in a Browser

A HID (Human Interface Device) barcode scanner presents itself to the operating system as a standard keyboard. When a barcode is scanned, the device:

1. Rapidly sends a sequence of `keydown`/`keypress`/`keyup` events for each character in the barcode string.
2. Terminates the sequence with an `Enter` keystroke (`\n` / `\r`).

This happens much faster than any human typist — typically the full string arrives within **50–100 ms**, compared to the hundreds of milliseconds between human keystrokes.

Because the scanner looks identical to a keyboard at the OS level, the browser has no native API to distinguish scanner input from typed input. All differentiation logic must be implemented in JavaScript.

---

## The Core Detection Problem

The challenge has two parts:

**1. Detecting that input came from a scanner, not a human**

Since both arrive as keyboard events, a heuristic must be used — most commonly **inter-keystroke timing**. A scanner fires all characters in a burst; a human does not.

**2. Capturing scanner input without disrupting the UI**

If the scanner fires while a text field has focus, the characters will appear in that field. If no element has focus, characters may be silently dropped or trigger hotkeys. The goal is to intercept the scanner stream regardless of focus state, and use it to trigger an application action rather than text entry.

---

## Proposed Approaches

### Approach 1: Global `keydown` Listener with Timing Heuristic

**How it works**

Attach a `keydown` event listener to the `document` (or `window`). Buffer incoming characters. If the full buffer (ending with `Enter`) arrives within a configurable time window (e.g. 100 ms), treat it as a barcode scan. Reset the buffer on timeout or on `Enter`.

```typescript
// Conceptual sketch
let buffer = '';
let lastKeyTime = 0;
const SCAN_THRESHOLD_MS = 100;
const MIN_BARCODE_LENGTH = 4;

document.addEventListener('keydown', (e) => {
  const now = Date.now();
  if (now - lastKeyTime > SCAN_THRESHOLD_MS) {
    buffer = ''; // timeout — reset, treat as new input
  }
  lastKeyTime = now;

  if (e.key === 'Enter') {
    if (buffer.length >= MIN_BARCODE_LENGTH) {
      onBarcodeScanned(buffer);
    }
    buffer = '';
    return;
  }

  if (e.key.length === 1) {
    buffer += e.key;
  }
});
```

**Advantages**
- Simple to implement and understand.
- Works regardless of which element has focus.
- No DOM manipulation required.
- Easy to add to a Svelte layout component so it is always active.

**Drawbacks**
- **False positives**: A fast typist can occasionally trigger it. The 100 ms threshold is a trade-off; lowering it reduces false positives but risks missing slow scanners.
- **Interferes with text fields**: If a search input or form field is focused, the characters will be inserted into the field *and* buffered by the global listener. Double-handling logic is required to suppress this.
- **Modifier key edge cases**: Some scanners are configured to prefix output with modifier keys (e.g., `Shift` + digit for symbols). The listener must handle these correctly.
- **`e.key` vs `e.code` locale issues**: `e.key` reflects the user's keyboard layout. On non-English layouts, character values may differ from the raw barcode digits. Using `e.code` + a lookup table is more robust but more complex.

---

### Approach 2: Hidden Focused Input Element

**How it works**

Render a visually hidden `<input>` element that holds programmatic focus at all times. Because it has focus, it naturally captures all keyboard input (including scanner bursts). Monitor its `input` or `keydown` events to detect a completed barcode, then clear the field and dispatch the action.

```svelte
<!-- Svelte example -->
<input
  bind:this={hiddenInput}
  class="sr-only"
  aria-hidden="true"
  tabindex="-1"
  on:keydown={handleScannerInput}
/>
```

Maintain focus by re-focusing the hidden input whenever focus moves away (e.g., on `blur` or `focusout` on the document body), *except* when focus moves to a legitimate interactive element like a search box or modal.

**Advantages**
- Scanner input is cleanly isolated to one element — no risk of characters appearing in the wrong place.
- The input field can be `aria-hidden` and visually hidden with `sr-only` (Tailwind), making it invisible to users.
- Simpler double-handling problem: if a real interactive element has focus, the hidden input is deliberately *not* re-focused, so the scanner fires into that element instead (desirable, e.g., in a patron search field).

**Drawbacks**
- **Focus management is fragile**: Aggressively re-focusing a hidden element can break keyboard navigation for accessibility (WCAG 2.1 requirement). Users navigating by keyboard (`Tab`) may find their focus unexpectedly stolen.
- **Modal and overlay conflicts**: When a modal or dropdown is open, re-focus logic must be suspended or it will pull focus away from the modal.
- **`aria-hidden` + focusable is an a11y anti-pattern**: An element that is `aria-hidden="true"` should not be focusable. Screen readers may behave unpredictably. This requires careful implementation (e.g., using `tabindex="-1"` only and never including it in the tab order).
- **Not truly opaque to the user**: On some browsers/OSes, a brief cursor flicker may be visible even in `sr-only` positioned elements.

---

### Approach 3: Web HID API

**How it works**

The [Web HID API](https://developer.mozilla.org/en-US/docs/Web/API/WebHID_API) allows a web page to communicate directly with HID devices via a low-level protocol, bypassing the OS keyboard emulation layer entirely. The browser prompts the user to select and grant access to a specific device. Once granted, raw HID reports are delivered to JavaScript.

```typescript
// Conceptual sketch
const [device] = await navigator.hid.requestDevice({ filters: [] });
await device.open();
device.addEventListener('inputreport', (event) => {
  // Raw HID report — requires parsing the usage page/report descriptor
  const data = new Uint8Array(event.data.buffer);
  // ... parse USB HID keyboard usage codes into characters
});
```

**Advantages**
- Completely separate from keyboard events — zero interference with text fields or focus state.
- Eliminates the timing heuristic entirely; the source of the input is unambiguous.
- Could theoretically support scanners that emit non-keyboard HID reports.

**Drawbacks**
- **Requires a user permission gesture**: The browser must display a device-picker dialog. This cannot be pre-authorized; the user must interact with it, which is disruptive in a librarian workflow.
- **Chromium-only**: Web HID is only available in Chrome and Edge as of 2026. Firefox and Safari do not support it. This is a hard blocker for any non-Chromium deployment.
- **Complex report parsing**: Raw HID keyboard reports use USB HID usage codes, not characters. Translating usage codes to glyphs requires implementing a keymap — a significant amount of low-level code that is fragile across keyboard layouts.
- **Overkill**: Since scanners emulate keyboards by design, bypassing that emulation gains little practical benefit and introduces significant complexity.

---

### Approach 4: Dedicated Scanner Input Field (Explicit Focus Mode)

**How it works**

Rather than trying to intercept scanner input globally, the UI dedicates a visible (but perhaps small or subtle) input field specifically for barcode scanning. The field is clearly labelled (e.g., "Scan barcode…") and the librarian ensures the scanner fires into it. On `Enter`, the value is processed as a barcode.

This is the simplest possible approach — it is just a normal `<input>` with an `onchange`/`onkeydown` handler.

**Advantages**
- Zero complexity. No timing heuristics, no focus management, no hidden elements.
- Completely accessible — no a11y trade-offs.
- Visually explicit: the librarian always knows where scan input will go.
- Scanner and keyboard input are naturally separated by UI context.

**Drawbacks**
- **Requires deliberate focus management by the librarian**: The librarian must ensure the barcode field has focus before scanning. If they have been typing a patron name, they must click away first.
- **Not truly opaque**: The field is visible and a distraction in the UI.
- **Poor UX for high-frequency scanning**: In a fast-paced convention environment, requiring the librarian to click a field before every scan adds friction.

---

## Recommendation for This Project

Given the project's constraints (Svelte + TypeScript frontend, librarian-focused UX, WCAG 2.1 AA compliance, desktop-first), the recommended approach is a **combination of Approach 1 and Approach 4**:

### Primary: Explicit Barcode Input Field (Approach 4)

The core implementation uses **Approach 4**: a dedicated, visible barcode input field (`BarcodeInput.svelte`) that is always present in the check-in and check-out toolbars when `isBarcodeEnabled()` is true. This field is the "official" target for barcode scanner input and is the fallback that works 100% reliably.

- The field is small, subtle, and clearly labelled with a muted all-caps "BARCODE" prefix
- Librarians can always manually click into it and type/scan a barcode
- It follows WCAG 2.1 AA accessibility standards with no focus-stealing side effects

### Enhancement: Global Listener (Approach 1) for Convenience

The global keydown listener (implemented as a Svelte action `barcodeScanner`) acts as a **convenience enhancement** to Approach 4:

- When the page has no focused interactive element, a barcode scan is detected via timing heuristics (80ms threshold, ≥4 characters)
- Instead of buffering the scan globally, the listener **automatically focuses the visible barcode input field and inputs the data into it**
- This allows librarians to scan without having to manually click into the field
- The explicit field then processes the barcode normally (calling `onGameFound` or `onError` handlers)

**In practice**: The librarian can pick up a game/patron barcode scanner and scan directly; the global listener detects the burst, focuses the barcode input field, and inputs the barcode there. The field's own handlers then perform the API call and check-in/check-out logic.

### Why This Combination Works

1. **Approach 4 (explicit field) is always available**: It's never hidden, never depends on timing heuristics, and always works regardless of page state
2. **Approach 1 (global listener) is a convenience layer**: It speeds up the happy path (scan with no manual clicking) but falls back gracefully to Approach 4 if the timing heuristic misfire or is disabled
3. **No accessibility trade-offs**: The global listener never steals focus from user navigation; it only focuses an already-visible, appropriately-labelled field
4. **No interference with text fields**: When a real input (search bar, patron name field) has focus, the listener explicitly suppresses itself and lets the characters type naturally into that field

### Why Not Approach 2 (hidden input)?

The hidden focused input is a tempting middle ground, but the focus-stealing behaviour creates real accessibility and usability problems with modals and keyboard navigation that are difficult to solve without significant ongoing complexity. The combination of visible Approach 4 field + convenience Approach 1 listener achieves the same goal with a simpler, more robust model.

### Why Not Approach 3 (Web HID)?

The permission flow and Chromium-only limitation make it unsuitable as a primary strategy. It could be considered as an **optional progressive enhancement** in a future iteration if the deployment environment is guaranteed to be Chromium-based.

---

## Implementation Details

### Global Listener Behavior (Approach 1)

The `barcodeScanner` action:
- Listens globally for `keydown` events
- Buffers printable characters (excluding modifier key combinations)
- Resets the buffer if >80ms passes between keystrokes (timing heuristic)
- When `Enter` is detected:
  - If the buffer is ≥4 characters AND no interactive element is focused:
    - Calls `onScan(barcode)` callback with the accumulated buffer
    - The callback should **focus the barcode input field and set its value**, allowing the field's own handlers to complete the transaction
  - Otherwise: does nothing, lets the `Enter` key behave normally

### Barcode Input Field Behavior (Approach 4)

The `BarcodeInput` component:
- Renders a small, muted barcode input field with label "BARCODE"
- Accepts manual text input or pasted data
- On `Enter`: calls `onGameFound(game)` or `onError(message)` based on the API result
- Is always available as a fallback, even if the global listener is disabled or misfires
- Exports `barcodeInputElement` reference for parent components to focus/set value programmatically

### Integration in Views

When a barcode scanner is detected by the global listener:

```typescript
async function onScanComplete(barcode: string) {
  // 1. Focus and populate the visible barcode input field
  if (barcodeInputElement) {
    barcodeInputElement.focus();
    barcodeInputElement.value = barcode;
  }
  
  // 2. The field's own handlers then process the barcode via API call
  // (This could be done by triggering the field's onKeyDown(Enter) or by calling handleScan directly)
  
  // 3. Success/error is handled by the field's onGameFound/onError handlers
  // (converting to toasts, state updates, etc.)
}
```

---

## Interaction with Modals and Focus Management

This section analyses how the Approach 1 + 4 combination behaves when elements with focus are present.

### Scanning a game barcode while LoanModal is open — ✅ Safe

The `LoanModal` contains a focused patron name `<Input>` field. When the modal is open and that field has focus:
- The global listener's suppression logic kicks in: it sees an INPUT element has focus
- The listener discards its buffer and **does not** call the `onScan` callback
- Characters naturally flow into the patron name field as keyboard input
- The barcode is NOT scanned into the check-in/check-out flow

This is the correct behavior — the modal is mid-transaction; scanning at this point should not trigger an unrelated game barcode lookup.

### Scanning without a focused element — ✅ Scan detected and focused into barcode field

When no interactive element has focus (user is viewing the check-in table):
- Global listener detects the scanner burst by timing
- Calls `onScan(barcode)` with the complete barcode string
- The handler focuses the visible `BarcodeInput` field and populates its value
- The field's own `Enter` handler fires, calling `onGameFound` or `onError`

### Scanning into the search bar — ✅ Characters type into search, no barcode scan

When the user has the search bar focused:
- Global listener sees an INPUT element has focus
- It suppresses itself, discards the buffer, does not call `onScan`
- Barcode characters naturally type into the search field
- User can then press Enter to search (or clear and try again with the barcode field)

### Summary

| Context | Global listener active? | Behavior |
|---|---|---|
| No element focused, barcode scanned | ✅ Yes | Calls `onScan(barcode)`, handler focuses and populates barcode field |
| Search bar focused, barcode scanned | 🚫 Suppressed | Characters type into search bar, no barcode action |
| Patron name field focused, barcode scanned | 🚫 Suppressed | Characters append to patron name, no barcode action |
| Barcode field focused manually, barcode scanned | ✅ Yes (but suppressed by focus check) | Actual behavior: focus check sees INPUT, so suppressed; barcode types directly into field |

---

## Implementation Notes

- **Feature flag**: Always check `isBarcodeEnabled()` before rendering the `BarcodeInput` component or registering the `barcodeScanner` action. See the [Feature Flag](#feature-flag) section above.
- **Global listener scope**: The `barcodeScanner` action should be attached to `<svelte:window>` in the view component (e.g., `CheckInTable.svelte`) so it is only active when that view is mounted.
- **Timing threshold**: The 80ms inter-keystroke threshold is configurable. Different scanner models may have slightly different burst speeds; adjust if needed via the `SCAN_THRESHOLD_MS` constant in `barcodeScannerAction.ts`.
- **Minimum barcode length**: The 4-character minimum prevents accidental triggering on very short key sequences. Adjust `MIN_BARCODE_LENGTH` if your barcodes are shorter.
- **Barcode values**: Scanned barcodes are passed to `apiClient.getGameByBarcode()` or `apiClient.getPatronByBarcode()` — no additional sanitization is needed on the frontend (per the project's separation-of-concerns policy: backend handles all sanitization).
- **Fallback mechanism**: If the global listener ever misfires or is disabled, users can always fall back to manually clicking the barcode input field and typing/pasting the barcode. The field is always available and always works.

---

## Interaction with Modals

This section analyses how the Approach 1 + 4 combination behaves when a modal is open, specifically the existing `LoanModal` (patron name search + loan initiation) and the planned barcode conflict resolution modal.

### Scanning a game barcode while LoanModal is open — ✅ Safe

The `LoanModal` contains a focused patron name `<Input>` field. When the modal is open and that field has focus, the global listener's suppression logic kicks in: `document.activeElement` is the patron name input, so the listener discards its buffer and lets the characters flow into the field. The scanner burst will not trigger a game barcode lookup.

However, this means a librarian **cannot** scan a game barcode while the loan modal is open. This is desirable — the modal is mid-transaction. Scanning a new game at that point would be an error, not a valid workflow step.

### The Enter key conflict in LoanModal — ⚠️ Requires care

`LoanModal` has its own `handleKeydown` listener on the patron name `<Input>` that calls `handleLoan()` on `Enter`. This is also the termination character a barcode scanner sends.

If a librarian accidentally scans *anything* while the patron name field has focus:
- The barcode characters will be appended to the patron name field (since the global listener is suppressed).
- The `Enter` at the end will fire `handleLoan()` with garbled patron name input.

This is a **pre-existing UX hazard** that exists independently of the barcode feature. Mitigation options at implementation time:
- Validate that the patron name field contains only plausible name characters before submitting on `Enter`.
- Debounce the `Enter` handler so a burst-speed `Enter` is distinguishable from a deliberate keypress (checking elapsed time since the last character).

### Scanning a patron barcode for conflict resolution — ✅ Handled by Approach 4

The barcode conflict resolution modal (triggered when multiple copies of a game are checked out) requires the librarian to identify the returning patron. This is the explicit "Scan patron barcode" input field (Approach 4).

Because this field takes focus when the modal opens, the global listener is naturally suppressed. The scanner burst flows directly into the dedicated patron barcode field, which processes the value on `Enter` and calls `getPatronByBarcode()`. No cross-interference with the global listener occurs.

This is the cleanest interaction in the entire barcode workflow: Approach 4's explicit field handles the one place where intentional scanner input inside a modal is expected, and the global listener stays out of the way.

### Summary

| Context | Global listener active? | Scanner input goes to… | Result |
|---|---|---|---|
| No modal open, no field focused | ✅ Yes | Global buffer | Game barcode lookup triggered |
| No modal open, search field focused | 🚫 Suppressed | Search field | Characters appear in search input |
| LoanModal open, patron field focused | 🚫 Suppressed | Patron name field | Characters appended to patron name; `Enter` submits loan |
| Conflict resolution modal open | 🚫 Suppressed | Dedicated patron barcode field | `getPatronByBarcode()` triggered |

---

## Implementation Notes

- **Feature flag**: Always check `isBarcodeEnabled()` before registering any listener or rendering any barcode UI. See the [Feature Flag](#feature-flag) section above.
- The global listener should be implemented as a **Svelte action** (`use:barcodeScanner`) or a **rune-based snippet** so it can be cleanly scoped to individual views and automatically removed when the component is destroyed.
- The timing threshold should be **configurable** (e.g., via `public/config.js`) to accommodate different scanner models, which can vary in burst speed.
- Barcode values should be passed to the existing `apiClient.getGameByBarcode()` or `apiClient.getPatronByBarcode()` functions directly — no additional sanitization is needed on the frontend (per the project's separation-of-concerns policy: backend handles all sanitization).
- In the global listener, call `e.preventDefault()` on buffered keystrokes only if you are confident they came from a scanner (i.e., after the `Enter` arrives and the threshold was met). Do **not** call `preventDefault` speculatively, as this would break normal typing.




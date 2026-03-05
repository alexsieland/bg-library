# Barcode Scanner Input Handling

This document researches and evaluates approaches for integrating a USB/Bluetooth HID barcode scanner into the frontend checkout workflow, with no dependency on external native applications or browser extensions.

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

### Primary: Global listener (Approach 1) with field suppression

Implement a global `keydown` listener as a Svelte snippet or action, active whenever the checkout or check-in view is mounted. Use a timing threshold (suggested: **80 ms** between keystrokes) and a minimum barcode length (suggested: **4 characters**) to distinguish scanner bursts from typing.

Add suppression logic: if the current `document.activeElement` is a recognized interactive input (search bar, modal field), do **not** process the buffer as a barcode — let the characters flow naturally into that field instead.

### Fallback: Explicit scan field (Approach 4) in conflict resolution UI

When the conflict resolution modal is open (multiple copies checked out — see `barcode-conflict-workflow.md`), render an explicit "Scan patron barcode" input field that takes focus automatically when the modal opens. This is the one context where the librarian is expected to scan a second barcode, and making it explicit removes ambiguity.

### Why not Approach 2 (hidden input)?

The hidden focused input is a tempting middle ground, but the focus-stealing behaviour creates real accessibility and usability problems with modals and keyboard navigation that are difficult to solve without significant ongoing complexity. The global listener achieves the same goal with a simpler model.

### Why not Approach 3 (Web HID)?

The permission flow and Chromium-only limitation make it unsuitable as a primary strategy. It could be considered as an **optional progressive enhancement** in a future iteration if the deployment environment is guaranteed to be Chromium-based.

---

## Implementation Notes

- The global listener should be implemented as a **Svelte action** (`use:barcodeScanner`) or a **rune-based snippet** so it can be cleanly scoped to individual views and automatically removed when the component is destroyed.
- The timing threshold should be **configurable** (e.g., via `public/config.js`) to accommodate different scanner models, which can vary in burst speed.
- Barcode values should be passed to the existing `apiClient.getGameByBarcode()` or `apiClient.getPatronByBarcode()` functions directly — no additional sanitization is needed on the frontend (per the project's separation-of-concerns policy: backend handles all sanitization).
- In the global listener, call `e.preventDefault()` on buffered keystrokes only if you are confident they came from a scanner (i.e., after the `Enter` arrives and the threshold was met). Do **not** call `preventDefault` speculatively, as this would break normal typing.


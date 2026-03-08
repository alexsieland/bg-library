import { isBarcodeEnabled } from "./config";

const SCAN_THRESHOLD_MS = 80;
const MIN_BARCODE_LENGTH = 4;

export interface BarcodeScannerOptions {
  onScan: (barcode: string) => void;
}

/**
 * Svelte action for detecting barcode scanner input via global keydown listener.
 * Buffers characters and detects scanner bursts by inter-keystroke timing.
 * Suppresses itself when an interactive element has focus.
 *
 * This is Approach 1 (global listener) and serves as a convenience enhancement
 * to Approach 4 (explicit barcode input fields). When a scan is detected:
 * - The onScan callback is invoked with the barcode string
 * - The callback MUST focus the visible barcode input field and set its value
 * - The field's own handlers then process the barcode (API call, check-in/out, etc.)
 *
 * Usage:
 * ```svelte
 * <svelte:window use:barcodeScanner={{ onScan: handleBarcodeScan }} />
 *
 * async function handleBarcodeScan(barcode: string) {
 *   // Focus and populate the explicit barcode input field (Approach 4 fallback)
 *   if (barcodeInputElement) {
 *     barcodeInputElement.focus();
 *     barcodeInputElement.value = barcode;
 *   }
 *   // The field's onEnter handler will then process the barcode
 * }
 * ```
 */
export function barcodeScanner(node: Window, options: BarcodeScannerOptions) {
  let scanBuffer = "";
  let lastKeyTime = 0;

  function isInteractiveElementFocused(): boolean {
    const el = document.activeElement;
    if (!el) return false;
    const tag = el.tagName;
    return (
      tag === "INPUT" ||
      tag === "TEXTAREA" ||
      (el as HTMLElement).isContentEditable
    );
  }

  function handleGlobalKeydown(e: KeyboardEvent) {
    if (!isBarcodeEnabled()) return;

    const now = Date.now();

    if (e.key === "Enter") {
      const barcode = scanBuffer;
      scanBuffer = "";
      lastKeyTime = 0;
      // Only treat as a scan if timing threshold was met for the whole burst
      // and no interactive element is focused
      if (
        barcode.length >= MIN_BARCODE_LENGTH &&
        !isInteractiveElementFocused()
      ) {
        e.preventDefault();
        options.onScan(barcode);
      }
      return;
    }

    // Suppress if a real input has focus — let characters go to that field
    if (isInteractiveElementFocused()) {
      scanBuffer = "";
      lastKeyTime = 0;
      return;
    }

    // Reset buffer if too much time has passed since last keystroke
    if (now - lastKeyTime > SCAN_THRESHOLD_MS) {
      scanBuffer = "";
    }
    lastKeyTime = now;

    // Only buffer printable single characters
    if (e.key.length === 1 && !e.ctrlKey && !e.metaKey && !e.altKey) {
      scanBuffer += e.key;
    }
  }

  node.addEventListener("keydown", handleGlobalKeydown);

  return {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    destroy() {
      node.removeEventListener("keydown", handleGlobalKeydown);
    },
  };
}

# End-User Documentation Style Guide

This guide sets the rules for all user-facing documentation in the `docs/manual/` directory. Follow these rules any time you create or edit a manual page.

---

## Who These Docs Are For

These documents are written for **librarians** — the people who help patrons check out and return games. They are **not** written for developers, system administrators, or technical staff.

- Do **not** include setup steps, configuration flags, server settings, or code snippets.
- Do **not** explain how the software works internally (e.g., timing thresholds, database behavior, API calls).
- Do **not** mention admin-only features unless they are directly part of the librarian's task.

---

## Reading Level

All text must be written at a **7th grade reading level** or lower.

- Use short sentences.
- Use common, everyday words. Avoid technical terms.
- If a technical word must be used, explain it in plain language right after.
- Examples of word swaps:
  - "authenticate" → "sign in"
  - "query" → "search"
  - "modal" → "pop-up window" (or "window" for repeat uses)
  - "patron registry" → "patron list"

---

## Step Coverage

Every guide must walk the librarian through the **complete process**, from the moment they open the app to the moment the task is done.

- Always include a step for navigating to the right page.
- Never assume the librarian already knows where to go.
- Use numbered steps for sequential actions.
- Use bullet points only for lists of options, not for sequential steps.

---

## Accessibility

This app meets **WCAG 2.1 Level AA** accessibility standards. Documentation must reflect this.

**Any time a button or clickable element is described, also describe how to reach it with a keyboard.** Use this pattern:

> Click the **Check Out** button — or press **Tab** to move to it, then press **Enter** or **Space** to select it.

Apply this note the **first time** a button is mentioned in a document. For repeated references to the same button, the click-only phrasing is fine.

Page tabs and navigation items follow the same rule — include the keyboard shortcut or Tab-navigation note on first mention.

**Do not describe UI elements by color alone.** Color can be used as a helpful extra detail, but always pair it with the element's label or another identifier. This ensures the instructions work for users who cannot distinguish colors.

> ✅ Click the **New Patron** button (shown in green).  
> ❌ Click the green button.

---

## Error Handling

Every document must include the following note in the **Notes** section at the bottom, word-for-word (or close to it):

> If a red error message appears at the bottom of the screen and the reason is not clear, ask for help or write down what you were doing so someone can look into it later.

This text must always appear as a bullet point in the Notes section.

---

## Images

- All screenshots go in `docs/manual/img/`.
- Always include an `alt` text description with the image tag.
- Use a `<!-- TODO: screenshot — FileName.png -->` comment above any image that has not yet been taken, so it is easy to find and fill in later.
- Image filenames use `PascalCase` (e.g., `Rent_Page.png`).

---

## Barcode Scanner Sections

For barcode-assisted workflows:

- Do **not** mention configuration flags (`BARCODE_ENABLED`) or `config.js`.
- Instead, write: *"If barcode scanning is set up for your station..."* or *"If your station has a barcode scanner..."*
- Do **not** describe timing thresholds, global listeners, or internal detection logic.
- Focus only on what the librarian sees and does: scan the label, read the result, confirm or fix.

---

## Technical Terms to Avoid

| Instead of… | Write… |
|---|---|
| `BARCODE_ENABLED = true` | *barcode scanning is set up for your station* |
| global listener / timing heuristic | *(omit entirely)* |
| modal | pop-up window (first use), window (after) |
| soft delete | *(omit; not relevant to librarians)* |
| patron registry | patron list |
| HID / USB HID | barcode scanner |
| audit trail | *(omit; not relevant to librarians)* |
| conflict resolution | *more than one copy is checked out* |
| `Enter` key (raw code) | **Enter** key |

---

## Document Structure

Every manual page must follow this structure:

```
# Title

One-sentence plain-language description of what this guide covers.

---

## What You Need  (optional — only if there are physical requirements like a scanner)

---

## Steps

### 1. [Action]
### 2. [Action]
...

---

## Notes

- Specific notes relevant to librarians only.
- If a red error message appears at the bottom of the screen and the reason is not clear, ask for help or write down what you were doing so someone can look into it later.

---

*See also: links to related guides*
```

---

## Cross-Links

Always end each document with a *See also* line linking to related guides. Use plain descriptive link text, not technical names.

---

*This style guide applies to all files in `docs/manual/`. Last updated: May 2026.*


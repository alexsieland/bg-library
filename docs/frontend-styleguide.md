# Frontend Style Guide

This document outlines the visual identity and user interface standards for the Board Game Library frontend. It is designed to ensure a consistent, accessible, and responsive experience across all supported devices.

## 1. Design Philosophy
The UI follows **Material Design** principles, prioritizing clarity, clean layouts, and meaningful transitions. The application is a Single Page Application (SPA) with a focus on speed and ease of use for Librarians and Admins.

## 2. Color Palette
The color scheme leans toward **neutral and cool colors** to create a calm and professional interface. Warmer colors are used sparingly for specific actions or statuses and should always be **muted** rather than bright.

### 2.1 Theme Support
- **Dynamic Theming**: The application must support both **Light** and **Dark** modes.
- **Default Behavior**: The theme should default to the user's operating system (OS) preference using the `prefers-color-scheme` media query.

### 2.2 Functional Colors (Muted Palette)
- **Primary/Action**: Cool blues (e.g., Search buttons, active navigation).
- **Success/Available**: Muted greens (e.g., "Available" status, "Loan" confirmation).
- **Error/Conflict/Occupied**: Muted reds/pinks (e.g., "John Smith" (checked out) status, "Delete" actions).
- **Warning/Edit**: Muted yellows/ambers (e.g., "Edit" buttons).
- **Neutral**: Grays and off-whites/deep grays for backgrounds, borders, and secondary text.

### 2.3 Color Standardization
To maintain consistency across the application, specific Tailwind color variants are standardized:
- **Red → Rose**: Use `rose-*` variants (e.g., `text-rose-500`, `bg-rose-500`) instead of `red-*` for errors, conflicts, and negative actions.
- **Green → Emerald**: Use `emerald-*` variants (e.g., `text-emerald-500`, `bg-emerald-500`) instead of `green-*` for success states and positive confirmations.

These muted variants align with the design philosophy and provide better visual harmony across light and dark themes.

## 3. Typography
- Use standard sans-serif fonts that are highly legible (e.g., Roboto, Inter, or system defaults).
- Maintain high contrast between text and background to satisfy accessibility requirements.

## 4. Accessibility (A11y)
The application must follow **WCAG 2.1 Level AA** standards.
- **Contrast**: Ensure all text meets the minimum contrast ratio.
- **Interactives**: Buttons and form fields must have clear focus states.
- **Semantics**: Use proper HTML tags (e.g., `<button>`, `<nav>`, `<main>`) to support screen readers.

## 5. Responsiveness
The UI must render correctly on **Mobile, PC, and Tablets**.
- **Priority**: Design is prioritized for **PC (Desktop)** usage, as Librarians are expected to use laptops or workstations.
- **Mobile Experience**: Must be at least a **usable experience**, ensuring touch targets are large enough and layouts stack vertically where appropriate.
- **Scrolling**: Use "Unlimited scroll" (infinite scroll or efficient pagination) for long lists of games and patrons.

## 6. Implementation Strategy
- **Styling Framework**: Use **Tailwind CSS** for consistent and efficient styling.
- **Components**: Leverage a Svelte-compatible Material Design library or build custom components following Tailwind-based Material patterns.

## 7. Mockup References
Refer to the mockups in `docs/mockups/` for the layout of the three primary pages:
- **Checkout Page**: Primary interface for finding games and initiating loans via a modal.
- **Check-in Page**: Focused view of currently checked-out games for quick returns.
- **Admin Page**: Management interface for the game catalog and patron database, including Edit/Delete workflows.
import { render, screen } from '@testing-library/svelte';
import AdminPatronsTab from './AdminPatronsTab.svelte';
import { describe, it, expect, vi } from 'vitest';

// Use a real mock Svelte component file that exports a valid component
// for Svelte 5. This avoids shape mismatches when testing.
vi.mock('./PatronsManagementTable.svelte', async () => {
  const mod = await import('./PatronsManagementTable.mock.svelte');
  return { default: mod.default };
});

describe('AdminPatronsTab', () => {
  it('Should render the PatronsManagementTable component', () => {
    render(AdminPatronsTab);

    expect(screen.getByTestId('patrons-management-table')).toBeInTheDocument();
  });
});

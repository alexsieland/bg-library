import { render, screen } from '@testing-library/svelte';
import AdminPatronsTab from './AdminPatronsTab.svelte';
import { describe, it, expect, vi } from 'vitest';
import AdminGamesTab from './AdminGamesTab.svelte';

// Use a real mock Svelte component file that exports a valid component
// for Svelte 5. This avoids shape mismatches when testing.
vi.mock('./GamesManagementTable.svelte', async () => {
  const mod = await import('./GamesManagementTable.mock.svelte');
  return { default: mod.default };
});

describe('AdminGamesTab', () => {
  it('Should render the GamesManagementTable component', () => {
    render(AdminGamesTab);

    expect(screen.getByTestId('games-management-table')).toBeInTheDocument();
  });
});

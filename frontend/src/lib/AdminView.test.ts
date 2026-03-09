import { render, screen } from '@testing-library/svelte';
import AdminView from './AdminView.svelte';
import { describe, it, expect, vi } from 'vitest';

vi.mock('./config', () => ({
  getBackendUrl: () => 'http://localhost:8080',
  isBarcodeEnabled: vi.fn().mockReturnValue(false),
}));

vi.mock('./api-client', async (importOriginal) => {
  const actual = await importOriginal<any>();
  return {
    ...actual,
    apiClient: {
      addGame: vi.fn(),
      bulkAddGames: vi.fn(),
      addPatron: vi.fn(),
      getPatronByBarcode: vi.fn(),
    },
  };
});

vi.mock('./toast-store', () => ({
  toasts: { add: vi.fn() },
}));

describe('AdminView', () => {
  it('Should render the Games tab as active by default', () => {
    render(AdminView);
    expect(screen.getByRole('button', { name: 'Add Game' })).toBeInTheDocument();
  });

  it('Should render the Games tab label', () => {
    render(AdminView);
    expect(screen.getByText('Games')).toBeInTheDocument();
  });

  it('Should render the Patrons tab label', () => {
    render(AdminView);
    expect(screen.getByText('Patrons')).toBeInTheDocument();
  });
});

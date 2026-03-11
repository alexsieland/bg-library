import { render, screen } from '@testing-library/svelte';
import App from './App.svelte';
import { describe, it, expect, vi } from 'vitest';

vi.mock('./lib/config', () => ({
  getBackendUrl: () => 'http://localhost:8080',
  isBarcodeEnabled: vi.fn().mockReturnValue(false),
}));

vi.mock('./lib/api-client', async (importOriginal) => {
  const actual = await importOriginal<any>();
  return {
    ...actual,
    apiClient: {
      listGames: vi.fn().mockResolvedValue({ games: [] }),
      listPlayToWinGames: vi.fn().mockResolvedValue({ games: [] }),
      addGame: vi.fn(),
      bulkAddGames: vi.fn(),
      addPatron: vi.fn(),
      getPatronByBarcode: vi.fn(),
    },
  };
});

// Mock child components to focus on App's own behavior
vi.mock('./lib/CheckOutTable.svelte', () => ({
  default: vi.fn(() => null),
}));

vi.mock('./lib/CheckInTable.svelte', () => ({
  default: vi.fn(() => null),
}));

vi.mock('./lib/PlayToWinTable.svelte', () => ({
  default: vi.fn(() => null),
}));

vi.mock('./lib/AdminView.svelte', () => ({
  default: vi.fn(() => null),
}));

vi.mock('./lib/ToastContainer.svelte', () => ({
  default: vi.fn(() => null),
}));

vi.mock('./lib/AppNavbar.svelte', () => ({
  default: vi.fn(() => null),
}));

describe('App', () => {
  it('Should render with checkout tab active by default', () => {
    render(App);
    // Test that App renders the checkout heading (from App.svelte itself, not child components)
    const heading = screen.getByRole('heading', { level: 1 });
    expect(heading).toHaveTextContent('Checkout Games');
  });

  it('Should render Check In Games heading when checkin tab is active', () => {
    render(App, { props: { activeTab: 'checkin' } });
    const heading = screen.getByRole('heading', { level: 1 });
    expect(heading).toHaveTextContent('Check In Games');
  });

  it('Should render Play To Win heading when ptw tab is active', () => {
    render(App, { props: { activeTab: 'ptw' } });
    const heading = screen.getByRole('heading', { level: 1 });
    expect(heading).toHaveTextContent('Play To Win');
  });

  it('Should render Admin heading when admin tab is active', () => {
    render(App, { props: { activeTab: 'admin' } });
    const heading = screen.getByRole('heading', { level: 1 });
    expect(heading).toHaveTextContent('Admin');
  });
});

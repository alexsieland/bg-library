import { render, screen, waitFor } from '@testing-library/svelte';
import AdminGamesTab from './AdminGamesTab.svelte';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { apiClient } from './api-client';

vi.mock('./config', () => ({
  getBackendUrl: () => 'http://localhost:8080',
  isBarcodeEnabled: () => false,
}));

vi.mock('./api-client', async (importOriginal) => {
  const actual = await importOriginal<any>();
  return {
    ...actual,
    apiClient: {
      listGames: vi.fn(),
      deleteGame: vi.fn(),
      getGame: vi.fn(),
      updateGame: vi.fn(),
      addGame: vi.fn(),
      getGameByBarcode: vi.fn(),
      bulkAddGames: vi.fn(),
    },
  };
});

vi.mock('./toast-store', () => ({
  toasts: {
    add: vi.fn(),
  },
}));

const mockGames = {
  games: [
    {
      game: { gameId: 'g1', title: 'Catan', barcode: '9780307455925', isPlayToWin: false },
      patron: undefined,
    },
  ],
};

describe('AdminGamesTab', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(apiClient.listGames).mockResolvedValue(mockGames);
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  it('Should render the GamesManagementTable component', async () => {
    render(AdminGamesTab);

    await waitFor(() => {
      expect(screen.getByText('Catan')).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'Add Game' })).toBeInTheDocument();
    });
  });

  it('Should render table headers', async () => {
    render(AdminGamesTab);

    await waitFor(() => {
      expect(screen.getByText('Game Title')).toBeInTheDocument();
      expect(screen.getByText('Actions')).toBeInTheDocument();
    });
  });

  it('Should render search input', async () => {
    render(AdminGamesTab);

    expect(screen.getByPlaceholderText('Search by game title')).toBeInTheDocument();
  });

  it('Should load and display games on mount', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValue(mockGames);

    render(AdminGamesTab);

    await waitFor(() => {
      expect(apiClient.listGames).toHaveBeenCalled();
      expect(screen.getByText('Catan')).toBeInTheDocument();
    });
  });

  it('Should render Edit and Delete buttons for each game', async () => {
    render(AdminGamesTab);

    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'Edit' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'Delete' })).toBeInTheDocument();
    });
  });
});

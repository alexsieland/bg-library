import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';
import GamesManagementTable from './GamesManagementTable.svelte';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { apiClient } from './api-client';
import { toasts } from './toast-store';

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
    {
      game: { gameId: 'g2', title: 'Ticket to Ride', barcode: '9780387455926', isPlayToWin: true },
      patron: undefined,
    },
    {
      game: { gameId: 'g3', title: 'Azul', barcode: undefined, isPlayToWin: false },
      patron: undefined,
    },
  ],
};

describe('GamesManagementTable', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  it('Should render the games table with search input', () => {
    vi.mocked(apiClient.listGames).mockResolvedValue(mockGames);

    render(GamesManagementTable);

    expect(screen.getByPlaceholderText('Search games by title...')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Add Game' })).toBeInTheDocument();
  });

  it('Should load and display games on mount', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValue(mockGames);

    render(GamesManagementTable);

    await waitFor(() => {
      expect(screen.getByText('Catan')).toBeInTheDocument();
      expect(screen.getByText('Ticket to Ride')).toBeInTheDocument();
      expect(screen.getByText('Azul')).toBeInTheDocument();
    });
  });

  it('Should display P2W badge for Play to Win games', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValue(mockGames);

    render(GamesManagementTable);

    await waitFor(() => {
      const p2wBadges = screen.getAllByText('P2W');
      expect(p2wBadges.length).toBe(1); // Only one P2W game
    });
  });

  it('Should filter games by search term', async () => {
    vi.mocked(apiClient.listGames).mockImplementation(async (params) => {
      if (params?.title?.toLowerCase().includes('ticket')) {
        return {
          games: [
            {
              game: {
                gameId: 'g2',
                title: 'Ticket to Ride',
                barcode: '9780387455926',
                isPlayToWin: true,
              },
              patron: undefined,
            },
          ],
        };
      }
      return mockGames;
    });

    render(GamesManagementTable);

    await waitFor(() => {
      expect(screen.getByText('Catan')).toBeInTheDocument();
    });

    const searchInput = screen.getByPlaceholderText('Search games by title...');
    await fireEvent.input(searchInput, { target: { value: 'ticket' } });

    await waitFor(() => {
      expect(screen.queryByText('Catan')).not.toBeInTheDocument();
      expect(screen.getByText('Ticket to Ride')).toBeInTheDocument();
    });
  });

  it('Should show message when no games match search', async () => {
    vi.mocked(apiClient.listGames).mockImplementation(async (params) => {
      if (params?.title) {
        return { games: [] };
      }
      return mockGames;
    });

    render(GamesManagementTable);

    const searchInput = screen.getByPlaceholderText('Search games by title...');
    await fireEvent.input(searchInput, { target: { value: 'nonexistent' } });

    await waitFor(() => {
      expect(screen.getByText('No games found.')).toBeInTheDocument();
    });
  });

  it('Should show loading state initially', () => {
    vi.mocked(apiClient.listGames).mockReturnValue(new Promise(() => {})); // Never resolves

    render(GamesManagementTable);

    expect(screen.getByText('Loading games...')).toBeInTheDocument();
  });

  it('Should show error message when fetch fails', async () => {
    const errorMessage = 'Failed to fetch games';
    vi.mocked(apiClient.listGames).mockRejectedValue(new Error(errorMessage));

    render(GamesManagementTable);

    await waitFor(() => {
      expect(screen.getByText(errorMessage)).toBeInTheDocument();
    });

    expect(toasts.add).toHaveBeenCalledWith(`Failed to load games: ${errorMessage}`, 'error');
  });

  it('Should open AddGameModal when Add Game button is clicked', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValue(mockGames);

    render(GamesManagementTable);

    const addButton = screen.getByRole('button', { name: 'Add Game' });
    await fireEvent.click(addButton);

    await waitFor(() => {
      expect(screen.getByPlaceholderText('Enter game title')).toBeInTheDocument();
    });
  });

  it('Should open AddGameModal in edit mode when Edit button is clicked', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValue(mockGames);

    render(GamesManagementTable);

    await waitFor(() => {
      expect(screen.getByText('Catan')).toBeInTheDocument();
    });

    const editButtons = screen.getAllByRole('button', { name: 'Edit' });
    await fireEvent.click(editButtons[0]);

    await waitFor(() => {
      const titles = screen.getAllByText('Update Game');
      expect(titles[0]).toBeInTheDocument(); // Modal title
    });
  });

  it('Should open delete confirmation when Delete button is clicked', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValue(mockGames);

    render(GamesManagementTable);

    await waitFor(() => {
      expect(screen.getByText('Catan')).toBeInTheDocument();
    });

    const deleteButtons = screen.getAllByRole('button', { name: 'Delete' });
    await fireEvent.click(deleteButtons[0]);

    await waitFor(() => {
      expect(screen.getByText(/Are you sure you want to delete/)).toBeInTheDocument();
    });
  });

  it('Should render edit and delete buttons for each game', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValue(mockGames);

    render(GamesManagementTable);

    await waitFor(() => {
      expect(screen.getByText('Catan')).toBeInTheDocument();
    });

    const editButtons = screen.getAllByRole('button', { name: 'Edit' });
    const deleteButtons = screen.getAllByRole('button', { name: 'Delete' });

    expect(editButtons.length).toBe(3); // One for each game
    expect(deleteButtons.length).toBe(3);
  });

  it('Should show no games found when list is empty', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValue({ games: [] });

    render(GamesManagementTable);

    await waitFor(() => {
      expect(screen.getByText('No games found.')).toBeInTheDocument();
    });
  });
});

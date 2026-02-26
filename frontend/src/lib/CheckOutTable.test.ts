import { render, screen, waitFor, fireEvent } from '@testing-library/svelte';
import CheckOutTable from './CheckOutTable.svelte';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { apiClient } from './api-client';

// Mock getBackendUrl to return a consistent URL
vi.mock('./config', () => ({
  getBackendUrl: () => 'http://localhost:8080'
}));

// Mock apiClient
vi.mock('./api-client', async (importOriginal) => {
  const actual = await importOriginal<any>();
  return {
    ...actual,
    apiClient: {
      listGames: vi.fn(),
    }
  };
});

const mockGamesResponse = {
  games: [
    {
      game: { gameId: '1', title: 'Catan' },
      patron: null
    },
    {
      game: { gameId: '2', title: 'Ticket to Ride' },
      patron: { patronId: '101', name: 'John Doe' }
    }
  ]
};

describe('CheckOutTable', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Suppress console.logs during tests
    vi.spyOn(console, 'log').mockImplementation(() => {});
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  it('Should fetch games on mount', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValue(mockGamesResponse);

    render(CheckOutTable);

    expect(apiClient.listGames).toHaveBeenCalled();
    
    await waitFor(() => {
      expect(screen.getByText('Catan')).toBeInTheDocument();
      expect(screen.getByText('Ticket to Ride')).toBeInTheDocument();
    });
  });

  it('Should show loading state initially', async () => {
    // Return a promise that doesn't resolve immediately
    vi.mocked(apiClient.listGames).mockReturnValue(new Promise(() => {}));

    render(CheckOutTable);

    expect(screen.getByText('Loading games...')).toBeInTheDocument();
  });

  it('Should display "Available" badge for games without a patron', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValue(mockGamesResponse);

    render(CheckOutTable);

    await waitFor(() => {
      expect(screen.getByText('Available')).toBeInTheDocument();
    });
  });

  it('Should display patron name for checked out games', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValue(mockGamesResponse);

    render(CheckOutTable);

    await waitFor(() => {
      expect(screen.getByText('John Doe')).toBeInTheDocument();
    });
  });

  it('Should show error message when fetch fails', async () => {
    vi.mocked(apiClient.listGames).mockRejectedValue(new Error('Internal Server Error'));

    render(CheckOutTable);

    await waitFor(() => {
      expect(screen.getByText('Internal Server Error')).toBeInTheDocument();
    });
  });

  it('Should call fetch with title param when searching', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValue({ games: [] });

    render(CheckOutTable);
    
    // Wait for initial fetch
    await waitFor(() => expect(apiClient.listGames).toHaveBeenCalledTimes(1));

    const input = screen.getByRole('searchbox');
    await fireEvent.input(input, { target: { value: 'catan' } });
    
    // Press Enter to trigger immediate search
    await fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() => {
      expect(apiClient.listGames).toHaveBeenCalledWith(expect.objectContaining({ title: 'catan' }));
    });
  });
});

import { render, screen, waitFor, fireEvent } from '@testing-library/svelte';
import CheckOutTable from './CheckOutTable.svelte';
import { describe, it, expect, vi, beforeEach } from 'vitest';

// Mock getBackendUrl to return a consistent URL
vi.mock('./config', () => ({
  getBackendUrl: () => 'http://localhost:8080'
}));

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
    vi.stubGlobal('fetch', vi.fn());
    // Suppress console.logs during tests
    vi.spyOn(console, 'log').mockImplementation(() => {});
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  it('Should fetch games on mount', async () => {
    (fetch as any).mockResolvedValue({
      ok: true,
      json: async () => mockGamesResponse
    });

    render(CheckOutTable);

    expect(fetch).toHaveBeenCalledWith(expect.stringContaining('/api/v1/library/games'));
    
    await waitFor(() => {
      expect(screen.getByText('Catan')).toBeInTheDocument();
      expect(screen.getByText('Ticket to Ride')).toBeInTheDocument();
    });
  });

  it('Should show loading state initially', async () => {
    // Return a promise that doesn't resolve immediately
    (fetch as any).mockReturnValue(new Promise(() => {}));

    render(CheckOutTable);

    expect(screen.getByText('Loading games...')).toBeInTheDocument();
  });

  it('Should display "Available" badge for games without a patron', async () => {
    (fetch as any).mockResolvedValue({
      ok: true,
      json: async () => mockGamesResponse
    });

    render(CheckOutTable);

    await waitFor(() => {
      expect(screen.getByText('Available')).toBeInTheDocument();
    });
  });

  it('Should display patron name for checked out games', async () => {
    (fetch as any).mockResolvedValue({
      ok: true,
      json: async () => mockGamesResponse
    });

    render(CheckOutTable);

    await waitFor(() => {
      expect(screen.getByText('John Doe')).toBeInTheDocument();
    });
  });

  it('Should show error message when fetch fails', async () => {
    (fetch as any).mockResolvedValue({
      ok: false,
      statusText: 'Internal Server Error'
    });

    render(CheckOutTable);

    await waitFor(() => {
      expect(screen.getByText('Failed to fetch games: Internal Server Error')).toBeInTheDocument();
    });
  });

  it('Should call fetch with title param when searching', async () => {
    (fetch as any).mockResolvedValue({
      ok: true,
      json: async () => ({ games: [] })
    });

    render(CheckOutTable);
    
    // Wait for initial fetch
    await waitFor(() => expect(fetch).toHaveBeenCalledTimes(1));

    const input = screen.getByRole('searchbox');
    await fireEvent.input(input, { target: { value: 'catan' } });
    
    // Press Enter to trigger immediate search
    await fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() => {
      expect(fetch).toHaveBeenCalledWith(expect.stringContaining('title=catan'));
    });
  });
});

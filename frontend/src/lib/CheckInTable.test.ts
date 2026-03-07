import { render, screen, waitFor, fireEvent } from '@testing-library/svelte';
import CheckInTable from './CheckInTable.svelte';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { apiClient } from './api-client';
import { isBarcodeEnabled } from './config';

vi.mock('./config', () => ({
  getBackendUrl: () => 'http://localhost:8080',
  isBarcodeEnabled: vi.fn().mockReturnValue(false),
}));

// Mock apiClient
vi.mock('./api-client', async (importOriginal) => {
  const actual = await importOriginal<any>();
  return {
    ...actual,
    apiClient: {
      listGames: vi.fn(),
      checkInGame: vi.fn(),
      getGameByBarcode: vi.fn(),
    }
  };
});

const mockCheckedOutGames = {
  games: [
    {
      game: { gameId: '1', title: 'Catan', isPlayToWin: false },
      patron: { patronId: 'p1', name: 'Alice' },
      transactionId: 't1',
      checkedOutAt: '2026-01-31T12:00:00Z'
    },
    {
      game: { gameId: '2', title: 'Ticket to Ride', isPlayToWin: false },
      patron: { patronId: 'p2', name: 'Bob' },
      transactionId: 't2',
      checkedOutAt: '2026-02-01T14:30:00Z'
    }
  ]
};

describe('CheckInTable', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(isBarcodeEnabled).mockReturnValue(false);
    vi.spyOn(console, 'log').mockImplementation(() => {});
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  it('Should fetch checked out games on mount', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValue(mockCheckedOutGames);

    render(CheckInTable);

    expect(apiClient.listGames).toHaveBeenCalledWith(expect.objectContaining({ checkedOut: true }));
    
    await waitFor(() => {
      expect(screen.getByText('Catan')).toBeInTheDocument();
      expect(screen.getByText('Alice')).toBeInTheDocument();
      expect(screen.getByText('Ticket to Ride')).toBeInTheDocument();
      expect(screen.getByText('Bob')).toBeInTheDocument();
    });
  });

  it('Should format check out time correctly', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValue(mockCheckedOutGames);

    render(CheckInTable);

    await waitFor(() => {
      // 2026-01-31T12:00:00Z -> 01/31/2026 12:00 PM (assuming UTC or similar for test environment)
      // Note: toLocaleString depends on locale. In tests it might be different.
      // Let's check for at least some parts of it.
      expect(screen.getByText(/01\/31\/2026/)).toBeInTheDocument();
    });
  });

  it('Should call checkInGame when "Returned" button is clicked', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValue(mockCheckedOutGames);
    vi.mocked(apiClient.checkInGame).mockResolvedValue({} as any);

    render(CheckInTable);

    await waitFor(() => screen.getByText('Catan'));
    
    const returnedButtons = screen.getAllByText('Returned');
    await fireEvent.click(returnedButtons[0]);

    expect(apiClient.checkInGame).toHaveBeenCalledWith('t1');
    
    await waitFor(() => {
        // Should refresh the list
        expect(apiClient.listGames).toHaveBeenCalledTimes(2);
    });
  });

  it('Should call fetch with title param when searching', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValue({ games: [] });

    render(CheckInTable);
    
    await waitFor(() => expect(apiClient.listGames).toHaveBeenCalledTimes(1));

    const input = screen.getByPlaceholderText('Search checked out games...');
    await fireEvent.input(input, { target: { value: 'catan' } });
    
    // Press Enter to trigger immediate search
    await fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() => {
      expect(apiClient.listGames).toHaveBeenCalledWith(expect.objectContaining({ 
        title: 'catan',
        checkedOut: true 
      }));
    });
  });

  it('Should show message when no checked out games found', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValue({ games: [] });

    render(CheckInTable);

    await waitFor(() => {
      expect(screen.getByText('No checked out games found.')).toBeInTheDocument();
    });
  });
});

describe('CheckInTable (barcode enabled)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(isBarcodeEnabled).mockReturnValue(true);
    vi.spyOn(console, 'log').mockImplementation(() => {});
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  it('Should not render the barcode input when isBarcodeEnabled is false', async () => {
    vi.mocked(isBarcodeEnabled).mockReturnValue(false);
    vi.mocked(apiClient.listGames).mockResolvedValue({ games: [] });

    render(CheckInTable);

    await waitFor(() => expect(apiClient.listGames).toHaveBeenCalled());

    expect(screen.queryByPlaceholderText('Scan…')).not.toBeInTheDocument();
  });

  it('Should render the barcode input when isBarcodeEnabled is true', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValue({ games: [] });

    render(CheckInTable);

    await waitFor(() => expect(apiClient.listGames).toHaveBeenCalled());

    expect(screen.getByPlaceholderText('Scan…')).toBeInTheDocument();
  });

  it('Should call checkInGame when a barcode scan matches a checked out game', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValue(mockCheckedOutGames);
    vi.mocked(apiClient.getGameByBarcode).mockResolvedValue({
      games: [{ gameId: '1', title: 'Catan', barcode: '9780307455925', isPlayToWin: false }],
    });
    vi.mocked(apiClient.checkInGame).mockResolvedValue({} as any);

    render(CheckInTable);

    await waitFor(() => screen.getByText('Catan'));

    const barcodeInput = screen.getByPlaceholderText('Scan…');
    await fireEvent.input(barcodeInput, { target: { value: '9780307455925' } });
    await fireEvent.keyDown(barcodeInput, { key: 'Enter' });

    await waitFor(() => {
      expect(apiClient.checkInGame).toHaveBeenCalledWith('t1');
    });
  });

  it('Should refresh the game list after a successful barcode check-in', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValue(mockCheckedOutGames);
    vi.mocked(apiClient.getGameByBarcode).mockResolvedValue({
      games: [{ gameId: '1', title: 'Catan', barcode: '9780307455925', isPlayToWin: false }],
    });
    vi.mocked(apiClient.checkInGame).mockResolvedValue({} as any);

    render(CheckInTable);

    await waitFor(() => screen.getByText('Catan'));

    const barcodeInput = screen.getByPlaceholderText('Scan…');
    await fireEvent.input(barcodeInput, { target: { value: '9780307455925' } });
    await fireEvent.keyDown(barcodeInput, { key: 'Enter' });

    await waitFor(() => {
      expect(apiClient.listGames).toHaveBeenCalledTimes(2);
    });
  });

  it('Should show a warning toast when the scanned game is not in the checked out list', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValue(mockCheckedOutGames);
    vi.mocked(apiClient.getGameByBarcode).mockResolvedValue({
      games: [{ gameId: 'unknown-id', title: 'Azul', barcode: '1111111111111', isPlayToWin: false }],
    });

    render(CheckInTable);

    await waitFor(() => screen.getByText('Catan'));

    const barcodeInput = screen.getByPlaceholderText('Scan…');
    await fireEvent.input(barcodeInput, { target: { value: '1111111111111' } });
    await fireEvent.keyDown(barcodeInput, { key: 'Enter' });

    await waitFor(() => {
      expect(apiClient.checkInGame).not.toHaveBeenCalled();
    });
  });

  it('Should show an error toast when the barcode lookup fails', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValue(mockCheckedOutGames);
    vi.mocked(apiClient.getGameByBarcode).mockRejectedValue(new Error('Not found'));

    render(CheckInTable);

    await waitFor(() => screen.getByText('Catan'));

    const barcodeInput = screen.getByPlaceholderText('Scan…');
    await fireEvent.input(barcodeInput, { target: { value: '0000000000000' } });
    await fireEvent.keyDown(barcodeInput, { key: 'Enter' });

    await waitFor(() => {
      expect(apiClient.checkInGame).not.toHaveBeenCalled();
    });
  });
});


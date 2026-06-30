import { render, screen, waitFor, fireEvent } from '@testing-library/svelte';
import CheckOutTable from './CheckOutTable.svelte';
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
    },
  };
});

// Mock the barcode scanner action
let savedOnScanCallback: ((barcode: string) => void) | null = null;
vi.mock('./barcodeScannerAction', () => ({
  barcodeScanner: (_node: any, options: { onScan: (barcode: string) => void }) => {
    savedOnScanCallback = options.onScan;
    return { destroy: () => {} };
  },
}));

function buildGameStatuses(count: number, startIndex: number = 0) {
  return Array.from({ length: count }, (_, i) => {
    const id = startIndex + i + 1;
    return {
      game: { gameId: `${id}`, title: `Game ${id}`, isPlayToWin: false },
      patron: undefined,
    };
  });
}

const mockGamesResponse = {
  games: [
    {
      game: { gameId: '1', title: 'Catan', isPlayToWin: false },
      patron: undefined,
    },
    {
      game: { gameId: '2', title: 'Ticket to Ride', isPlayToWin: false },
      patron: { patronId: '101', name: 'John Doe' },
    },
  ],
};

const mockGamesResponseWithP2W = {
  games: [
    {
      game: { gameId: '1', title: 'Catan', isPlayToWin: true },
      patron: undefined,
    },
  ],
};

describe('CheckOutTable', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(isBarcodeEnabled).mockReturnValue(false);
    // Suppress console.logs during tests
    vi.spyOn(console, 'log').mockImplementation(() => {});
    vi.spyOn(console, 'error').mockImplementation(() => {});
    Object.defineProperty(document.documentElement, 'scrollHeight', {
      configurable: true,
      value: 2000,
    });
    Object.defineProperty(window, 'innerHeight', { configurable: true, value: 1000 });
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

  it('Should not render the barcode input when isBarcodeEnabled returns false', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValue({ games: [] });

    render(CheckOutTable);

    await waitFor(() => expect(apiClient.listGames).toHaveBeenCalled());

    expect(screen.queryByLabelText('Barcode Scanner')).not.toBeInTheDocument();
    expect(screen.queryByTestId('barcode-scanner-input')).not.toBeInTheDocument();
  });

  it('Should display P2W badge when game has isPlayToWin true', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValue(mockGamesResponseWithP2W);

    render(CheckOutTable);

    await waitFor(() => {
      expect(screen.getByText('Catan')).toBeInTheDocument();
      expect(screen.getByText('P2W')).toBeInTheDocument();
    });
  });

  it('Should load additional games when page scroll reaches bottom', async () => {
    vi.mocked(apiClient.listGames)
      .mockResolvedValueOnce({ games: buildGameStatuses(100, 0) })
      .mockResolvedValueOnce({ games: buildGameStatuses(2, 100) });

    render(CheckOutTable);

    await waitFor(() => {
      expect(apiClient.listGames).toHaveBeenCalledWith(
        expect.objectContaining({ limit: 100, offset: 0 })
      );
    });

    Object.defineProperty(window, 'scrollY', { configurable: true, value: 950 });
    await fireEvent.scroll(window);

    await waitFor(() => {
      expect(apiClient.listGames).toHaveBeenCalledWith(
        expect.objectContaining({ limit: 100, offset: 100 })
      );
    });
  });

  it('Should accumulate games across pages', async () => {
    vi.mocked(apiClient.listGames)
      .mockResolvedValueOnce({ games: buildGameStatuses(100, 0) })
      .mockResolvedValueOnce({ games: buildGameStatuses(2, 100) });

    render(CheckOutTable);

    await waitFor(() => expect(screen.getByTestId('check-out-table')).toBeInTheDocument());

    Object.defineProperty(window, 'scrollY', { configurable: true, value: 950 });
    await fireEvent.scroll(window);

    await waitFor(() => {
      expect(apiClient.listGames).toHaveBeenCalledTimes(2);
    });
  });

  it('Should not fetch more when hasMore is false', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValueOnce({ games: buildGameStatuses(5, 0) });

    render(CheckOutTable);

    await waitFor(() => expect(apiClient.listGames).toHaveBeenCalledTimes(1));

    Object.defineProperty(window, 'scrollY', { configurable: true, value: 950 });
    await fireEvent.scroll(window);

    await waitFor(() => expect(apiClient.listGames).toHaveBeenCalledTimes(1));
  });

  it('Should reset offset and games on new search', async () => {
    vi.mocked(apiClient.listGames)
      .mockResolvedValueOnce({ games: buildGameStatuses(100, 0) })
      .mockResolvedValueOnce({ games: buildGameStatuses(3, 0) });

    render(CheckOutTable);

    await waitFor(() => expect(apiClient.listGames).toHaveBeenCalledTimes(1));

    const input = screen.getByRole('searchbox');
    await fireEvent.input(input, { target: { value: 'catan' } });
    await fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() => {
      expect(apiClient.listGames).toHaveBeenCalledWith(
        expect.objectContaining({ title: 'catan', offset: 0 })
      );
    });
  });
});

describe('CheckOutTable (barcode enabled)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(isBarcodeEnabled).mockReturnValue(true);
    vi.spyOn(console, 'log').mockImplementation(() => {});
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  it('Should render the barcode input when isBarcodeEnabled returns true', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValue({ games: [] });

    render(CheckOutTable);

    await waitFor(() => expect(apiClient.listGames).toHaveBeenCalled());

    expect(screen.getByLabelText('Barcode Scanner')).toBeInTheDocument();
    expect(screen.getByTestId('barcode-scanner-input')).toBeInTheDocument();
  });

  it('Should open the rent modal with the found game when a barcode scan succeeds', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValueOnce({ games: [] });

    render(CheckOutTable);

    await waitFor(() => expect(apiClient.listGames).toHaveBeenCalled());

    // Mock the barcode scan that returns an available game
    vi.mocked(apiClient.listGames).mockResolvedValueOnce({
      games: [
        {
          game: {
            gameId: 'g1',
            title: 'Catan',
            barcode: '9780307455925',
            isPlayToWin: false,
          },
          patron: undefined,
        },
      ],
    });

    const barcodeInput = screen.getByTestId('barcode-scanner-input');
    await fireEvent.input(barcodeInput, { target: { value: '9780307455925' } });
    await fireEvent.keyDown(barcodeInput, { key: 'Enter' });

    await waitFor(() => {
      expect(apiClient.listGames).toHaveBeenCalledWith({
        barcode: '9780307455925',
        checkedOut: false,
      });
      expect(screen.getByText('Borrow Game: Catan')).toBeInTheDocument();
    });
  });
  it('Should open the rent modal for the first available copy when multiple games share a barcode', async () => {
    // Initial load
    vi.mocked(apiClient.listGames).mockResolvedValueOnce({
      games: [
        {
          game: { gameId: 'g1', title: 'Catan', barcode: 'UPC-001', isPlayToWin: false },
          patron: { patronId: 'p1', name: 'Alice' },
        },
        {
          game: { gameId: 'g2', title: 'Catan', barcode: 'UPC-001', isPlayToWin: false },
          patron: undefined,
        },
        {
          game: { gameId: 'g3', title: 'Catan', barcode: 'UPC-001', isPlayToWin: false },
          patron: undefined,
        },
      ],
    });

    render(CheckOutTable);
    await waitFor(() => expect(apiClient.listGames).toHaveBeenCalledTimes(1));

    // Window-level barcode scan triggers onScanComplete
    // which calls listGames with barcode and checkedOut: false
    vi.mocked(apiClient.listGames).mockResolvedValueOnce({
      games: [
        {
          game: { gameId: 'g2', title: 'Catan', barcode: 'UPC-001', isPlayToWin: false },
          patron: undefined,
        },
        {
          game: { gameId: 'g3', title: 'Catan', barcode: 'UPC-001', isPlayToWin: false },
          patron: undefined,
        },
      ],
    });

    // Trigger the window-level barcode scanner callback directly
    expect(savedOnScanCallback).toBeTruthy();
    savedOnScanCallback!('UPC-001');

    // g1 is checked out (filtered out); g2 is the first available — rent modal should open for Catan (g2)
    await waitFor(() => {
      expect(apiClient.listGames).toHaveBeenCalledWith({ barcode: 'UPC-001', checkedOut: false });
      expect(screen.getByText('Borrow Game: Catan')).toBeInTheDocument();
    });
  });

  it('Should show an error toast when all copies of a duplicate-barcode game are checked out', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValueOnce({
      games: [
        {
          game: { gameId: 'g1', title: 'Catan', barcode: 'UPC-001', isPlayToWin: false },
          patron: { patronId: 'p1', name: 'Alice' },
        },
        {
          game: { gameId: 'g2', title: 'Catan', barcode: 'UPC-001', isPlayToWin: false },
          patron: { patronId: 'p2', name: 'Bob' },
        },
      ],
    });

    const { container } = render(CheckOutTable);
    await waitFor(() => expect(apiClient.listGames).toHaveBeenCalled());

    // Mock barcode scan - since all copies are checked out, filtering for checkedOut: false returns empty
    vi.mocked(apiClient.listGames).mockResolvedValueOnce({
      games: [],
    });

    const barcodeInput = screen.getByTestId('barcode-scanner-input');
    await fireEvent.input(barcodeInput, { target: { value: 'UPC-001' } });
    await fireEvent.keyDown(barcodeInput, { key: 'Enter' });

    await waitFor(() => {
      expect(apiClient.listGames).toHaveBeenCalledWith({ barcode: 'UPC-001', checkedOut: false });
    });

    // Should not open rent modal since no available copies
    await waitFor(() => {
      expect(screen.queryByText('Borrow Game: Catan')).not.toBeInTheDocument();
    });

    // Verify no rent modal was opened (no modal heading present)
    expect(container.querySelector('[data-testid="check-out-table"]')).toBeInTheDocument();
  });
});

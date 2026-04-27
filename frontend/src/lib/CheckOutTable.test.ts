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
      getGameByBarcode: vi.fn(),
    },
  };
});

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

  it('Should open the loan modal with the found game when a barcode scan succeeds', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValue({ games: [] });
    vi.mocked(apiClient.getGameByBarcode).mockResolvedValue({
      games: [
        {
          gameId: 'g1',
          title: 'Catan',
          barcode: '9780307455925',
          isPlayToWin: false,
        },
      ],
    });

    render(CheckOutTable);

    await waitFor(() => expect(apiClient.listGames).toHaveBeenCalled());

    const barcodeInput = screen.getByTestId('barcode-scanner-input');
    await fireEvent.input(barcodeInput, { target: { value: '9780307455925' } });
    await fireEvent.keyDown(barcodeInput, { key: 'Enter' });

    await waitFor(() => {
      expect(apiClient.getGameByBarcode).toHaveBeenCalledWith('9780307455925');
      expect(screen.getByText('Loan Game: Catan')).toBeInTheDocument();
    });
  });

  it('Should open the loan modal for the first available copy when multiple games share a barcode', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValue({
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
    vi.mocked(apiClient.getGameByBarcode).mockResolvedValue({
      games: [
        { gameId: 'g1', title: 'Catan', barcode: 'UPC-001', isPlayToWin: false },
        { gameId: 'g2', title: 'Catan', barcode: 'UPC-001', isPlayToWin: false },
        { gameId: 'g3', title: 'Catan', barcode: 'UPC-001', isPlayToWin: false },
      ],
    });

    render(CheckOutTable);
    await waitFor(() => expect(apiClient.listGames).toHaveBeenCalled());

    const barcodeInput = screen.getByTestId('barcode-scanner-input');
    await fireEvent.input(barcodeInput, { target: { value: 'UPC-001' } });
    await fireEvent.keyDown(barcodeInput, { key: 'Enter' });

    // g1 is checked out; g2 is the first available — loan modal should open for Catan (g2)
    await waitFor(() => {
      expect(screen.getByText('Loan Game: Catan')).toBeInTheDocument();
    });
  });

  it('Should show an error toast when all copies of a duplicate-barcode game are checked out', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValue({
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
    vi.mocked(apiClient.getGameByBarcode).mockResolvedValue({
      games: [
        { gameId: 'g1', title: 'Catan', barcode: 'UPC-001', isPlayToWin: false },
        { gameId: 'g2', title: 'Catan', barcode: 'UPC-001', isPlayToWin: false },
      ],
    });

    const { container } = render(CheckOutTable);
    await waitFor(() => expect(apiClient.listGames).toHaveBeenCalled());

    const barcodeInput = screen.getByTestId('barcode-scanner-input');
    await fireEvent.input(barcodeInput, { target: { value: 'UPC-001' } });
    await fireEvent.keyDown(barcodeInput, { key: 'Enter' });

    await waitFor(() => {
      expect(screen.queryByText('Loan Game: Catan')).not.toBeInTheDocument();
    });

    // Verify no loan modal was opened (no modal heading present)
    expect(container.querySelector('[data-testid="check-out-table"]')).toBeInTheDocument();
  });
});

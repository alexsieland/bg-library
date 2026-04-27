import { render, screen, waitFor, fireEvent } from '@testing-library/svelte';
import BarcodeInput from './BarcodeInput.svelte';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { apiClient } from './api-client';

vi.mock('./config', () => ({
  getBackendUrl: () => 'http://localhost:8080',
  isBarcodeEnabled: () => true,
}));

vi.mock('./api-client', async (importOriginal) => {
  const actual = await importOriginal<any>();
  return {
    ...actual,
    apiClient: {
      getGameByBarcode: vi.fn(),
    },
  };
});

const mockGame = {
  gameId: 'g1',
  title: 'Catan',
  barcode: '9780307455925',
  isPlayToWin: false,
};

describe('BarcodeInput', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  it('Should render the barcode input field and label', () => {
    render(BarcodeInput);

    expect(screen.getByLabelText('Barcode Scanner')).toBeInTheDocument();
    expect(screen.getByTestId('barcode-scanner-input')).toBeInTheDocument();
  });

  it('Should call getGameByBarcode with the scanned value when Enter is pressed', async () => {
    vi.mocked(apiClient.getGameByBarcode).mockResolvedValue({
      games: [mockGame],
    });

    render(BarcodeInput);

    const input = screen.getByTestId('barcode-scanner-input');
    await fireEvent.input(input, { target: { value: '9780307455925' } });
    await fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() => {
      expect(apiClient.getGameByBarcode).toHaveBeenCalledWith('9780307455925');
    });
  });

  it('Should clear the input field after a scan', async () => {
    vi.mocked(apiClient.getGameByBarcode).mockResolvedValue({
      games: [mockGame],
    });

    render(BarcodeInput);

    const input = screen.getByTestId('barcode-scanner-input') as HTMLInputElement;
    await fireEvent.input(input, { target: { value: '9780307455925' } });
    await fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() => {
      expect(input.value).toBe('');
    });
  });

  it('Should call onGameFound with the matched game when exactly one game is returned', async () => {
    vi.mocked(apiClient.getGameByBarcode).mockResolvedValue({
      games: [mockGame],
    });
    const onGameFound = vi.fn();

    render(BarcodeInput, {
      props: {
        onGameFound,
        barcodeInputElement: undefined,
      },
    });

    const input = screen.getByTestId('barcode-scanner-input');
    await fireEvent.input(input, { target: { value: '9780307455925' } });
    await fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() => {
      expect(onGameFound).toHaveBeenCalledWith(mockGame);
    });
  });

  it('Should call onError with the fallback conflict message when multiple games are returned and no resolveConflict is provided', async () => {
    const conflictGames = [
      {
        gameId: 'g1',
        title: 'Catan',
        barcode: '9780307455925',
        isPlayToWin: false,
      },
      {
        gameId: 'g2',
        title: 'Catan (2nd Edition)',
        barcode: '9780307455925',
        isPlayToWin: false,
      },
    ];
    vi.mocked(apiClient.getGameByBarcode).mockResolvedValue({
      games: conflictGames,
    });
    const onError = vi.fn();

    render(BarcodeInput, {
      props: {
        onError,
        barcodeInputElement: undefined,
      },
    });

    const input = screen.getByTestId('barcode-scanner-input');
    await fireEvent.input(input, { target: { value: '9780307455925' } });
    await fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() => {
      expect(onError).toHaveBeenCalledWith(
        'Barcode conflict handling not yet implemented. Please manually trigger the check out.'
      );
    });
  });

  it('Should call onGameFound with the resolved game when resolveConflict returns a game', async () => {
    const conflictGames = [
      { gameId: 'g1', title: 'Catan', barcode: 'UPC-001', isPlayToWin: false },
      { gameId: 'g2', title: 'Catan', barcode: 'UPC-001', isPlayToWin: false },
    ];
    vi.mocked(apiClient.getGameByBarcode).mockResolvedValue({ games: conflictGames });
    const onGameFound = vi.fn();
    const resolveConflict = vi.fn().mockReturnValue(conflictGames[1]);

    render(BarcodeInput, {
      props: { onGameFound, resolveConflict, barcodeInputElement: undefined },
    });

    const input = screen.getByTestId('barcode-scanner-input');
    await fireEvent.input(input, { target: { value: 'UPC-001' } });
    await fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() => {
      expect(resolveConflict).toHaveBeenCalledWith(conflictGames);
      expect(onGameFound).toHaveBeenCalledWith(conflictGames[1]);
    });
  });

  it('Should call onError when resolveConflict returns null (all copies checked out)', async () => {
    const conflictGames = [
      { gameId: 'g1', title: 'Catan', barcode: 'UPC-001', isPlayToWin: false },
      { gameId: 'g2', title: 'Catan', barcode: 'UPC-001', isPlayToWin: false },
    ];
    vi.mocked(apiClient.getGameByBarcode).mockResolvedValue({ games: conflictGames });
    const onError = vi.fn();
    const resolveConflict = vi.fn().mockReturnValue(null);

    render(BarcodeInput, { props: { onError, resolveConflict, barcodeInputElement: undefined } });

    const input = screen.getByTestId('barcode-scanner-input');
    await fireEvent.input(input, { target: { value: 'UPC-001' } });
    await fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() => {
      expect(onError).toHaveBeenCalledWith('All copies of this game are currently checked out.');
    });
  });

  it('Should not call onGameFound when resolveConflict returns null', async () => {
    const conflictGames = [
      { gameId: 'g1', title: 'Catan', barcode: 'UPC-001', isPlayToWin: false },
      { gameId: 'g2', title: 'Catan', barcode: 'UPC-001', isPlayToWin: false },
    ];
    vi.mocked(apiClient.getGameByBarcode).mockResolvedValue({ games: conflictGames });
    const onGameFound = vi.fn();
    const resolveConflict = vi.fn().mockReturnValue(null);

    render(BarcodeInput, {
      props: { onGameFound, resolveConflict, barcodeInputElement: undefined },
    });

    const input = screen.getByTestId('barcode-scanner-input');
    await fireEvent.input(input, { target: { value: 'UPC-001' } });
    await fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() => expect(resolveConflict).toHaveBeenCalled());
    expect(onGameFound).not.toHaveBeenCalled();
  });

  it('Should call onError when the barcode is not found (404)', async () => {
    vi.mocked(apiClient.getGameByBarcode).mockRejectedValue(new Error('Not found'));
    const onError = vi.fn();

    render(BarcodeInput, {
      props: {
        onError,
        barcodeInputElement: undefined,
      },
    });

    const input = screen.getByTestId('barcode-scanner-input');
    await fireEvent.input(input, { target: { value: '0000000000000' } });
    await fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() => {
      expect(onError).toHaveBeenCalledWith('Not found');
    });
  });

  it('Should not call getGameByBarcode when Enter is pressed with an empty input', async () => {
    render(BarcodeInput);

    const input = screen.getByTestId('barcode-scanner-input');
    await fireEvent.keyDown(input, { key: 'Enter' });

    expect(apiClient.getGameByBarcode).not.toHaveBeenCalled();
  });

  it('Should not call getGameByBarcode when a key other than Enter is pressed', async () => {
    render(BarcodeInput);

    const input = screen.getByTestId('barcode-scanner-input');
    await fireEvent.input(input, { target: { value: '9780307455925' } });
    await fireEvent.keyDown(input, { key: 'a' });

    expect(apiClient.getGameByBarcode).not.toHaveBeenCalled();
  });

  it('Should show a spinner while the barcode lookup is in progress', async () => {
    // Never resolves — holds loading state open
    vi.mocked(apiClient.getGameByBarcode).mockReturnValue(new Promise(() => {}));

    render(BarcodeInput);

    const input = screen.getByTestId('barcode-scanner-input');
    await fireEvent.input(input, { target: { value: '9780307455925' } });
    await fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() => {
      expect(input).toBeDisabled();
    });
  });
});

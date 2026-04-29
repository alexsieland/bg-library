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
      listGames: vi.fn(),
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

  it('Should call listGames with the scanned barcode when Enter is pressed', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValue({
      games: [{ game: mockGame, patron: undefined }],
    });

    render(BarcodeInput, { props: { checkedOut: false } });

    const input = screen.getByTestId('barcode-scanner-input');
    await fireEvent.input(input, { target: { value: '9780307455925' } });
    await fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() => {
      expect(apiClient.listGames).toHaveBeenCalledWith({
        barcode: '9780307455925',
        checkedOut: false,
      });
    });
  });

  it('Should call listGames with checkedOut: true when that prop is provided', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValue({
      games: [{ game: mockGame, patron: { patronId: 'p1', name: 'Alice' } }],
    });

    render(BarcodeInput, { props: { checkedOut: true } });

    const input = screen.getByTestId('barcode-scanner-input');
    await fireEvent.input(input, { target: { value: '9780307455925' } });
    await fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() => {
      expect(apiClient.listGames).toHaveBeenCalledWith({
        barcode: '9780307455925',
        checkedOut: true,
      });
    });
  });

  it('Should clear the input field after a scan', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValue({
      games: [{ game: mockGame, patron: undefined }],
    });

    render(BarcodeInput, { props: { checkedOut: false } });

    const input = screen.getByTestId('barcode-scanner-input') as HTMLInputElement;
    await fireEvent.input(input, { target: { value: '9780307455925' } });
    await fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() => {
      expect(input.value).toBe('');
    });
  });

  it('Should call onStatusesFound with all returned statuses', async () => {
    const game2 = { gameId: 'g2', title: 'Catan', barcode: '9780307455925', isPlayToWin: false };
    const statuses = [
      { game: mockGame, patron: undefined },
      { game: game2, patron: undefined },
    ];
    vi.mocked(apiClient.listGames).mockResolvedValue({ games: statuses });
    const onStatusesFound = vi.fn();

    render(BarcodeInput, {
      props: { checkedOut: false, onStatusesFound, barcodeInputElement: undefined },
    });

    const input = screen.getByTestId('barcode-scanner-input');
    await fireEvent.input(input, { target: { value: '9780307455925' } });
    await fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() => {
      expect(onStatusesFound).toHaveBeenCalledWith(statuses);
    });
  });

  it('Should call onError when no games are found', async () => {
    vi.mocked(apiClient.listGames).mockResolvedValue({ games: [] });
    const onError = vi.fn();

    render(BarcodeInput, {
      props: { onError, checkedOut: false, barcodeInputElement: undefined },
    });

    const input = screen.getByTestId('barcode-scanner-input');
    await fireEvent.input(input, { target: { value: '0000000000000' } });
    await fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() => {
      expect(onError).toHaveBeenCalledWith(
        'All copies of this game are currently checked out or no game found with this barcode.'
      );
    });
  });

  it('Should call onError when the barcode lookup fails', async () => {
    vi.mocked(apiClient.listGames).mockRejectedValue(new Error('Not found'));
    const onError = vi.fn();

    render(BarcodeInput, {
      props: { onError, checkedOut: false, barcodeInputElement: undefined },
    });

    const input = screen.getByTestId('barcode-scanner-input');
    await fireEvent.input(input, { target: { value: '0000000000000' } });
    await fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() => {
      expect(onError).toHaveBeenCalledWith('Not found');
    });
  });

  it('Should not call listGames when Enter is pressed with an empty input', async () => {
    render(BarcodeInput);

    const input = screen.getByTestId('barcode-scanner-input');
    await fireEvent.keyDown(input, { key: 'Enter' });

    expect(apiClient.listGames).not.toHaveBeenCalled();
  });

  it('Should not call listGames when a key other than Enter is pressed', async () => {
    render(BarcodeInput);

    const input = screen.getByTestId('barcode-scanner-input');
    await fireEvent.input(input, { target: { value: '9780307455925' } });
    await fireEvent.keyDown(input, { key: 'a' });

    expect(apiClient.listGames).not.toHaveBeenCalled();
  });

  it('Should show a spinner while the barcode lookup is in progress', async () => {
    vi.mocked(apiClient.listGames).mockReturnValue(new Promise(() => {}));

    render(BarcodeInput, { props: { checkedOut: false } });

    const input = screen.getByTestId('barcode-scanner-input');
    await fireEvent.input(input, { target: { value: '9780307455925' } });
    await fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() => {
      expect(input).toBeDisabled();
    });
  });
});

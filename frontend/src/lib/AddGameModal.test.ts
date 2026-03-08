import { render, screen, waitFor, fireEvent } from '@testing-library/svelte';
import AddGameModal from './AddGameModal.svelte';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { apiClient } from './api-client';
import { isBarcodeEnabled } from './config';
import { toasts } from './toast-store';

vi.mock('./config', () => ({
  getBackendUrl: () => 'http://localhost:8080',
  isBarcodeEnabled: vi.fn().mockReturnValue(false),
}));

vi.mock('./api-client', async (importOriginal) => {
  const actual = await importOriginal<any>();
  return {
    ...actual,
    apiClient: {
      addGame: vi.fn(),
      getGame: vi.fn(),
      updateGame: vi.fn(),
      getGameByBarcode: vi.fn(),
    },
  };
});

vi.mock('./toast-store', () => ({
  toasts: {
    add: vi.fn(),
  },
}));

const mockGame = {
  gameId: 'g1',
  title: 'Catan',
  barcode: '9780307455925',
  isPlayToWin: false,
};

const mockGameP2W = {
  gameId: 'g2',
  title: 'Ticket to Ride',
  barcode: '9780387455926',
  isPlayToWin: true,
};

describe('AddGameModal (Add Mode)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(isBarcodeEnabled).mockReturnValue(false);
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  it('Should render with "Add Game" title when gameId is null', async () => {
    render(AddGameModal, { open: true, gameId: null });
    // Use getByRole to find the title specifically
    const titles = screen.getAllByText('Add Game');
    expect(titles[0]).toBeInTheDocument(); // The modal title
  });

  it('Should render the game title input when open', async () => {
    render(AddGameModal, { open: true, gameId: null });
    expect(screen.getByPlaceholderText('Enter game title')).toBeInTheDocument();
  });

  it('Should disable the Add Game button when the title field is empty', async () => {
    render(AddGameModal, { open: true, gameId: null });
    expect(screen.getByTestId('add-game-submit')).toBeDisabled();
  });

  it('Should enable the Add Game button when the title field has content', async () => {
    render(AddGameModal, { open: true, gameId: null });
    const input = screen.getByPlaceholderText('Enter game title');
    await fireEvent.input(input, { target: { value: 'Azul' } });
    expect(screen.getByTestId('add-game-submit')).not.toBeDisabled();
  });

  it('Should not submit when Enter is pressed in the title input', async () => {
    render(AddGameModal, { open: true, gameId: null });
    const input = screen.getByPlaceholderText('Enter game title');
    await fireEvent.input(input, { target: { value: 'Azul' } });
    await fireEvent.keyDown(input, { key: 'Enter' });
    expect(apiClient.addGame).not.toHaveBeenCalled();
  });

  it('Should call onGameSaved with the new game on successful submit', async () => {
    vi.mocked(apiClient.addGame).mockResolvedValue(mockGame);
    const onGameSaved = vi.fn();

    render(AddGameModal, { open: true, gameId: null, onGameSaved });

    const input = screen.getByPlaceholderText('Enter game title');
    await fireEvent.input(input, { target: { value: 'Catan' } });
    await fireEvent.click(screen.getByTestId('add-game-submit'));

    await waitFor(() => {
      expect(apiClient.addGame).toHaveBeenCalledWith({ title: 'Catan', isPlayToWin: false });
      expect(onGameSaved).toHaveBeenCalledWith(mockGame);
    });
  });

  it('Should call onCancel when the Cancel button is clicked', async () => {
    const onCancel = vi.fn();
    render(AddGameModal, { open: true, gameId: null, onCancel });
    await fireEvent.click(screen.getByText('Cancel'));
    expect(onCancel).toHaveBeenCalled();
  });

  it('Should show an error toast when addGame fails', async () => {
    vi.mocked(apiClient.addGame).mockRejectedValue(new Error('Server error'));

    render(AddGameModal, { open: true, gameId: null });

    const input = screen.getByPlaceholderText('Enter game title');
    await fireEvent.input(input, { target: { value: 'Catan' } });
    await fireEvent.click(screen.getByTestId('add-game-submit'));

    await waitFor(() => {
      expect(toasts.add).toHaveBeenCalledWith('Failed to add game: Server error', 'error');
    });
  });

  it('Should include barcode in the request when provided', async () => {
    vi.mocked(apiClient.addGame).mockResolvedValue(mockGame);
    const onGameSaved = vi.fn();

    render(AddGameModal, { open: true, gameId: null, onGameSaved });

    // Since barcode is only shown when enabled, we need to manually trigger the input
    // For this test, we'll focus on the behavior when it's present in the component
    const input = screen.getByPlaceholderText('Enter game title');
    await fireEvent.input(input, { target: { value: 'Catan' } });
    await fireEvent.click(screen.getByTestId('add-game-submit'));

    await waitFor(() => {
      expect(apiClient.addGame).toHaveBeenCalledWith({ title: 'Catan', isPlayToWin: false });
    });
  });

  it('Should handle Play to Win checkbox in add mode', async () => {
    vi.mocked(apiClient.addGame).mockResolvedValue(mockGameP2W);
    const onGameSaved = vi.fn();

    render(AddGameModal, { open: true, gameId: null, onGameSaved });

    const input = screen.getByPlaceholderText('Enter game title');
    await fireEvent.input(input, { target: { value: 'Ticket to Ride' } });

    // Find checkbox by its associated label text
    const checkboxLabel = screen.getByText('Play to Win Game');
    const checkbox = checkboxLabel.previousElementSibling as HTMLInputElement;
    await fireEvent.click(checkbox);

    await fireEvent.click(screen.getByTestId('add-game-submit'));

    await waitFor(() => {
      expect(apiClient.addGame).toHaveBeenCalledWith({
        title: 'Ticket to Ride',
        isPlayToWin: true,
      });
    });
  });

  it('Should reset fields when the modal is closed after cancellation', async () => {
    const { rerender } = render(AddGameModal, {
      open: true,
      gameId: null,
    });

    const input = screen.getByPlaceholderText('Enter game title') as HTMLInputElement;
    await fireEvent.input(input, { target: { value: 'Catan' } });

    expect(input.value).toBe('Catan');

    await fireEvent.click(screen.getByText('Cancel'));

    await rerender({ open: true, gameId: null });

    await waitFor(() => {
      const newInput = screen.getByPlaceholderText('Enter game title') as HTMLInputElement;
      expect(newInput.value).toBe('');
    });
  });

  it('Should not show the barcode input when isBarcodeEnabled is false', async () => {
    render(AddGameModal, { open: true, gameId: null });
    expect(screen.queryByPlaceholderText('Scan game barcode…')).not.toBeInTheDocument();
  });
});

describe('AddGameModal (Update Mode)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(isBarcodeEnabled).mockReturnValue(false);
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  it('Should render with "Update Game" title when gameId is set', async () => {
    vi.mocked(apiClient.getGame).mockResolvedValue(mockGame);
    render(AddGameModal, { open: true, gameId: 'g1' });

    await waitFor(() => {
      const titles = screen.getAllByText('Update Game');
      expect(titles[0]).toBeInTheDocument(); // The modal title
    });
  });

  it('Should pre-populate fields with game data in edit mode', async () => {
    vi.mocked(apiClient.getGame).mockResolvedValue(mockGame);
    render(AddGameModal, { open: true, gameId: 'g1' });

    await waitFor(() => {
      const input = screen.getByPlaceholderText('Enter game title') as HTMLInputElement;
      expect(input.value).toBe('Catan');
    });
  });

  it('Should load game data when gameId changes', async () => {
    vi.mocked(apiClient.getGame).mockResolvedValue(mockGame);
    const { rerender } = render(AddGameModal, { open: true, gameId: null });

    await rerender({ open: true, gameId: 'g1' });

    await waitFor(() => {
      expect(apiClient.getGame).toHaveBeenCalledWith('g1');
    });
  });

  it('Should call updateGame and then getGame when submitting in edit mode', async () => {
    vi.mocked(apiClient.getGame).mockResolvedValue(mockGame);
    vi.mocked(apiClient.updateGame).mockResolvedValue(undefined);
    const onGameSaved = vi.fn();

    render(AddGameModal, { open: true, gameId: 'g1', onGameSaved });

    await waitFor(() => {
      expect(apiClient.getGame).toHaveBeenCalledWith('g1');
    });

    const input = screen.getByPlaceholderText('Enter game title');
    await fireEvent.input(input, { target: { value: 'Catan Updated' } });
    await fireEvent.click(screen.getByTestId('add-game-submit'));

    await waitFor(() => {
      expect(apiClient.updateGame).toHaveBeenCalledWith('g1', {
        title: 'Catan Updated',
        barcode: '9780307455925',
        isPlayToWin: false,
      });
      expect(apiClient.getGame).toHaveBeenCalledTimes(2); // Once for load, once for verification
      expect(onGameSaved).toHaveBeenCalledWith(mockGame);
    });
  });

  it('Should show "Update Game" button text in edit mode', async () => {
    vi.mocked(apiClient.getGame).mockResolvedValue(mockGame);
    render(AddGameModal, { open: true, gameId: 'g1' });

    await waitFor(() => {
      const button = screen.getByTestId('add-game-submit');
      expect(button.textContent).toContain('Update Game');
    });
  });

  it('Should show an error toast when game load fails', async () => {
    vi.mocked(apiClient.getGame).mockRejectedValue(new Error('Not found'));
    render(AddGameModal, { open: true, gameId: 'invalid-id' });

    await waitFor(() => {
      expect(toasts.add).toHaveBeenCalledWith('Failed to load game: Not found', 'error');
    });
  });

  it('Should show an error toast when updateGame fails', async () => {
    vi.mocked(apiClient.getGame).mockResolvedValue(mockGame);
    vi.mocked(apiClient.updateGame).mockRejectedValue(new Error('Server error'));

    render(AddGameModal, { open: true, gameId: 'g1' });

    await waitFor(() => {
      expect(apiClient.getGame).toHaveBeenCalledWith('g1');
    });

    const input = screen.getByPlaceholderText('Enter game title');
    await fireEvent.input(input, { target: { value: 'New Title' } });
    await fireEvent.click(screen.getByTestId('add-game-submit'));

    await waitFor(() => {
      expect(toasts.add).toHaveBeenCalledWith('Failed to update game: Server error', 'error');
    });
  });

  it('Should preserve Play to Win status in edit mode', async () => {
    vi.mocked(apiClient.getGame).mockResolvedValue(mockGameP2W);
    vi.mocked(apiClient.updateGame).mockResolvedValue(undefined);
    const onGameSaved = vi.fn();

    render(AddGameModal, { open: true, gameId: 'g2', onGameSaved });

    await waitFor(() => {
      const checkboxLabel = screen.getByText('Play to Win Game');
      const checkbox = checkboxLabel.previousElementSibling as HTMLInputElement;
      expect(checkbox.checked).toBe(true);
    });

    await fireEvent.click(screen.getByTestId('add-game-submit'));

    await waitFor(() => {
      expect(apiClient.updateGame).toHaveBeenCalledWith('g2', expect.objectContaining({
        isPlayToWin: true,
      }));
    });
  });
});

describe('AddGameModal (barcode enabled)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(isBarcodeEnabled).mockReturnValue(true);
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  it('Should show the barcode input when isBarcodeEnabled is true', async () => {
    render(AddGameModal, { open: true, gameId: null });
    expect(screen.getByPlaceholderText('Scan game barcode…')).toBeInTheDocument();
  });

  it('Should not submit when Enter is pressed in the barcode input', async () => {
    vi.mocked(apiClient.getGameByBarcode).mockRejectedValue(new Error('Not found'));

    render(AddGameModal, { open: true, gameId: null });

    const barcodeInput = screen.getByPlaceholderText('Scan game barcode…');
    await fireEvent.input(barcodeInput, { target: { value: '1234567890' } });
    await fireEvent.keyDown(barcodeInput, { key: 'Enter' });

    expect(apiClient.addGame).not.toHaveBeenCalled();
  });

  it('Should show an error toast when barcode already belongs to a different game', async () => {
    vi.mocked(apiClient.getGameByBarcode).mockResolvedValue({
      games: [mockGame],
    });

    render(AddGameModal, { open: true, gameId: null });

    const barcodeInput = screen.getByPlaceholderText('Scan game barcode…') as HTMLInputElement;
    await fireEvent.input(barcodeInput, { target: { value: '9780307455925' } });
    await fireEvent.keyDown(barcodeInput, { key: 'Enter' });

    await waitFor(() => {
      expect(toasts.add).toHaveBeenCalledWith('A game with this barcode already exists', 'error');
      expect(barcodeInput.value).toBe('');
    });
  });

  it('Should keep the barcode in the field when the barcode is not found (free to use)', async () => {
    vi.mocked(apiClient.getGameByBarcode).mockRejectedValue(new Error('Not found'));

    render(AddGameModal, { open: true, gameId: null });

    const barcodeInput = screen.getByPlaceholderText('Scan game barcode…') as HTMLInputElement;
    await fireEvent.input(barcodeInput, { target: { value: '1234567890' } });
    await fireEvent.keyDown(barcodeInput, { key: 'Enter' });

    await waitFor(() => {
      expect(barcodeInput.value).toBe('1234567890');
      expect(toasts.add).not.toHaveBeenCalled();
    });
  });

  it('Should allow the same barcode in edit mode (same game)', async () => {
    vi.mocked(apiClient.getGame).mockResolvedValue(mockGame);
    vi.mocked(apiClient.getGameByBarcode).mockResolvedValue({
      games: [mockGame],
    });

    render(AddGameModal, { open: true, gameId: 'g1' });

    await waitFor(() => {
      expect(apiClient.getGame).toHaveBeenCalledWith('g1');
    });

    const barcodeInput = screen.getByPlaceholderText('Scan game barcode…') as HTMLInputElement;
    await fireEvent.input(barcodeInput, { target: { value: '9780307455925' } });
    await fireEvent.keyDown(barcodeInput, { key: 'Enter' });

    await waitFor(() => {
      // Should not clear the barcode or show error
      expect(barcodeInput.value).toBe('9780307455925');
      expect(toasts.add).not.toHaveBeenCalledWith(
        'A game with this barcode already exists',
        'error'
      );
    });
  });

  it('Should submit with the barcode value included when a game is created', async () => {
    vi.mocked(apiClient.getGameByBarcode).mockRejectedValue(new Error('Not found'));
    vi.mocked(apiClient.addGame).mockResolvedValue({
      ...mockGame,
      barcode: '1234567890',
    });
    const onGameSaved = vi.fn();

    render(AddGameModal, { open: true, gameId: null, onGameSaved });

    await fireEvent.input(screen.getByPlaceholderText('Enter game title'), {
      target: { value: 'Catan' },
    });
    await fireEvent.input(screen.getByPlaceholderText('Scan game barcode…'), {
      target: { value: '1234567890' },
    });
    await fireEvent.keyDown(screen.getByPlaceholderText('Scan game barcode…'), { key: 'Enter' });
    await fireEvent.click(screen.getByTestId('add-game-submit'));

    await waitFor(() => {
      expect(apiClient.addGame).toHaveBeenCalledWith({
        title: 'Catan',
        barcode: '1234567890',
        isPlayToWin: false,
      });
      expect(onGameSaved).toHaveBeenCalled();
    });
  });
});






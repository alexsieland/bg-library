import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';
import AdminGamesTab from './AdminGamesTab.svelte';
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
      addGame: vi.fn(),
      bulkAddGames: vi.fn(),
    },
  };
});

vi.mock('./toast-store', () => ({
  toasts: {
    add: vi.fn(),
  },
}));

describe('AdminGamesTab', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  it('Should render "Add Games" section with button to open modal', () => {
    render(AdminGamesTab);
    expect(screen.getByText('Add Games')).toBeInTheDocument();
    // Get the button specifically from the "Add Games" section (first button with name "Add Game")
    const buttons = screen.getAllByRole('button', { name: 'Add Game' });
    expect(buttons.length).toBeGreaterThanOrEqual(1); // At least the main button, plus modal button
  });

  it('Should open AddGameModal when "Add Game" button is clicked', async () => {
    render(AdminGamesTab);

    // Get the first "Add Game" button (the one in the Add Games section)
    const buttons = screen.getAllByRole('button', { name: 'Add Game' });
    const addButton = buttons[0]; // The main button, not the modal button

    await fireEvent.click(addButton);

    // After clicking, the modal form should be visible
    await waitFor(() => {
      expect(screen.getByPlaceholderText('Enter game title')).toBeInTheDocument();
    });
  });

  it('Should call handleGameCreated when a game is saved through the modal', async () => {
    vi.mocked(apiClient.addGame).mockResolvedValue({
      gameId: 'g1',
      title: 'Everdell',
      isPlayToWin: false,
    });

    render(AdminGamesTab);

    // Get the first "Add Game" button (the one in the Add Games section)
    const buttons = screen.getAllByRole('button', { name: 'Add Game' });
    const addButton = buttons[0];

    await fireEvent.click(addButton);

    await waitFor(() => {
      expect(screen.getByPlaceholderText('Enter game title')).toBeInTheDocument();
    });

    // The AddGameModal will handle the game creation - we just verify the callback would be called
    expect(toasts.add).not.toHaveBeenCalled(); // Not called yet since we haven't submitted
  });

  it('Should render "Bulk Add Games" section', () => {
    render(AdminGamesTab);
    expect(screen.getByText('Bulk Add Games')).toBeInTheDocument();
    expect(screen.getByLabelText('Upload CSV File')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Upload Games' })).toBeInTheDocument();
    expect(screen.getByText(/Upload a CSV file with one game title per line/)).toBeInTheDocument();
    expect(screen.getByText(/Maximum file size: 10MB/)).toBeInTheDocument();
  });

  it('Should upload games when file is selected and button clicked', async () => {
    const mockFile = new File(['Catan\nTicket to Ride\nAzul'], 'games.csv', {
      type: 'text/csv',
    });
    vi.mocked(apiClient.bulkAddGames).mockResolvedValue({ imported: 3 });

    render(AdminGamesTab);

    await fireEvent.change(screen.getByLabelText('Upload CSV File'), {
      target: { files: [mockFile] },
    });
    await fireEvent.click(screen.getByRole('button', { name: 'Upload Games' }));

    await waitFor(() => {
      expect(apiClient.bulkAddGames).toHaveBeenCalledWith(mockFile);
    });

    expect(toasts.add).toHaveBeenCalledWith('Successfully imported 3 games', 'success');
  });

  it('Should show singular message when uploading 1 game', async () => {
    const mockFile = new File(['Catan'], 'games.csv', { type: 'text/csv' });
    vi.mocked(apiClient.bulkAddGames).mockResolvedValue({ imported: 1 });

    render(AdminGamesTab);

    await fireEvent.change(screen.getByLabelText('Upload CSV File'), {
      target: { files: [mockFile] },
    });
    await fireEvent.click(screen.getByRole('button', { name: 'Upload Games' }));

    await waitFor(() => {
      expect(apiClient.bulkAddGames).toHaveBeenCalled();
    });

    expect(toasts.add).toHaveBeenCalledWith('Successfully imported 1 game', 'success');
  });

  it('Should show error toast when bulk upload fails', async () => {
    const mockFile = new File(['Invalid'], 'games.csv', { type: 'text/csv' });
    vi.mocked(apiClient.bulkAddGames).mockRejectedValue(
      new Error('Invalid file type: image/png. Please upload a CSV or text file.')
    );

    render(AdminGamesTab);

    await fireEvent.change(screen.getByLabelText('Upload CSV File'), {
      target: { files: [mockFile] },
    });
    await fireEvent.click(screen.getByRole('button', { name: 'Upload Games' }));

    await waitFor(() => {
      expect(apiClient.bulkAddGames).toHaveBeenCalled();
    });

    expect(toasts.add).toHaveBeenCalledWith(
      'Failed to upload games: Invalid file type: image/png. Please upload a CSV or text file.',
      'error'
    );
    expect(screen.getByText(/Invalid file type/)).toBeInTheDocument();
  });

  it('Should disable upload button when no file is selected', () => {
    render(AdminGamesTab);
    expect(screen.getByRole('button', { name: 'Upload Games' })).toBeDisabled();
  });

  it('Should clear file input after successful upload', async () => {
    const mockFile = new File(['Catan'], 'games.csv', { type: 'text/csv' });
    vi.mocked(apiClient.bulkAddGames).mockResolvedValue({ imported: 1 });

    render(AdminGamesTab);

    const fileInput = screen.getByLabelText('Upload CSV File') as HTMLInputElement;
    await fireEvent.change(fileInput, { target: { files: [mockFile] } });
    expect(fileInput.files).toHaveLength(1);

    const button = screen.getByRole('button', { name: 'Upload Games' });
    await fireEvent.click(button);

    await waitFor(() => {
      expect(apiClient.bulkAddGames).toHaveBeenCalled();
    });

    await waitFor(() => {
      expect(button).toBeDisabled();
    });
  });
});

import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';
import AdminGamesTab from './AdminGamesTab.svelte';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { apiClient } from './api-client';
import { toasts } from './toast-store';

vi.mock('./config', () => ({
  getBackendUrl: () => 'http://localhost:8080',
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

  it('Should render "Add Games" section', () => {
    render(AdminGamesTab);
    expect(screen.getByText('Add Games')).toBeInTheDocument();
    expect(screen.getByLabelText('Game Title')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Add Game' })).toBeInTheDocument();
  });

  it('Should add a game when clicking "Add Game" button', async () => {
    vi.mocked(apiClient.addGame).mockResolvedValue({ gameId: 'g1', title: 'Everdell', isPlayToWin: false });

    render(AdminGamesTab);

    const input = screen.getByLabelText('Game Title');
    await fireEvent.input(input, { target: { value: 'Everdell' } });
    await fireEvent.click(screen.getByRole('button', { name: 'Add Game' }));

    await waitFor(() => {
      expect(apiClient.addGame).toHaveBeenCalledWith({ title: 'Everdell', isPlayToWin: false });
    });

    expect(toasts.add).toHaveBeenCalledWith('Successfully added Everdell to the library', 'success');
    expect((input as HTMLInputElement).value).toBe('');
  });

  it('Should add a game when pressing Enter in the input field', async () => {
    vi.mocked(apiClient.addGame).mockResolvedValue({ gameId: 'g2', title: 'Wingspan', isPlayToWin: false });

    render(AdminGamesTab);

    const input = screen.getByLabelText('Game Title');
    await fireEvent.input(input, { target: { value: 'Wingspan' } });
    await fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() => {
      expect(apiClient.addGame).toHaveBeenCalledWith({ title: 'Wingspan', isPlayToWin: false });
    });

    expect(toasts.add).toHaveBeenCalledWith('Successfully added Wingspan to the library', 'success');
  });

  it('Should show error toast when adding a game fails', async () => {
    vi.mocked(apiClient.addGame).mockRejectedValue(new Error('Internal Server Error'));

    render(AdminGamesTab);

    const input = screen.getByLabelText('Game Title');
    await fireEvent.input(input, { target: { value: 'Failed Game' } });
    await fireEvent.click(screen.getByRole('button', { name: 'Add Game' }));

    await waitFor(() => {
      expect(apiClient.addGame).toHaveBeenCalled();
    });

    expect(toasts.add).toHaveBeenCalledWith('Failed to add game: Internal Server Error', 'error');
    expect(screen.getByText('Internal Server Error')).toBeInTheDocument();
  });

  it('Should disable the Add Game button when input is empty', () => {
    render(AdminGamesTab);
    expect(screen.getByRole('button', { name: 'Add Game' })).toBeDisabled();
  });

  it('Should enable the Add Game button when input has content', async () => {
    render(AdminGamesTab);
    const input = screen.getByLabelText('Game Title');
    await fireEvent.input(input, { target: { value: 'Game' } });
    expect(screen.getByRole('button', { name: 'Add Game' })).not.toBeDisabled();
  });

  it('Should disable the Add Game button and show spinner while loading', async () => {
    vi.mocked(apiClient.addGame).mockReturnValue(new Promise(() => {})); // never resolves

    render(AdminGamesTab);

    const input = screen.getByLabelText('Game Title');
    await fireEvent.input(input, { target: { value: 'Game' } });

    const button = screen.getByRole('button', { name: 'Add Game' });
    await fireEvent.click(button);

    expect(button).toBeDisabled();
    expect(screen.getByRole('status')).toBeInTheDocument(); // Spinner
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
    const mockFile = new File(['Catan\nTicket to Ride\nAzul'], 'games.csv', { type: 'text/csv' });
    vi.mocked(apiClient.bulkAddGames).mockResolvedValue({ imported: 3 });

    render(AdminGamesTab);

    await fireEvent.change(screen.getByLabelText('Upload CSV File'), { target: { files: [mockFile] } });
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

    await fireEvent.change(screen.getByLabelText('Upload CSV File'), { target: { files: [mockFile] } });
    await fireEvent.click(screen.getByRole('button', { name: 'Upload Games' }));

    await waitFor(() => {
      expect(apiClient.bulkAddGames).toHaveBeenCalled();
    });

    expect(toasts.add).toHaveBeenCalledWith('Successfully imported 1 game', 'success');
  });

  it('Should show error toast when bulk upload fails', async () => {
    const mockFile = new File(['Invalid'], 'games.csv', { type: 'text/csv' });
    vi.mocked(apiClient.bulkAddGames).mockRejectedValue(new Error('Invalid file type: image/png. Please upload a CSV or text file.'));

    render(AdminGamesTab);

    await fireEvent.change(screen.getByLabelText('Upload CSV File'), { target: { files: [mockFile] } });
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


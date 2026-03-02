import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';
import AdminView from './AdminView.svelte';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { apiClient } from './api-client';
import { toasts } from './toast-store';

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
      addGame: vi.fn(),
    }
  };
});

// Mock toasts
vi.mock('./toast-store', () => ({
  toasts: {
    add: vi.fn(),
  }
}));

describe('AdminView', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('Should render "Add Games" section', () => {
    render(AdminView);
    expect(screen.getByText('Add Games')).toBeInTheDocument();
    expect(screen.getByLabelText('Game Title')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Add Game' })).toBeInTheDocument();
  });

  it('Should add a game when clicking "Add Game" button', async () => {
    const mockGame = { gameId: 'g1', title: 'Everdell' };
    vi.mocked(apiClient.addGame).mockResolvedValue(mockGame);

    render(AdminView);

    const input = screen.getByLabelText('Game Title');
    await fireEvent.input(input, { target: { value: 'Everdell' } });

    const button = screen.getByRole('button', { name: 'Add Game' });
    await fireEvent.click(button);

    await waitFor(() => {
      expect(apiClient.addGame).toHaveBeenCalledWith({ title: 'Everdell' });
    });

    expect(toasts.add).toHaveBeenCalledWith('Successfully added Everdell to the library', 'success');
    expect((input as HTMLInputElement).value).toBe('');
  });

  it('Should add a game when pressing Enter in the input field', async () => {
    const mockGame = { gameId: 'g2', title: 'Wingspan' };
    vi.mocked(apiClient.addGame).mockResolvedValue(mockGame);

    render(AdminView);

    const input = screen.getByLabelText('Game Title');
    await fireEvent.input(input, { target: { value: 'Wingspan' } });
    await fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() => {
      expect(apiClient.addGame).toHaveBeenCalledWith({ title: 'Wingspan' });
    });

    expect(toasts.add).toHaveBeenCalledWith('Successfully added Wingspan to the library', 'success');
  });

  it('Should show error toast when adding a game fails', async () => {
    vi.mocked(apiClient.addGame).mockRejectedValue(new Error('Internal Server Error'));

    render(AdminView);

    const input = screen.getByLabelText('Game Title');
    await fireEvent.input(input, { target: { value: 'Failed Game' } });

    const button = screen.getByRole('button', { name: 'Add Game' });
    await fireEvent.click(button);

    await waitFor(() => {
      expect(apiClient.addGame).toHaveBeenCalled();
    });

    expect(toasts.add).toHaveBeenCalledWith('Failed to add game: Internal Server Error', 'error');
    expect(screen.getByText('Internal Server Error')).toBeInTheDocument();
  });

  it('Should disable button when input is empty or loading', async () => {
    // Initially empty
    render(AdminView);
    const button = screen.getByRole('button', { name: 'Add Game' });
    expect(button).toBeDisabled();

    // With input
    const input = screen.getByLabelText('Game Title');
    await fireEvent.input(input, { target: { value: 'Game' } });
    expect(button).not.toBeDisabled();

    // While loading
    vi.mocked(apiClient.addGame).mockReturnValue(new Promise(() => {})); // Never resolves
    await fireEvent.click(button);
    expect(button).toBeDisabled();
    expect(screen.getByRole('status')).toBeInTheDocument(); // Spinner
  });
});

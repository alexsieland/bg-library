import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import ResetRaffleConfirmationPrompt from './ResetRaffleConfirmationPrompt.svelte';
import { apiClient } from './api-client';

vi.mock('./api-client', () => ({
  apiClient: {
    resetPlayToWinGameRaffle: vi.fn(),
  },
}));

describe('ResetRaffleConfirmationPrompt', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('Should render the confirmation prompt when open', () => {
    render(ResetRaffleConfirmationPrompt, { props: { open: true } });

    expect(screen.getByTestId('reset-raffle-confirmation-prompt')).toBeInTheDocument();
    expect(screen.getByTestId('reset-raffle-cancel-button')).toBeInTheDocument();
    expect(screen.getByTestId('reset-raffle-confirm-button')).toBeInTheDocument();
  });

  it('Should call onCancel and not call reset API when cancel is clicked', async () => {
    const onCancel = vi.fn();
    render(ResetRaffleConfirmationPrompt, { props: { open: true, onCancel } });

    await fireEvent.click(screen.getByTestId('reset-raffle-cancel-button'));

    await waitFor(() => {
      expect(onCancel).toHaveBeenCalledOnce();
      expect(apiClient.resetPlayToWinGameRaffle).not.toHaveBeenCalled();
    });
  });

  it('Should call reset API when confirm is clicked', async () => {
    vi.mocked(apiClient.resetPlayToWinGameRaffle).mockResolvedValue({} as never);
    render(ResetRaffleConfirmationPrompt, { props: { open: true } });

    await fireEvent.click(screen.getByTestId('reset-raffle-confirm-button'));

    await waitFor(() => {
      expect(apiClient.resetPlayToWinGameRaffle).toHaveBeenCalledOnce();
    });
  });

  it('Should disable buttons while reset request is in progress', async () => {
    let resolveRequest: (() => void) | undefined;
    const pendingRequest = new Promise<void>((resolve) => {
      resolveRequest = resolve;
    });
    vi.mocked(apiClient.resetPlayToWinGameRaffle).mockReturnValue(pendingRequest as never);

    render(ResetRaffleConfirmationPrompt, { props: { open: true } });

    const cancelButton = screen.getByTestId('reset-raffle-cancel-button');
    const confirmButton = screen.getByTestId('reset-raffle-confirm-button');

    expect(cancelButton).toBeEnabled();
    expect(confirmButton).toBeEnabled();

    await fireEvent.click(confirmButton);

    await waitFor(() => {
      expect(cancelButton).toBeDisabled();
      expect(confirmButton).toBeDisabled();
    });

    resolveRequest?.();

    await waitFor(() => {
      expect(apiClient.resetPlayToWinGameRaffle).toHaveBeenCalledOnce();
    });
  });
});

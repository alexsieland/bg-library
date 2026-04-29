import { render, screen, waitFor, fireEvent } from '@testing-library/svelte';
import ReturnModal from './ReturnModal.svelte';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { apiClient } from './api-client';
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
      checkInGame: vi.fn(),
    },
  };
});

// ---------------------------------------------------------------------------
// Shared fixtures
// ---------------------------------------------------------------------------

const mockStatuses = [
  {
    game: { gameId: 'g1', title: 'Catan', isPlayToWin: false },
    patron: { patronId: 'p1', name: 'Alice', barcode: 'P-001' },
    transactionId: 't1',
    checkedOutAt: '2026-01-31T12:00:00Z',
  },
  {
    game: { gameId: 'g1', title: 'Catan', isPlayToWin: false },
    patron: { patronId: 'p2', name: 'Bob', barcode: 'P-002' },
    transactionId: 't2',
    checkedOutAt: '2026-02-01T09:00:00Z',
  },
];

const singleStatus = [mockStatuses[0]];

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** Type into the barcode field and fire an input event. */
async function typeBarcodeValue(value: string) {
  const input = screen.getByTestId('return-barcode-input') as HTMLInputElement;
  await fireEvent.input(input, { target: { value } });
  return input;
}

/** Type a barcode value then press Enter. */
async function scanBarcode(value: string) {
  const input = await typeBarcodeValue(value);
  await fireEvent.keyDown(input, { key: 'Enter' });
  return input;
}

// ---------------------------------------------------------------------------
// Main suite
// ---------------------------------------------------------------------------

describe('ReturnModal', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, 'log').mockImplementation(() => {});
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  // --- Rendering ---

  it('Should render the barcode input field when open', () => {
    render(ReturnModal, { open: true, statuses: mockStatuses });
    expect(screen.getByTestId('return-barcode-input')).toBeInTheDocument();
  });

  it('Should render all checked-out patron options in the dropdown', () => {
    render(ReturnModal, { open: true, statuses: mockStatuses });
    expect(screen.getByTestId('return-patron-select')).toBeInTheDocument();
    expect(screen.getByText('Alice')).toBeInTheDocument();
    expect(screen.getByText('Bob')).toBeInTheDocument();
  });

  it('Should render the no-copies message when statuses is empty', () => {
    render(ReturnModal, { open: true, statuses: [] });
    expect(screen.getByTestId('return-no-copies-message')).toBeInTheDocument();
  });

  it('Should disable the barcode input when statuses is empty', () => {
    render(ReturnModal, { open: true, statuses: [] });
    expect(screen.getByTestId('return-barcode-input')).toBeDisabled();
  });

  it('Should disable the Return button when no transaction is selected', () => {
    render(ReturnModal, { open: true, statuses: mockStatuses });
    expect(screen.getByTestId('return-modal-submit')).toBeDisabled();
  });

  // --- Auto-selection when exactly one status ---

  it('Should pre-select the transaction when there is exactly one status', async () => {
    render(ReturnModal, { open: true, statuses: singleStatus });
    await waitFor(() => {
      const select = screen.getByTestId('return-patron-select') as HTMLSelectElement;
      expect(select.value).toBe('t1');
    });
  });

  it('Should enable the Return button when there is exactly one status (auto-selected)', async () => {
    render(ReturnModal, { open: true, statuses: singleStatus });
    await waitFor(() => {
      expect(screen.getByTestId('return-modal-submit')).not.toBeDisabled();
    });
  });

  it('Should not pre-select a transaction when there are multiple statuses', () => {
    render(ReturnModal, { open: true, statuses: mockStatuses });
    const select = screen.getByTestId('return-patron-select') as HTMLSelectElement;
    expect(select.value).toBe('');
  });

  // --- Barcode scan — successful match ---

  it('Should call checkInGame with the correct transaction id when a matching barcode is scanned', async () => {
    vi.mocked(apiClient.checkInGame).mockResolvedValue(undefined);
    render(ReturnModal, { open: true, statuses: mockStatuses });

    await scanBarcode('P-001');

    await waitFor(() => {
      expect(apiClient.checkInGame).toHaveBeenCalledWith('t1');
    });
  });

  it('Should match the second patron when their barcode is scanned', async () => {
    vi.mocked(apiClient.checkInGame).mockResolvedValue(undefined);
    render(ReturnModal, { open: true, statuses: mockStatuses });

    await scanBarcode('P-002');

    await waitFor(() => {
      expect(apiClient.checkInGame).toHaveBeenCalledWith('t2');
    });
  });

  it('Should invoke onReturnSuccess after a successful barcode-triggered return', async () => {
    vi.mocked(apiClient.checkInGame).mockResolvedValue(undefined);
    const onReturnSuccess = vi.fn();
    render(ReturnModal, { open: true, statuses: mockStatuses, onReturnSuccess });

    await scanBarcode('P-001');

    await waitFor(() => {
      expect(onReturnSuccess).toHaveBeenCalled();
    });
  });

  it('Should not show a barcode error after a successful scan', async () => {
    vi.mocked(apiClient.checkInGame).mockResolvedValue(undefined);
    render(ReturnModal, { open: true, statuses: mockStatuses });

    await scanBarcode('P-001');

    await waitFor(() => expect(apiClient.checkInGame).toHaveBeenCalled());
    expect(screen.queryByTestId('return-barcode-error')).not.toBeInTheDocument();
  });

  // --- Barcode scan — no match ---

  it('Should show an inline error when the scanned barcode does not match any patron', async () => {
    render(ReturnModal, { open: true, statuses: mockStatuses });

    await scanBarcode('UNKNOWN-BARCODE');

    expect(screen.getByTestId('return-barcode-error')).toBeInTheDocument();
    expect(apiClient.checkInGame).not.toHaveBeenCalled();
  });

  it('Should include the unmatched barcode value in the error message', async () => {
    render(ReturnModal, { open: true, statuses: mockStatuses });

    await scanBarcode('UNKNOWN-123');

    expect(screen.getByTestId('return-barcode-error').textContent).toContain('UNKNOWN-123');
  });

  it('Should not call checkInGame when the barcode does not match any patron', async () => {
    render(ReturnModal, { open: true, statuses: mockStatuses });

    await scanBarcode('NO-MATCH');

    expect(apiClient.checkInGame).not.toHaveBeenCalled();
  });

  // --- Barcode scan — empty input ---

  it('Should not call checkInGame when Enter is pressed on an empty barcode field', async () => {
    render(ReturnModal, { open: true, statuses: mockStatuses });

    const input = screen.getByTestId('return-barcode-input');
    await fireEvent.keyDown(input, { key: 'Enter' });

    expect(apiClient.checkInGame).not.toHaveBeenCalled();
  });

  it('Should not show a barcode error when Enter is pressed on an empty barcode field', async () => {
    render(ReturnModal, { open: true, statuses: mockStatuses });

    const input = screen.getByTestId('return-barcode-input');
    await fireEvent.keyDown(input, { key: 'Enter' });

    expect(screen.queryByTestId('return-barcode-error')).not.toBeInTheDocument();
  });

  // --- Barcode scan — non-Enter key ---

  it('Should not call checkInGame when a non-Enter key is pressed in the barcode field', async () => {
    render(ReturnModal, { open: true, statuses: mockStatuses });

    const input = await typeBarcodeValue('P-001');
    await fireEvent.keyDown(input, { key: 'Tab' });

    expect(apiClient.checkInGame).not.toHaveBeenCalled();
  });

  // --- Manual dropdown + Return button ---

  it('Should enable the Return button after manually selecting a patron from the dropdown', async () => {
    render(ReturnModal, { open: true, statuses: mockStatuses });

    const select = screen.getByTestId('return-patron-select') as HTMLSelectElement;
    await fireEvent.change(select, { target: { value: 't1' } });

    expect(screen.getByTestId('return-modal-submit')).not.toBeDisabled();
  });

  it('Should call checkInGame with the selected transaction id when Return is clicked', async () => {
    vi.mocked(apiClient.checkInGame).mockResolvedValue(undefined);
    render(ReturnModal, { open: true, statuses: mockStatuses });

    const select = screen.getByTestId('return-patron-select') as HTMLSelectElement;
    await fireEvent.change(select, { target: { value: 't2' } });
    await fireEvent.click(screen.getByTestId('return-modal-submit'));

    await waitFor(() => {
      expect(apiClient.checkInGame).toHaveBeenCalledWith('t2');
    });
  });

  it('Should invoke onReturnSuccess after a successful manual return', async () => {
    vi.mocked(apiClient.checkInGame).mockResolvedValue(undefined);
    const onReturnSuccess = vi.fn();
    render(ReturnModal, { open: true, statuses: mockStatuses, onReturnSuccess });

    const select = screen.getByTestId('return-patron-select') as HTMLSelectElement;
    await fireEvent.change(select, { target: { value: 't1' } });
    await fireEvent.click(screen.getByTestId('return-modal-submit'));

    await waitFor(() => {
      expect(onReturnSuccess).toHaveBeenCalled();
    });
  });

  // --- Error toast on checkInGame failure ---

  it('Should add an error toast when checkInGame rejects', async () => {
    vi.mocked(apiClient.checkInGame).mockRejectedValue(new Error('Server error'));
    render(ReturnModal, { open: true, statuses: singleStatus });

    await fireEvent.click(screen.getByTestId('return-modal-submit'));

    await waitFor(() => {
      let messages: string[] = [];
      toasts.subscribe((t) => {
        messages = t.map((x) => x.message);
      })();
      expect(messages.some((m) => m.includes('Failed to return game'))).toBe(true);
    });
  });

  it('Should not invoke onReturnSuccess when checkInGame rejects', async () => {
    vi.mocked(apiClient.checkInGame).mockRejectedValue(new Error('Server error'));
    const onReturnSuccess = vi.fn();
    render(ReturnModal, { open: true, statuses: singleStatus, onReturnSuccess });

    await fireEvent.click(screen.getByTestId('return-modal-submit'));

    await waitFor(() => expect(apiClient.checkInGame).toHaveBeenCalled());
    expect(onReturnSuccess).not.toHaveBeenCalled();
  });

  // --- State reset when modal re-opens ---

  it('Should clear a previous barcode error when the modal re-opens', async () => {
    const { rerender } = render(ReturnModal, { open: true, statuses: mockStatuses });

    await scanBarcode('NO-MATCH');
    expect(screen.getByTestId('return-barcode-error')).toBeInTheDocument();

    // Close then re-open
    await rerender({ open: false, statuses: mockStatuses });
    await rerender({ open: true, statuses: mockStatuses });

    await waitFor(() => {
      expect(screen.queryByTestId('return-barcode-error')).not.toBeInTheDocument();
    });
  });

  it('Should clear the barcode input value when the modal re-opens', async () => {
    const { rerender } = render(ReturnModal, { open: true, statuses: mockStatuses });

    await typeBarcodeValue('P-001');

    await rerender({ open: false, statuses: mockStatuses });
    await rerender({ open: true, statuses: mockStatuses });

    await waitFor(() => {
      const input = screen.getByTestId('return-barcode-input') as HTMLInputElement;
      expect(input.value).toBe('');
    });
  });
});

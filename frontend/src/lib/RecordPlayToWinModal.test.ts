import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';
import RecordPlayToWinModal from './RecordPlayToWinModal.svelte';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { apiClient, type PlayToWinGame } from './api-client';

vi.mock('./config', () => ({
  getBackendUrl: () => 'http://localhost:8080',
  isBarcodeEnabled: vi.fn().mockReturnValue(false),
  getPlayToWinIdLabel: vi.fn().mockReturnValue('Leaderboard ID'),
}));

vi.mock('./api-client', async (importOriginal) => {
  const actual = await importOriginal<any>();
  return {
    ...actual,
    apiClient: {
      addPlayToWinSession: vi.fn().mockResolvedValue({}),
    },
  };
});

vi.mock('./toast-store', () => ({
  toasts: { add: vi.fn() },
}));

const mockGame: PlayToWinGame = {
  playToWinId: 'ptw-1',
  gameId: 'game-1',
  title: 'Test Game',
};

describe('RecordPlayToWinModal', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('Should not render playtime input when open', () => {
    render(RecordPlayToWinModal, { props: { open: true, playToWinGame: mockGame } });
    expect(screen.queryByTestId('ptw-playtime-input')).not.toBeInTheDocument();
  });

  it('Should render one player entry field pair by default', () => {
    render(RecordPlayToWinModal, { props: { open: true, playToWinGame: mockGame } });
    expect(screen.getByTestId('ptw-entry-0')).toBeInTheDocument();
    expect(screen.getByTestId('ptw-entrant-name-0')).toBeInTheDocument();
    expect(screen.getByTestId('ptw-entrant-id-0')).toBeInTheDocument();
  });

  it('Should show add button on the last entry row', () => {
    render(RecordPlayToWinModal, { props: { open: true, playToWinGame: mockGame } });
    expect(screen.getByTestId('ptw-add-entry-button')).toBeInTheDocument();
  });

  it('Should add a new entry field when plus button is clicked', async () => {
    render(RecordPlayToWinModal, { props: { open: true, playToWinGame: mockGame } });
    const addButton = screen.getByTestId('ptw-add-entry-button');

    await fireEvent.click(addButton);

    await waitFor(() => {
      expect(screen.getByTestId('ptw-entry-1')).toBeInTheDocument();
    });
  });

  it('Should show remove button on non-last entry rows', async () => {
    render(RecordPlayToWinModal, { props: { open: true, playToWinGame: mockGame } });
    const addButton = screen.getByTestId('ptw-add-entry-button');

    await fireEvent.click(addButton);

    await waitFor(() => {
      expect(screen.getByTestId('ptw-remove-entry-button-0')).toBeInTheDocument();
    });
  });

  it('Should remove an entry when minus button is clicked', async () => {
    render(RecordPlayToWinModal, { props: { open: true, playToWinGame: mockGame } });
    const addButton = screen.getByTestId('ptw-add-entry-button');

    // Add one entry
    await fireEvent.click(addButton);

    await waitFor(() => {
      expect(screen.getByTestId('ptw-entry-1')).toBeInTheDocument();
    });

    const removeButton = screen.getByTestId('ptw-remove-entry-button-0');
    expect(removeButton).toBeInTheDocument();
    await fireEvent.click(removeButton);

    // After removal, we should have only 1 entry left
    await waitFor(() => {
      expect(screen.queryByTestId('ptw-remove-entry-button-0')).not.toBeInTheDocument();
    });
  });

  it('Should disable submit button by default when required fields are empty', () => {
    render(RecordPlayToWinModal, { props: { open: true, playToWinGame: mockGame } });
    expect(screen.getByTestId('ptw-record-submit-button')).toBeDisabled();
  });

  it('Should enable submit button when all required fields are filled', async () => {
    render(RecordPlayToWinModal, { props: { open: true, playToWinGame: mockGame } });

    await fireEvent.input(screen.getByTestId('ptw-entrant-name-0'), {
      target: { value: 'Jane Doe' },
    });
    await fireEvent.input(screen.getByTestId('ptw-entrant-id-0'), {
      target: { value: 'ID-001' },
    });

    await waitFor(() => {
      expect(screen.getByTestId('ptw-record-submit-button')).toBeEnabled();
    });
  });

  it('Should keep submit button disabled when any added row has incomplete required fields', async () => {
    render(RecordPlayToWinModal, { props: { open: true, playToWinGame: mockGame } });

    await fireEvent.input(screen.getByTestId('ptw-entrant-name-0'), {
      target: { value: 'Jane Doe' },
    });
    await fireEvent.input(screen.getByTestId('ptw-entrant-id-0'), {
      target: { value: 'ID-001' },
    });
    await fireEvent.click(screen.getByTestId('ptw-add-entry-button'));

    await waitFor(() => {
      expect(screen.getByTestId('ptw-record-submit-button')).toBeDisabled();
    });
  });

  it('Should call addPlayToWinSession with entrant rows when submit is clicked', async () => {
    render(RecordPlayToWinModal, { props: { open: true, playToWinGame: mockGame } });

    await fireEvent.input(screen.getByTestId('ptw-entrant-name-0'), {
      target: { value: 'Jane Doe' },
    });
    await fireEvent.input(screen.getByTestId('ptw-entrant-id-0'), {
      target: { value: 'ID-001' },
    });

    await fireEvent.click(screen.getByTestId('ptw-record-submit-button'));

    await waitFor(() => {
      expect(apiClient.addPlayToWinSession).toHaveBeenCalledWith('ptw-1', [
        { entrantName: 'Jane Doe', entrantUniqueId: 'ID-001' },
      ]);
    });
  });

  it('Should show submit button', () => {
    render(RecordPlayToWinModal, { props: { open: true, playToWinGame: mockGame } });
    expect(screen.getByTestId('ptw-record-submit-button')).toBeInTheDocument();
  });

  it('Should show cancel button', () => {
    render(RecordPlayToWinModal, { props: { open: true, playToWinGame: mockGame } });
    expect(screen.getByTestId('ptw-record-cancel-button')).toBeInTheDocument();
  });

  it('Should use custom ID label from config', () => {
    render(RecordPlayToWinModal, { props: { open: true, playToWinGame: mockGame } });
    expect(screen.getByText('Leaderboard ID')).toBeInTheDocument();
  });
});

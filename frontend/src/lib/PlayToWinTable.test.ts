import { render, screen, waitFor, fireEvent } from '@testing-library/svelte';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import PlayToWinTable from './PlayToWinTable.svelte';
import { apiClient } from './api-client';

vi.mock('./config', () => ({
  getBackendUrl: () => 'http://localhost:8080',
  isBarcodeEnabled: vi.fn().mockReturnValue(false),
}));

vi.mock('./api-client', async (importOriginal) => {
  const actual = await importOriginal<any>();
  return {
    ...actual,
    apiClient: {
      listPlayToWinGames: vi.fn(),
    },
  };
});

vi.mock('./toast-store', () => ({
  toasts: { add: vi.fn() },
}));

vi.mock('./RecordPlayToWinModal.svelte', () => ({
  default: vi.fn(),
}));

function buildGames(count: number, startIndex: number = 0) {
  return Array.from({ length: count }, (_, index) => {
    const id = startIndex + index + 1;
    return {
      playToWinId: `ptw-${id}`,
      gameId: `game-${id}`,
      title: `Game ${id}`,
    };
  });
}

describe('PlayToWinTable', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    Object.defineProperty(document.documentElement, 'scrollHeight', {
      configurable: true,
      value: 2000,
    });
    Object.defineProperty(window, 'innerHeight', {
      configurable: true,
      value: 1000,
    });
  });

  it('Should render without an internal scroll container', async () => {
    vi.mocked(apiClient.listPlayToWinGames).mockResolvedValue({ games: buildGames(1) });

    render(PlayToWinTable);

    await waitFor(() => {
      expect(screen.getByTestId('ptw-table')).toBeInTheDocument();
    });

    expect(screen.queryByTestId('ptw-scroll-container')).not.toBeInTheDocument();
  });

  it('Should load additional games when page scroll reaches bottom', async () => {
    vi.mocked(apiClient.listPlayToWinGames)
      .mockResolvedValueOnce({ games: buildGames(100, 0) })
      .mockResolvedValueOnce({ games: buildGames(2, 100) });

    render(PlayToWinTable);

    await waitFor(() => {
      expect(apiClient.listPlayToWinGames).toHaveBeenCalledWith('', 100, 0);
    });

    Object.defineProperty(window, 'scrollY', {
      configurable: true,
      value: 950,
    });

    await fireEvent.scroll(window);

    await waitFor(() => {
      expect(apiClient.listPlayToWinGames).toHaveBeenCalledWith('', 100, 100);
    });
  });
});

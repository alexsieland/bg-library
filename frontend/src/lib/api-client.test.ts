import { describe, it, expect, vi, beforeEach } from 'vitest';
import { apiClient } from './api-client';

// Mock getBackendUrl to return a consistent URL
vi.mock('./config', () => ({
  getBackendUrl: () => 'http://localhost:8080'
}));

describe('ApiClient', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.stubGlobal('fetch', vi.fn());
  });

  describe('request helper (via health check)', () => {
    it('should make a request with default headers', async () => {
      vi.mocked(fetch).mockResolvedValue({
        ok: true,
        status: 204,
      } as Response);

      await apiClient.health();

      expect(fetch).toHaveBeenCalledWith(
        'http://localhost:8080/health',
        expect.objectContaining({
          headers: {
            'Content-Type': 'application/json',
          },
        })
      );
    });

    it('should throw an error when response is not ok', async () => {
      vi.mocked(fetch).mockResolvedValue({
        ok: false,
        status: 500,
        json: async () => ({ message: 'Something went wrong' }),
      } as Response);

      await expect(apiClient.health()).rejects.toThrow('Something went wrong');
    });

    it('should throw a default error message when response is not ok and json fails', async () => {
      vi.mocked(fetch).mockResolvedValue({
        ok: false,
        status: 404,
        json: async () => { throw new Error('No JSON'); },
      } as unknown as Response);

      await expect(apiClient.health()).rejects.toThrow('Request failed with status 404');
    });

    it('should return empty object for 204 status', async () => {
      vi.mocked(fetch).mockResolvedValue({
        ok: true,
        status: 204,
      } as Response);

      const result = await apiClient.health();
      expect(result).toEqual({});
    });
  });

  describe('Games API', () => {
    it('listGames should call the correct URL with query params', async () => {
      vi.mocked(fetch).mockResolvedValue({
        ok: true,
        status: 200,
        json: async () => ({ games: [] }),
      } as Response);

      await apiClient.listGames({ title: 'Catan', checkedOut: false });

      const url = new URL(vi.mocked(fetch).mock.calls[0][0] as string);
      expect(url.pathname).toBe('/api/v1/library/games');
      expect(url.searchParams.get('title')).toBe('Catan');
      expect(url.searchParams.get('checkedOut')).toBe('false');
    });

    it('addGame should make a POST request with game data', async () => {
      const mockGame = { title: 'New Game' };
      vi.mocked(fetch).mockResolvedValue({
        ok: true,
        status: 201,
        json: async () => ({ ...mockGame, gameId: '123' }),
      } as Response);

      const result = await apiClient.addGame(mockGame as any);

      expect(fetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/v1/library/game',
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify(mockGame),
        })
      );
      expect(result).toHaveProperty('gameId', '123');
    });

    it('getGame should make a GET request', async () => {
      vi.mocked(fetch).mockResolvedValue({
        ok: true,
        status: 200,
        json: async () => ({ gameId: '123', title: 'Test Game' }),
      } as Response);

      const result = await apiClient.getGame('123');

      expect(fetch).toHaveBeenCalledWith('http://localhost:8080/api/v1/library/game/123', expect.any(Object));
      expect(result).toHaveProperty('gameId', '123');
    });

    it('updateGame should make a PUT request', async () => {
      const updateData = { title: 'Updated Title' };
      vi.mocked(fetch).mockResolvedValue({
        ok: true,
        status: 204,
      } as Response);

      await apiClient.updateGame('123', updateData as any);

      expect(fetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/v1/library/game/123',
        expect.objectContaining({
          method: 'PUT',
          body: JSON.stringify(updateData),
        })
      );
    });

    it('deleteGame should make a DELETE request', async () => {
      vi.mocked(fetch).mockResolvedValue({
        ok: true,
        status: 204,
      } as Response);

      await apiClient.deleteGame('123');

      expect(fetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/v1/library/game/123',
        expect.objectContaining({
          method: 'DELETE',
        })
      );
    });
  });

  describe('Patrons API', () => {
    it('listPatrons should call the correct URL', async () => {
      vi.mocked(fetch).mockResolvedValue({
        ok: true,
        status: 200,
        json: async () => ({ patrons: [] }),
      } as Response);

      await apiClient.listPatrons();

      expect(fetch).toHaveBeenCalledWith('http://localhost:8080/api/v1/library/patrons', expect.any(Object));
    });

    it('addPatron should make a POST request', async () => {
      const mockPatron = { name: 'John Doe' };
      vi.mocked(fetch).mockResolvedValue({
        ok: true,
        status: 201,
        json: async () => ({ ...mockPatron, patronId: 'p1' }),
      } as Response);

      const result = await apiClient.addPatron(mockPatron as any);

      expect(fetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/v1/library/patron',
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify(mockPatron),
        })
      );
      expect(result).toHaveProperty('patronId', 'p1');
    });

    it('getPatron should make a GET request', async () => {
      vi.mocked(fetch).mockResolvedValue({
        ok: true,
        status: 200,
        json: async () => ({ patronId: 'p1', name: 'John' }),
      } as Response);

      await apiClient.getPatron('p1');

      expect(fetch).toHaveBeenCalledWith('http://localhost:8080/api/v1/library/patron/p1', expect.any(Object));
    });

    it('updatePatron should make a PUT request', async () => {
      const updateData = { name: 'John Updated' };
      vi.mocked(fetch).mockResolvedValue({
        ok: true,
        status: 204,
      } as Response);

      await apiClient.updatePatron('p1', updateData as any);

      expect(fetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/v1/library/patron/p1',
        expect.objectContaining({
          method: 'PUT',
          body: JSON.stringify(updateData),
        })
      );
    });

    it('deletePatron should make a DELETE request', async () => {
      vi.mocked(fetch).mockResolvedValue({
        ok: true,
        status: 204,
      } as Response);

      await apiClient.deletePatron('p1');

      expect(fetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/v1/library/patron/p1',
        expect.objectContaining({
          method: 'DELETE',
        })
      );
    });
  });

  describe('Transactions API', () => {
    it('checkOutGame should make a POST request with transaction data', async () => {
      const checkoutRequest = { gameId: 'g1', patronId: 'p1' };
      vi.mocked(fetch).mockResolvedValue({
        ok: true,
        status: 201,
        json: async () => ({ transactionId: 't1', ...checkoutRequest }),
      } as Response);

      const result = await apiClient.checkOutGame(checkoutRequest as any);

      expect(fetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/v1/library/checkout',
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify(checkoutRequest),
        })
      );
      expect(result).toHaveProperty('transactionId', 't1');
    });

    it('checkInGame should make a POST request with transactionId in query', async () => {
      vi.mocked(fetch).mockResolvedValue({
        ok: true,
        status: 204,
      } as Response);

      await apiClient.checkInGame('t1');

      const url = new URL(vi.mocked(fetch).mock.calls[0][0] as string);
      expect(url.pathname).toBe('/api/v1/library/checkin');
      expect(url.searchParams.get('transactionId')).toBe('t1');
      expect(vi.mocked(fetch).mock.calls[0][1]?.method).toBe('POST');
    });
  });
});

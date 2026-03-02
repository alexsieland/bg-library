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
    // Re-initialize the client to use the newly mocked global fetch
    (apiClient as any).init();
  });

  const mockResponse = (status: number, data?: any, ok: boolean = true) => ({
    ok,
    status,
    headers: new Headers({
      'Content-Type': 'application/json'
    }),
    json: async () => data,
    text: async () => JSON.stringify(data),
  } as Response);

  describe('request helper (via health check)', () => {
    it('should make a request with default headers', async () => {
      vi.mocked(fetch).mockResolvedValue(mockResponse(204, undefined));

      await apiClient.health();

      expect(fetch).toHaveBeenCalled();
      const firstCall = vi.mocked(fetch).mock.calls[0];
      const request = firstCall[0] as Request;
      expect(request.url).toContain('/health');
    });

    it('should throw an error when response is not ok', async () => {
      vi.mocked(fetch).mockResolvedValue(mockResponse(500, { message: 'Something went wrong' }, false));

      await expect(apiClient.health()).rejects.toThrow('Something went wrong');
    });

    it('should throw a default error message when response is not ok and json fails', async () => {
      // Mock error response as openapi-fetch expects it
      vi.mocked(fetch).mockResolvedValue({
        ok: false,
        status: 404,
        headers: new Headers({ 'Content-Type': 'application/json' }),
        json: async () => { throw new Error('No JSON'); },
        text: async () => { return JSON.stringify({ message: 'Request failed with status 404' }); },
      } as unknown as Response);

      await expect(apiClient.health()).rejects.toThrow('Request failed with status 404');
    });

    it('should return empty object for 204 status', async () => {
      vi.mocked(fetch).mockResolvedValue(mockResponse(204, undefined));

      const result = await apiClient.health();
      expect(result).toEqual({});
    });
  });

  describe('Games API', () => {
    it('listGames should call the correct URL with query params', async () => {
      vi.mocked(fetch).mockResolvedValue(mockResponse(200, { games: [] }));

      await apiClient.listGames({ title: 'Catan', checkedOut: false });

      expect(fetch).toHaveBeenCalled();
      const firstCall = vi.mocked(fetch).mock.calls[0];
      const request = firstCall[0] as Request;
      const url = new URL(request.url);
      expect(url.pathname).toBe('/api/v1/library/games');
      expect(url.searchParams.get('title')).toBe('Catan');
      expect(url.searchParams.get('checkedOut')).toBe('false');
    });

    it('addGame should make a POST request with game data', async () => {
      const mockGame = { title: 'New Game' };
      vi.mocked(fetch).mockResolvedValue(mockResponse(201, { ...mockGame, gameId: '123' }));

      const result = await apiClient.addGame(mockGame as any);

      expect(fetch).toHaveBeenCalled();
      const firstCall = vi.mocked(fetch).mock.calls[0];
      const request = firstCall[0] as Request;
      expect(request.url).toContain('/api/v1/library/game');
      expect(request.method).toBe('POST');
      
      const body = await request.json();
      expect(body).toEqual(mockGame);
      expect(result).toHaveProperty('gameId', '123');
    });

    it('getGame should make a GET request', async () => {
      vi.mocked(fetch).mockResolvedValue(mockResponse(200, { gameId: '123', title: 'Test Game' }));

      const result = await apiClient.getGame('123');

      expect(fetch).toHaveBeenCalled();
      const firstCall = vi.mocked(fetch).mock.calls[0];
      const request = firstCall[0] as Request;
      expect(request.url).toContain('/api/v1/library/game/123');
      expect(result).toHaveProperty('gameId', '123');
    });

    it('updateGame should make a PUT request', async () => {
      const updateData = { title: 'Updated Title' };
      vi.mocked(fetch).mockResolvedValue(mockResponse(204, undefined));

      await apiClient.updateGame('123', updateData as any);

      expect(fetch).toHaveBeenCalled();
      const firstCall = vi.mocked(fetch).mock.calls[0];
      const request = firstCall[0] as Request;
      expect(request.url).toContain('/api/v1/library/game/123');
      expect(request.method).toBe('PUT');
      
      const body = await request.json();
      expect(body).toEqual(updateData);
    });

    it('deleteGame should make a DELETE request', async () => {
      vi.mocked(fetch).mockResolvedValue(mockResponse(204, undefined));

      await apiClient.deleteGame('123');

      expect(fetch).toHaveBeenCalled();
      const firstCall = vi.mocked(fetch).mock.calls[0];
      const request = firstCall[0] as Request;
      expect(request.url).toContain('/api/v1/library/game/123');
      expect(request.method).toBe('DELETE');
    });
  });

  describe('Patrons API', () => {
    it('listPatrons should call the correct URL', async () => {
      vi.mocked(fetch).mockResolvedValue(mockResponse(200, { patrons: [] }));

      await apiClient.listPatrons();

      expect(fetch).toHaveBeenCalled();
      const firstCall = vi.mocked(fetch).mock.calls[0];
      const request = firstCall[0] as Request;
      const url = new URL(request.url);
      expect(url.pathname).toBe('/api/v1/library/patrons');
    });

    it('addPatron should make a POST request', async () => {
      const mockPatron = { name: 'John Doe' };
      vi.mocked(fetch).mockResolvedValue(mockResponse(201, { ...mockPatron, patronId: 'p1' }));

      const result = await apiClient.addPatron(mockPatron as any);

      expect(fetch).toHaveBeenCalled();
      const firstCall = vi.mocked(fetch).mock.calls[0];
      const request = firstCall[0] as Request;
      expect(request.url).toContain('/api/v1/library/patron');
      expect(request.method).toBe('POST');
      
      const body = await request.json();
      expect(body).toEqual(mockPatron);
      expect(result).toHaveProperty('patronId', 'p1');
    });

    it('getPatron should make a GET request', async () => {
      vi.mocked(fetch).mockResolvedValue(mockResponse(200, { patronId: 'p1', name: 'John' }));

      const result = await apiClient.getPatron('p1');

      expect(fetch).toHaveBeenCalled();
      const firstCall = vi.mocked(fetch).mock.calls[0];
      const request = firstCall[0] as Request;
      expect(request.url).toContain('/api/v1/library/patron/p1');
      expect(result).toHaveProperty('patronId', 'p1');
    });

    it('updatePatron should make a PUT request', async () => {
      const updateData = { name: 'John Updated' };
      vi.mocked(fetch).mockResolvedValue(mockResponse(204, undefined));

      await apiClient.updatePatron('p1', updateData as any);

      expect(fetch).toHaveBeenCalled();
      const firstCall = vi.mocked(fetch).mock.calls[0];
      const request = firstCall[0] as Request;
      expect(request.url).toContain('/api/v1/library/patron/p1');
      expect(request.method).toBe('PUT');
      
      const body = await request.json();
      expect(body).toEqual(updateData);
    });

    it('deletePatron should make a DELETE request', async () => {
      vi.mocked(fetch).mockResolvedValue(mockResponse(204, undefined));

      await apiClient.deletePatron('p1');

      expect(fetch).toHaveBeenCalled();
      const firstCall = vi.mocked(fetch).mock.calls[0];
      const request = firstCall[0] as Request;
      expect(request.url).toContain('/api/v1/library/patron/p1');
      expect(request.method).toBe('DELETE');
    });
  });

  describe('Transactions API', () => {
    it('checkOutGame should make a POST request with transaction data', async () => {
      const checkoutRequest = { gameId: 'g1', patronId: 'p1' };
      vi.mocked(fetch).mockResolvedValue(mockResponse(201, { transactionId: 't1', ...checkoutRequest }));

      const result = await apiClient.checkOutGame(checkoutRequest as any);

      expect(fetch).toHaveBeenCalled();
      const firstCall = vi.mocked(fetch).mock.calls[0];
      const request = firstCall[0] as Request;
      expect(request.url).toContain('/api/v1/library/checkout');
      expect(request.method).toBe('POST');
      
      const body = await request.json();
      expect(body).toEqual(checkoutRequest);
      expect(result).toHaveProperty('transactionId', 't1');
    });

    it('checkInGame should make a POST request with transactionId in query', async () => {
      vi.mocked(fetch).mockResolvedValue(mockResponse(204, undefined));

      await apiClient.checkInGame('t1');

      expect(fetch).toHaveBeenCalled();
      const firstCall = vi.mocked(fetch).mock.calls[0];
      const request = firstCall[0] as Request;
      const url = new URL(request.url);
      expect(url.pathname).toBe('/api/v1/library/checkin');
      expect(url.searchParams.get('transactionId')).toBe('t1');
      expect(request.method).toBe('POST');
    });
  });
});

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { apiClient } from './api-client';

// Mock getBackendUrl to return a consistent URL
vi.mock('./config', () => ({
  getBackendUrl: () => 'http://localhost:8080',
}));

describe('ApiClient', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.stubGlobal('fetch', vi.fn());
    // Re-initialize the client to use the newly mocked global fetch
    (apiClient as any).init();
  });

  const mockResponse = (status: number, data?: any, ok: boolean = true) =>
    ({
      ok,
      status,
      headers: new Headers({
        'Content-Type': 'application/json',
      }),
      json: async () => data,
      text: async () => JSON.stringify(data),
    }) as Response;

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
      vi.mocked(fetch).mockResolvedValue(
        mockResponse(500, { message: 'Something went wrong' }, false)
      );

      await expect(apiClient.health()).rejects.toThrow('Something went wrong');
    });

    it('should throw a default error message when response is not ok and json fails', async () => {
      // Mock error response as openapi-fetch expects it
      vi.mocked(fetch).mockResolvedValue({
        ok: false,
        status: 404,
        headers: new Headers({ 'Content-Type': 'application/json' }),
        json: async () => {
          throw new Error('No JSON');
        },
        text: async () => {
          return JSON.stringify({ message: 'Request failed with status 404' });
        },
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
      expect(request.url).toContain('/api/v1/library/game/id/123');
      expect(result).toHaveProperty('gameId', '123');
    });

    describe('getGameByBarcode', () => {
      it('Should make a GET request to the barcode URL with the provided barcode', async () => {
        const mockGames = {
          games: [{ gameId: '123', title: 'Catan', barcode: '9780307455925' }],
        };
        vi.mocked(fetch).mockResolvedValue(mockResponse(200, mockGames));

        await apiClient.getGameByBarcode('9780307455925');

        expect(fetch).toHaveBeenCalled();
        const request = vi.mocked(fetch).mock.calls[0][0] as Request;
        expect(request.url).toContain('/api/v1/library/game/barcode/9780307455925');
        expect(request.method).toBe('GET');
      });

      it('Should return a GameList when one game matches the barcode', async () => {
        const mockGames = {
          games: [{ gameId: '123', title: 'Catan', barcode: '9780307455925' }],
        };
        vi.mocked(fetch).mockResolvedValue(mockResponse(200, mockGames));

        const result = await apiClient.getGameByBarcode('9780307455925');

        expect(result.games).toHaveLength(1);
        expect(result.games[0]).toHaveProperty('gameId', '123');
        expect(result.games[0]).toHaveProperty('title', 'Catan');
      });

      it('Should return a GameList with multiple games when the barcode matches more than one game', async () => {
        const mockGames = {
          games: [
            { gameId: '123', title: 'Catan', barcode: '9780307455925' },
            {
              gameId: '456',
              title: 'Catan (2nd Edition)',
              barcode: '9780307455925',
            },
          ],
        };
        vi.mocked(fetch).mockResolvedValue(mockResponse(200, mockGames));

        const result = await apiClient.getGameByBarcode('9780307455925');

        expect(result.games).toHaveLength(2);
      });

      it('Should throw an error when no games are found for the barcode', async () => {
        vi.mocked(fetch).mockResolvedValue({
          ok: false,
          status: 404,
          headers: new Headers({ 'Content-Type': 'application/json' }),
          json: async () => ({ message: 'Not found' }),
          text: async () => JSON.stringify({ message: 'Not found' }),
        } as Response);

        await expect(apiClient.getGameByBarcode('0000000000000')).rejects.toThrow();
      });
    });

    it('updateGame should make a PUT request', async () => {
      const updateData = { title: 'Updated Title' };
      vi.mocked(fetch).mockResolvedValue(mockResponse(204, undefined));

      await apiClient.updateGame('123', updateData as any);

      expect(fetch).toHaveBeenCalled();
      const firstCall = vi.mocked(fetch).mock.calls[0];
      const request = firstCall[0] as Request;
      expect(request.url).toContain('/api/v1/library/game/id/123');
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
      expect(request.url).toContain('/api/v1/library/game/id/123');
      expect(request.method).toBe('DELETE');
    });

    describe('bulkAddGames', () => {
      it('Should successfully upload a CSV file and return imported count when file is valid', async () => {
        const csvContent = 'Catan\nTicket to Ride\nAzul';
        const mockFile = new File([csvContent], 'games.csv', {
          type: 'text/csv',
        });
        vi.mocked(fetch).mockResolvedValue(mockResponse(201, { imported: 3 }));

        const result = await apiClient.bulkAddGames(mockFile);

        expect(fetch).toHaveBeenCalled();
        const firstCall = vi.mocked(fetch).mock.calls[0];
        const request = firstCall[0] as Request;
        expect(request.url).toContain('/api/v1/library/games');
        expect(request.method).toBe('POST');
        expect(result).toEqual({ imported: 3 });

        // Verify the body is base64 encoded
        const body = await request.text();
        expect(body).toBeTruthy();
        // Decode and verify content
        const decoded = atob(body);
        expect(decoded).toBe(csvContent);
      });

      it('Should reject when file type is not CSV or text', async () => {
        const mockFile = new File(['data'], 'image.png', { type: 'image/png' });

        await expect(apiClient.bulkAddGames(mockFile)).rejects.toThrow(
          'Invalid file type: image/png'
        );
      });

      it('Should reject when file is empty', async () => {
        const mockFile = new File([], 'empty.csv', { type: 'text/csv' });

        await expect(apiClient.bulkAddGames(mockFile)).rejects.toThrow('File is empty');
      });

      it('Should reject when file exceeds 10MB limit', async () => {
        // Create a file larger than 10MB
        const largeContent = 'x'.repeat(11 * 1024 * 1024); // 11MB
        const mockFile = new File([largeContent], 'large.csv', {
          type: 'text/csv',
        });

        await expect(apiClient.bulkAddGames(mockFile)).rejects.toThrow(
          'File size exceeds 10MB limit'
        );
      });

      it('Should accept text/plain MIME type', async () => {
        const csvContent = 'Game1\nGame2';
        const mockFile = new File([csvContent], 'games.txt', {
          type: 'text/plain',
        });
        vi.mocked(fetch).mockResolvedValue(mockResponse(201, { imported: 2 }));

        const result = await apiClient.bulkAddGames(mockFile);

        expect(result).toEqual({ imported: 2 });
      });

      it('Should accept application/csv MIME type', async () => {
        const csvContent = 'Game1';
        const mockFile = new File([csvContent], 'games.csv', {
          type: 'application/csv',
        });
        vi.mocked(fetch).mockResolvedValue(mockResponse(201, { imported: 1 }));

        const result = await apiClient.bulkAddGames(mockFile);

        expect(result).toEqual({ imported: 1 });
      });

      it('Should handle UTF-8 characters correctly', async () => {
        const csvContent = 'Catan\nCafé International\nPuerto Rico';
        const mockFile = new File([csvContent], 'games.csv', {
          type: 'text/csv',
        });
        vi.mocked(fetch).mockResolvedValue(mockResponse(201, { imported: 3 }));

        const result = await apiClient.bulkAddGames(mockFile);

        expect(result).toEqual({ imported: 3 });

        const firstCall = vi.mocked(fetch).mock.calls[0];
        const request = firstCall[0] as Request;
        const body = await request.text();

        // Properly decode base64 -> UTF-8
        const binaryString = atob(body);
        const bytes = new Uint8Array(binaryString.length);
        for (let i = 0; i < binaryString.length; i++) {
          bytes[i] = binaryString.charCodeAt(i);
        }
        const decoded = new TextDecoder().decode(bytes);

        // Verify UTF-8 encoding works
        expect(decoded).toContain('Café');
      });

      it('Should propagate API errors when backend returns error', async () => {
        const mockFile = new File(['Game1'], 'games.csv', { type: 'text/csv' });
        vi.mocked(fetch).mockResolvedValue({
          ok: false,
          status: 400,
          headers: new Headers({ 'Content-Type': 'application/json' }),
          json: async () => ({ message: 'Validation failed' }),
          text: async () => JSON.stringify({ message: 'Validation failed' }),
        } as Response);

        await expect(apiClient.bulkAddGames(mockFile)).rejects.toThrow('Validation failed');
      });
    });
  });

  describe('Patrons API', () => {
    it('listPatrons should call the correct URL with query params', async () => {
      vi.mocked(fetch).mockResolvedValue(mockResponse(200, { patrons: [] }));

      await apiClient.listPatrons({ name: 'John Smith' });

      expect(fetch).toHaveBeenCalled();
      const firstCall = vi.mocked(fetch).mock.calls[0];
      const request = firstCall[0] as Request;
      const url = new URL(request.url);
      expect(url.pathname).toBe('/api/v1/library/patrons');
      expect(url.searchParams.get('name')).toBe('John Smith');
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
      expect(request.url).toContain('/api/v1/library/patron/id/p1');
      expect(result).toHaveProperty('patronId', 'p1');
    });

    describe('getPatronByBarcode', () => {
      it('Should make a GET request to the barcode URL with the provided barcode', async () => {
        const mockPatron = {
          patronId: 'p1',
          name: 'John Doe',
          barcode: 'P-12345',
        };
        vi.mocked(fetch).mockResolvedValue(mockResponse(200, mockPatron));

        await apiClient.getPatronByBarcode('P-12345');

        expect(fetch).toHaveBeenCalled();
        const request = vi.mocked(fetch).mock.calls[0][0] as Request;
        expect(request.url).toContain('/api/v1/library/patron/barcode/P-12345');
        expect(request.method).toBe('GET');
      });

      it('Should return a Patron when the barcode matches', async () => {
        const mockPatron = {
          patronId: 'p1',
          name: 'John Doe',
          barcode: 'P-12345',
        };
        vi.mocked(fetch).mockResolvedValue(mockResponse(200, mockPatron));

        const result = await apiClient.getPatronByBarcode('P-12345');

        expect(result).toHaveProperty('patronId', 'p1');
        expect(result).toHaveProperty('name', 'John Doe');
        expect(result).toHaveProperty('barcode', 'P-12345');
      });

      it('Should throw an error when no patron is found for the barcode', async () => {
        vi.mocked(fetch).mockResolvedValue({
          ok: false,
          status: 404,
          headers: new Headers({ 'Content-Type': 'application/json' }),
          json: async () => ({ message: 'Not found' }),
          text: async () => JSON.stringify({ message: 'Not found' }),
        } as Response);

        await expect(apiClient.getPatronByBarcode('INVALID-BARCODE')).rejects.toThrow();
      });
    });

    it('updatePatron should make a PUT request', async () => {
      const updateData = { name: 'John Updated' };
      vi.mocked(fetch).mockResolvedValue(mockResponse(204, undefined));

      await apiClient.updatePatron('p1', updateData as any);

      expect(fetch).toHaveBeenCalled();
      const firstCall = vi.mocked(fetch).mock.calls[0];
      const request = firstCall[0] as Request;
      expect(request.url).toContain('/api/v1/library/patron/id/p1');
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
      expect(request.url).toContain('/api/v1/library/patron/id/p1');
      expect(request.method).toBe('DELETE');
    });

    describe('bulkAddPatrons', () => {
      it('Should successfully upload a CSV file and return imported count when file is valid', async () => {
        const csvContent = 'John Smith\nJane Doe\nAlice Baker';
        const mockFile = new File([csvContent], 'patrons.csv', {
          type: 'text/csv',
        });
        vi.mocked(fetch).mockResolvedValue(mockResponse(201, { imported: 3 }));

        const result = await apiClient.bulkAddPatrons(mockFile);

        expect(fetch).toHaveBeenCalled();
        const firstCall = vi.mocked(fetch).mock.calls[0];
        const request = firstCall[0] as Request;
        expect(request.url).toContain('/api/v1/library/patrons');
        expect(request.method).toBe('POST');
        expect(result).toEqual({ imported: 3 });

        // Verify the body is base64 encoded
        const body = await request.text();
        expect(body).toBeTruthy();
        // Decode and verify content
        const decoded = atob(body);
        expect(decoded).toBe(csvContent);
      });

      it('Should reject when file type is not CSV or text', async () => {
        const mockFile = new File(['data'], 'document.pdf', {
          type: 'application/pdf',
        });

        await expect(apiClient.bulkAddPatrons(mockFile)).rejects.toThrow(
          'Invalid file type: application/pdf'
        );
      });

      it('Should reject when file is empty', async () => {
        const mockFile = new File([], 'empty.csv', { type: 'text/csv' });

        await expect(apiClient.bulkAddPatrons(mockFile)).rejects.toThrow('File is empty');
      });

      it('Should reject when file exceeds 10MB limit', async () => {
        // Create a file larger than 10MB
        const largeContent = 'x'.repeat(11 * 1024 * 1024); // 11MB
        const mockFile = new File([largeContent], 'large.csv', {
          type: 'text/csv',
        });

        await expect(apiClient.bulkAddPatrons(mockFile)).rejects.toThrow(
          'File size exceeds 10MB limit'
        );
      });

      it('Should accept text/plain MIME type', async () => {
        const csvContent = 'Patron1\nPatron2';
        const mockFile = new File([csvContent], 'patrons.txt', {
          type: 'text/plain',
        });
        vi.mocked(fetch).mockResolvedValue(mockResponse(201, { imported: 2 }));

        const result = await apiClient.bulkAddPatrons(mockFile);

        expect(result).toEqual({ imported: 2 });
      });

      it('Should accept application/csv MIME type', async () => {
        const csvContent = 'Patron1';
        const mockFile = new File([csvContent], 'patrons.csv', {
          type: 'application/csv',
        });
        vi.mocked(fetch).mockResolvedValue(mockResponse(201, { imported: 1 }));

        const result = await apiClient.bulkAddPatrons(mockFile);

        expect(result).toEqual({ imported: 1 });
      });

      it('Should handle UTF-8 characters correctly', async () => {
        const csvContent = 'José García\nMüller Schmidt\nFrançois Dupont';
        const mockFile = new File([csvContent], 'patrons.csv', {
          type: 'text/csv',
        });
        vi.mocked(fetch).mockResolvedValue(mockResponse(201, { imported: 3 }));

        const result = await apiClient.bulkAddPatrons(mockFile);

        expect(result).toEqual({ imported: 3 });

        const firstCall = vi.mocked(fetch).mock.calls[0];
        const request = firstCall[0] as Request;
        const body = await request.text();

        // Properly decode base64 -> UTF-8
        const binaryString = atob(body);
        const bytes = new Uint8Array(binaryString.length);
        for (let i = 0; i < binaryString.length; i++) {
          bytes[i] = binaryString.charCodeAt(i);
        }
        const decoded = new TextDecoder().decode(bytes);

        // Verify UTF-8 encoding works
        expect(decoded).toContain('José');
        expect(decoded).toContain('Müller');
        expect(decoded).toContain('François');
      });

      it('Should propagate API errors when backend returns error', async () => {
        const mockFile = new File(['Patron1'], 'patrons.csv', {
          type: 'text/csv',
        });
        vi.mocked(fetch).mockResolvedValue({
          ok: false,
          status: 400,
          headers: new Headers({ 'Content-Type': 'application/json' }),
          json: async () => ({ message: 'Validation failed' }),
          text: async () => JSON.stringify({ message: 'Validation failed' }),
        } as Response);

        await expect(apiClient.bulkAddPatrons(mockFile)).rejects.toThrow('Validation failed');
      });
    });
  });

  describe('Transactions API', () => {
    it('checkOutGame should make a POST request with transaction data', async () => {
      const checkoutRequest = { gameId: 'g1', patronId: 'p1' };
      vi.mocked(fetch).mockResolvedValue(
        mockResponse(201, { transactionId: 't1', ...checkoutRequest })
      );

      const result = await apiClient.checkOutGame(
        checkoutRequest.gameId as string,
        checkoutRequest.patronId as string
      );

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

  describe('Play To Win API', () => {
    describe('listPlayToWinGames', () => {
      it('Should make a GET request with title, limit, and offset query params', async () => {
        const mockGames = {
          games: [{ playToWinId: 'ptw-1', gameId: 'g1', title: 'Azul' }],
        };
        vi.mocked(fetch).mockResolvedValue(mockResponse(200, mockGames));

        const result = await apiClient.listPlayToWinGames('Azul', 25, 10);

        expect(fetch).toHaveBeenCalled();
        const request = vi.mocked(fetch).mock.calls[0][0] as Request;
        const url = new URL(request.url);
        expect(url.pathname).toBe('/api/v1/ptw/games');
        expect(url.searchParams.get('title')).toBe('Azul');
        expect(url.searchParams.get('limit')).toBe('25');
        expect(url.searchParams.get('offset')).toBe('10');
        expect(request.method).toBe('GET');
        expect(result.games).toHaveLength(1);
        expect(result.games[0]).toHaveProperty('playToWinId', 'ptw-1');
      });

      it('Should propagate API errors when backend returns an error response', async () => {
        vi.mocked(fetch).mockResolvedValue({
          ok: false,
          status: 400,
          headers: new Headers({ 'Content-Type': 'application/json' }),
          json: async () => ({
            error: {
              code: 'VALIDATION_ERROR',
              message: 'Invalid pagination params',
              details: [],
            },
          }),
          text: async () =>
            JSON.stringify({
              error: {
                code: 'VALIDATION_ERROR',
                message: 'Invalid pagination params',
                details: [],
              },
            }),
        } as Response);

        await expect(apiClient.listPlayToWinGames('Azul', 0, -1)).rejects.toThrow(
          'Invalid pagination params'
        );
      });
    });

    describe('getPlayToWinGame', () => {
      it('Should make a GET request to the ptw game URL with the provided ptwId', async () => {
        const mockGame = { playToWinId: 'ptw-1', gameId: 'g1', title: 'Azul' };
        vi.mocked(fetch).mockResolvedValue(mockResponse(200, mockGame));

        const result = await apiClient.getPlayToWinGame('ptw-1');

        expect(fetch).toHaveBeenCalled();
        const request = vi.mocked(fetch).mock.calls[0][0] as Request;
        expect(request.url).toContain('/api/v1/ptw/game/ptwId/ptw-1');
        expect(request.method).toBe('GET');
        expect(result).toHaveProperty('playToWinId', 'ptw-1');
      });

      it('Should propagate API errors when backend returns an error response', async () => {
        vi.mocked(fetch).mockResolvedValue({
          ok: false,
          status: 404,
          headers: new Headers({ 'Content-Type': 'application/json' }),
          json: async () => ({
            error: {
              code: 'NOT_FOUND',
              message: 'Play to win game not found',
              details: [],
            },
          }),
          text: async () =>
            JSON.stringify({
              error: {
                code: 'NOT_FOUND',
                message: 'Play to win game not found',
                details: [],
              },
            }),
        } as Response);

        await expect(apiClient.getPlayToWinGame('missing-id')).rejects.toThrow(
          'Play to win game not found'
        );
      });
    });

    describe('updatePlayToWinGame', () => {
      it('Should make a PUT request with the ptw game payload', async () => {
        const updatePayload = {
          playToWinId: 'ptw-1',
          gameId: 'g1',
          title: 'Azul - Updated',
        };
        vi.mocked(fetch).mockResolvedValue(mockResponse(200, updatePayload));

        const result = await apiClient.updatePlayToWinGame('ptw-1', updatePayload as any);

        expect(fetch).toHaveBeenCalled();
        const request = vi.mocked(fetch).mock.calls[0][0] as Request;
        expect(request.url).toContain('/api/v1/ptw/game/ptwId/ptw-1');
        expect(request.method).toBe('PUT');
        expect(await request.json()).toEqual(updatePayload);
        expect(result).toHaveProperty('title', 'Azul - Updated');
      });
    });

    describe('deletePlayToWinGameByPlayToWinId', () => {
      it('Should make a DELETE request with remove request payload', async () => {
        const reqBody = { reason: 'Damaged prize copy' };
        vi.mocked(fetch).mockResolvedValue(mockResponse(204, undefined));

        await apiClient.deletePlayToWinGameByPlayToWinId('ptw-1', reqBody as any);

        expect(fetch).toHaveBeenCalled();
        const request = vi.mocked(fetch).mock.calls[0][0] as Request;
        expect(request.url).toContain('/api/v1/ptw/game/ptwId/ptw-1');
        expect(request.method).toBe('DELETE');
        expect(await request.json()).toEqual(reqBody);
      });
    });

    describe('resetPlayToWinGameRaffle', () => {
      it('Should make a POST request to reset raffle endpoint', async () => {
        vi.mocked(fetch).mockResolvedValue(mockResponse(204, undefined));

        await apiClient.resetPlayToWinGameRaffle();

        expect(fetch).toHaveBeenCalled();
        const request = vi.mocked(fetch).mock.calls[0][0] as Request;
        expect(request.url).toContain('/api/v1/ptw/raffle/reset');
        expect(request.method).toBe('POST');
      });
    });

    describe('drawPlayToWinRaffle', () => {
      it('Should make a POST request to draw endpoint with the provided ptwId', async () => {
        const mockWinner = {
          entryId: 'entry-1',
          entrantName: 'Jane Doe',
          entrantUniqueId: 'P-007',
        };
        vi.mocked(fetch).mockResolvedValue(mockResponse(200, mockWinner));

        const result = await apiClient.drawPlayToWinRaffle('ptw-1');

        expect(fetch).toHaveBeenCalled();
        const request = vi.mocked(fetch).mock.calls[0][0] as Request;
        expect(request.url).toContain('/api/v1/ptw/raffle/ptwId/ptw-1');
        expect(request.method).toBe('POST');
        expect(result).toHaveProperty('entryId', 'entry-1');
      });
    });

    describe('getPlayToWinEntries', () => {
      it('Should make a GET request to the entries URL with the provided playToWinId', async () => {
        const mockEntries = {
          entries: [{ entryId: 'e1', entrantName: 'John Smith', entrantUniqueId: 'ABC123' }],
        };
        vi.mocked(fetch).mockResolvedValue(mockResponse(200, mockEntries));

        const result = await apiClient.getPlayToWinEntries('ptw-1');

        expect(fetch).toHaveBeenCalled();
        const request = vi.mocked(fetch).mock.calls[0][0] as Request;
        expect(request.url).toContain('/api/v1/ptw/entries/playToWinId/ptw-1');
        expect(request.method).toBe('GET');
        expect(result.entries).toHaveLength(1);
        expect(result.entries[0]).toHaveProperty('entrantName', 'John Smith');
      });

      it('Should return an empty entries list when there are no entries', async () => {
        vi.mocked(fetch).mockResolvedValue(mockResponse(200, { entries: [] }));

        const result = await apiClient.getPlayToWinEntries('ptw-empty');

        expect(result.entries).toEqual([]);
      });

      it('Should propagate API errors when backend returns an error response', async () => {
        vi.mocked(fetch).mockResolvedValue({
          ok: false,
          status: 404,
          headers: new Headers({ 'Content-Type': 'application/json' }),
          json: async () => ({
            error: {
              code: 'NOT_FOUND',
              message: 'Play to win game not found',
              details: [],
            },
          }),
          text: async () =>
            JSON.stringify({
              error: {
                code: 'NOT_FOUND',
                message: 'Play to win game not found',
                details: [],
              },
            }),
        } as Response);

        await expect(apiClient.getPlayToWinEntries('missing-id')).rejects.toThrow(
          'Play to win game not found'
        );
      });
    });

    describe('addPlayToWinSession', () => {
      it('Should make a POST request with playToWin session payload', async () => {
        const entries = [
          { entrantName: 'Jane Doe', entrantUniqueId: 'P-001' },
          { entrantName: 'John Smith', entrantUniqueId: 'P-002' },
        ];
        const mockSession = {
          sessionId: 's1',
          playtimeMinutes: 45,
          playToWinEntries: [
            { entryId: 'e1', entrantName: 'Jane Doe', entrantUniqueId: 'P-001' },
            { entryId: 'e2', entrantName: 'John Smith', entrantUniqueId: 'P-002' },
          ],
        };
        vi.mocked(fetch).mockResolvedValue(mockResponse(201, mockSession));

        const result = await apiClient.addPlayToWinSession('ptw-1', entries, 45);

        expect(fetch).toHaveBeenCalled();
        const request = vi.mocked(fetch).mock.calls[0][0] as Request;
        expect(request.url).toContain('/api/v1/ptw/session');
        expect(request.method).toBe('POST');

        const body = await request.json();
        expect(body).toEqual({
          playToWinId: 'ptw-1',
          playtimeMinutes: 45,
          entries,
        });
        expect(result).toHaveProperty('sessionId', 's1');
        expect(result.playToWinEntries).toHaveLength(2);
      });

      it('Should propagate API errors when backend returns validation error', async () => {
        const entries = [{ entrantName: 'Jane Doe', entrantUniqueId: 'P-001' }];
        vi.mocked(fetch).mockResolvedValue({
          ok: false,
          status: 400,
          headers: new Headers({ 'Content-Type': 'application/json' }),
          json: async () => ({
            error: {
              code: 'VALIDATION_ERROR',
              message: 'Validation failed',
              details: [{ field: 'playtimeMinutes', message: 'must be non-negative' }],
            },
          }),
          text: async () =>
            JSON.stringify({
              error: {
                code: 'VALIDATION_ERROR',
                message: 'Validation failed',
                details: [{ field: 'playtimeMinutes', message: 'must be non-negative' }],
              },
            }),
        } as Response);

        await expect(apiClient.addPlayToWinSession('ptw-1', entries, -1)).rejects.toThrow(
          'Validation failed'
        );
      });
    });
  });
});

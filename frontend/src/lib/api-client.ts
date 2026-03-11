import createClient from 'openapi-fetch';
import type { paths, components, operations } from '../generated/library-api';
import { getBackendUrl } from './config';

type ApiPath = keyof paths;

const API_PATHS = {
  health: '/health',
  addGame: '/api/v1/library/game',
  listGames: '/api/v1/library/games',
  bulkAddGames: '/api/v1/library/games',
  getGame: '/api/v1/library/game/id/{gameId}',
  updateGame: '/api/v1/library/game/id/{gameId}',
  deleteGame: '/api/v1/library/game/id/{gameId}',
  getGameByBarcode: '/api/v1/library/game/barcode/{gameBarcode}',
  addPatron: '/api/v1/library/patron',
  listPatrons: '/api/v1/library/patrons',
  bulkAddPatrons: '/api/v1/library/patrons',
  getPatron: '/api/v1/library/patron/id/{patronId}',
  updatePatron: '/api/v1/library/patron/id/{patronId}',
  deletePatron: '/api/v1/library/patron/id/{patronId}',
  getPatronByBarcode: '/api/v1/library/patron/barcode/{patronBarcode}',
  checkInGame: '/api/v1/library/checkin',
  checkOutGame: '/api/v1/library/checkout',
  listPlayToWinGames: '/api/v1/ptw/games',
  getPlayToWinEntries: '/api/v1/ptw/entries/playToWinId/{playToWinId}',
  addPlayToWinSession: '/api/v1/ptw/session',
} as const satisfies Record<string, ApiPath>;

export type Game = components['schemas']['Game'];
export type GameList = components['schemas']['GameList'];
export type GameStatusList = components['schemas']['GameStatusList'];
export type GameStatus = components['schemas']['GameStatus'];
export type Patron = components['schemas']['Patron'];
export type CreateGameRequest = components['schemas']['CreateGameRequest'];
export type CreatePatronRequest = components['schemas']['CreatePatronRequest'];
export type CheckOutRequest = components['schemas']['CheckOutRequest'];
export type LibraryTransaction = components['schemas']['LibraryTransaction'];
export type PlayToWinGameList = components['schemas']['PlayToWinGameList'];
export type PlayToWinGame = components['schemas']['PlayToWinGame'];
export type PlayToWinSession = components['schemas']['PlayToWinSession'];
export type CreatePlayToWinSessionRequest = components['schemas']['CreatePlayToWinSessionRequest'];
export type CreatePlayToWinSessionEntry =
  components['schemas']['CreatePlayToWinSessionRequest']['entries'][number];
export type PlayToWinEntryList = components['schemas']['PlayToWinEntryList'];
export type PlayToWinEntry = components['schemas']['PlayToWinEntry'];
export type ErrorResponse = components['schemas']['ErrorResponse'];
export type BulkAddResponse = components['schemas']['BulkAddResponse'];

/**
 * Validates and encodes a CSV file to base64 for bulk upload operations.
 *
 * @param file - The CSV file to encode
 * @throws Error if file is not a text/CSV file, is empty, or exceeds 10MB
 * @returns Promise resolving to base64-encoded string
 */
async function encodeCsvFile(file: File): Promise<string> {
  // Validate MIME type
  const validMimeTypes = ['text/csv', 'text/plain', 'application/csv'];
  if (!validMimeTypes.includes(file.type)) {
    throw new Error(`Invalid file type: ${file.type}. Please upload a CSV or text file.`);
  }

  // Validate file size (10MB max)
  const maxSizeBytes = 10 * 1024 * 1024; // 10MB
  if (file.size === 0) {
    throw new Error('File is empty. Please upload a file with content.');
  }
  if (file.size > maxSizeBytes) {
    throw new Error(
      `File size exceeds 10MB limit. File size: ${(file.size / (1024 * 1024)).toFixed(2)}MB`
    );
  }

  // Read file as text and encode to base64
  return new Promise((resolve, reject) => {
    const reader = new FileReader();

    reader.onload = () => {
      try {
        const text = reader.result as string;

        // Properly encode UTF-8 string to base64
        // First encode the string as UTF-8, then convert to base64
        const utf8Bytes = new TextEncoder().encode(text);

        // Convert Uint8Array to binary string using Array.from
        const binaryString = Array.from(utf8Bytes, (byte) => String.fromCharCode(byte)).join('');

        // Now encode to base64
        const base64 = btoa(binaryString);

        resolve(base64);
      } catch (error) {
        reject(
          new Error(
            `Failed to encode file: ${error instanceof Error ? error.message : 'Unknown error'}`
          )
        );
      }
    };

    reader.onerror = () => {
      reject(new Error('Failed to read file'));
    };

    reader.readAsText(file);
  });
}

class ApiClient {
  private client!: ReturnType<typeof createClient<paths>>;

  constructor() {
    this.init();
  }

  public init() {
    this.client = createClient<paths>({
      baseUrl: getBackendUrl(),
      fetch: (input: RequestInfo | URL, init?: RequestInit) => {
        return fetch(input, init);
      },
    });
  }

  private async handleResponse<T>(response: any): Promise<T> {
    if (response.error) {
      const errorMessage =
        response.error?.error?.message ??
        response.error?.message ??
        `Request failed with status ${response.response?.status}`;
      console.error('Request failed: ', errorMessage, response.error.response.data);
      throw new Error(errorMessage);
    }
    if (response.response && response.response.status === 204) {
      return {} as T;
    }
    return (response.data ?? {}) as T;
  }

  // Games
  async listGames(query?: operations['listGames']['parameters']['query']): Promise<GameStatusList> {
    const res = await this.client.GET(API_PATHS.listGames, {
      params: { query },
    });
    return this.handleResponse(res);
  }

  async addGame(game: CreateGameRequest): Promise<Game> {
    const res = await this.client.POST(API_PATHS.addGame, {
      body: game,
    });
    return this.handleResponse(res);
  }

  async getGame(gameId: string): Promise<Game> {
    const res = await this.client.GET(API_PATHS.getGame, {
      params: { path: { gameId } },
    });
    return this.handleResponse(res);
  }

  async getGameByBarcode(gameBarcode: string): Promise<GameList> {
    const res = await this.client.GET(API_PATHS.getGameByBarcode, {
      params: { path: { gameBarcode } },
    });
    return this.handleResponse(res);
  }

  async updateGame(gameId: string, game: CreateGameRequest): Promise<void> {
    const res = await this.client.PUT(API_PATHS.updateGame, {
      params: { path: { gameId } },
      body: game,
    });
    return this.handleResponse(res);
  }

  async deleteGame(gameId: string): Promise<void> {
    const res = await this.client.DELETE(API_PATHS.deleteGame, {
      params: { path: { gameId } },
    });
    return this.handleResponse(res);
  }

  async bulkAddGames(csvFile: File): Promise<BulkAddResponse> {
    const base64Content = await encodeCsvFile(csvFile);
    const res = await this.client.POST(API_PATHS.bulkAddGames, {
      body: base64Content,
      bodySerializer: (body) => body as string,
    });
    return this.handleResponse(res);
  }

  // Patrons
  async listPatrons(
    query?: operations['listPatrons']['parameters']['query']
  ): Promise<components['schemas']['PatronList']> {
    const res = await this.client.GET(API_PATHS.listPatrons, {
      params: { query },
    });
    return this.handleResponse(res);
  }

  async addPatron(patron: CreatePatronRequest): Promise<Patron> {
    const res = await this.client.POST(API_PATHS.addPatron, {
      body: patron,
    });
    return this.handleResponse(res);
  }

  async getPatron(patronId: string): Promise<Patron> {
    const res = await this.client.GET(API_PATHS.getPatron, {
      params: { path: { patronId } },
    });
    return this.handleResponse(res);
  }

  async getPatronByBarcode(patronBarcode: string): Promise<Patron> {
    const res = await this.client.GET(API_PATHS.getPatronByBarcode, {
      params: { path: { patronBarcode } },
    });
    return this.handleResponse(res);
  }

  async updatePatron(patronId: string, patron: CreatePatronRequest): Promise<void> {
    const res = await this.client.PUT(API_PATHS.updatePatron, {
      params: { path: { patronId } },
      body: patron,
    });
    return this.handleResponse(res);
  }

  async deletePatron(patronId: string): Promise<void> {
    const res = await this.client.DELETE(API_PATHS.deletePatron, {
      params: { path: { patronId } },
    });
    return this.handleResponse(res);
  }

  async bulkAddPatrons(csvFile: File): Promise<BulkAddResponse> {
    const base64Content = await encodeCsvFile(csvFile);
    const res = await this.client.POST(API_PATHS.bulkAddPatrons, {
      body: base64Content,
      bodySerializer: (body) => body as string,
    });
    return this.handleResponse(res);
  }

  // Transactions
  async checkOutGame(gameId: string, patronId: string): Promise<LibraryTransaction> {
    const reqBody: CheckOutRequest = { gameId, patronId };
    const res = await this.client.POST(API_PATHS.checkOutGame, {
      body: reqBody,
    });
    return this.handleResponse(res);
  }

  async checkInGame(transactionId: string): Promise<void> {
    const res = await this.client.POST(API_PATHS.checkInGame, {
      params: { query: { transactionId } },
    });
    return this.handleResponse(res);
  }

  // Play To Win
  async listPlayToWinGames(
    title: string,
    limit: number,
    offset: number
  ): Promise<PlayToWinGameList> {
    const res = await this.client.GET(API_PATHS.listPlayToWinGames, {
      params: { query: { title, limit, offset } },
    });
    return this.handleResponse(res);
  }

  async getPlayToWinEntries(playToWinId: string): Promise<PlayToWinEntryList> {
    const res = await this.client.GET(API_PATHS.getPlayToWinEntries, {
      params: { path: { playToWinId } },
    });
    return this.handleResponse(res);
  }

  async addPlayToWinSession(
    playToWinId: string,
    playtimeMinutes: number,
    entries: CreatePlayToWinSessionEntry[]
  ): Promise<PlayToWinSession> {
    const reqBody: CreatePlayToWinSessionRequest = { playToWinId, playtimeMinutes, entries };
    const res = await this.client.POST(API_PATHS.addPlayToWinSession, {
      body: reqBody,
    });
    return this.handleResponse(res);
  }

  // Health
  async health(): Promise<void> {
    const res = await this.client.GET(API_PATHS.health);
    return this.handleResponse(res);
  }
}

export const apiClient = new ApiClient();

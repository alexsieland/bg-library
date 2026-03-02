import createClient from 'openapi-fetch';
import type { paths, components, operations } from '../generated/library-api';
import { getBackendUrl } from './config';

export type Game = components["schemas"]["Game"];
export type GameList = components["schemas"]["GameList"];
export type GameStatus = components["schemas"]["GameStatus"];
export type Patron = components["schemas"]["Patron"];
export type CreateGameRequest = components["schemas"]["CreateGameRequest"];
export type CreatePatronRequest = components["schemas"]["CreatePatronRequest"];
export type CheckOutRequest = components["schemas"]["CheckOutRequest"];
export type LibraryTransaction = components["schemas"]["LibraryTransaction"];
export type ErrorResponse = components["schemas"]["ErrorResponse"];

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
      }
    });
  }

  private async handleResponse<T>(response: any): Promise<T> {
    if (response.error) {
      throw new Error(response.error.message || `Request failed with status ${response.response?.status}`);
    }
    if (response.response && response.response.status === 204) {
      return {} as T;
    }
    return (response.data ?? {}) as T;
  }

  // Games
  async listGames(query?: operations["listGames"]["parameters"]["query"]): Promise<GameList> {
    const res = await this.client.GET('/api/v1/library/games', {
      params: {
        query: query
      }
    });
    return this.handleResponse(res);
  }

  async addGame(game: CreateGameRequest): Promise<Game> {
    const res = await this.client.POST('/api/v1/library/game', {
      body: game
    });
    return this.handleResponse(res);
  }

  async getGame(gameId: string): Promise<Game> {
    const res = await this.client.GET('/api/v1/library/game/{gameId}', {
      params: {
        path: { gameId }
      }
    });
    return this.handleResponse(res);
  }

  async updateGame(gameId: string, game: CreateGameRequest): Promise<void> {
    const res = await this.client.PUT('/api/v1/library/game/{gameId}', {
      params: {
        path: { gameId }
      },
      body: game
    });
    return this.handleResponse(res);
  }

  async deleteGame(gameId: string): Promise<void> {
    const res = await this.client.DELETE('/api/v1/library/game/{gameId}', {
      params: {
        path: { gameId }
      }
    });
    return this.handleResponse(res);
  }

  // Patrons
  async listPatrons(query?: operations["listPatrons"]["parameters"]["query"]): Promise<components["schemas"]["PatronList"]> {
    const res = await this.client.GET('/api/v1/library/patrons',{
      params: {
        query: query
      }
    });
    return this.handleResponse(res);
  }

  async addPatron(patron: CreatePatronRequest): Promise<Patron> {
    const res = await this.client.POST('/api/v1/library/patron', {
      body: patron
    });
    return this.handleResponse(res);
  }

  async getPatron(patronId: string): Promise<Patron> {
    const res = await this.client.GET('/api/v1/library/patron/{patronId}', {
      params: {
        path: { patronId }
      }
    });
    return this.handleResponse(res);
  }

  async updatePatron(patronId: string, patron: CreatePatronRequest): Promise<void> {
    const res = await this.client.PUT('/api/v1/library/patron/{patronId}', {
      params: {
        path: { patronId }
      },
      body: patron
    });
    return this.handleResponse(res);
  }

  async deletePatron(patronId: string): Promise<void> {
    const res = await this.client.DELETE('/api/v1/library/patron/{patronId}', {
      params: {
        path: { patronId }
      }
    });
    return this.handleResponse(res);
  }

  // Transactions
  async checkOutGame(gameId: string, patronId: string): Promise<LibraryTransaction> {
    const reqBody: CheckOutRequest = {
      gameId,
      patronId
    }
    const res = await this.client.POST('/api/v1/library/checkout', {
      body: reqBody
    });
    return this.handleResponse(res);
  }

  async checkInGame(transactionId: string): Promise<void> {
    const res = await this.client.POST('/api/v1/library/checkin', {
      params: {
        query: { transactionId }
      }
    });
    return this.handleResponse(res);
  }

  // Health
  async health(): Promise<void> {
    const res = await this.client.GET('/health');
    return this.handleResponse(res);
  }
}

export const apiClient = new ApiClient();

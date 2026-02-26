import type { components, operations } from '../generated/library-api';
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
  private baseUrl: string;

  constructor() {
    this.baseUrl = getBackendUrl();
  }

  private async request<T>(path: string, options?: RequestInit): Promise<T> {
    const url = new URL(path, this.baseUrl);
    const response = await fetch(url.toString(), {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers,
      },
    });

    if (!response.ok) {
      let errorData;
      try {
        errorData = await response.json();
      } catch {
        // ignore
      }
      throw new Error(errorData?.message || `Request failed with status ${response.status}`);
    }

    if (response.status === 204) {
      return {} as T;
    }

    return response.json();
  }

  // Games
  async listGames(query?: operations["listGames"]["parameters"]["query"]): Promise<GameList> {
    const path = new URL('/api/v1/library/games', this.baseUrl);
    if (query) {
      if (query.title) path.searchParams.append('title', query.title);
      if (query.checkedOut !== undefined) path.searchParams.append('checkedOut', String(query.checkedOut));
    }
    return this.request<GameList>(path.pathname + path.search);
  }

  async addGame(game: CreateGameRequest): Promise<Game> {
    return this.request<Game>('/api/v1/library/game', {
      method: 'POST',
      body: JSON.stringify(game),
    });
  }

  async getGame(gameId: string): Promise<Game> {
    return this.request<Game>(`/api/v1/library/game/${gameId}`);
  }

  async updateGame(gameId: string, game: CreateGameRequest): Promise<void> {
    return this.request<void>(`/api/v1/library/game/${gameId}`, {
      method: 'PUT',
      body: JSON.stringify(game),
    });
  }

  async deleteGame(gameId: string): Promise<void> {
    return this.request<void>(`/api/v1/library/game/${gameId}`, {
      method: 'DELETE',
    });
  }

  // Patrons
  async listPatrons(): Promise<components["schemas"]["PatronList"]> {
    return this.request<components["schemas"]["PatronList"]>('/api/v1/library/patrons');
  }

  async addPatron(patron: CreatePatronRequest): Promise<Patron> {
    return this.request<Patron>('/api/v1/library/patron', {
      method: 'POST',
      body: JSON.stringify(patron),
    });
  }

  async getPatron(patronId: string): Promise<Patron> {
    return this.request<Patron>(`/api/v1/library/patron/${patronId}`);
  }

  async updatePatron(patronId: string, patron: CreatePatronRequest): Promise<void> {
    return this.request<void>(`/api/v1/library/patron/${patronId}`, {
      method: 'PUT',
      body: JSON.stringify(patron),
    });
  }

  async deletePatron(patronId: string): Promise<void> {
    return this.request<void>(`/api/v1/library/patron/${patronId}`, {
      method: 'DELETE',
    });
  }

  // Transactions
  async checkOutGame(request: CheckOutRequest): Promise<LibraryTransaction> {
    return this.request<LibraryTransaction>('/api/v1/library/checkout', {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }

  async checkInGame(transactionId: string): Promise<void> {
    const path = new URL('/api/v1/library/checkin', this.baseUrl);
    path.searchParams.append('transactionId', transactionId);
    return this.request<void>(path.pathname + path.search, {
      method: 'POST',
    });
  }

  // Health
  async health(): Promise<void> {
    return this.request<void>('/health');
  }
}

export const apiClient = new ApiClient();

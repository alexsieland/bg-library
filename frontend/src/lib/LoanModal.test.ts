import { render, screen, waitFor, fireEvent } from '@testing-library/svelte';
import LoanModal from './LoanModal.svelte';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { apiClient } from './api-client';

vi.mock('./config', () => ({
  getBackendUrl: () => 'http://localhost:8080'
}));

// Mock apiClient
vi.mock('./api-client', async (importOriginal) => {
  const actual = await importOriginal<any>();
  return {
    ...actual,
    apiClient: {
      listPatrons: vi.fn(),
      addPatron: vi.fn(),
      checkOutGame: vi.fn(),
    }
  };
});

const mockGame = { gameId: 'g1', title: 'Catan' };
const mockPatronData = {
  patrons: [
    { patronId: 'p1', name: 'Alice' },
    { patronId: 'p2', name: 'Bob' }
  ]
};

describe('LoanModal', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, 'log').mockImplementation(() => {});
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  it('Should search patrons when typing 3 or more characters', async () => {
    vi.mocked(apiClient.listPatrons).mockImplementation(async (params) => {
      if (params?.name === 'Ali') {
        return { patrons: [{ patronId: 'p1', name: 'Alice' }] };
      }
      return { patrons: [] };
    });

    render(LoanModal, { open: true, game: mockGame });

    const input = screen.getByPlaceholderText('Enter patron name');
    await fireEvent.input(input, { target: { value: 'Ali' } });

    // Wait for debounce (2000ms in component)
    await waitFor(() => {
      expect(apiClient.listPatrons).toHaveBeenCalledWith({ name: 'Ali' });
    }, { timeout: 3000 });

    await waitFor(() => {
      expect(screen.getByText('Alice')).toBeInTheDocument();
      // Bob should NOT be in the document because we now filter at the backend,
      // and we will mock the return value to ONLY include Alice if name 'Ali' is passed.
    });
  });

  it('Should limit search results to 5 patrons', async () => {
    const manyPatrons = {
      patrons: Array.from({ length: 10 }, (_, i) => ({ patronId: `${i}`, name: `Patron ${i}` }))
    };
    vi.mocked(apiClient.listPatrons).mockImplementation(async (params) => {
      if (params?.name === 'Patron') {
        return manyPatrons;
      }
      return { patrons: [] };
    });

    render(LoanModal, { open: true, game: mockGame });

    const input = screen.getByPlaceholderText('Enter patron name');
    await fireEvent.input(input, { target: { value: 'Patron' } });

    await waitFor(() => {
      expect(apiClient.listPatrons).toHaveBeenCalledWith({ name: 'Patron' });
      // After it's called, the list should be displayed.
      expect(screen.getByText('Patron 0')).toBeInTheDocument();
    }, { timeout: 3000 });

    const allButtons = screen.getAllByRole('button', { hidden: true });
    const patronButtons = allButtons.filter(b => b.textContent?.trim().startsWith('Patron'));
    expect(patronButtons.length).toBe(5);
  });

  it('Should checkout to existing patron when selected', async () => {
    vi.mocked(apiClient.listPatrons).mockResolvedValue(mockPatronData);
    vi.mocked(apiClient.checkOutGame).mockResolvedValue({} as any);

    const onLoanSuccess = vi.fn();
    render(LoanModal, { open: true, game: mockGame, onLoanSuccess });

    const input = screen.getByPlaceholderText('Enter patron name');
    await fireEvent.input(input, { target: { value: 'Ali' } });

    await waitFor(() => screen.getByText('Alice'), { timeout: 3000 });
    
    const aliceItem = screen.getByText('Alice');
    await fireEvent.click(aliceItem);

    const loanButton = screen.getByText('Loan');
    await fireEvent.click(loanButton);

    await waitFor(() => {
      expect(apiClient.checkOutGame).toHaveBeenCalledWith('g1', 'p1');
    });
    expect(onLoanSuccess).toHaveBeenCalled();
  });

  it('Should create new patron and checkout if not found', async () => {
    // 1. Initial search (returns nothing matching 'Charlie')
    vi.mocked(apiClient.listPatrons).mockImplementation(async (params) => {
      if (params?.name === 'Charlie') {
        return { patrons: [] };
      }
      return { patrons: [] };
    });
    // 2. Create patron
    vi.mocked(apiClient.addPatron).mockResolvedValue({ patronId: 'p-new', name: 'Charlie' });
    // 3. Checkout
    vi.mocked(apiClient.checkOutGame).mockResolvedValue({} as any);

    render(LoanModal, { open: true, game: mockGame });

    const input = screen.getByPlaceholderText('Enter patron name');
    await fireEvent.input(input, { target: { value: 'Charlie' } });

    const loanButton = screen.getByText('Loan');
    await fireEvent.click(loanButton);

    await waitFor(() => {
      // Verify listPatrons call with name
      expect(apiClient.listPatrons).toHaveBeenCalledWith({ name: 'Charlie' });
      // Verify Create Patron call
      expect(apiClient.addPatron).toHaveBeenCalledWith({ name: 'Charlie' });
      // Verify Checkout call
      expect(apiClient.checkOutGame).toHaveBeenCalledWith('g1', 'p-new');
    });
  });
});

import { render, screen, waitFor, fireEvent } from '@testing-library/svelte';
import LoanModal from './LoanModal.svelte';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('./config', () => ({
  getBackendUrl: () => 'http://localhost:8080'
}));

const mockGame = { gameId: 'g1', title: 'Catan' };
const mockPatrons = {
  patrons: [
    { patronId: 'p1', name: 'Alice' },
    { patronId: 'p2', name: 'Bob' }
  ]
};

describe('LoanModal', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn());
    vi.spyOn(console, 'log').mockImplementation(() => {});
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  it('Should search patrons when typing 3 or more characters', async () => {
    (fetch as any).mockResolvedValue({
      ok: true,
      json: async () => mockPatrons
    });

    render(LoanModal, { open: true, game: mockGame });

    const input = screen.getByPlaceholderText('Enter patron name');
    await fireEvent.input(input, { target: { value: 'Ali' } });

    // Wait for debounce (2000ms in component)
    await waitFor(() => {
      expect(fetch).toHaveBeenCalledWith(expect.stringContaining('/api/v1/library/patrons?name=Ali'));
    }, { timeout: 3000 });

    await waitFor(() => {
      expect(screen.getByText('Alice')).toBeInTheDocument();
      expect(screen.getByText('Bob')).toBeInTheDocument();
    });
  });

  it('Should limit search results to 5 patrons', async () => {
    const manyPatrons = {
      patrons: Array.from({ length: 10 }, (_, i) => ({ patronId: `${i}`, name: `Patron ${i}` }))
    };
    (fetch as any).mockResolvedValue({
      ok: true,
      json: async () => manyPatrons
    });

    render(LoanModal, { open: true, game: mockGame });

    const input = screen.getByPlaceholderText('Enter patron name');
    await fireEvent.input(input, { target: { value: 'Patron' } });

    await waitFor(() => {
      expect(fetch).toHaveBeenCalled();
    }, { timeout: 3000 });

    await waitFor(() => {
      const allButtons = screen.getAllByRole('button');
      const patronButtons = allButtons.filter(b => b.textContent?.includes('Patron'));
      expect(patronButtons.length).toBe(5);
    });
  });

  it('Should checkout to existing patron when selected', async () => {
    // 1. Search fetch
    (fetch as any).mockResolvedValueOnce({
      ok: true,
      json: async () => mockPatrons
    });
    // 2. Checkout fetch
    (fetch as any).mockResolvedValueOnce({
      ok: true,
      json: async () => ({ id: 't1' })
    });

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
      expect(fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/library/checkout'),
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify({ gameId: 'g1', patronId: 'p1' })
        })
      );
    });
    expect(onLoanSuccess).toHaveBeenCalled();
  });

  it('Should create new patron and checkout if not found', async () => {
    // 1. Search fetch (no exact match)
    (fetch as any).mockResolvedValueOnce({
      ok: true,
      json: async () => ({ patrons: [] })
    });
    // 2. Exact match check (no match)
    (fetch as any).mockResolvedValueOnce({
      ok: true,
      json: async () => ({ patrons: [] })
    });
    // 3. Create patron
    (fetch as any).mockResolvedValueOnce({
      ok: true,
      json: async () => ({ patronId: 'p-new', name: 'Charlie' })
    });
    // 4. Checkout
    (fetch as any).mockResolvedValueOnce({
      ok: true,
      json: async () => ({ id: 't-new' })
    });

    render(LoanModal, { open: true, game: mockGame });

    const input = screen.getByPlaceholderText('Enter patron name');
    await fireEvent.input(input, { target: { value: 'Charlie' } });

    const loanButton = screen.getByText('Loan');
    await fireEvent.click(loanButton);

    await waitFor(() => {
      // Verify Create Patron call
      expect(fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/library/patron'),
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify({ name: 'Charlie' })
        })
      );
      // Verify Checkout call
      expect(fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/library/checkout'),
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify({ gameId: 'g1', patronId: 'p-new' })
        })
      );
    });
  });
});

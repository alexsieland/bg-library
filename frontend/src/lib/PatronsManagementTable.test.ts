import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';
import PatronsManagementTable from './PatronsManagementTable.svelte';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { apiClient } from './api-client';
import { toasts } from './toast-store';
vi.mock('./config', () => ({
  getBackendUrl: () => 'http://localhost:8080',
  isBarcodeEnabled: () => false,
}));
vi.mock('./api-client', async (importOriginal) => {
  const actual = await importOriginal<any>();
  return {
    ...actual,
    apiClient: {
      listPatrons: vi.fn(),
      deletePatron: vi.fn(),
      addPatron: vi.fn(),
      updatePatron: vi.fn(),
      getPatron: vi.fn(),
      bulkAddPatrons: vi.fn(),
    },
  };
});
vi.mock('./toast-store', () => ({
  toasts: {
    add: vi.fn(),
  },
}));
const mockPatrons = {
  patrons: [
    { patronId: 'p1', name: 'Alice' },
    { patronId: 'p2', name: 'Bob' },
    { patronId: 'p3', name: 'Charlie' },
  ],
};
describe('PatronsManagementTable', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(apiClient.listPatrons).mockResolvedValue(mockPatrons);
    vi.mocked(apiClient.getPatron).mockImplementation(async (patronId: string) => {
      if (patronId === 'p1') {
        return { patronId: 'p1', name: 'Alice (latest)' };
      }
      return { patronId, name: 'Unknown Patron' };
    });
    vi.spyOn(console, 'error').mockImplementation(() => {});
    Object.defineProperty(document.documentElement, 'scrollHeight', {
      configurable: true,
      value: 2000,
    });
    Object.defineProperty(window, 'innerHeight', { configurable: true, value: 1000 });
  });
  // -------------------------------------------------------------------------
  // Initial render
  // -------------------------------------------------------------------------
  it('Should render the search input on mount', () => {
    render(PatronsManagementTable);
    expect(screen.getByPlaceholderText('Search patrons by name...')).toBeInTheDocument();
  });
  it('Should render the Add Patron button on mount', () => {
    render(PatronsManagementTable);
    expect(screen.getByRole('button', { name: /Add Patron/i })).toBeInTheDocument();
  });
  it('Should render the Actions button on mount', () => {
    render(PatronsManagementTable);
    expect(screen.getByRole('button', { name: /Actions/i })).toBeInTheDocument();
  });
  it('Should render the Patron Name and Action column headers', async () => {
    render(PatronsManagementTable);
    await waitFor(() => {
      expect(screen.getByText('Patron Name')).toBeInTheDocument();
      expect(screen.getByText('Action')).toBeInTheDocument();
    });
  });
  // -------------------------------------------------------------------------
  // Data loading
  // -------------------------------------------------------------------------
  it('Should show a loading indicator before patrons are fetched', () => {
    vi.mocked(apiClient.listPatrons).mockReturnValue(new Promise(() => {})); // Never resolves
    render(PatronsManagementTable);
    expect(screen.getByText('Loading patrons...')).toBeInTheDocument();
  });
  it('Should call listPatrons on mount', async () => {
    render(PatronsManagementTable);
    await waitFor(() => {
      expect(apiClient.listPatrons).toHaveBeenCalledTimes(1);
    });
  });
  it('Should display all patrons returned from the API', async () => {
    render(PatronsManagementTable);
    await waitFor(() => {
      expect(screen.getByText('Alice')).toBeInTheDocument();
      expect(screen.getByText('Bob')).toBeInTheDocument();
      expect(screen.getByText('Charlie')).toBeInTheDocument();
    });
  });
  it('Should show "No patrons found." when the API returns an empty list', async () => {
    vi.mocked(apiClient.listPatrons).mockResolvedValue({ patrons: [] });
    render(PatronsManagementTable);
    await waitFor(() => {
      expect(screen.getByText('No patrons found.')).toBeInTheDocument();
    });
  });
  it('Should show an error message when listPatrons fails', async () => {
    const errorMessage = 'Network error';
    vi.mocked(apiClient.listPatrons).mockRejectedValue(new Error(errorMessage));
    render(PatronsManagementTable);
    await waitFor(() => {
      expect(screen.getByText(errorMessage)).toBeInTheDocument();
    });
  });
  it('Should show an error toast when listPatrons fails', async () => {
    const errorMessage = 'Network error';
    vi.mocked(apiClient.listPatrons).mockRejectedValue(new Error(errorMessage));
    render(PatronsManagementTable);
    await waitFor(() => {
      expect(toasts.add).toHaveBeenCalledWith(`Failed to load patrons: ${errorMessage}`, 'error');
    });
  });
  // -------------------------------------------------------------------------
  // Row actions
  // -------------------------------------------------------------------------
  it('Should render Edit and Delete buttons for each patron', async () => {
    render(PatronsManagementTable);
    await waitFor(() => {
      expect(screen.getByText('Alice')).toBeInTheDocument();
    });
    const editButtons = screen.getAllByRole('button', { name: 'Edit' });
    const deleteButtons = screen.getAllByRole('button', { name: 'Delete' });
    expect(editButtons).toHaveLength(mockPatrons.patrons.length);
    expect(deleteButtons).toHaveLength(mockPatrons.patrons.length);
  });
  it('Should open the Add Patron modal when the Add Patron button is clicked', async () => {
    render(PatronsManagementTable);
    await fireEvent.click(screen.getByRole('button', { name: /Add Patron/i }));
    await waitFor(() => {
      expect(screen.getByTestId('add-patron-name-input')).toBeInTheDocument();
    });
  });
  it('Should open the Edit modal pre-populated with the latest patron name when Edit is clicked', async () => {
    render(PatronsManagementTable);
    await waitFor(() => {
      expect(screen.getByText('Alice')).toBeInTheDocument();
    });

    const editButtons = screen.getAllByRole('button', { name: 'Edit' });
    await fireEvent.click(editButtons[0]);

    await waitFor(() => {
      expect(apiClient.getPatron).toHaveBeenCalledWith('p1');
      const input = screen.getByTestId('add-patron-name-input') as HTMLInputElement;
      expect(input.value).toBe('Alice (latest)');
    });
  });
  it('Should open the delete confirmation when the Delete button is clicked', async () => {
    render(PatronsManagementTable);
    await waitFor(() => {
      expect(screen.getByText('Alice')).toBeInTheDocument();
    });
    const deleteButtons = screen.getAllByRole('button', { name: 'Delete' });
    await fireEvent.click(deleteButtons[0]);
    await waitFor(() => {
      expect(screen.getByTestId('delete-confirmation-modal')).toBeInTheDocument();
    });
  });
  // -------------------------------------------------------------------------
  // Search
  // -------------------------------------------------------------------------
  it('Should call listPatrons with the search query when a search is performed', async () => {
    vi.mocked(apiClient.listPatrons).mockImplementation(async (params) => {
      if (params?.name?.toLowerCase().includes('alice')) {
        return { patrons: [{ patronId: 'p1', name: 'Alice' }] };
      }
      return mockPatrons;
    });
    render(PatronsManagementTable);
    await waitFor(() => {
      expect(screen.getByText('Alice')).toBeInTheDocument();
    });
    const searchInput = screen.getByPlaceholderText('Search patrons by name...');
    await fireEvent.input(searchInput, { target: { value: 'alice' } });
    // Fire Enter to immediately trigger the search (bypasses debounce).
    await fireEvent.keyDown(searchInput, { key: 'Enter' });
    await waitFor(() => {
      expect(apiClient.listPatrons).toHaveBeenCalledWith(expect.objectContaining({ name: 'alice' }));
    });
  });
  it('Should show "No patrons found." when a search returns no results', async () => {
    vi.mocked(apiClient.listPatrons).mockImplementation(async (params) => {
      if (params?.name) {
        return { patrons: [] };
      }
      return mockPatrons;
    });
    render(PatronsManagementTable);
    const searchInput = screen.getByPlaceholderText('Search patrons by name...');
    await fireEvent.input(searchInput, { target: { value: 'zzz' } });
    await fireEvent.keyDown(searchInput, { key: 'Enter' });
    await waitFor(() => {
      expect(screen.getByText('No patrons found.')).toBeInTheDocument();
    });
  });
  // -------------------------------------------------------------------------
  // Add Patron modal wiring
  // -------------------------------------------------------------------------
  it('Should show a success toast and refresh the list when a patron is successfully added', async () => {
    vi.mocked(apiClient.addPatron).mockResolvedValue({ patronId: 'p4', name: 'Diana' });
    render(PatronsManagementTable);
    await fireEvent.click(screen.getByRole('button', { name: /Add Patron/i }));
    const nameInput = screen.getByTestId('add-patron-name-input');
    await fireEvent.input(nameInput, { target: { value: 'Diana' } });
    await fireEvent.click(screen.getByTestId('add-patron-submit'));
    await waitFor(() => {
      expect(toasts.add).toHaveBeenCalledWith('Patron saved successfully', 'success');
      expect(apiClient.listPatrons).toHaveBeenCalledTimes(2);
    });
  });
  // -------------------------------------------------------------------------
  // Infinite scroll
  // -------------------------------------------------------------------------
  it('Should load additional patrons when page scroll reaches bottom', async () => {
    const page1 = { patrons: Array.from({ length: 100 }, (_, i) => ({ patronId: `p${i + 1}`, name: `Patron ${i + 1}` })) };
    const page2 = { patrons: [{ patronId: 'p101', name: 'Patron 101' }] };

    vi.mocked(apiClient.listPatrons)
      .mockResolvedValueOnce(page1)
      .mockResolvedValueOnce(page2);

    render(PatronsManagementTable);

    await waitFor(() => {
      expect(apiClient.listPatrons).toHaveBeenCalledWith(
        expect.objectContaining({ limit: 100, offset: 0 })
      );
    });

    Object.defineProperty(window, 'scrollY', { configurable: true, value: 950 });
    await fireEvent.scroll(window);

    await waitFor(() => {
      expect(apiClient.listPatrons).toHaveBeenCalledWith(
        expect.objectContaining({ limit: 100, offset: 100 })
      );
    });
  });

  it('Should not fetch more patrons when result is less than PAGE_LIMIT', async () => {
    render(PatronsManagementTable);

    await waitFor(() => expect(apiClient.listPatrons).toHaveBeenCalledTimes(1));

    Object.defineProperty(window, 'scrollY', { configurable: true, value: 950 });
    await fireEvent.scroll(window);

    await waitFor(() => expect(apiClient.listPatrons).toHaveBeenCalledTimes(1));
  });

  it('Should reset offset and patrons on new search', async () => {
    vi.mocked(apiClient.listPatrons)
      .mockResolvedValueOnce(mockPatrons)
      .mockResolvedValueOnce({ patrons: [{ patronId: 'p1', name: 'Alice' }] });

    render(PatronsManagementTable);

    await waitFor(() => expect(apiClient.listPatrons).toHaveBeenCalledTimes(1));

    const input = screen.getByPlaceholderText('Search patrons by name...');
    await fireEvent.input(input, { target: { value: 'alice' } });
    await fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() => {
      expect(apiClient.listPatrons).toHaveBeenCalledWith(
        expect.objectContaining({ name: 'alice', offset: 0 })
      );
    });
  });
});

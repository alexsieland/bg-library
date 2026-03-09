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
      getPatron: vi.fn(),
      updatePatron: vi.fn(),
      addPatron: vi.fn(),
      getPatronByBarcode: vi.fn(),
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
    {
      patronId: 'p1',
      name: 'Alice Smith',
    },
    {
      patronId: 'p2',
      name: 'Bob Johnson',
    },
    {
      patronId: 'p3',
      name: 'Charlie Brown',
    },
  ],
};

describe('PatronsManagementTable', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  it('Should render the patrons table with search input', () => {
    vi.mocked(apiClient.listPatrons).mockResolvedValue(mockPatrons);

    render(PatronsManagementTable);

    expect(screen.getByPlaceholderText('Search patrons by name...')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Add Patron' })).toBeInTheDocument();
  });

  it('Should load and display patrons on mount', async () => {
    vi.mocked(apiClient.listPatrons).mockResolvedValue(mockPatrons);

    render(PatronsManagementTable);

    await waitFor(() => {
      expect(screen.getByText('Alice Smith')).toBeInTheDocument();
      expect(screen.getByText('Bob Johnson')).toBeInTheDocument();
      expect(screen.getByText('Charlie Brown')).toBeInTheDocument();
    });
  });

  it('Should filter patrons by search term', async () => {
    vi.mocked(apiClient.listPatrons).mockImplementation(async (params) => {
      if (params?.name?.toLowerCase().includes('alice')) {
        return {
          patrons: [mockPatrons.patrons[0]],
        };
      }
      return mockPatrons;
    });

    render(PatronsManagementTable);

    await waitFor(() => {
      expect(screen.getByText('Alice Smith')).toBeInTheDocument();
    });

    const searchInput = screen.getByPlaceholderText('Search patrons by name...');
    await fireEvent.input(searchInput, { target: { value: 'alice' } });

    await waitFor(() => {
      expect(screen.queryByText('Bob Johnson')).not.toBeInTheDocument();
      expect(screen.getByText('Alice Smith')).toBeInTheDocument();
    });
  });

  it('Should show message when no patrons match search', async () => {
    vi.mocked(apiClient.listPatrons).mockImplementation(async (params) => {
      if (params?.name) {
        return { patrons: [] };
      }
      return mockPatrons;
    });

    render(PatronsManagementTable);

    const searchInput = screen.getByPlaceholderText('Search patrons by name...');
    await fireEvent.input(searchInput, { target: { value: 'nonexistent' } });

    await waitFor(() => {
      expect(screen.getByText('No patrons found.')).toBeInTheDocument();
    });
  });

  it('Should show loading state initially', () => {
    vi.mocked(apiClient.listPatrons).mockReturnValue(new Promise(() => {})); // Never resolves

    render(PatronsManagementTable);

    expect(screen.getByText('Loading patrons...')).toBeInTheDocument();
  });

  it('Should show error message when fetch fails', async () => {
    const errorMessage = 'Failed to fetch patrons';
    vi.mocked(apiClient.listPatrons).mockRejectedValue(new Error(errorMessage));

    render(PatronsManagementTable);

    await waitFor(() => {
      expect(screen.getByText(errorMessage)).toBeInTheDocument();
    });

    expect(toasts.add).toHaveBeenCalledWith(`Failed to load patrons: ${errorMessage}`, 'error');
  });

  it('Should open AddPatronModal when Add Patron button is clicked', async () => {
    vi.mocked(apiClient.listPatrons).mockResolvedValue(mockPatrons);

    render(PatronsManagementTable);

    const addButton = screen.getByRole('button', { name: 'Add Patron' });
    await fireEvent.click(addButton);

    await waitFor(() => {
      expect(screen.getByPlaceholderText('Enter patron name')).toBeInTheDocument();
    });
  });

  it('Should open AddPatronModal in edit mode when Edit button is clicked', async () => {
    vi.mocked(apiClient.listPatrons).mockResolvedValue(mockPatrons);

    render(PatronsManagementTable);

    await waitFor(() => {
      expect(screen.getByText('Alice Smith')).toBeInTheDocument();
    });

    const editButtons = screen.getAllByRole('button', { name: 'Edit' });
    await fireEvent.click(editButtons[0]);

    await waitFor(() => {
      expect(screen.getByPlaceholderText('Enter patron name')).toBeInTheDocument();
    });
  });

  it('Should open delete confirmation when Delete button is clicked', async () => {
    vi.mocked(apiClient.listPatrons).mockResolvedValue(mockPatrons);

    render(PatronsManagementTable);

    await waitFor(() => {
      expect(screen.getByText('Alice Smith')).toBeInTheDocument();
    });

    const deleteButtons = screen.getAllByRole('button', { name: 'Delete' });
    await fireEvent.click(deleteButtons[0]);

    await waitFor(() => {
      expect(screen.getByText(/Are you sure you want to delete/)).toBeInTheDocument();
    });
  });

  it('Should delete patron when confirmed', async () => {
    vi.mocked(apiClient.listPatrons).mockResolvedValueOnce(mockPatrons);
    vi.mocked(apiClient.listPatrons).mockResolvedValueOnce({
      patrons: [mockPatrons.patrons[1], mockPatrons.patrons[2]], // Alice removed
    });
    vi.mocked(apiClient.deletePatron).mockResolvedValue(undefined);

    render(PatronsManagementTable);

    await waitFor(() => {
      expect(screen.getByText('Alice Smith')).toBeInTheDocument();
    });

    const deleteButtons = screen.getAllByRole('button', { name: 'Delete' });
    await fireEvent.click(deleteButtons[0]);

    await waitFor(() => {
      expect(screen.getByText(/Are you sure you want to delete/)).toBeInTheDocument();
    });

    // Click the confirm button rendered by the real DeleteConfirmationPrompt modal
    const confirmButton = screen.getByText("Yes, I'm sure");
    await fireEvent.click(confirmButton);

    await waitFor(() => {
      expect(apiClient.deletePatron).toHaveBeenCalledWith('p1');
      expect(toasts.add).toHaveBeenCalledWith('Deleted Alice Smith from the library', 'success');
    });
  });

  it('Should show error when delete fails', async () => {
    vi.mocked(apiClient.listPatrons).mockResolvedValue(mockPatrons);
    const deleteError = new Error('Delete failed');
    vi.mocked(apiClient.deletePatron).mockRejectedValue(deleteError);

    render(PatronsManagementTable);

    await waitFor(() => {
      expect(screen.getByText('Alice Smith')).toBeInTheDocument();
    });

    const deleteButtons = screen.getAllByRole('button', { name: 'Delete' });
    await fireEvent.click(deleteButtons[0]);

    await waitFor(() => {
      expect(screen.getByText(/Are you sure you want to delete/)).toBeInTheDocument();
    });

    const confirmButton = screen.getByText("Yes, I'm sure");
    await fireEvent.click(confirmButton);

    await waitFor(() => {
      expect(toasts.add).toHaveBeenCalledWith('Failed to delete patron: Delete failed', 'error');
    });
  });

  it('Should render edit and delete buttons for each patron', async () => {
    vi.mocked(apiClient.listPatrons).mockResolvedValue(mockPatrons);

    render(PatronsManagementTable);

    await waitFor(() => {
      expect(screen.getByText('Alice Smith')).toBeInTheDocument();
    });

    const editButtons = screen.getAllByRole('button', { name: 'Edit' });
    const deleteButtons = screen.getAllByRole('button', { name: 'Delete' });

    expect(editButtons.length).toBe(3); // One for each patron
    expect(deleteButtons.length).toBe(3);
  });

  it('Should show no patrons found when list is empty', async () => {
    vi.mocked(apiClient.listPatrons).mockResolvedValue({ patrons: [] });

    render(PatronsManagementTable);

    await waitFor(() => {
      expect(screen.getByText('No patrons found.')).toBeInTheDocument();
    });
  });

  it('Should have an Actions dropdown with Bulk Add option', async () => {
    vi.mocked(apiClient.listPatrons).mockResolvedValue(mockPatrons);

    render(PatronsManagementTable);

    const actionsButton = screen.getByRole('button', { name: /Actions/ });
    expect(actionsButton).toBeInTheDocument();

    // Flowbite renders dropdown items in a hidden div until opened;
    // confirm the item exists in the DOM tree
    expect(screen.getByText('Bulk Add', { hidden: true })).toBeInTheDocument();
  });
});

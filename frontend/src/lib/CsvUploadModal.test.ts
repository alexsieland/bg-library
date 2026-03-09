import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';
import CsvUploadModal from './CsvUploadModal.svelte';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { toasts } from './toast-store';

vi.mock('./toast-store', () => ({
  toasts: {
    add: vi.fn(),
  },
}));

describe('CsvUploadModal', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  it('Should render the modal with title and file input', () => {
    const onUpload = vi.fn();
    render(CsvUploadModal, {
      open: true,
      title: 'Upload Games',
      onUpload,
    });

    expect(screen.getByText('Upload Games')).toBeInTheDocument();
    expect(screen.getByText(/Upload a CSV file/)).toBeInTheDocument();
  });

  it('Should disable upload button when no file is selected', () => {
    const onUpload = vi.fn();
    render(CsvUploadModal, {
      open: true,
      onUpload,
    });

    const uploadButton = screen.getByText('Upload');
    expect(uploadButton).toBeDisabled();
  });

  it('Should enable upload button when file is selected', async () => {
    const onUpload = vi.fn();
    render(CsvUploadModal, {
      open: true,
      onUpload,
    });

    const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement;
    const file = new File(['game1\ngame2'], 'games.csv', { type: 'text/csv' });

    await fireEvent.change(fileInput, { target: { files: [file] } });

    const uploadButton = screen.getByText('Upload');
    expect(uploadButton).not.toBeDisabled();
  });

  it('Should show success message when upload succeeds', async () => {
    const onUpload = vi.fn().mockResolvedValue({ imported: 3 });
    render(CsvUploadModal, {
      open: true,
      title: 'Upload Games',
      successMessage: (count) => `Imported ${count} games`,
      onUpload,
    });

    const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement;
    const file = new File(['game1\ngame2\ngame3'], 'games.csv', { type: 'text/csv' });

    await fireEvent.change(fileInput, { target: { files: [file] } });

    const uploadButton = screen.getByText('Upload');
    await fireEvent.click(uploadButton);

    await waitFor(() => {
      expect(toasts.add).toHaveBeenCalledWith('Imported 3 games', 'success');
    });
  });

  it('Should call onSuccess callback after successful upload', async () => {
    const onUpload = vi.fn().mockResolvedValue({ imported: 5 });
    const onSuccess = vi.fn();

    render(CsvUploadModal, {
      open: true,
      onUpload,
      onSuccess,
    });

    const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement;
    const file = new File(['data'], 'test.csv', { type: 'text/csv' });

    await fireEvent.change(fileInput, { target: { files: [file] } });

    const uploadButton = screen.getByText('Upload');
    await fireEvent.click(uploadButton);

    await waitFor(() => {
      expect(onSuccess).toHaveBeenCalledWith(5);
    });
  });

  it('Should show error message when upload fails', async () => {
    const onUpload = vi.fn().mockRejectedValue(new Error('File format invalid'));
    render(CsvUploadModal, {
      open: true,
      onUpload,
    });

    const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement;
    const file = new File(['invalid data'], 'test.csv', { type: 'text/csv' });

    await fireEvent.change(fileInput, { target: { files: [file] } });

    const uploadButton = screen.getByText('Upload');
    await fireEvent.click(uploadButton);

    await waitFor(() => {
      expect(toasts.add).toHaveBeenCalledWith('Failed to upload: File format invalid', 'error');
    });
  });

  it('Should disable buttons while loading', async () => {
    const onUpload = vi.fn().mockReturnValue(new Promise(() => {})); // Never resolves
    render(CsvUploadModal, {
      open: true,
      onUpload,
    });

    const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement;
    const file = new File(['data'], 'test.csv', { type: 'text/csv' });

    await fireEvent.change(fileInput, { target: { files: [file] } });

    const uploadButton = screen.getByText('Upload');
    const cancelButton = screen.getByText('Cancel');

    expect(uploadButton).not.toBeDisabled();
    expect(cancelButton).not.toBeDisabled();

    await fireEvent.click(uploadButton);

    await waitFor(() => {
      expect(uploadButton).toBeDisabled();
      expect(cancelButton).toBeDisabled();
    });
  });

  it('Should call onCancel when cancel button is clicked', async () => {
    const onUpload = vi.fn();
    const onCancel = vi.fn();

    render(CsvUploadModal, {
      open: true,
      onUpload,
      onCancel,
    });

    const cancelButton = screen.getByText('Cancel');
    await fireEvent.click(cancelButton);

    await waitFor(() => {
      expect(onCancel).toHaveBeenCalled();
    });
  });

  it('Should accept only CSV and text files', () => {
    const onUpload = vi.fn();
    render(CsvUploadModal, {
      open: true,
      onUpload,
    });

    const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement;
    expect(fileInput.accept).toBe('.csv,text/csv,text/plain');
  });

  it('Should show singular vs plural in default success message', async () => {
    const onUpload = vi.fn().mockResolvedValue({ imported: 1 });
    render(CsvUploadModal, {
      open: true,
      onUpload,
    });

    const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement;
    const file = new File(['game'], 'test.csv', { type: 'text/csv' });

    await fireEvent.change(fileInput, { target: { files: [file] } });

    const uploadButton = screen.getByText('Upload');
    await fireEvent.click(uploadButton);

    await waitFor(() => {
      expect(toasts.add).toHaveBeenCalledWith('Successfully imported 1 item', 'success');
    });
  });
});

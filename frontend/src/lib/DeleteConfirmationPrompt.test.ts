import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';
import DeleteConfirmationPrompt from './DeleteConfirmationPrompt.svelte';
import { describe, it, expect, vi, beforeEach } from 'vitest';

describe('DeleteConfirmationPrompt', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  it('Should render the prompt with delete confirmation message', () => {
    const onConfirm = vi.fn();
    render(DeleteConfirmationPrompt, {
      open: true,
      itemName: 'Catan',
      onConfirm,
    });

    expect(screen.getByText(/Are you sure you want to delete/)).toBeInTheDocument();
    expect(screen.getAllByText('Catan')[0]).toBeInTheDocument();
  });

  it('Should render buttons for cancellation and confirmation', () => {
    const onConfirm = vi.fn();
    render(DeleteConfirmationPrompt, {
      open: true,
      itemName: 'Test Item',
      onConfirm,
    });

    expect(screen.getByText('No, cancel')).toBeInTheDocument();
    expect(screen.getByText("Yes, I'm sure")).toBeInTheDocument();
  });

  it('Should display custom item name in confirmation message', () => {
    const onConfirm = vi.fn();
    render(DeleteConfirmationPrompt, {
      open: true,
      itemName: 'My Custom Game',
      onConfirm,
    });

    expect(screen.getByText('My Custom Game')).toBeInTheDocument();
  });

  it('Should use default item name if not provided', () => {
    const onConfirm = vi.fn();
    render(DeleteConfirmationPrompt, {
      open: true,
      onConfirm,
    });

    expect(screen.getByText('this item')).toBeInTheDocument();
  });

  it('Should handle async onConfirm callback', async () => {
    const onConfirm = vi.fn().mockResolvedValue(undefined);

    render(DeleteConfirmationPrompt, {
      open: true,
      itemName: 'Test Item',
      onConfirm,
    });

    const confirmButton = screen.getByText("Yes, I'm sure");
    await fireEvent.click(confirmButton);

    await waitFor(() => {
      expect(onConfirm).toHaveBeenCalled();
    });
  });
});

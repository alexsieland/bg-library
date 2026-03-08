import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/svelte';
import { toasts } from './toast-store';
import ToastContainer from './ToastContainer.svelte';
import { get } from 'svelte/store';

describe('Toast System', () => {
  beforeEach(() => {
    // Clear all toasts before each test
    toasts.subscribe((t) => {
      t.forEach((toast) => toasts.remove(toast.id));
    })();
    vi.useFakeTimers();
  });

  it('should add and remove toasts via store', () => {
    const id = toasts.add('Test message', 'success');
    let currentToasts = get(toasts);
    expect(currentToasts).toHaveLength(1);
    expect(currentToasts[0].message).toBe('Test message');
    expect(currentToasts[0].type).toBe('success');

    toasts.remove(id);
    expect(get(toasts)).toHaveLength(0);
  });

  it('should auto-close success toasts after 3 seconds by default', () => {
    toasts.add('Auto close', 'success');
    expect(get(toasts)).toHaveLength(1);

    vi.advanceTimersByTime(3000);
    expect(get(toasts)).toHaveLength(0);
  });

  it('should NOT auto-close error toasts by default', () => {
    toasts.add('Manual close', 'error');
    expect(get(toasts)).toHaveLength(1);

    vi.advanceTimersByTime(10000);
    expect(get(toasts)).toHaveLength(1);
  });

  it('should auto-close warn toasts after 10 seconds by default', () => {
    toasts.add('Warning message', 'warn');
    expect(get(toasts)).toHaveLength(1);

    vi.advanceTimersByTime(9999);
    expect(get(toasts)).toHaveLength(1);

    vi.advanceTimersByTime(1);
    expect(get(toasts)).toHaveLength(0);
  });

  it('should NOT auto-close warn toasts before 10 seconds', () => {
    toasts.add('Warning message', 'warn');
    expect(get(toasts)).toHaveLength(1);

    vi.advanceTimersByTime(3000);
    expect(get(toasts)).toHaveLength(1);
  });

  it('should render warn toasts with yellow styling', async () => {
    render(ToastContainer);

    toasts.add('Watch out', 'warn');

    await waitFor(() => {
      expect(screen.getByText('Watch out')).toBeInTheDocument();
      const alert = screen.getByRole('alert');
      expect(alert.className).toContain('bg-yellow-500');
    });
  });

  it('should render toasts in ToastContainer', async () => {
    render(ToastContainer);

    toasts.add('Visible toast', 'success');

    await waitFor(() => {
      expect(screen.getByText('Visible toast')).toBeInTheDocument();
    });
  });

  it('should allow manual closing of toasts', async () => {
    render(ToastContainer);

    toasts.add('Closable', 'error');

    let closeButton: HTMLElement;
    await waitFor(() => {
      closeButton = screen.getByRole('button');
    });

    await fireEvent.click(closeButton!);

    await waitFor(() => {
      expect(get(toasts)).toHaveLength(0);
    });
  });
});

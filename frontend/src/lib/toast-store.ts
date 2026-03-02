import { writable } from 'svelte/store';

export type ToastType = 'success' | 'error';

export interface Toast {
  id: string;
  message: string;
  type: ToastType;
  dismissible: boolean;
  timeout?: number;
}

const { subscribe, update } = writable<Toast[]>([]);

export const toasts = {
  subscribe,
  add: (message: string, type: ToastType = 'success', timeout?: number) => {
    const id = crypto.randomUUID();
    const dismissible = true;
    
    // Default success toasts to 3 seconds auto-close if not specified
    if (type === 'success' && timeout === undefined) {
      timeout = 3000;
    }

    const toast: Toast = { id, message, type, dismissible, timeout };
    update(all => [...all, toast]);

    if (timeout) {
      setTimeout(() => {
        toasts.remove(id);
      }, timeout);
    }

    return id;
  },
  remove: (id: string) => {
    update(all => all.filter(t => t.id !== id));
  }
};

import { writable } from 'svelte/store';

export type ToastType = 'success' | 'error' | 'warn';

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

    // Default timeouts by type
    if (type === 'success' && timeout === undefined) {
      timeout = 3000;
    } else if (type === 'warn' && timeout === undefined) {
      timeout = 10000;
    }

    const toast: Toast = { id, message, type, dismissible, timeout };
    update((all) => [...all, toast]);

    if (timeout) {
      setTimeout(() => {
        toasts.remove(id);
      }, timeout);
    }

    return id;
  },
  remove: (id: string) => {
    update((all) => all.filter((t) => t.id !== id));
  },
};

<script lang="ts">
  import { toasts, type Toast } from './toast-store';
  import { CloseButton } from 'flowbite-svelte';
  import {
    CheckCircleSolid,
    CloseCircleSolid,
    ExclamationCircleSolid,
  } from 'flowbite-svelte-icons';
  import { fade, slide } from 'svelte/transition';

  export let toast: Toast;

  const close = () => toasts.remove(toast.id);
</script>

<div
  role="alert"
  in:slide={{ axis: 'y', duration: 300 }}
  out:fade={{ duration: 200 }}
  class="pointer-events-auto flex w-full max-w-full items-center justify-between p-4 shadow-lg
    {toast.type === 'success'
    ? 'bg-emerald-500 text-white'
    : toast.type === 'warn'
      ? 'bg-yellow-500 text-white'
      : 'bg-rose-500 text-white'}"
>
  <div class="flex items-center space-x-3">
    {#if toast.type === 'success'}
      <CheckCircleSolid class="h-6 w-6" />
    {:else if toast.type === 'warn'}
      <ExclamationCircleSolid class="h-6 w-6" />
    {:else}
      <CloseCircleSolid class="h-6 w-6" />
    {/if}
    <span class="text-lg font-medium">{toast.message}</span>
  </div>

  {#if toast.dismissible}
    <CloseButton
      color="none"
      class="text-white/80 hover:bg-white/10 hover:text-white"
      onclick={close}
    />
  {/if}
</div>

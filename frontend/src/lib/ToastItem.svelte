<script lang="ts">
  import { toasts, type Toast } from './toast-store';
  import { CloseButton } from 'flowbite-svelte';
  import { CheckCircleSolid, CloseCircleSolid, ExclamationCircleSolid } from 'flowbite-svelte-icons';
  import { fade, slide } from 'svelte/transition';

  export let toast: Toast;

  const close = () => toasts.remove(toast.id);
</script>

<div
  role="alert"
  in:slide={{ axis: 'y', duration: 300 }}
  out:fade={{ duration: 200 }}
  class="w-full max-w-full p-4 flex items-center justify-between shadow-lg pointer-events-auto
    {toast.type === 'success' ? 'bg-emerald-500 text-white' : toast.type === 'warn' ? 'bg-yellow-500 text-white' : 'bg-rose-500 text-white'}"
>
  <div class="flex items-center space-x-3">
    {#if toast.type === 'success'}
      <CheckCircleSolid class="w-6 h-6" />
    {:else if toast.type === 'warn'}
      <ExclamationCircleSolid class="w-6 h-6" />
    {:else}
      <CloseCircleSolid class="w-6 h-6" />
    {/if}
    <span class="font-medium text-lg">{toast.message}</span>
  </div>
  
  {#if toast.dismissible}
    <CloseButton 
      color="none" 
      class="text-white/80 hover:text-white hover:bg-white/10" 
      onclick={close} 
    />
  {/if}
</div>

<script lang="ts">
  import { toasts } from './toast-store';
  import ToastItem from './ToastItem.svelte';
</script>

<div 
  class="fixed bottom-0 left-0 right-0 z-[100] flex flex-col pointer-events-none overflow-hidden"
  aria-live="polite"
>
  {#each $toasts as toast (toast.id)}
    <ToastItem {toast} />
  {/each}
</div>

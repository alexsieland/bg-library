<script lang="ts">
  import { toasts } from './toast-store';
  import ToastItem from './ToastItem.svelte';
</script>

<div
  class="pointer-events-none fixed right-0 bottom-0 left-0 z-[100] flex flex-col overflow-hidden"
  aria-live="polite"
>
  {#each $toasts as toast (toast.id)}
    <ToastItem {toast} />
  {/each}
</div>

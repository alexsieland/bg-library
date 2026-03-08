<script lang="ts">
  import { Spinner } from 'flowbite-svelte';
  import { apiClient } from './api-client';
  import type { components } from '../generated/library-api';

  type Game = components['schemas']['Game'];

  export let onGameFound: (game: Game) => void = () => {};
  export let onError: (message: string) => void = () => {};
  export let barcodeInputElement: HTMLInputElement | undefined;

  let barcode = '';
  let loading = false;

  async function handleScan() {
    const value = barcode.trim();
    barcode = '';

    if (!value) return;

    loading = true;
    try {
      const result = await apiClient.getGameByBarcode(value);

      if (result.games.length > 1) {
        onError(
          'Barcode conflict handling not yet implemented. Please manually trigger the check out.'
        );
        return;
      }

      // getGameByBarcode returns 404 (throws) when empty, so result.games[0] is always defined here
      onGameFound(result.games[0]);
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to look up barcode';
      onError(message);
    } finally {
      loading = false;
    }
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter') {
      handleScan();
    }
  }
</script>

<div class="flex items-center gap-2">
  <span
    class="text-xs font-medium tracking-wide whitespace-nowrap text-slate-400 uppercase select-none dark:text-slate-500"
  >
    Barcode
  </span>
  <div class="relative">
    <input
      bind:this={barcodeInputElement}
      id="barcode-input"
      type="text"
      bind:value={barcode}
      onkeydown={handleKeydown}
      placeholder="Scan…"
      aria-label="Barcode Scanner"
      autocomplete="off"
      disabled={loading}
      class="w-36 rounded-lg border border-slate-200 bg-white
             px-3 py-2
             text-sm text-slate-500 placeholder:text-slate-300
             focus:border-slate-400 focus:ring-1
             focus:ring-slate-300 focus:outline-none
             disabled:opacity-50 dark:border-slate-600
             dark:bg-slate-800 dark:text-slate-400 dark:placeholder:text-slate-600 dark:focus:border-slate-500
             dark:focus:ring-slate-500"
    />
    {#if loading}
      <div class="pointer-events-none absolute inset-y-0 inset-e-0 flex items-center pe-2">
        <Spinner size="4" />
      </div>
    {/if}
  </div>
</div>

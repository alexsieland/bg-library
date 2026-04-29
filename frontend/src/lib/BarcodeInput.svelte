<script lang="ts">
  import { Spinner } from 'flowbite-svelte';
  import { apiClient } from './api-client';
  import type { GameStatus } from './api-client';

  export let onStatusesFound: (statuses: GameStatus[]) => void = () => {};
  export let onError: (message: string) => void = () => {};
  export let checkedOut: boolean = false;
  export let barcodeInputElement: HTMLInputElement | undefined;

  let barcode = '';
  let loading = false;

  async function handleScan() {
    const value = barcode.trim();
    barcode = '';

    if (!value) return;

    loading = true;
    try {
      const result = await apiClient.listGames({ barcode: value, checkedOut });

      if (result.games.length === 0) {
        if (checkedOut) {
          onError('No checked out games with this barcode.');
        } else {
          onError('No available games with this barcode.');
        }
        return;
      }

      onStatusesFound(result.games);
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
      data-testid="barcode-scanner-input"
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

<script lang="ts">
  import { Input, Label, Spinner } from 'flowbite-svelte';
  import { apiClient } from './api-client';
  import type { components } from '../generated/library-api';

  type Game = components["schemas"]["Game"];

  export let onGameFound: (game: Game) => void = () => {};
  export let onError: (message: string) => void = () => {};

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
        onError('Barcode conflict handling not yet implemented. Please manually trigger the check out.');
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

<div class="flex items-end gap-2">
  <div class="relative w-56">
    <Label for="barcode-input" class="mb-1 text-xs font-medium text-slate-400 dark:text-slate-500 uppercase tracking-wide">
      Barcode Scanner
    </Label>
    <Input
      id="barcode-input"
      bind:value={barcode}
      onkeydown={handleKeydown}
      placeholder="Scan barcode…"
      autocomplete="off"
      disabled={loading}
      class="w-full text-sm border-slate-300 dark:border-slate-600 text-slate-500 dark:text-slate-400 placeholder-slate-300 dark:placeholder-slate-600 focus:border-slate-400 focus:ring-slate-300"
    />
    {#if loading}
      <div class="absolute inset-y-0 inset-e-0 flex items-center pe-3 pointer-events-none top-6">
        <Spinner size="4" />
      </div>
    {/if}
    <p class="mt-1 text-xs text-slate-400 dark:text-slate-600">
      Connect a barcode scanner to use
    </p>
  </div>
</div>

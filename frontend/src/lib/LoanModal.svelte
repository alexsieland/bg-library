<script lang="ts">
  import { Modal, Button, Input, Label, Spinner } from 'flowbite-svelte';
  import { apiClient } from './api-client';
  import type { components } from '../generated/library-api';
  import Debounce from './snippets/debounce.svelte';
  import { toasts } from './toast-store';
  import { isBarcodeEnabled } from './config';

  export let open = false;
  export let game: components["schemas"]["Game"] | null = null;
  export let onLoanSuccess: () => void = () => {};

  let patronName = '';
  let patrons: components["schemas"]["Patron"][] = [];
  let loading = false;
  let loaning = false;
  let error: string | null = null;
  let lastValueRef = { v: '' };

  // This is used by the Debounce snippet to skip updates if we manually trigger a search
  let cancelKey = 0;

  let patronBarcode = '';
  let barcodeLoading = false;

  async function handlePatronBarcodeScan() {
    const value = patronBarcode.trim();
    patronBarcode = '';

    if (!value) return;

    barcodeLoading = true;
    try {
      const patron = await apiClient.getPatronByBarcode(value);
      selectPatron(patron);
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Patron not found';
      toasts.add(`Barcode scan failed: ${message}`, 'error');
    } finally {
      barcodeLoading = false;
    }
  }

  function handleBarcodeKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter') {
      handlePatronBarcodeScan();
    }
  }

  async function searchPatrons(name: string) {
    if (name.length < 3) {
      patrons = [];
      return;
    }
    loading = true;
    error = null;
    try {
      const data = await apiClient.listPatrons({ name });
      patrons = data.patrons.slice(0, 5);
    } catch (e) {
      console.error('Error searching patrons:', e);
      error = e instanceof Error ? e.message : 'Search failed';
    } finally {
      loading = false;
    }
  }

  async function handleLoan() {
    if (!game || !patronName.trim()) return;
    loaning = true;
    error = null;
    try {
      // 1. Check if patron exists
      let patron = patrons.find(p => p.name.toLowerCase() === patronName.trim().toLowerCase());
      
      if (!patron) {
        // If not in the current search results, try to find the exact patron from backend
        const searchData = await apiClient.listPatrons({ name: patronName.trim() });
        patron = searchData.patrons.find(p => p.name.toLowerCase() === patronName.trim().toLowerCase());
      }

      let patronId: string;
      if (!patron) {
        // 2. Add patron
        const newPatron = await apiClient.addPatron({ name: patronName.trim() });
        patronId = newPatron.patronId;
      } else {
        patronId = patron.patronId;
      }

      // 3. Initiate checkout
      await apiClient.checkOutGame(game.gameId, patronId);

      toasts.add(`Successfully loaned ${game.title} to ${patronName.trim()}`, 'success');
      onLoanSuccess();
      open = false;
      patronName = '';
      patrons = [];
    } catch (e) {
      console.error('Error during loan process:', e);
      const errorMessage = e instanceof Error ? e.message : 'Loan process failed';
      error = errorMessage;
      toasts.add(`Failed to loan game: ${errorMessage}`, 'error');
    } finally {
      loaning = false;
    }
  }

  function selectPatron(patron: components["schemas"]["Patron"]) {
    patronName = patron.name;
    patrons = [];
    lastValueRef.v = patronName;
    cancelKey++;
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter') {
      handleLoan();
    }
  }
</script>

<Modal bind:open title={`Loan Game: ${game?.title || ''}`} size="md" autoclose={false} class="w-full">
  <div class="space-y-4 min-h-[300px]">
    <div class="flex items-end space-x-2">
      <div class="flex-grow relative">
        <Label for="patronName" class="mb-2">Patron Name</Label>
        <div class="relative">
          <Input 
            id="patronName" 
            placeholder="Enter patron name" 
            bind:value={patronName} 
            onkeydown={handleKeydown}
            autocomplete="off"
            maxlength={100}
            class="w-full"
          />
          {#if loading}
            <div class="absolute inset-y-0 end-0 flex items-center pe-3 pointer-events-none">
              <Spinner size="4" />
            </div>
          {/if}
        </div>
        
        {#if patrons.length > 0}
          <ul class="absolute z-50 w-full mt-1 shadow-lg bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden">
            {#each patrons as patron}
              <li>
                <button 
                  type="button" 
                  class="w-full text-left px-4 py-2 hover:bg-gray-100 dark:hover:bg-gray-700"
                  onclick={() => selectPatron(patron)}
                >
                  {patron.name}
                </button>
              </li>
            {/each}
          </ul>
        {/if}
      </div>

      <Button onclick={handleLoan} disabled={loaning || !patronName.trim()} class="mb-0">
        {#if loaning}
          <Spinner size="4" class="me-2" />
        {/if}
        Loan
      </Button>
    </div>

    {#if isBarcodeEnabled()}
      <div class="flex justify-end">
        <div class="flex items-center gap-2">
          <span class="text-xs font-medium tracking-wide text-slate-400 dark:text-slate-500 uppercase whitespace-nowrap select-none">
            Barcode
          </span>
          <div class="relative">
            <input
              type="text"
              bind:value={patronBarcode}
              onkeydown={handleBarcodeKeydown}
              placeholder="Scan…"
              aria-label="Patron Barcode Scanner"
              autocomplete="off"
              disabled={barcodeLoading}
              class="w-36 rounded-lg border border-slate-200 dark:border-slate-600
                     bg-white dark:bg-slate-800
                     px-3 py-2 text-sm
                     text-slate-500 dark:text-slate-400
                     placeholder:text-slate-300 dark:placeholder:text-slate-600
                     focus:border-slate-400 dark:focus:border-slate-500
                     focus:outline-none focus:ring-1 focus:ring-slate-300 dark:focus:ring-slate-500
                     disabled:opacity-50"
            />
            {#if barcodeLoading}
              <div class="absolute inset-y-0 inset-e-0 flex items-center pe-2 pointer-events-none">
                <Spinner size="4" />
              </div>
            {/if}
          </div>
        </div>
      </div>
    {/if}

    {#if error}
      <p class="text-sm text-rose-500">{error}</p>
    {/if}
  </div>
</Modal>

<Debounce value={patronName} onTrigger={searchPatrons} delay={300} minLength={3} {lastValueRef} {cancelKey} />

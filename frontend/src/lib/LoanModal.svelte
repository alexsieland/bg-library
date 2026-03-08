<script lang="ts">
  import { Modal, Button, Input, Label, Spinner } from 'flowbite-svelte';
  import { UserSolid, UserAddSolid } from 'flowbite-svelte-icons';
  import { apiClient, type Patron } from './api-client';
  import type { components } from '../generated/library-api';
  import Debounce from './snippets/debounce.svelte';
  import { toasts } from './toast-store';
  import { isBarcodeEnabled } from './config';
  import AddPatronModal from './AddPatronModal.svelte';

  export let open = false;
  export let game: components['schemas']['Game'] | null = null;
  export let onLoanSuccess: () => void = () => {};

  let patronName = '';
  let patrons: Patron[] = [];
  let selectedPatron: Patron | null = null;
  let loading = false;
  let loaning = false;
  let error: string | null = null;
  let lastValueRef = { v: '' };

  // This is used by the Debounce snippet to skip updates if we manually trigger a search
  let cancelKey = 0;

  let patronBarcode = '';
  let barcodeLoading = false;
  let addPatronModalOpen = false;

  // Close AddPatronModal if the loan modal is closed programmatically
  $: if (!open) {
    addPatronModalOpen = false;
  }

  // Show "New Patron" button when 3+ chars typed and no patron selected
  $: showNewPatronButton = patronName.trim().length >= 3 && !selectedPatron;

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
      // Deduplicate by name (case-insensitive), keeping first occurrence, then slice to 5
      const seen = new Set<string>();
      const deduplicated = data.patrons.filter((p) => {
        const key = p.name.toLowerCase();
        if (seen.has(key)) return false;
        seen.add(key);
        return true;
      });
      patrons = deduplicated.slice(0, 5);
    } catch (e) {
      console.error('Error searching patrons:', e);
      error = e instanceof Error ? e.message : 'Search failed';
    } finally {
      loading = false;
    }
  }

  async function handleLoan() {
    if (!game || !selectedPatron) return;
    loaning = true;
    error = null;
    try {
      await apiClient.checkOutGame(game.gameId, selectedPatron.patronId);

      toasts.add(`Successfully loaned ${game.title} to ${selectedPatron.name}`, 'success');
      onLoanSuccess();
      open = false;
      patronName = '';
      patrons = [];
      selectedPatron = null;
    } catch (e) {
      console.error('Error during loan process:', e);
      const errorMessage = e instanceof Error ? e.message : 'Loan process failed';
      error = errorMessage;
      toasts.add(`Failed to loan game: ${errorMessage}`, 'error');
    } finally {
      loaning = false;
    }
  }

  function selectPatron(patron: Patron) {
    selectedPatron = patron;
    patronName = patron.name;
    patrons = [];
    lastValueRef.v = patronName;
    cancelKey++;
  }

  function handlePatronNameInput() {
    // If a patron was previously selected from the dropdown and the librarian
    // modifies the name input, deselect immediately — they are signalling new intent.
    if (selectedPatron) {
      selectedPatron = null;
    }
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter') {
      // Only trigger loan if a patron is already selected (fast-path for keyboard users).
      // When no patron is selected, Enter is suppressed to prevent accidental patron creation
      // via HID barcode scanner terminator keystrokes.
      if (selectedPatron) {
        handleLoan();
      }
      event.preventDefault();
    }
  }

  function openAddPatronModal() {
    addPatronModalOpen = true;
  }

  function handleNewPatronCreated(patron: Patron) {
    selectedPatron = patron;
    patronName = patron.name;
    patrons = [];
    lastValueRef.v = patron.name;
    cancelKey++;
    addPatronModalOpen = false;
  }
</script>

<Modal
  bind:open
  title={`Loan Game: ${game?.title || ''}`}
  size="md"
  autoclose={false}
  class="w-full"
>
  <div class="min-h-[300px] space-y-4">
    <div class="flex items-end space-x-2">
      <div class="relative flex-grow">
        <Label for="patronName" class="mb-2">Patron Name</Label>
        <div class="relative">
          <Input
            id="patronName"
            type="text"
            placeholder="Enter patron name"
            bind:value={patronName}
            onkeydown={handleKeydown}
            oninput={handlePatronNameInput}
            autocomplete="off"
            maxlength={100}
            class="ps-9"
          >
            {#snippet left()}
              {#if loading}
                <Spinner size="5" />
              {:else}
                <UserSolid class="h-5 w-5 shrink-0" />
              {/if}
            {/snippet}
            {#snippet right()}
              {#if showNewPatronButton}
                <Button size="xs" color="emerald" onclick={openAddPatronModal} class="gap-1">
                  <UserAddSolid class="h-5 w-5 shrink-0" /> New Patron
                </Button>
              {/if}
            {/snippet}
          </Input>
        </div>

        {#if patrons.length > 0}
          <ul
            class="absolute z-50 mt-1 w-full overflow-hidden rounded-lg border border-gray-200 bg-white shadow-lg dark:border-gray-700 dark:bg-gray-800"
          >
            {#each patrons as patron}
              <li>
                <button
                  type="button"
                  class="w-full px-4 py-2 text-left hover:bg-gray-100 dark:hover:bg-gray-700"
                  onclick={() => selectPatron(patron)}
                >
                  {patron.name}
                </button>
              </li>
            {/each}
          </ul>
        {/if}
      </div>

      <Button onclick={handleLoan} disabled={loaning || !selectedPatron} class="mb-0">
        {#if loaning}
          <Spinner size="4" class="me-2" />
        {/if}
        Loan
      </Button>
    </div>

    {#if isBarcodeEnabled()}
      <div class="flex justify-end">
        <div class="flex items-center gap-2">
          <span
            class="text-xs font-medium tracking-wide whitespace-nowrap text-slate-400 uppercase select-none dark:text-slate-500"
          >
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
              class="w-36 rounded-lg border border-slate-200 bg-white
                     px-3 py-2
                     text-sm text-slate-500 placeholder:text-slate-300
                     focus:border-slate-400 focus:ring-1
                     focus:ring-slate-300 focus:outline-none
                     disabled:opacity-50 dark:border-slate-600
                     dark:bg-slate-800 dark:text-slate-400 dark:placeholder:text-slate-600 dark:focus:border-slate-500
                     dark:focus:ring-slate-500"
            />
            {#if barcodeLoading}
              <div class="pointer-events-none absolute inset-y-0 inset-e-0 flex items-center pe-2">
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

<AddPatronModal
  bind:open={addPatronModalOpen}
  initialName={patronName}
  onPatronCreated={handleNewPatronCreated}
  onCancel={() => {
    addPatronModalOpen = false;
  }}
/>

<Debounce
  value={patronName}
  onTrigger={searchPatrons}
  delay={300}
  minLength={3}
  {lastValueRef}
  {cancelKey}
/>

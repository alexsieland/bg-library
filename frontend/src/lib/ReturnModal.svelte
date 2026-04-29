<script lang="ts">
  import { tick } from 'svelte';
  import { Modal, Button, Select, Label, Spinner } from 'flowbite-svelte';
  import { apiClient, type GameStatus } from './api-client';
  import { toasts } from './toast-store';

  export let open = false;
  export let statuses: GameStatus[] = [];
  export let onReturnSuccess: () => void = () => {};

  let selectedTransactionId = '';
  let returning = false;
  let barcodeValue = '';
  let barcodeError = '';
  let barcodeInputEl: HTMLInputElement | null = null;

  function captureBarcodeInput(node: HTMLInputElement) {
    barcodeInputEl = node;
    return {
      destroy() {
        barcodeInputEl = null;
      },
    };
  }

  $: gameTitle = statuses[0]?.game.title ?? '';
  $: hasCheckedOutGames = statuses.length > 0;

  // Reset state whenever the modal opens with new statuses or closes
  $: if (open) {
    barcodeValue = '';
    barcodeError = '';
    if (statuses.length === 1 && statuses[0].transactionId) {
      selectedTransactionId = statuses[0].transactionId;
    } else {
      selectedTransactionId = '';
    }
    // Focus the barcode input after the DOM updates
    tick().then(() => {
      barcodeInputEl?.focus();
    });
  } else {
    selectedTransactionId = '';
    barcodeValue = '';
    barcodeError = '';
  }

  function handleBarcodeKeydown(e: KeyboardEvent) {
    if (e.key !== 'Enter') return;
    e.preventDefault();

    const trimmed = barcodeValue.trim();
    if (!trimmed) return;

    const match = statuses.find((s) => s.patron?.barcode === trimmed);
    if (!match?.transactionId) {
      barcodeError = `No checked-out patron found with barcode "${trimmed}"`;
      return;
    }

    barcodeError = '';
    selectedTransactionId = match.transactionId;
    handleReturn();
  }

  async function handleReturn() {
    const status = statuses.find((s) => s.transactionId === selectedTransactionId);
    if (!status?.transactionId) return;
    returning = true;
    try {
      await apiClient.checkInGame(status.transactionId);
      toasts.add(`Successfully returned ${gameTitle}`, 'success');
      onReturnSuccess();
      open = false;
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Return failed';
      toasts.add(`Failed to return game: ${message}`, 'error');
    } finally {
      returning = false;
    }
  }
</script>

<Modal bind:open title={`Return Game: ${gameTitle}`} size="sm" autoclose={false}>
  <div class="space-y-4">
    <div>
      <Label for="returnBarcodeInput" class="mb-2">Scan Patron Barcode</Label>
      <input
        id="returnBarcodeInput"
        data-testid="return-barcode-input"
        use:captureBarcodeInput
        bind:value={barcodeValue}
        placeholder="Scan…"
        disabled={!hasCheckedOutGames || returning}
        onkeydown={handleBarcodeKeydown}
        autocomplete="off"
        class="block w-full rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-sm text-gray-900 focus:border-blue-500 focus:ring-blue-500 disabled:cursor-not-allowed disabled:opacity-50 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:placeholder-gray-400 dark:focus:border-blue-500 dark:focus:ring-blue-500"
      />
      {#if barcodeError}
        <p class="mt-1 text-sm text-red-500 dark:text-red-400" data-testid="return-barcode-error">
          {barcodeError}
        </p>
      {/if}
    </div>

    <div>
      <Label for="returnPatronSelect" class="mb-2">Patron</Label>
      <Select
        id="returnPatronSelect"
        data-testid="return-patron-select"
        bind:value={selectedTransactionId}
        disabled={!hasCheckedOutGames || returning}
      >
        {#if !hasCheckedOutGames}
          <option value="">No copies currently checked out</option>
        {:else}
          <option value="">Select a patron…</option>
          {#each statuses as status}
            <option value={status.transactionId ?? ''}>{status.patron?.name ?? 'Unknown'}</option>
          {/each}
        {/if}
      </Select>
      {#if !hasCheckedOutGames}
        <p
          class="mt-1 text-sm text-slate-500 dark:text-slate-400"
          data-testid="return-no-copies-message"
        >
          All copies of this game are currently available.
        </p>
      {/if}
    </div>

    <div class="flex justify-end gap-2 pt-2">
      <Button
        color="alternative"
        onclick={() => {
          open = false;
        }}
        disabled={returning}
      >
        Cancel
      </Button>
      <Button
        data-testid="return-modal-submit"
        onclick={handleReturn}
        disabled={returning || !selectedTransactionId}
        color="emerald"
      >
        {#if returning}
          <Spinner size="4" class="me-2" />
        {/if}
        Return
      </Button>
    </div>
  </div>
</Modal>

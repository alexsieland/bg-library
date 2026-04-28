<script lang="ts">
  import { Modal, Button, Select, Label, Spinner } from 'flowbite-svelte';
  import { apiClient, type GameStatus } from './api-client';
  import { toasts } from './toast-store';

  export let open = false;
  export let statuses: GameStatus[] = [];
  export let onReturnSuccess: () => void = () => {};

  let selectedTransactionId = '';
  let returning = false;

  $: gameTitle = statuses[0]?.game.title ?? '';
  $: hasCheckedOutGames = statuses.length > 0;

  // Reset selection whenever the modal opens with new statuses or closes
  $: if (open) {
    if (statuses.length === 1 && statuses[0].transactionId) {
      selectedTransactionId = statuses[0].transactionId;
    } else {
      selectedTransactionId = '';
    }
  } else {
    selectedTransactionId = '';
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

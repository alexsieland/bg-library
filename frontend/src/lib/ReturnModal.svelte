<script lang="ts">
  import { Modal, Button, Select, Label, Spinner } from 'flowbite-svelte';
  import { apiClient, type GameStatus } from './api-client';
  import type { components } from '../generated/library-api';
  import { toasts } from './toast-store';

  type Game = components['schemas']['Game'];

  export let open = false;
  export let games: Game[] = [];
  export let onReturnSuccess: () => void = () => {};

  let checkedOutStatuses: GameStatus[] = [];
  let selectedTransactionId = '';
  let loading = false;
  let returning = false;

  $: gameTitle = games[0]?.title ?? '';
  $: hasCheckedOutGames = checkedOutStatuses.length > 0;

  $: if (open && games.length > 0) {
    loadStatuses();
  } else if (!open) {
    reset();
  }

  async function loadStatuses() {
    const gameIds = new Set(games.map((g) => g.gameId));
    loading = true;
    try {
      const result = await apiClient.listGames({ checkedOut: true });
      checkedOutStatuses = result.games.filter((gs) => gameIds.has(gs.game.gameId));
      // Auto-select when exactly one copy is checked out
      if (checkedOutStatuses.length === 1 && checkedOutStatuses[0].transactionId) {
        selectedTransactionId = checkedOutStatuses[0].transactionId;
      } else {
        selectedTransactionId = '';
      }
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to load game status';
      toasts.add(`Failed to load game status: ${message}`, 'error');
    } finally {
      loading = false;
    }
  }

  async function handleReturn() {
    const status = checkedOutStatuses.find((s) => s.transactionId === selectedTransactionId);
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

  function reset() {
    checkedOutStatuses = [];
    selectedTransactionId = '';
  }
</script>

<Modal bind:open title={`Return Game: ${gameTitle}`} size="sm" autoclose={false}>
  <div class="space-y-4">
    {#if loading}
      <div class="flex justify-center py-8" data-testid="return-modal-loading">
        <Spinner size="8" />
      </div>
    {:else}
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
            {#each checkedOutStatuses as status}
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
    {/if}
  </div>
</Modal>

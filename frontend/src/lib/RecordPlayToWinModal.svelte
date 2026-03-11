<script lang="ts">
  import { Button, Modal, Label, Input, Spinner } from 'flowbite-svelte';
  import { CirclePlusSolid, CircleMinusSolid } from 'flowbite-svelte-icons';
  import { apiClient, type PlayToWinGame, type CreatePlayToWinSessionEntry } from './api-client';
  import { getPlayToWinIdLabel } from './config';
  import { toasts } from './toast-store';

  export let open = false;
  export let playToWinGame: PlayToWinGame | null = null;

  const MAX_ENTRIES = 12;

  interface Entry {
    id: string;
    entrantName: string;
    entrantUniqueId: string;
  }

  let entries: Entry[] = [{ id: crypto.randomUUID(), entrantName: '', entrantUniqueId: '' }];
  let submitting = false;
  let canSubmit = false;

  function addEntry() {
    if (entries.length < MAX_ENTRIES) {
      entries = [...entries, { id: crypto.randomUUID(), entrantName: '', entrantUniqueId: '' }];
    }
  }

  function removeEntry(id: string) {
    entries = entries.filter((e) => e.id !== id);
  }

  async function handleSubmit() {
    if (!playToWinGame || !canSubmit) return;

    // Validate at least one entry
    if (entries.length === 0) {
      toasts.add('Please add at least one player entry', 'error');
      return;
    }

    // Validate all entries have names and IDs
    const allValid = entries.every((e) => e.entrantName.trim() && e.entrantUniqueId.trim());
    if (!allValid) {
      toasts.add('Please fill in all player names and IDs', 'error');
      return;
    }

    submitting = true;
    try {
      const sessionEntries: CreatePlayToWinSessionEntry[] = entries.map((e) => ({
        entrantName: e.entrantName.trim(),
        entrantUniqueId: e.entrantUniqueId.trim(),
      }));

      await apiClient.addPlayToWinSession(playToWinGame.playToWinId, sessionEntries);

      toasts.add(`Successfully recorded session for ${playToWinGame.title}`, 'success');
      open = false;
      resetForm();
    } catch (e) {
      const errorMessage = e instanceof Error ? e.message : 'Failed to record session';
      toasts.add(`Error: ${errorMessage}`, 'error');
    } finally {
      submitting = false;
    }
  }

  function resetForm() {
    entries = [{ id: crypto.randomUUID(), entrantName: '', entrantUniqueId: '' }];
  }

  // Reset form when modal is opened/closed
  $: if (open) {
    resetForm();
  }

  $: canSubmit =
    entries.length > 0 &&
    entries.every((e) => e.entrantName.trim().length > 0 && e.entrantUniqueId.trim().length > 0);

  const idLabel = getPlayToWinIdLabel();
</script>

<Modal
  bind:open
  title={`Record Session: ${playToWinGame?.title || ''}`}
  size="lg"
  autoclose={false}
  dismissable={false}
  outsideclose={false}
  class="w-full"
>
  <div class="space-y-6" data-testid="ptw-record-modal">
    {#if playToWinGame}
      <!-- Entries Section -->
      <div class="max-h-96 space-y-3 overflow-y-auto">
        <div class="text-sm font-semibold text-slate-700 dark:text-slate-300">Players</div>

        {#each entries as entry, index (entry.id)}
          <div class="flex items-end gap-3" data-testid={`ptw-entry-${index}`}>
            <!-- Player Name -->
            <div class="flex-1">
              <Label for={`entrant-name-${entry.id}`} class="text-xs">Player Name</Label>
              <Input
                id={`entrant-name-${entry.id}`}
                type="text"
                placeholder="e.g., John Smith"
                bind:value={entry.entrantName}
                disabled={submitting}
                maxlength={100}
                data-testid={`ptw-entrant-name-${index}`}
              />
            </div>

            <!-- ID Field -->
            <div class="flex-1">
              <Label for={`entrant-id-${entry.id}`} class="text-xs">{idLabel}</Label>
              <Input
                id={`entrant-id-${entry.id}`}
                type="text"
                placeholder={`e.g., ${idLabel}`}
                bind:value={entry.entrantUniqueId}
                disabled={submitting}
                maxlength={100}
                data-testid={`ptw-entrant-id-${index}`}
              />
            </div>

            <!-- Add/Remove Buttons -->
            <div class="flex gap-1">
              {#if index === entries.length - 1}
                <!-- Plus button on last row -->
                <button
                  type="button"
                  onclick={addEntry}
                  disabled={entries.length >= MAX_ENTRIES || submitting}
                  class="rounded-lg p-2 text-emerald-500 transition hover:bg-emerald-50 disabled:cursor-not-allowed disabled:opacity-50 dark:hover:bg-emerald-900/20"
                  aria-label="Add player"
                  data-testid="ptw-add-entry-button"
                >
                  <CirclePlusSolid size="lg" />
                </button>
              {:else}
                <!-- Minus button on non-last rows -->
                <button
                  type="button"
                  onclick={() => removeEntry(entry.id)}
                  disabled={submitting}
                  class="rounded-lg p-2 text-rose-500 transition hover:bg-rose-50 disabled:cursor-not-allowed disabled:opacity-50 dark:hover:bg-rose-900/20"
                  aria-label="Remove player"
                  data-testid={`ptw-remove-entry-button-${index}`}
                >
                  <CircleMinusSolid size="lg" />
                </button>
              {/if}
            </div>
          </div>
        {/each}
      </div>

      <!-- Form Actions -->
      <div class="flex justify-end gap-3 border-t border-slate-200 pt-4 dark:border-slate-700">
        <Button
          color="alternative"
          disabled={submitting}
          onclick={() => (open = false)}
          data-testid="ptw-record-cancel-button"
        >
          Cancel
        </Button>
        <Button
          color="emerald"
          disabled={submitting || !canSubmit}
          onclick={handleSubmit}
          data-testid="ptw-record-submit-button"
        >
          {#if submitting}
            <Spinner size="4" class="me-2" />
            Recording...
          {:else}
            Record Play to Win
          {/if}
        </Button>
      </div>
    {/if}
  </div>
</Modal>

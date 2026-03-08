<script lang="ts">
  import { Button, Input, Label, Spinner, Fileupload, Helper } from 'flowbite-svelte';
  import { apiClient, type Game } from './api-client';
  import { toasts } from './toast-store';
  import AddPatronModal from './AddPatronModal.svelte';
  import AddGameModal from './AddGameModal.svelte';

  let gameTitle = '';
  let loading = false;
  let error: string | null = null;

  let bulkUploadFile: FileList | undefined;
  let bulkLoading = false;
  let bulkError: string | null = null;

  let addGameModalOpen = false;

  function handleGameCreated(game: Game) {
    toasts.add(`Successfully added game: ${game.title}`, 'success');
  }

  async function handleBulkUpload() {
    if (!bulkUploadFile || bulkUploadFile.length === 0) return;

    bulkLoading = true;
    bulkError = null;

    try {
      const file = bulkUploadFile[0];
      const result = await apiClient.bulkAddGames(file);
      toasts.add(
        `Successfully imported ${result.imported} game${result.imported !== 1 ? 's' : ''}`,
        'success'
      );
      bulkUploadFile = undefined;
    } catch (e) {
      console.error('Error uploading games:', e);
      const errorMessage = e instanceof Error ? e.message : 'Failed to upload games';
      bulkError = errorMessage;
      toasts.add(`Failed to upload games: ${errorMessage}`, 'error');
    } finally {
      bulkLoading = false;
    }
  }
</script>

<div class="p-6">
  <div class="grid grid-cols-1 gap-8 lg:grid-cols-2">
    <section>
      <h2 class="mb-4 text-xl font-semibold text-slate-900 dark:text-slate-100">Add Games</h2>
      <div class="space-y-2">
        <Button onclick={() => (addGameModalOpen = true)}>Add Game</Button>
      </div>
    </section>

    <section>
      <h2 class="mb-4 text-xl font-semibold text-slate-900 dark:text-slate-100">Bulk Add Games</h2>
      <div class="space-y-4">
        <div>
          <Label for="bulkGameUpload" class="mb-2">Upload CSV File</Label>
          <div class="flex items-center space-x-2">
            <div class="grow">
              <Fileupload
                id="bulkGameUpload"
                bind:files={bulkUploadFile}
                accept=".csv,text/csv,text/plain"
              />
            </div>
            <Button
              onclick={handleBulkUpload}
              disabled={bulkLoading || !bulkUploadFile || bulkUploadFile.length === 0}
            >
              {#if bulkLoading}
                <Spinner size="4" class="me-2" />
              {/if}
              Upload Games
            </Button>
          </div>
          <Helper class="mt-2">
            Upload a CSV file with one game title per line. Maximum file size: 10MB.
          </Helper>
        </div>
        {#if bulkError}
          <p class="text-sm text-rose-500">{bulkError}</p>
        {/if}
      </div>
    </section>
  </div>
</div>

<AddGameModal bind:open={addGameModalOpen} onGameSaved={handleGameCreated} />

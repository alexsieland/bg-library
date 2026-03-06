<script lang="ts">
  import { Button, Input, Label, Spinner, Fileupload, Helper } from 'flowbite-svelte';
  import { apiClient } from './api-client';
  import { toasts } from './toast-store';

  let gameTitle = '';
  let loading = false;
  let error: string | null = null;

  let bulkUploadFile: FileList | undefined;
  let bulkLoading = false;
  let bulkError: string | null = null;

  async function handleAddGame() {
    if (!gameTitle.trim()) return;
    loading = true;
    error = null;
    try {
      const newGame = await apiClient.addGame({ title: gameTitle.trim() });
      toasts.add(`Successfully added ${newGame.title} to the library`, 'success');
      gameTitle = '';
    } catch (e) {
      console.error('Error adding game:', e);
      const errorMessage = e instanceof Error ? e.message : 'Failed to add game';
      error = errorMessage;
      toasts.add(`Failed to add game: ${errorMessage}`, 'error');
    } finally {
      loading = false;
    }
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter') {
      handleAddGame();
    }
  }

  async function handleBulkUpload() {
    if (!bulkUploadFile || bulkUploadFile.length === 0) return;

    bulkLoading = true;
    bulkError = null;

    try {
      const file = bulkUploadFile[0];
      const result = await apiClient.bulkAddGames(file);
      toasts.add(`Successfully imported ${result.imported} game${result.imported !== 1 ? 's' : ''}`, 'success');
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
  <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
    <section>
      <h2 class="text-xl font-semibold text-slate-900 dark:text-slate-100 mb-4">Add Games</h2>
      <div class="space-y-2">
        <Label for="gameTitle">Game Title</Label>
        <div class="flex items-center space-x-2">
          <div class="grow">
            <Input
              id="gameTitle"
              placeholder="Enter game title"
              bind:value={gameTitle}
              onkeydown={handleKeydown}
              autocomplete="off"
              maxlength={100}
            />
          </div>
          <Button onclick={handleAddGame} disabled={loading || !gameTitle.trim()}>
            {#if loading}
              <Spinner size="4" class="me-2" />
            {/if}
            Add Game
          </Button>
        </div>
      </div>
      {#if error}
        <p class="mt-2 text-sm text-rose-500">{error}</p>
      {/if}
    </section>

    <section>
      <h2 class="text-xl font-semibold text-slate-900 dark:text-slate-100 mb-4">Bulk Add Games</h2>
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


<script lang="ts">
  import { Button, Helper } from 'flowbite-svelte';
  import { apiClient, type Game } from './api-client';
  import { toasts } from './toast-store';
  import AddGameModal from './AddGameModal.svelte';
  import CsvUploadModal from './CsvUploadModal.svelte';

  let addGameModalOpen = false;
  let csvUploadModalOpen = false;

  function handleGameCreated(game: Game) {
    toasts.add(`Successfully added game: ${game.title}`, 'success');
  }

  async function handleBulkGamesUpload(file: File) {
    return await apiClient.bulkAddGames(file);
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
        <Button onclick={() => (csvUploadModalOpen = true)}>Upload CSV</Button>
        <Helper>Upload a CSV file with one game title per line. Maximum file size: 10MB.</Helper>
      </div>
    </section>
  </div>
</div>

<AddGameModal bind:open={addGameModalOpen} onGameSaved={handleGameCreated} />

<CsvUploadModal
  bind:open={csvUploadModalOpen}
  title="Bulk Add Games"
  successMessage={(count) => `Successfully imported ${count} game${count !== 1 ? 's' : ''}`}
  onUpload={handleBulkGamesUpload}
  onCancel={() => {
    // Handle cancel if needed
  }}
/>

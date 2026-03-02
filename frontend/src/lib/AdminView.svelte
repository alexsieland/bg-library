<script lang="ts">
  import { Button, Input, Label, Spinner } from 'flowbite-svelte';
  import { apiClient } from './api-client';
  import { toasts } from './toast-store';

  let gameTitle = '';
  let loading = false;
  let error: string | null = null;

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
</script>

<div class="p-6 space-y-8">
  <section class="max-w-2xl">
    <h2 class="text-xl font-semibold text-slate-900 dark:text-slate-100 mb-4">Add Games</h2>
    <div class="flex items-end space-x-2">
      <div class="flex-grow">
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
    {#if error}
      <p class="mt-2 text-sm text-red-500">{error}</p>
    {/if}
  </section>
</div>

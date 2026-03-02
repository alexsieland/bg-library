<script lang="ts">
  import { Table, TableBody, TableBodyCell, TableBodyRow, TableHead, TableHeadCell, Button } from 'flowbite-svelte';
  import SearchBar from './SearchBar.svelte';
  import { apiClient, type GameList } from './api-client';
  import { onMount } from 'svelte';
  import { toasts } from './toast-store';

  let searchQuery = '';
  let gameList: GameList = { games: [] };
  let error: string | null = null;
  let loading = true;

  async function fetchCheckedOutGames() {
    loading = true;
    error = null;
    try {
      gameList = await apiClient.listGames({ 
        title: searchQuery || undefined,
        checkedOut: true
      });
    } catch (e) {
      const errorMessage = e instanceof Error ? e.message : 'An unknown error occurred';
      error = errorMessage;
      toasts.add(`Failed to load checked out games: ${errorMessage}`, 'error');
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    fetchCheckedOutGames();
  });

  async function handleReturn(transactionId: string | undefined, gameTitle: string) {
    if (!transactionId) {
      toasts.add(`Cannot return ${gameTitle}: Missing transaction ID`, 'error');
      return;
    }
    try {
      await apiClient.checkInGame(transactionId);
      toasts.add(`Successfully returned ${gameTitle}`, 'success');
      fetchCheckedOutGames();
    } catch (e) {
      const errorMessage = e instanceof Error ? e.message : 'An unknown error occurred';
      toasts.add(`Failed to return ${gameTitle}: ${errorMessage}`, 'error');
    }
  }

  function handleSearch(query: string) {
    searchQuery = query;
    fetchCheckedOutGames();
  }

  function formatDate(dateString: string | undefined) {
    if (!dateString) return '-';
    const date = new Date(dateString);
    return date.toLocaleString('en-US', {
      month: '2-digit',
      day: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      hour12: true
    });
  }
</script>

<div class="p-6 border-b border-slate-200 dark:border-slate-700 bg-slate-50/50 dark:bg-slate-800/50">
  <SearchBar bind:searchQuery placeholder="Search checked out games..." onSearch={handleSearch} />
</div>

{#if loading && gameList.games.length === 0}
  <div class="p-8 text-center text-slate-500 dark:text-slate-400">Loading checked out games...</div>
{:else if error}
  <div class="p-8 text-center text-red-500">{error}</div>
{:else}
  <Table shadow hoverable={true} class="w-full">
    <TableHead>
      <TableHeadCell>Game Title</TableHeadCell>
      <TableHeadCell>Borrower</TableHeadCell>
      <TableHeadCell>Check Out Time</TableHeadCell>
      <TableHeadCell>Action</TableHeadCell>
    </TableHead>
    <TableBody class="divide-y">
      {#each gameList.games as gameStatus (gameStatus.game.gameId)}
        <TableBodyRow>
          <TableBodyCell class="text-lg font-medium text-slate-900 dark:text-slate-100">{gameStatus.game.title}</TableBodyCell>
          <TableBodyCell class="text-slate-700 dark:text-slate-300">
            {gameStatus.patron?.name || 'Unknown'}
          </TableBodyCell>
          <TableBodyCell class="text-slate-600 dark:text-slate-400">
            {formatDate(gameStatus.checkedOutAt)}
          </TableBodyCell>
          <TableBodyCell>
            <Button
              onclick={() => handleReturn(gameStatus.transactionId, gameStatus.game.title)}
              color="emerald"
              size="sm"
            >
              Returned
            </Button>
          </TableBodyCell>
        </TableBodyRow>
      {/each}
      {#if gameList.games.length === 0}
        <TableBodyRow>
          <TableBodyCell colspan={4} class="px-6 py-12 text-center text-slate-500 dark:text-slate-400">
            No checked out games found.
          </TableBodyCell>
        </TableBodyRow>
      {/if}
    </TableBody>
  </Table>
{/if}

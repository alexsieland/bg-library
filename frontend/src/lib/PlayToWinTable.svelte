<script lang="ts">
  import {
    Table,
    TableBody,
    TableBodyCell,
    TableBodyRow,
    TableHead,
    TableHeadCell,
    Button,
  } from 'flowbite-svelte';
  import SearchBar from './SearchBar.svelte';
  import { apiClient, type PlayToWinGame, type PlayToWinGameList } from './api-client';
  import { onMount } from 'svelte';
  import { toasts } from './toast-store';
  import RecordPlayToWinModal from './RecordPlayToWinModal.svelte';

  const PAGE_LIMIT = 100;

  let searchQuery = '';
  let recordModalOpen = false;
  let selectedPlayToWinGame: PlayToWinGame | null = null;

  let playToWinGames: PlayToWinGameList = { games: [] };
  let error: string | null = null;
  let loading = true;

  async function fetchPlayToWinGames() {
    loading = true;
    error = null;
    try {
      playToWinGames = await apiClient.listPlayToWinGames(searchQuery, PAGE_LIMIT, 0);
    } catch (e) {
      const errorMessage = e instanceof Error ? e.message : 'An unknown error occurred';
      error = errorMessage;
      toasts.add(`Failed to load Play To Win games: ${errorMessage}`, 'error');
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    fetchPlayToWinGames();
  });

  function handleSearch(query: string) {
    searchQuery = query;
    fetchPlayToWinGames();
  }

  function handleRecord(game: PlayToWinGame) {
    selectedPlayToWinGame = game;
    recordModalOpen = true;
  }
</script>

<div
  class="border-b border-slate-200 bg-slate-50/50 px-6 py-4 dark:border-slate-700 dark:bg-slate-800/50"
>
  <div class="flex items-center justify-between gap-4">
    <div class="flex-1">
      <SearchBar
        bind:searchQuery
        placeholder="Search Play To Win games..."
        onSearch={handleSearch}
      />
    </div>
  </div>
</div>

<RecordPlayToWinModal bind:open={recordModalOpen} playToWinGame={selectedPlayToWinGame} />

{#if loading && playToWinGames.games.length === 0}
  <div class="p-8 text-center text-slate-500 dark:text-slate-400">Loading Play To Win games...</div>
{:else if error}
  <div class="p-8 text-center text-rose-500">{error}</div>
{:else}
  <Table shadow hoverable={true} class="w-full" data-testid="ptw-table">
    <TableHead>
      <TableHeadCell>Title</TableHeadCell>
      <TableHeadCell>Action</TableHeadCell>
    </TableHead>
    <TableBody class="divide-y">
      {#each playToWinGames.games as game (game.playToWinId)}
        <TableBodyRow data-testid={`ptw-row-${game.playToWinId}`}>
          <TableBodyCell class="text-lg font-medium text-slate-900 dark:text-slate-100">
            {game.title}
          </TableBodyCell>
          <TableBodyCell>
            <Button
              onclick={() => handleRecord(game)}
              color="primary"
              size="sm"
              data-testid={`ptw-record-button-${game.playToWinId}`}
            >
              Record
            </Button>
          </TableBodyCell>
        </TableBodyRow>
      {/each}
      {#if playToWinGames.games.length === 0}
        <TableBodyRow>
          <TableBodyCell
            colspan={2}
            class="px-6 py-12 text-center text-slate-500 dark:text-slate-400"
            data-testid="ptw-empty-state"
          >
            No Play To Win games found.
          </TableBodyCell>
        </TableBodyRow>
      {/if}
    </TableBody>
  </Table>
{/if}

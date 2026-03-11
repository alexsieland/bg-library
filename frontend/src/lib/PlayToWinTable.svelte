<script lang="ts">
  import {
    Table,
    TableBody,
    TableBodyCell,
    TableBodyRow,
    TableHead,
    TableHeadCell,
    Button,
    Spinner,
  } from 'flowbite-svelte';
  import SearchBar from './SearchBar.svelte';
  import { apiClient, type PlayToWinGame } from './api-client';
  import { onMount } from 'svelte';
  import { toasts } from './toast-store';
  import RecordPlayToWinModal from './RecordPlayToWinModal.svelte';

  const PAGE_LIMIT = 100;

  let searchQuery = '';
  let recordModalOpen = false;
  let selectedPlayToWinGame: PlayToWinGame | null = null;

  let games: PlayToWinGame[] = [];
  let error: string | null = null;
  let loading = true;
  let loadingMore = false;
  let offset = 0;
  let hasMore = true;
  let tableBodyElement: HTMLTableSectionElement | undefined;
  let lastSearchQuery = '';

  async function fetchPlayToWinGames(newOffset: number = 0) {
    const isNewSearch = searchQuery !== lastSearchQuery;
    const isInitialLoad = newOffset === 0;

    if (isNewSearch) {
      loading = true;
      error = null;
      games = [];
      offset = 0;
      lastSearchQuery = searchQuery;
    } else if (newOffset > 0) {
      loadingMore = true;
    }

    try {
      const result = await apiClient.listPlayToWinGames(searchQuery, PAGE_LIMIT, newOffset);
      if (isNewSearch) {
        games = result.games;
      } else {
        games = [...games, ...result.games];
      }
      offset = newOffset + result.games.length;
      hasMore = result.games.length === PAGE_LIMIT;
    } catch (e) {
      const errorMessage = e instanceof Error ? e.message : 'An unknown error occurred';
      error = errorMessage;
      toasts.add(`Failed to load Play To Win games: ${errorMessage}`, 'error');
    } finally {
      if (isInitialLoad) {
        loading = false;
      } else {
        loadingMore = false;
      }
    }
  }

  onMount(() => {
    fetchPlayToWinGames();
  });

  function handleSearch(query: string) {
    searchQuery = query;
    fetchPlayToWinGames(0);
  }

  function handleRecord(game: PlayToWinGame) {
    selectedPlayToWinGame = game;
    recordModalOpen = true;
  }

  function handleTableScroll() {
    if (!tableBodyElement || loadingMore || !hasMore || error) {
      return;
    }

    const { scrollTop, scrollHeight, clientHeight } = tableBodyElement;
    const isAtBottom = scrollHeight - scrollTop - clientHeight < 100;

    if (isAtBottom) {
      fetchPlayToWinGames(offset);
    }
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

{#if loading && games.length === 0}
  <div class="p-8 text-center text-slate-500 dark:text-slate-400">Loading Play To Win games...</div>
{:else if error}
  <div class="p-8 text-center text-rose-500">{error}</div>
{:else}
  <div class="max-h-96 overflow-y-auto">
    <Table shadow hoverable={true} class="w-full" data-testid="ptw-table">
      <TableHead>
        <TableHeadCell>Title</TableHeadCell>
        <TableHeadCell>Action</TableHeadCell>
      </TableHead>
      <TableBody
        class="divide-y"
        bind:this={tableBodyElement}
        on:scroll={handleTableScroll}
        data-testid="ptw-table-body"
      >
        {#each games as game (game.playToWinId)}
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
        {#if games.length === 0}
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
        {#if loadingMore}
          <TableBodyRow>
            <TableBodyCell colspan={2} class="px-6 py-4 text-center">
              <div class="flex items-center justify-center gap-2">
                <Spinner size="4" />
                <span class="text-sm text-slate-500 dark:text-slate-400">Loading more games...</span
                >
              </div>
            </TableBodyCell>
          </TableBodyRow>
        {/if}
      </TableBody>
    </Table>
  </div>
{/if}

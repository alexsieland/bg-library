<script lang="ts">
  import {
    TableBody,
    TableBodyCell,
    TableBodyRow,
    TableHead,
    TableHeadCell,
    Table,
    Button,
    Dropdown,
  } from 'flowbite-svelte';
  import { ChevronDownOutline } from 'flowbite-svelte-icons';
  import { apiClient, type DeletePlayToWinGameRequest, type PlayToWinGame } from './api-client';
  import { toasts } from './toast-store';
  import SearchBar from './SearchBar.svelte';
  import { onMount } from 'svelte';

  const PAGE_SIZE = 20;

  let searchQuery = $state('');
  let playToWinList: { games: PlayToWinGame[] } = $state({ games: [] });
  let loading = $state(true);
  let error: string | null = $state(null);
  let offset = $state(0);

  let filteredGames = $derived(playToWinList.games);
  let hasPreviousPage = $derived(offset > 0);
  let hasNextPage = $derived(playToWinList.games.length === PAGE_SIZE);

  async function fetchPlayToWinGames() {
    loading = true;
    error = null;
    try {
      playToWinList = await apiClient.listPlayToWinGames(searchQuery, PAGE_SIZE, offset);
    } catch (e) {
      const errorMessage = e instanceof Error ? e.message : 'An unknown error occurred';
      error = errorMessage;
      toasts.add(`Failed to load Play To Win games: ${errorMessage}`, 'error');
    } finally {
      loading = false;
    }
  }

  async function drawWinner(playToWinId: string, gameTitle: string) {
    const winner = await apiClient.drawPlayToWinRaffle(playToWinId);
    if (winner) {
      toasts.add(
        `Winner drawn for Play To Win game ${gameTitle}: ${winner.entrantName} (${winner.entrantUniqueId})`,
        'success'
      );
    } else {
      toasts.add(`Failed to draw winner for Play To Win game ${gameTitle}`, 'error');
    }
  }

  async function redrawWinner(playToWinId: string, winnerId: string, gameTitle: string) {
    // TODO eventually this should delete the current winner and draw a new one then redraw the winner from the new pool
    drawWinner(playToWinId, gameTitle);
  }

  async function claimRaffle(playToWinId: string, gameTitle: string) {
    const reqBody: DeletePlayToWinGameRequest = {
      RemovalReason: 'claimed',
    };
    await apiClient.deletePlayToWinGame(playToWinId, reqBody);
    toasts.add(`Raffle claimed for Play To Win game ${gameTitle}`, 'success');
  }

  function handleSearch(query: string) {
    searchQuery = query;
    offset = 0;
    fetchPlayToWinGames();
  }

  function previousPage() {
    if (!hasPreviousPage) return;
    offset = Math.max(0, offset - PAGE_SIZE);
    fetchPlayToWinGames();
  }

  function nextPage() {
    if (!hasNextPage) return;
    offset = offset + PAGE_SIZE;
    fetchPlayToWinGames();
  }

  onMount(() => {
    fetchPlayToWinGames();
  });
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
    <div class="flex items-center gap-2">
      <Button color="alternative" size="sm">
        Actions
        <ChevronDownOutline class="ml-2 h-3 w-3" />
      </Button>
      <Dropdown simple class="w-44 divide-y divide-gray-100">
        <span class="hidden" aria-hidden="true"></span>
      </Dropdown>
    </div>
  </div>
</div>

<div class="relative overflow-hidden bg-white shadow-md sm:rounded-lg dark:bg-slate-800">
  {#if loading && playToWinList.games.length === 0}
    <div class="p-8 text-center text-slate-500 dark:text-slate-400">
      Loading Play To Win games...
    </div>
  {:else if error}
    <div class="p-8 text-center text-rose-500">{error}</div>
  {:else}
    <Table shadow hoverable={true} class="w-full" data-testid="ptw-management-table">
      <TableHead>
        <TableHeadCell class="px-4 py-3" scope="col">Game Title</TableHeadCell>
        <TableHeadCell class="px-4 py-3" scope="col">Winner</TableHeadCell>
        <TableHeadCell class="px-4 py-3" scope="col">Action</TableHeadCell>
      </TableHead>

      <TableBody class="divide-y">
        {#if filteredGames.length === 0}
          <TableBodyRow>
            <TableBodyCell
              colspan={3}
              class="px-4 py-12 text-center text-slate-500 dark:text-slate-400"
              data-testid="ptw-management-empty-state"
            >
              No Play To Win games found.
            </TableBodyCell>
          </TableBodyRow>
        {:else}
          {#each filteredGames as game (game.playToWinId)}
            <TableBodyRow>
              <TableBodyCell
                class="px-4 py-3 text-lg font-medium text-slate-900 dark:text-slate-100"
                data-testid={`ptw-management-title-${game.playToWinId}`}
              >
                {game.title}
              </TableBodyCell>
              <TableBodyCell
                class="px-4 py-3 text-slate-700 dark:text-slate-200"
                data-testid={`ptw-management-winner-${game.playToWinId}`}
              >
                {#if game.winner}
                  {game.winner.entrantName} ({game.winner.entrantUniqueId})
                {:else}
                  &mdash;
                {/if}
              </TableBodyCell>
              <TableBodyCell class="px-4 py-3">
                <div class="flex gap-2">
                  <Button color="primary" size="sm">Draw Winner</Button>
                  <Button color="emerald" size="sm" disabled>Claimed</Button>
                </div>
              </TableBodyCell>
            </TableBodyRow>
          {/each}
        {/if}
      </TableBody>
    </Table>

    <nav
      class="flex items-center justify-end gap-2 border-t border-slate-200 px-4 py-3 dark:border-slate-700"
      aria-label="Pagination"
      data-testid="ptw-management-pagination-nav"
    >
      <Button color="light" size="xs" disabled={!hasPreviousPage} onclick={previousPage}
        >Previous</Button
      >
      <span class="text-sm text-slate-500 dark:text-slate-400">|</span>
      <Button color="light" size="xs" disabled={!hasNextPage} onclick={nextPage}>Next</Button>
    </nav>
  {/if}
</div>

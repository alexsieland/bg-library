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
    DropdownItem,
  } from 'flowbite-svelte';
  import { ChevronDownOutline } from 'flowbite-svelte-icons';
  import { apiClient, type DeletePlayToWinGameRequest, type PlayToWinGame } from './api-client';
  import { toasts } from './toast-store';
  import SearchBar from './SearchBar.svelte';
  import ResetRaffleConfirmationPrompt from './ResetRaffleConfirmationPrompt.svelte';
  import { onMount } from 'svelte';

  const PAGE_SIZE = 20;

  let searchQuery = $state('');
  let playToWinList: { games: PlayToWinGame[] } = $state({ games: [] });
  let loading = $state(true);
  let error: string | null = $state(null);
  let offset = $state(0);
  let resetRaffleConfirmationOpen = $state(false);
  let viewMode: 'available' | 'claimed' | 'all' = $state('available');

  let filteredGames = $derived(playToWinList.games);
  let hasPreviousPage = $derived(viewMode !== 'all' && offset > 0);
  let hasNextPage = $derived(viewMode !== 'all' && playToWinList.games.length === PAGE_SIZE);
  let isPaginationDisabled = $derived(viewMode === 'all');

  function formatDate(dateString: string | undefined) {
    if (!dateString) return '—';
    const date = new Date(dateString);
    return date.toLocaleString('en-US', {
      month: '2-digit',
      day: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      hour12: true,
    });
  }

  async function fetchPlayToWinGames() {
    loading = true;
    error = null;
    try {
      if (viewMode === 'all') {
        // Fetch both available and claimed games, concatenate with claimed games after available
        const [availableGames, claimedGames] = await Promise.all([
          apiClient.listPlayToWinGames(searchQuery, PAGE_SIZE, 0),
          apiClient.listPlayToWinGames(searchQuery, PAGE_SIZE, 0, 'claimed'),
        ]);
        playToWinList = {
          games: [...availableGames.games, ...claimedGames.games],
        };
      } else if (viewMode === 'claimed') {
        playToWinList = await apiClient.listPlayToWinGames(
          searchQuery,
          PAGE_SIZE,
          offset,
          'claimed'
        );
      } else {
        playToWinList = await apiClient.listPlayToWinGames(searchQuery, PAGE_SIZE, offset);
      }
    } catch (e) {
      const errorMessage = e instanceof Error ? e.message : 'An unknown error occurred';
      error = errorMessage;
      toasts.add(`Failed to load Play To Win games: ${errorMessage}`, 'error');
    } finally {
      loading = false;
    }
  }

  async function drawWinner(playToWinId: string, gameTitle: string) {
    try {
      const winner = await apiClient.drawPlayToWinRaffle(playToWinId);

      playToWinList = {
        ...playToWinList,
        games: playToWinList.games.map((game) =>
          game.playToWinId === playToWinId ? { ...game, winner } : game
        ),
      };

      toasts.add(
        `Winner drawn for Play To Win game ${gameTitle}: ${winner.entrantName} (${winner.entrantUniqueId})`,
        'success'
      );
    } catch (e) {
      const errorMessage = e instanceof Error ? e.message : 'An unknown error occurred';
      if (errorMessage.includes('Resource not found')) {
        console.warn(errorMessage);
        toasts.add(`No winner drawn for ${gameTitle}. Did anyone enter the raffle?`, 'warn');
        return;
      }
      console.error(errorMessage);
      toasts.add(
        `Failed to draw winner for Play To Win game ${gameTitle}: ${errorMessage}`,
        'error'
      );
    }
  }

  async function claimRaffle(playToWinId: string, gameTitle: string) {
    try {
      const reqBody: DeletePlayToWinGameRequest = {
        RemovalReason: 'claimed',
      };
      await apiClient.deletePlayToWinGame(playToWinId, reqBody);
      toasts.add(`Raffle claimed for Play To Win game ${gameTitle}`, 'success');
      await fetchPlayToWinGames();
    } catch (e) {
      const errorMessage = e instanceof Error ? e.message : 'An unknown error occurred';
      toasts.add(
        `Failed to claim raffle for Play To Win game ${gameTitle}: ${errorMessage}`,
        'error'
      );
    }
  }

  async function restoreGame(playToWinId: string, gameTitle: string) {
    try {
      await apiClient.restorePlayToWinGame(playToWinId);
      toasts.add(`Restored Play To Win game ${gameTitle}`, 'success');
      await fetchPlayToWinGames();
    } catch (e) {
      const errorMessage = e instanceof Error ? e.message : 'An unknown error occurred';
      toasts.add(`Failed to restore Play To Win game ${gameTitle}: ${errorMessage}`, 'error');
    }
  }

  function handleSearch(query: string) {
    searchQuery = query;
    offset = 0;
    fetchPlayToWinGames();
  }

  function handleViewModeChange(mode: 'available' | 'claimed' | 'all') {
    viewMode = mode;
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

<ResetRaffleConfirmationPrompt
  bind:open={resetRaffleConfirmationOpen}
  onConfirm={fetchPlayToWinGames}
/>

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
      <div
        class="inline-flex rounded-lg border border-slate-300 dark:border-slate-600"
        data-testid="ptw-view-mode-selector"
      >
        <Button
          color={viewMode === 'available' ? 'primary' : 'light'}
          size="sm"
          class="rounded-r-none"
          onclick={() => handleViewModeChange('available')}
        >
          Available
        </Button>
        <Button
          color={viewMode === 'claimed' ? 'primary' : 'light'}
          size="sm"
          class="rounded-none border-x"
          onclick={() => handleViewModeChange('claimed')}
        >
          Claimed
        </Button>
        <Button
          color={viewMode === 'all' ? 'primary' : 'light'}
          size="sm"
          class="rounded-l-none"
          onclick={() => handleViewModeChange('all')}
        >
          All
        </Button>
      </div>
      <Button color="alternative" size="sm">
        Actions
        <ChevronDownOutline class="ml-2 h-3 w-3" />
      </Button>
      <Dropdown simple class="w-44 divide-y divide-gray-100">
        <DropdownItem onclick={() => (resetRaffleConfirmationOpen = true)}
          >Restart Raffle</DropdownItem
        >
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
        <TableHeadCell class="px-4 py-3" scope="col">Claimed At</TableHeadCell>
        <TableHeadCell class="px-4 py-3" scope="col">Action</TableHeadCell>
      </TableHead>

      <TableBody class="divide-y">
        {#if filteredGames.length === 0}
          <TableBodyRow>
            <TableBodyCell
              colspan={4}
              class="px-4 py-12 text-center text-slate-500 dark:text-slate-400"
              data-testid="ptw-management-empty-state"
            >
              No Play To Win games found.
            </TableBodyCell>
          </TableBodyRow>
        {:else}
          {#each filteredGames as game (game.playToWinId)}
            {@const isClaimed = !!game.deletedAt}
            <TableBodyRow>
              <TableBodyCell
                class="px-4 py-3 text-lg font-medium text-slate-900 dark:text-slate-100"
                data-testid={`ptw-management-title-${game.playToWinId}`}
              >
                <span class:line-through={isClaimed}>{game.title}</span>
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
              <TableBodyCell
                class="px-4 py-3 text-slate-700 dark:text-slate-200"
                data-testid={`ptw-management-claimed-at-${game.playToWinId}`}
              >
                {formatDate(game.deletedAt)}
              </TableBodyCell>
              <TableBodyCell class="px-4 py-3">
                <div class="flex gap-2">
                  {#if isClaimed}
                    <Button
                      color="yellow"
                      size="sm"
                      onclick={() => restoreGame(game.playToWinId, game.title)}
                      data-testid={`ptw-management-restore-button-${game.playToWinId}`}
                      >Restore</Button
                    >
                  {:else}
                    <Button
                      color="primary"
                      size="sm"
                      onclick={() => drawWinner(game.playToWinId, game.title)}
                      data-testid={`ptw-management-draw-button-${game.playToWinId}`}
                      >Draw Winner</Button
                    >
                    <Button
                      color="emerald"
                      size="sm"
                      disabled={!game.winner}
                      onclick={() => claimRaffle(game.playToWinId, game.title)}
                      data-testid={`ptw-management-claim-button-${game.playToWinId}`}
                      >Claim Raffle</Button
                    >
                  {/if}
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
      <Button
        color="light"
        size="xs"
        disabled={!hasPreviousPage || isPaginationDisabled}
        onclick={previousPage}>Previous</Button
      >
      <span class="text-sm text-slate-500 dark:text-slate-400">|</span>
      <Button
        color="light"
        size="xs"
        disabled={!hasNextPage || isPaginationDisabled}
        onclick={nextPage}>Next</Button
      >
    </nav>
  {/if}
</div>

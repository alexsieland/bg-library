<script lang="ts">
  import {
    Table,
    TableBody,
    TableBodyCell,
    TableBodyRow,
    TableHead,
    TableHeadCell,
    Button,
    Badge,
    Spinner,
  } from 'flowbite-svelte';
  import SearchBar from './SearchBar.svelte';
  import BarcodeInput from './BarcodeInput.svelte';
  import { apiClient, type GameStatus } from './api-client';
  import { onMount } from 'svelte';
  import { toasts } from './toast-store';
  import { isBarcodeEnabled } from './config';
  import { barcodeScanner } from './barcodeScannerAction';
  import { createPageScrollHandler } from './snippets/pageScrollLoader';
  import ReturnModal from './ReturnModal.svelte';

  const PAGE_LIMIT = 100;

  let searchQuery = '';
  let games: GameStatus[] = [];
  let error: string | null = null;
  let loading = true;
  let loadingMore = false;
  let offset = 0;
  let hasMore = true;
  let barcodeInputElement: HTMLInputElement | undefined;
  let returnModalOpen = false;
  let returnStatuses: GameStatus[] = [];

  async function fetchCheckedOutGames(newOffset: number = 0) {
    if (newOffset === 0) {
      loading = true;
      error = null;
      games = [];
      offset = 0;
      hasMore = true;
    } else {
      loadingMore = true;
    }
    try {
      const result = await apiClient.listGames({
        title: searchQuery || undefined,
        checkedOut: true,
        limit: PAGE_LIMIT,
        offset: newOffset,
      });
      if (newOffset === 0) {
        games = result.games;
      } else {
        games = [...games, ...result.games];
      }
      offset = newOffset + result.games.length;
      hasMore = result.games.length === PAGE_LIMIT;
    } catch (e) {
      const errorMessage = e instanceof Error ? e.message : 'An unknown error occurred';
      error = errorMessage;
      toasts.add(`Failed to load checked out games: ${errorMessage}`, 'error');
    } finally {
      if (newOffset === 0) {
        loading = false;
      } else {
        loadingMore = false;
      }
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
    fetchCheckedOutGames(0);
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
      hour12: true,
    });
  }

  function handleBarcodeError(message: string) {
    toasts.add(message, 'error');
  }

  function handleStatusesFound(statuses: GameStatus[]) {
    returnStatuses = statuses;
    returnModalOpen = true;
  }

  async function onScanComplete(barcode: string) {
    if (barcodeInputElement) {
      barcodeInputElement.focus();
    }
    try {
      const result = await apiClient.listGames({ barcode, checkedOut: true });
      if (result.games.length === 0) {
        toasts.add('All copies of this game are currently available.', 'warn');
        return;
      }
      handleStatusesFound(result.games);
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to look up barcode';
      toasts.add(message, 'error');
    }
  }

  function handleWindowKeydown(event: KeyboardEvent) {
    if (!isBarcodeEnabled()) return;
    // Alt+B: focus the barcode input
    if (event.altKey && event.key === 'b') {
      event.preventDefault();
      if (barcodeInputElement) barcodeInputElement.focus();
    }
  }

  const handlePageScroll = createPageScrollHandler({
    isLoading: () => loading,
    isLoadingMore: () => loadingMore,
    hasMore: () => hasMore,
    hasError: () => error !== null,
    loadMore: () => fetchCheckedOutGames(offset),
  });
</script>

<svelte:window
  use:barcodeScanner={{ onScan: onScanComplete }}
  on:keydown={handleWindowKeydown}
  on:scroll={handlePageScroll}
/>

<div
  class="border-b border-slate-200 bg-slate-50/50 px-6 py-4 dark:border-slate-700 dark:bg-slate-800/50"
>
  <div class="flex items-center justify-between gap-4">
    <div class="flex-1">
      <SearchBar
        bind:searchQuery
        placeholder="Search checked out games..."
        onSearch={handleSearch}
      />
    </div>
    {#if isBarcodeEnabled()}
      <BarcodeInput
        bind:barcodeInputElement
        checkedOut={true}
        onStatusesFound={handleStatusesFound}
        onError={handleBarcodeError}
      />
    {/if}
  </div>
</div>

{#if loading && games.length === 0}
  <div class="p-8 text-center text-slate-500 dark:text-slate-400">Loading checked out games...</div>
{:else if error}
  <div class="p-8 text-center text-rose-500">{error}</div>
{:else}
  <Table shadow hoverable={true} class="w-full" data-testid="check-in-table">
    <TableHead>
      <TableHeadCell>Game Title</TableHeadCell>
      <TableHeadCell>Borrower</TableHeadCell>
      <TableHeadCell>Borrow Start Time</TableHeadCell>
      <TableHeadCell>Action</TableHeadCell>
    </TableHead>
    <TableBody class="divide-y">
      {#each games as gameStatus (gameStatus.game.gameId)}
        <TableBodyRow>
          <TableBodyCell class="text-lg font-medium text-slate-900 dark:text-slate-100">
            <div class="flex items-center gap-1">
              {gameStatus.game.title}
              {#if gameStatus.game.isPlayToWin}
                <Badge color="sky" class="ml-2">P2W</Badge>
              {/if}
            </div>
          </TableBodyCell>
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
      {#if games.length === 0}
        <TableBodyRow>
          <TableBodyCell
            colspan={4}
            class="px-6 py-12 text-center text-slate-500 dark:text-slate-400"
          >
            No checked out games found.
          </TableBodyCell>
        </TableBodyRow>
      {/if}
      {#if loadingMore}
        <TableBodyRow>
          <TableBodyCell colspan={4} class="px-6 py-4 text-center">
            <div class="flex items-center justify-center gap-2">
              <Spinner size="4" />
              <span class="text-sm text-slate-500 dark:text-slate-400">Loading more games...</span>
            </div>
          </TableBodyCell>
        </TableBodyRow>
      {/if}
    </TableBody>
  </Table>
{/if}

<ReturnModal
  bind:open={returnModalOpen}
  statuses={returnStatuses}
  onReturnSuccess={fetchCheckedOutGames}
/>

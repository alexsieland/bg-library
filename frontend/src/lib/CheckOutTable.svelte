<script lang="ts">
  import {
    Table,
    TableBody,
    TableBodyCell,
    TableBodyRow,
    TableHead,
    TableHeadCell,
    Badge,
    Button,
    Spinner,
  } from 'flowbite-svelte';
  import SearchBar from './SearchBar.svelte';
  import BarcodeInput from './BarcodeInput.svelte';
  import { apiClient, type GameStatus } from './api-client';
  import type { components } from '../generated/library-api';
  import { onMount } from 'svelte';
  import { toasts } from './toast-store';
  import { isBarcodeEnabled } from './config';
  import { barcodeScanner } from './barcodeScannerAction';
  import { createPageScrollHandler } from './snippets/pageScrollLoader';

  import LoanModal from './LoanModal.svelte';

  const PAGE_LIMIT = 100;

  let searchQuery = '';
  let loanModalOpen = false;
  let selectedGame: components['schemas']['Game'] | null = null;
  let loanInitialFocus: 'patron' | 'barcode' = 'patron';
  let barcodeInputElement: HTMLInputElement | undefined;

  let games: GameStatus[] = [];
  let error: string | null = null;
  let loading = true;
  let loadingMore = false;
  let offset = 0;
  let hasMore = true;

  async function fetchGames(newOffset: number = 0) {
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
      toasts.add(`Failed to load games: ${errorMessage}`, 'error');
    } finally {
      if (newOffset === 0) {
        loading = false;
      } else {
        loadingMore = false;
      }
    }
  }

  onMount(() => {
    fetchGames();
  });

  function handleCheckout(game: components['schemas']['Game']) {
    selectedGame = game;
    loanModalOpen = true;
  }

  function handleSearch(query: string) {
    searchQuery = query;
    fetchGames(0);
  }

  const handlePageScroll = createPageScrollHandler({
    isLoading: () => loading,
    isLoadingMore: () => loadingMore,
    hasMore: () => hasMore,
    hasError: () => error !== null,
    loadMore: () => fetchGames(offset),
  });

  function handleBarcodeFound(game: components['schemas']['Game']) {
    loanInitialFocus = 'barcode';
    handleCheckout(game);
  }

  function handleBarcodeError(message: string) {
    toasts.add(message, 'error');
  }

  function handleStatusesFound(statuses: GameStatus[]) {
    handleBarcodeFound(statuses[0].game);
  }

  async function onScanComplete(barcode: string) {
    try {
      if (barcodeInputElement) {
        barcodeInputElement.focus();
        barcodeInputElement.value = barcode;
        barcodeInputElement.dispatchEvent(new Event('input', { bubbles: true }));
      }
      const result = await apiClient.listGames({ barcode, checkedOut: false });
      if (result.games.length === 0) {
        toasts.add(
          'All copies of this game are currently checked out or no game found with this barcode.',
          'error'
        );
        return;
      }
      // Since we filtered for available games, just take the first one
      handleBarcodeFound(result.games[0].game);
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
    <!-- Primary: search -->
    <div class="flex-1">
      <SearchBar bind:searchQuery placeholder="Search games..." onSearch={handleSearch} />
    </div>

    <!-- Secondary: barcode scanner, right-aligned and visually de-emphasised -->
    {#if isBarcodeEnabled()}
      <BarcodeInput
        bind:barcodeInputElement
        onStatusesFound={handleStatusesFound}
        onError={handleBarcodeError}
      />
    {/if}
  </div>
</div>

<LoanModal
  bind:open={loanModalOpen}
  game={selectedGame}
  onLoanSuccess={fetchGames}
  initialFocus={loanInitialFocus}
/>

{#if loading && games.length === 0}
  <div class="p-8 text-center text-slate-500 dark:text-slate-400">Loading games...</div>
{:else if error}
  <div class="p-8 text-center text-rose-500">{error}</div>
{:else}
  <Table shadow hoverable={true} class="w-full" data-testid="check-out-table">
    <TableHead>
      <TableHeadCell>Game Title</TableHeadCell>
      <TableHeadCell>Borrower</TableHeadCell>
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
          <TableBodyCell>
            {#if gameStatus.patron}
              <div class="flex flex-col">
                <Badge large border color="rose" class="w-fit">{gameStatus.patron.name}</Badge>
              </div>
            {:else}
              <Badge large color="emerald" class="w-fit">Available</Badge>
            {/if}
          </TableBodyCell>
          <TableBodyCell>
            {#if !gameStatus.patron}
              <Button
                onclick={() => {
                  loanInitialFocus = 'patron';
                  handleCheckout(gameStatus.game);
                }}
                color="primary"
                size="sm"
              >
                Borrow
              </Button>
            {:else}
              <Button disabled color="alternative" size="sm">Borrow</Button>
            {/if}
          </TableBodyCell>
        </TableBodyRow>
      {/each}
      {#if games.length === 0}
        <TableBodyRow>
          <TableBodyCell
            colspan={3}
            class="px-6 py-12 text-center text-slate-500 dark:text-slate-400"
          >
            No games found.
          </TableBodyCell>
        </TableBodyRow>
      {/if}
      {#if loadingMore}
        <TableBodyRow>
          <TableBodyCell colspan={3} class="px-6 py-4 text-center">
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

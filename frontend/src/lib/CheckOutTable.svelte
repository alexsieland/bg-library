<script lang="ts">
  import { Table, TableBody, TableBodyCell, TableBodyRow, TableHead, TableHeadCell, Badge, Button } from 'flowbite-svelte';
  import SearchBar from './SearchBar.svelte';
  import BarcodeInput from './BarcodeInput.svelte';
  import { apiClient, type GameStatusList } from './api-client';
  import type { components } from '../generated/library-api';
  import { onMount } from 'svelte';
  import { toasts } from './toast-store';
  import { isBarcodeEnabled } from './config';
  import { barcodeScanner } from './barcodeScannerAction';

  import LoanModal from './LoanModal.svelte';

  let searchQuery = '';
  let loanModalOpen = false;
  let selectedGame: components["schemas"]["Game"] | null = null;

  let gameStatusList: GameStatusList = { games: [] };
  let error: string | null = null;
  let loading = true;
  let barcodeInputElement: HTMLInputElement;

  async function fetchGames() {
    console.log('fetchGames called with query:', searchQuery);
    loading = true;
    error = null;
    try {
      gameStatusList = await apiClient.listGames({ title: searchQuery || undefined });
      console.log('Fetched games:', gameStatusList.games.length);
    } catch (e) {
      const errorMessage = e instanceof Error ? e.message : 'An unknown error occurred';
      error = errorMessage;
      toasts.add(`Failed to load games: ${errorMessage}`, 'error');
      console.error('Error fetching games:', e);
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    console.log('CheckOutTable mounted');
    fetchGames();
  });

  function handleCheckout(game: components["schemas"]["Game"]) {
    console.log('Initiating checkout for game:', game.gameId);
    selectedGame = game;
    loanModalOpen = true;
  }

  function handleSearch(query: string) {
    console.log('handleSearch in CheckOutTable:', query);
    searchQuery = query;
    fetchGames();
  }

  function handleBarcodeFound(game: components["schemas"]["Game"]) {
    handleCheckout(game);
  }

  function handleBarcodeError(message: string) {
    toasts.add(message, 'error');
  }

  async function onScanComplete(barcode: string) {
    try {
      if (barcodeInputElement) {
        barcodeInputElement.focus();
        barcodeInputElement.value = barcode;
      }
      const result = await apiClient.getGameByBarcode(barcode);
      if (result.games.length > 1) {
        toasts.add('Barcode conflict handling not yet implemented. Please manually trigger the checkout.', 'error');
        return;
      }
      handleBarcodeFound(result.games[0]);
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to look up barcode';
      toasts.add(message, 'error');
    }
  }

  function handleWindowKeydown(event: KeyboardEvent) {
    // Alt+B: focus the barcode input
    if (event.altKey && event.key === 'b') {
      event.preventDefault();
      if (barcodeInputElement) barcodeInputElement.focus();
    }
  }
</script>

<svelte:window use:barcodeScanner={{ onScan: onScanComplete }} on:keydown={handleWindowKeydown} />

<div class="px-6 py-4 border-b border-slate-200 dark:border-slate-700 bg-slate-50/50 dark:bg-slate-800/50">
  <div class="flex items-center justify-between gap-4">
    <!-- Primary: search -->
    <div class="flex-1">
      <SearchBar bind:searchQuery placeholder="Search games..." onSearch={handleSearch} />
    </div>

    <!-- Secondary: barcode scanner, right-aligned and visually de-emphasised -->
    {#if isBarcodeEnabled()}
      <BarcodeInput
        bind:barcodeInputElement
        onGameFound={handleBarcodeFound}
        onError={handleBarcodeError}
      />
    {/if}
  </div>
</div>

<LoanModal bind:open={loanModalOpen} game={selectedGame} onLoanSuccess={fetchGames} />

{#if loading && gameStatusList.games.length === 0}
  <div class="p-8 text-center text-slate-500 dark:text-slate-400">Loading games...</div>
{:else if error}
  <div class="p-8 text-center text-rose-500">{error}</div>
{:else}
  <Table shadow hoverable={true} class="w-full">
    <TableHead>
      <TableHeadCell>Game Title</TableHeadCell>
      <TableHeadCell>Borrower</TableHeadCell>
      <TableHeadCell>Action</TableHeadCell>
    </TableHead>
    <TableBody class="divide-y">
      {#each gameStatusList.games as gameStatus (gameStatus.game.gameId)}
        <TableBodyRow>
          <TableBodyCell class="text-lg font-medium text-slate-900 dark:text-slate-100">{gameStatus.game.title}</TableBodyCell>
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
                onclick={() => handleCheckout(gameStatus.game)}
                color="primary"
                size="sm"
              >
                Loan
              </Button>
            {:else}
              <Button
                disabled
                color="alternative"
                size="sm"
              >
                Loan
              </Button>
            {/if}
          </TableBodyCell>
        </TableBodyRow>
      {/each}
      {#if gameStatusList.games.length === 0}
        <TableBodyRow>
          <TableBodyCell colspan={3} class="px-6 py-12 text-center text-slate-500 dark:text-slate-400">
            No games found.
          </TableBodyCell>
        </TableBodyRow>
      {/if}
    </TableBody>
  </Table>
{/if}

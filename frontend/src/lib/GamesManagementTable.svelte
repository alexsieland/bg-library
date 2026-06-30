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
    Badge,
    Spinner,
  } from 'flowbite-svelte';
  import { PlusOutline, ChevronDownOutline } from 'flowbite-svelte-icons';
  import { apiClient, type GameStatus, type Game } from './api-client';
  import { toasts } from './toast-store';
  import AddGameModal from './AddGameModal.svelte';
  import DeleteConfirmationPrompt from './DeleteConfirmationPrompt.svelte';
  import CsvUploadModal from './CsvUploadModal.svelte';
  import SearchBar from './SearchBar.svelte';
  import Debounce from './snippets/debounce.svelte';
  import { createPageScrollHandler } from './snippets/pageScrollLoader';
  import { onMount } from 'svelte';

  const PAGE_LIMIT = 100;

  let searchQuery = $state('');
  let games: GameStatus[] = $state([]);
  let loading = $state(true);
  let error: string | null = $state(null);
  let loadingMore = $state(false);
  let offset = $state(0);
  let hasMore = $state(true);

  let addGameModalOpen = $state(false);
  let deleteConfirmationOpen = $state(false);
  let csvUploadModalOpen = $state(false);
  let selectedGame: Game | null = $state(null);

  let cancelKey = 0;
  let lastValueRef = $state({ v: '' });

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

  function handleSearch(query: string) {
    searchQuery = query;
    fetchGames(0);
  }

  onMount(() => {
    fetchGames();
  });

  function openAddModal() {
    selectedGame = null;
    addGameModalOpen = true;
  }

  function openEditModal(game: Game) {
    selectedGame = game;
    addGameModalOpen = true;
  }

  function handleGameSaved() {
    toasts.add('Game saved successfully', 'success');
    fetchGames();
  }

  function openDeleteModal(game: Game) {
    selectedGame = game;
    deleteConfirmationOpen = true;
  }

  async function handleDeleteConfirmed() {
    if (!selectedGame) return;

    try {
      await apiClient.deleteGame(selectedGame.gameId);
      toasts.add(`Deleted ${selectedGame.title} from the library`, 'success');
      fetchGames();
    } catch (e) {
      const errorMessage = e instanceof Error ? e.message : 'An unknown error occurred';
      toasts.add(`Failed to delete game: ${errorMessage}`, 'error');
    }
  }

  async function handleBulkUpload(file: File) {
    return await apiClient.bulkAddGames(file);
  }

  const handlePageScroll = createPageScrollHandler({
    isLoading: () => loading,
    isLoadingMore: () => loadingMore,
    hasMore: () => hasMore,
    hasError: () => error !== null,
    loadMore: () => fetchGames(offset),
  });

  let filteredGames = $derived(games);
</script>

<svelte:window on:scroll={handlePageScroll} />

<AddGameModal
  bind:open={addGameModalOpen}
  gameId={selectedGame?.gameId ?? null}
  onGameSaved={handleGameSaved}
  onCancel={() => {
    selectedGame = null;
  }}
/>

<DeleteConfirmationPrompt
  bind:open={deleteConfirmationOpen}
  itemName={selectedGame?.title ?? 'this item'}
  onConfirm={handleDeleteConfirmed}
/>

<CsvUploadModal
  bind:open={csvUploadModalOpen}
  title="Bulk Add Games"
  successMessage={(count) => `Successfully imported ${count} game${count !== 1 ? 's' : ''}`}
  onUpload={handleBulkUpload}
  onSuccess={() => {
    fetchGames();
  }}
  onCancel={() => {
    // Handle cancel if needed
  }}
  exampleCsvHref="/example_games.csv"
/>

<!-- Search bar with debounce -->
<div
  class="border-b border-slate-200 bg-slate-50/50 px-6 py-4 dark:border-slate-700 dark:bg-slate-800/50"
>
  <div class="flex items-center justify-between gap-4">
    <div class="flex-1">
      <SearchBar bind:searchQuery placeholder="Search games by title..." onSearch={handleSearch} />
    </div>
    <div class="flex items-center gap-2">
      <Button onclick={openAddModal} color="primary" size="sm">
        <PlusOutline class="mr-2 h-3.5 w-3.5" />
        Add Game
      </Button>
      <Button color="alternative" size="sm">
        Actions
        <ChevronDownOutline class="ml-2 h-3 w-3" />
      </Button>
      <Dropdown simple class="w-44 divide-y divide-gray-100">
        <DropdownItem onclick={() => (csvUploadModalOpen = true)}>Bulk Add</DropdownItem>
      </Dropdown>
    </div>
  </div>
</div>

<Debounce value={searchQuery} onTrigger={handleSearch} delay={300} {lastValueRef} {cancelKey} />

<!-- Main table -->
<div class="relative overflow-hidden bg-white shadow-md sm:rounded-lg dark:bg-slate-800">
  {#if loading && games.length === 0}
    <div class="p-8 text-center text-slate-500 dark:text-slate-400">Loading games...</div>
  {:else if error}
    <div class="p-8 text-center text-rose-500">{error}</div>
  {:else}
    <Table shadow hoverable={true} class="w-full">
      <TableHead>
        <TableHeadCell class="px-4 py-3" scope="col">Game Title</TableHeadCell>
        <TableHeadCell class="px-4 py-3" scope="col">Action</TableHeadCell>
      </TableHead>

      <TableBody class="divide-y">
        {#if filteredGames.length === 0}
          <TableBodyRow>
            <TableBodyCell
              colspan={2}
              class="px-4 py-12 text-center text-slate-500 dark:text-slate-400"
            >
              No games found.
            </TableBodyCell>
          </TableBodyRow>
        {:else}
          {#each filteredGames as gameStatus (gameStatus.game.gameId)}
            <TableBodyRow>
              <TableBodyCell
                class="px-4 py-3 text-lg font-medium text-slate-900 dark:text-slate-100"
              >
                <div class="flex items-center gap-2">
                  {gameStatus.game.title}
                  {#if gameStatus.game.isPlayToWin}
                    <Badge color="sky">P2W</Badge>
                  {/if}
                </div>
              </TableBodyCell>
              <TableBodyCell class="px-4 py-3">
                <div class="flex gap-2">
                  <Button size="sm" color="yellow" onclick={() => openEditModal(gameStatus.game)}>
                    Edit
                  </Button>
                  <Button size="sm" color="red" onclick={() => openDeleteModal(gameStatus.game)}>
                    Delete
                  </Button>
                </div>
              </TableBodyCell>
            </TableBodyRow>
          {/each}
        {/if}
        {#if loadingMore}
          <TableBodyRow>
            <TableBodyCell colspan={2} class="px-4 py-4 text-center">
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
</div>

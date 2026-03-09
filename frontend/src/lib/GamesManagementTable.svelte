<script lang="ts">
  import {
    TableBody,
    TableBodyCell,
    TableBodyRow,
    TableHead,
    TableHeadCell,
    TableSearch,
    Button,
    Dropdown,
    DropdownItem,
    Badge,
  } from 'flowbite-svelte';
  import { PlusOutline, ChevronDownOutline } from 'flowbite-svelte-icons';
  import { apiClient, type GameStatusList, type Game } from './api-client';
  import { toasts } from './toast-store';
  import AddGameModal from './AddGameModal.svelte';
  import DeleteConfirmationPrompt from './DeleteConfirmationPrompt.svelte';
  import CsvUploadModal from './CsvUploadModal.svelte';
  import { onMount } from 'svelte';

  let divClass = 'bg-white dark:bg-slate-800 relative shadow-md sm:rounded-lg overflow-hidden';
  let innerDivClass =
    'flex flex-col md:flex-row items-center justify-between space-y-3 md:space-y-0 md:space-x-4 p-4';
  let searchClass = 'w-full md:w-1/2 relative';

  let searchTerm = $state('');
  let gameStatusList: GameStatusList = $state({ games: [] });
  let loading = $state(true);
  let error: string | null = $state(null);

  let addGameModalOpen = $state(false);
  let deleteConfirmationOpen = $state(false);
  let csvUploadModalOpen = $state(false);
  let selectedGame: Game | null = $state(null);

  async function fetchGames() {
    loading = true;
    error = null;
    try {
      gameStatusList = await apiClient.listGames({
        title: searchTerm || undefined,
      });
    } catch (e) {
      const errorMessage = e instanceof Error ? e.message : 'An unknown error occurred';
      error = errorMessage;
      toasts.add(`Failed to load games: ${errorMessage}`, 'error');
    } finally {
      loading = false;
    }
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

  let filteredGames = $derived(
    gameStatusList.games.filter(
      (gs) =>
        searchTerm === '' || gs.game.title.toLowerCase().indexOf(searchTerm.toLowerCase()) !== -1
    )
  );
</script>

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
/>

<div class={divClass}>
  <TableSearch
    placeholder="Search by game title"
    hoverable={true}
    bind:inputValue={searchTerm}
    {divClass}
    {innerDivClass}
    {searchClass}
  >
    {#snippet header()}
      <div
        class="flex w-full flex-shrink-0 flex-col items-stretch justify-end space-y-2 md:w-auto md:flex-row md:items-center md:space-y-0 md:space-x-3"
      >
        <Button onclick={openAddModal} color="primary">
          <PlusOutline class="mr-2 h-3.5 w-3.5" />
          Add Game
        </Button>
        <Button color="alternative">
          Actions
          <ChevronDownOutline class="ml-2 h-3 w-3 " />
        </Button>
        <Dropdown simple class="w-44 divide-y divide-gray-100">
          <DropdownItem onclick={() => (csvUploadModalOpen = true)}>Bulk Add</DropdownItem>
        </Dropdown>
      </div>
    {/snippet}

    <TableHead>
      <TableHeadCell class="px-4 py-3" scope="col">Game Title</TableHeadCell>
      <TableHeadCell class="px-4 py-3" scope="col">Actions</TableHeadCell>
    </TableHead>

    <TableBody class="divide-y">
      {#if loading}
        <TableBodyRow>
          <TableBodyCell
            colspan="2"
            class="px-4 py-12 text-center text-slate-500 dark:text-slate-400"
          >
            Loading games...
          </TableBodyCell>
        </TableBodyRow>
      {:else if error}
        <TableBodyRow>
          <TableBodyCell colspan="2" class="px-4 py-12 text-center text-rose-500">
            {error}
          </TableBodyCell>
        </TableBodyRow>
      {:else if filteredGames.length === 0}
        <TableBodyRow>
          <TableBodyCell
            colspan="2"
            class="px-4 py-12 text-center text-slate-500 dark:text-slate-400"
          >
            No games found.
          </TableBodyCell>
        </TableBodyRow>
      {:else}
        {#each filteredGames as gameStatus (gameStatus.game.gameId)}
          <TableBodyRow>
            <TableBodyCell class="px-4 py-3 text-lg font-medium text-slate-900 dark:text-slate-100">
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
    </TableBody>
  </TableSearch>
</div>

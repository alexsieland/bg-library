<script lang="ts">
  import { Table, TableBody, TableBodyCell, TableBodyRow, TableHead, TableHeadCell, Badge, Button } from 'flowbite-svelte';
  import SearchBar from './SearchBar.svelte';
  import { apiClient, type GameList } from './api-client';
  import { onMount } from 'svelte';
  
  let searchQuery = '';

  let gameList: GameList = { games: [] };
  let error: string | null = null;
  let loading = true;

  async function fetchGames() {
    console.log('fetchGames called with query:', searchQuery);
    loading = true;
    error = null;
    try {
      gameList = await apiClient.listGames({ title: searchQuery || undefined });
      console.log('Fetched games:', gameList.games.length);
    } catch (e) {
      error = e instanceof Error ? e.message : 'An unknown error occurred';
      console.error('Error fetching games:', e);
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    console.log('CheckOutTable mounted');
    fetchGames();
  });

  function handleCheckout(gameId: string) {
    console.log('Initiating checkout for game:', gameId);
  }

  function handleSearch(query: string) {
    console.log('handleSearch in CheckOutTable:', query);
    searchQuery = query;
    fetchGames();
  }
</script>

<div class="p-6 border-b border-slate-200 dark:border-slate-700 bg-slate-50/50 dark:bg-slate-800/50">
  <SearchBar bind:searchQuery placeholder="Search games..." onSearch={handleSearch} />
</div>

{#if loading && gameList.games.length === 0}
  <div class="p-8 text-center text-slate-500 dark:text-slate-400">Loading games...</div>
{:else if error}
  <div class="p-8 text-center text-red-500">{error}</div>
{:else}
  <Table shadow hoverable={true} class="w-full">
    <TableHead>
      <TableHeadCell>Game Title</TableHeadCell>
      <TableHeadCell>Borrower</TableHeadCell>
      <TableHeadCell>Action</TableHeadCell>
    </TableHead>
    <TableBody class="divide-y">
      {#each gameList.games as gameStatus (gameStatus.game.gameId)}
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
                onclick={() => handleCheckout(gameStatus.game.gameId)}
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
      {#if gameList.games.length === 0}
        <TableBodyRow>
          <TableBodyCell colspan={3} class="px-6 py-12 text-center text-slate-500 dark:text-slate-400">
            No games found.
          </TableBodyCell>
        </TableBodyRow>
      {/if}
    </TableBody>
  </Table>
{/if}

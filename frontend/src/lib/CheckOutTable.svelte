<script lang="ts">
  import { Table, TableBody, TableBodyCell, TableBodyRow, TableHead, TableHeadCell, Badge, Button } from 'flowbite-svelte';
  import type { components } from '../generated/library-api';
  import { onMount } from 'svelte';
  import { getBackendUrl } from './config';
  
  export let searchQuery = '';

  let gameList: components["schemas"]["GameList"] = { games: [] };
  let error: string | null = null;
  let loading = true;

  async function fetchGames() {
    loading = true;
    error = null;
    try {
      const url = new URL('/api/v1/library/games', getBackendUrl());
      if (searchQuery) {
        url.searchParams.append('title', searchQuery);
      }
      
      const response = await fetch(url.toString());
      if (!response.ok) {
        throw new Error(`Failed to fetch games: ${response.statusText}`);
      }
      gameList = await response.json();
    } catch (e) {
      error = e instanceof Error ? e.message : 'An unknown error occurred';
      console.error('Error fetching games:', e);
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    fetchGames();
  });

  $: if (searchQuery !== undefined) {
    fetchGames();
  }

  function handleCheckout(gameId: string) {
    console.log('Initiating checkout for game:', gameId);
  }
</script>

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

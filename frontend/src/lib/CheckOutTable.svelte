<script lang="ts">
  import { Table, TableBody, TableBodyCell, TableBodyRow, TableHead, TableHeadCell, Badge, Button } from 'flowbite-svelte';
  import type { components } from '../generated/library-api';
  
  export let games: components["schemas"]["GameList"] = { games: [] };

  function handleCheckout(gameId: string) {
    console.log('Initiating checkout for game:', gameId);
  }
</script>

<Table shadow hoverable={true} class="w-full">
  <TableHead>
    <TableHeadCell>Game Title</TableHeadCell>
    <TableHeadCell>Borrower</TableHeadCell>
    <TableHeadCell>Action</TableHeadCell>
  </TableHead>
  <TableBody class="divide-y">
    {#each games.games as gameStatus (gameStatus.game.gameId)}
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
              on:click={() => handleCheckout(gameStatus.game.gameId)}
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
    {#if games.games.length === 0}
      <TableBodyRow>
        <TableBodyCell colspan="3" class="px-6 py-12 text-center text-slate-500 dark:text-slate-400">
          No games found matching your search.
        </TableBodyCell>
      </TableBodyRow>
    {/if}
  </TableBody>
</Table>

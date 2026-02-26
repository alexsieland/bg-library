<script lang="ts">
  import { Table, TableBody, TableBodyCell, TableBodyRow, TableHead, TableHeadCell, Badge, Button } from 'flowbite-svelte';
  
  type GameStatus = {
    gameId: string;
    title: string;
    patronName?: string;
    checkedOutAt?: string;
  };

  export let games: GameStatus[] = [];

  function handleCheckout(gameId: string) {
    console.log('Initiating checkout for game:', gameId);
  }
</script>

<Table shadow hoverable={true} class="w-full">
  <TableHead>
    <TableHeadCell>Game Title</TableHeadCell>
    <TableHeadCell>Borrower</TableHeadCell>
    <TableHeadCell>Checked Out At</TableHeadCell>
    <TableHeadCell>Action</TableHeadCell>
  </TableHead>
  <TableBody class="divide-y">
    {#each games as game (game.gameId)}
      <TableBodyRow>
        <TableBodyCell class="text-lg font-medium text-slate-900 dark:text-slate-100">{game.title}</TableBodyCell>
        <TableBodyCell>
          {#if game.patronName}
            <div class="flex flex-col">
              <Badge large border color="rose" class="w-fit">{game.patronName}</Badge>
            </div>
          {:else}
            <Badge large color="emerald" class="w-fit">Available</Badge>
          {/if}
        </TableBodyCell>
        <TableBodyCell>
          {#if game.checkedOutAt}
            <span class="text-sm text-slate-500 dark:text-slate-400">{new Date(game.checkedOutAt).toLocaleString()}</span>
          {:else}
            <span class="text-sm text-slate-500 dark:text-slate-400">N/A</span>
          {/if}
        </TableBodyCell>
        <TableBodyCell>
          {#if !game.patronName}
            <Button
              on:click={() => handleCheckout(game.gameId)}
              color="primary"
              size="sm"
            >
              Check Out
            </Button>
          {:else}
            <Button
              disabled
              color="alternative"
              size="sm"
            >
              Check Out
            </Button>
          {/if}
        </TableBodyCell>
      </TableBodyRow>
    {/each}
    {#if games.length === 0}
      <TableBodyRow>
        <TableBodyCell colspan="3" class="px-6 py-12 text-center text-slate-500 dark:text-slate-400">
          No games found matching your search.
        </TableBodyCell>
      </TableBodyRow>
    {/if}
  </TableBody>
</Table>

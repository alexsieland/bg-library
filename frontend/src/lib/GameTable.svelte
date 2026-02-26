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

<Table shadow hoverable={true}>
  <TableHead>
    <TableHeadCell>Game Title</TableHeadCell>
    <TableHeadCell>Status</TableHeadCell>
    <TableHeadCell>Action</TableHeadCell>
  </TableHead>
  <TableBody class="divide-y">
    {#each games as game (game.gameId)}
      <TableBodyRow>
        <TableBodyCell class="text-lg font-medium text-slate-900 dark:text-slate-100">{game.title}</TableBodyCell>
        <TableBodyCell>
          {#if game.patronName}
            <div class="flex flex-col">
              <Badge color="red" class="w-fit">Checked Out</Badge>
              <span class="text-sm text-slate-500 dark:text-slate-400 mt-1">to {game.patronName}</span>
            </div>
          {:else}
            <Badge color="green" class="w-fit">Available</Badge>
          {/if}
        </TableBodyCell>
        <TableBodyCell>
          {#if !game.patronName}
            <Button
              on:click={() => handleCheckout(game.gameId)}
              color="blue"
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
              Unavailable
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

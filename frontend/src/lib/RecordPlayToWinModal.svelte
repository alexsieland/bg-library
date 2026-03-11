<script lang="ts">
  import { Button, Modal } from 'flowbite-svelte';
  import type { PlayToWinGame } from './api-client';

  export let open = false;
  export let playToWinGame: PlayToWinGame | null = null;
</script>

<Modal
  bind:open
  title={`Record Session: ${playToWinGame?.title || ''}`}
  size="md"
  autoclose={false}
  class="w-full"
>
  <div class="space-y-4" data-testid="ptw-record-modal">
    <p class="text-sm text-slate-600 dark:text-slate-300">
      Session recording is not implemented yet.
    </p>
    {#if playToWinGame}
      <p class="text-xs text-slate-500 dark:text-slate-400">
        Selected game: {playToWinGame.title}
      </p>
    {/if}
    <div class="flex justify-end">
      <Button color="alternative" onclick={() => (open = false)}>Close</Button>
    </div>
  </div>
</Modal>

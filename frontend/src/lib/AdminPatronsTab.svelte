<script lang="ts">
  import { Button } from 'flowbite-svelte';
  import AddPatronModal from './AddPatronModal.svelte';
  import { toasts } from './toast-store';
  import type { Patron } from './api-client';

  let addPatronModalOpen = false;

  function handlePatronCreated(patron: Patron) {
    toasts.add(`Successfully added patron: ${patron.name}`, 'success');
  }
</script>

<div class="p-6">
  <div class="mb-4 flex items-center justify-between">
    <h2 class="text-xl font-semibold text-slate-900 dark:text-slate-100">Patrons</h2>
    <Button onclick={() => (addPatronModalOpen = true)}>Add Patron</Button>
  </div>
</div>

<AddPatronModal bind:open={addPatronModalOpen} onPatronCreated={handlePatronCreated} />

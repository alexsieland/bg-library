<script lang="ts">
  import { TrashBinSolid } from 'flowbite-svelte-icons';
  import { Button, Modal } from 'flowbite-svelte';

  interface Props {
    open?: boolean;
    onCancel?: () => void;
  }

  let { open = $bindable(false), onCancel }: Props = $props();

  let isLoading = $state(false);

  async function handleConfirm() {
    isLoading = true;
    try {
      // TODO reset all unclaimed play to win raffles
    } finally {
      isLoading = false;
      open = false;
    }
  }

  function handleCancel() {
    onCancel?.();
    open = false;
  }
</script>

<Modal title="" bind:open autoclose={false} size="sm" class="w-full">
  <TrashBinSolid class="mx-auto h-11 w-11 shrink-0" />
  <p class="mb-4 text-center text-gray-500 dark:text-gray-300">
    This will reset the winners for all unclaimed play to win raffles. This action cannot be undone.
    Restart raffle?
  </p>
  <div class="flex items-center justify-center space-x-4">
    <Button color="light" onclick={handleCancel} disabled={isLoading}>No, cancel</Button>
    <Button color="rose" onclick={handleConfirm} disabled={isLoading}>Yes, I'm sure</Button>
  </div>
</Modal>

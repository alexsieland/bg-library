<script lang="ts">
  import { TrashBinSolid } from 'flowbite-svelte-icons';
  import { Button, Modal } from 'flowbite-svelte';

  interface Props {
    open?: boolean;
    itemName?: string;
    onConfirm?: () => void | Promise<void>;
    onCancel?: () => void;
  }

  let { open = $bindable(false), itemName = 'this item', onConfirm, onCancel }: Props = $props();

  let isLoading = $state(false);

  async function handleConfirm() {
    isLoading = true;
    try {
      await onConfirm?.();
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

<Modal
  title=""
  bind:open
  autoclose={false}
  size="sm"
  class="w-full"
  data-testid="delete-confirmation-modal"
>
  <TrashBinSolid class="mx-auto h-11 w-11 shrink-0" />
  <p class="mb-4 text-center text-gray-500 dark:text-gray-300">
    Are you sure you want to delete <span class="font-semibold">{itemName}</span>?
  </p>
  <div class="flex items-center justify-center space-x-4">
    <Button color="light" onclick={handleCancel} disabled={isLoading}>No, cancel</Button>
    <Button color="emerald" onclick={handleConfirm} disabled={isLoading}>Yes, I'm sure</Button>
  </div>
</Modal>

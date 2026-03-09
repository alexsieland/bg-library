<script lang="ts">
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

<Modal title="" bind:open autoclose={false} size="sm" class="w-full">
  <svg
    class="mx-auto mb-3.5 h-11 w-11 text-gray-400 dark:text-gray-500"
    aria-hidden="true"
    fill="currentColor"
    viewBox="0 0 20 20"
    xmlns="http://www.w3.org/2000/svg"
    ><path
      fill-rule="evenodd"
      d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z"
      clip-rule="evenodd"
    ></path></svg
  >
  <p class="mb-4 text-center text-gray-500 dark:text-gray-300">
    Are you sure you want to delete <span class="font-semibold">{itemName}</span>?
  </p>
  <div class="flex items-center justify-center space-x-4">
    <Button color="light" onclick={handleCancel} disabled={isLoading}>No, cancel</Button>
    <Button color="red" onclick={handleConfirm} disabled={isLoading}>Yes, I'm sure</Button>
  </div>
</Modal>

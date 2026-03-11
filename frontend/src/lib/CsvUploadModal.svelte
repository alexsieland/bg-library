<script lang="ts">
  import { Modal, Button, Fileupload, Helper, Spinner } from 'flowbite-svelte';
  import { toasts } from './toast-store';

  interface Props {
    open?: boolean;
    title?: string;
    successMessage?: (count: number) => string;
    onCancel?: () => void;
    onUpload: (file: File) => Promise<{ imported: number }>;
    onSuccess?: (count: number) => void;
    exampleCsvHref?: string;
  }

  let {
    open = $bindable(false),
    title = 'Upload CSV',
    successMessage = (count) => `Successfully imported ${count} item${count !== 1 ? 's' : ''}`,
    onCancel,
    onUpload,
    onSuccess,
    exampleCsvHref,
  }: Props = $props();

  let bulkUploadFile: FileList | undefined = $state();
  let loading = $state(false);
  let error: string | null = $state(null);

  async function handleUpload() {
    if (!bulkUploadFile || bulkUploadFile.length === 0) return;

    loading = true;
    error = null;

    try {
      const file = bulkUploadFile[0];
      const result = await onUpload(file);
      const successMsg = successMessage(result.imported);
      toasts.add(successMsg, 'success');
      bulkUploadFile = undefined;
      onSuccess?.(result.imported);
      open = false;
    } catch (e) {
      const errorMessage = e instanceof Error ? e.message : 'Failed to upload';
      error = errorMessage;
      toasts.add(`Failed to upload: ${errorMessage}`, 'error');
    } finally {
      loading = false;
    }
  }

  function handleCancel() {
    bulkUploadFile = undefined;
    error = null;
    onCancel?.();
    open = false;
  }
</script>

<Modal {title} bind:open autoclose={false} size="sm">
  <div class="space-y-4">
    <div>
      <Fileupload
        bind:files={bulkUploadFile}
        accept=".csv,text/csv,text/plain"
        disabled={loading}
      />
      <Helper class="mt-2">
        Upload a CSV file with one item per line. Maximum file size: 10MB.
      </Helper>
      {#if exampleCsvHref}
        <Helper class="mt-1">
          Download example file
          <a
            class="text-blue-600 underline hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300"
            href={exampleCsvHref}
            target="_blank"
            rel="noopener noreferrer"
            data-testid="csv-example-download-link"
          >
            here
          </a>
        </Helper>
      {/if}
    </div>

    {#if error}
      <p class="text-sm text-rose-500">{error}</p>
    {/if}

    <div class="flex justify-end gap-2 pt-2">
      <Button color="alternative" onclick={handleCancel} disabled={loading}>Cancel</Button>
      <Button
        onclick={handleUpload}
        disabled={loading || !bulkUploadFile || bulkUploadFile.length === 0}
      >
        {#if loading}
          <Spinner size="4" class="me-2" />
        {/if}
        Upload
      </Button>
    </div>
  </div>
</Modal>

<script lang="ts">
  import { Modal, Button, Input, Label, Spinner } from 'flowbite-svelte';
  import { apiClient, type Patron } from './api-client';
  import { toasts } from './toast-store';
  import { isBarcodeEnabled } from './config';
  export let open = false;
  export let patronId: string | null = null;
  export let initialName: string = '';
  export let onPatronCreated: (patron: Patron) => void = () => {};
  export let onCancel: () => void = () => {};
  let patronName = '';
  let patronBarcode = '';
  let barcodeLoading = false;
  let loading = false;
  // Re-apply initialName and reset fields every time the modal opens
  $: if (open && patronId) {
    loadPatronData();
  } else if (open) {
    patronName = initialName;
  } else {
    patronName = '';
    patronBarcode = '';
  }

  async function loadPatronData() {
    if (!patronId) return;
    try {
      const patron = await apiClient.getPatron(patronId);
      patronName = patron.name;
      patronBarcode = patron.barcode || '';
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to load game';
      toasts.add(`Failed to load game: ${message}`, 'error');
    }
  }

  async function handleSubmit() {
    if (!patronName.trim()) return;
    loading = true;
    try {
      const newPatron = await apiClient.addPatron({
        name: patronName.trim(),
        ...(patronBarcode.trim() ? { barcode: patronBarcode.trim() } : {}),
      });
      onPatronCreated(newPatron);
      open = false;
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to add patron';
      toasts.add(`Failed to add patron: ${message}`, 'error');
    } finally {
      loading = false;
    }
  }
  function handleCancel() {
    onCancel();
    open = false;
  }
  // Suppress Enter on the name input — HID barcode scanners use Enter as their
  // terminator and would trigger submission before the librarian can review.
  function handleNameKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter') {
      event.preventDefault();
    }
  }
  async function handleBarcodeKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter') {
      event.preventDefault();
      await handleBarcodeScan();
    }
  }
  async function handleBarcodeScan() {
    const value = patronBarcode.trim();
    if (!value) return;
    barcodeLoading = true;
    try {
      // Check if the barcode already belongs to an existing patron
      await apiClient.getPatronByBarcode(value);
      // If we get here, a patron already has this barcode
      toasts.add('A patron with this barcode already exists', 'error');
      patronBarcode = '';
    } catch (e) {
      // A 404 means the barcode is free to use — keep it in the field
      const message = e instanceof Error ? e.message : '';
      if (!message.toLowerCase().includes('not found') && !message.includes('404')) {
        toasts.add(`Barcode lookup failed: ${message}`, 'error');
        patronBarcode = '';
      }
      // else: barcode is available, leave it in the field
    } finally {
      barcodeLoading = false;
    }
  }
</script>

<Modal bind:open title="Add Patron" size="sm" autoclose={false}>
  <div class="space-y-4">
    <div>
      <Label for="addPatronName" class="mb-2">Patron Name</Label>
      <Input
        id="addPatronName"
        data-testid="add-patron-name-input"
        placeholder="Enter patron name"
        bind:value={patronName}
        onkeydown={handleNameKeydown}
        autocomplete="off"
        maxlength={100}
        disabled={loading}
      />
    </div>
    {#if isBarcodeEnabled()}
      <div>
        <Label for="addPatronBarcode" class="mb-2">
          <span
            class="text-xs font-medium tracking-wide text-slate-400 uppercase dark:text-slate-500"
          >
            Patron Barcode
          </span>
        </Label>
        <div class="relative">
          <Input
            id="addPatronBarcode"
            data-testid="add-patron-barcode-input"
            placeholder="Scan patron barcode…"
            bind:value={patronBarcode}
            onkeydown={handleBarcodeKeydown}
            autocomplete="off"
            maxlength={48}
            disabled={loading || barcodeLoading}
          />
          {#if barcodeLoading}
            <div class="pointer-events-none absolute inset-y-0 inset-e-0 flex items-center pe-3">
              <Spinner size="4" />
            </div>
          {/if}
        </div>
        <p class="mt-1 text-xs text-slate-400 dark:text-slate-500">
          Optional. Scan a convention badge or ID to associate a barcode with this patron.
        </p>
      </div>
    {/if}

    <div class="flex justify-end gap-2 pt-2">
      <Button color="alternative" onclick={handleCancel} disabled={loading}>Cancel</Button>
      <Button
        data-testid="add-patron-submit"
        onclick={handleSubmit}
        disabled={loading || barcodeLoading || !patronName.trim()}
      >
        {#if loading}
          <Spinner size="4" class="me-2" />
        {/if}
        Add Patron
      </Button>
    </div>
  </div>
</Modal>

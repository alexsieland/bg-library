<script lang="ts">
  import {
    TableBody,
    TableBodyCell,
    TableBodyRow,
    TableHead,
    TableHeadCell,
    Table,
    Button,
    Dropdown,
    DropdownItem,
  } from 'flowbite-svelte';
  import { PlusOutline, ChevronDownOutline } from 'flowbite-svelte-icons';
  import { apiClient, type Patron } from './api-client';
  import { toasts } from './toast-store';
  import AddPatronModal from './AddPatronModal.svelte';
  import DeleteConfirmationPrompt from './DeleteConfirmationPrompt.svelte';
  import CsvUploadModal from './CsvUploadModal.svelte';
  import SearchBar from './SearchBar.svelte';
  import Debounce from './snippets/debounce.svelte';
  import { onMount } from 'svelte';

  let searchQuery = $state('');
  let patronList: { patrons: Patron[] } = $state({ patrons: [] });
  let loading = $state(true);
  let error: string | null = $state(null);

  let addPatronModalOpen = $state(false);
  let deleteConfirmationOpen = $state(false);
  let csvUploadModalOpen = $state(false);
  let selectedPatron: Patron | null = $state(null);

  let cancelKey = 0;
  let lastValueRef = $state({ v: '' });

  async function fetchPatrons() {
    loading = true;
    error = null;
    try {
      patronList = await apiClient.listPatrons({
        name: searchQuery || undefined,
      });
    } catch (e) {
      const errorMessage = e instanceof Error ? e.message : 'An unknown error occurred';
      error = errorMessage;
      toasts.add(`Failed to load patrons: ${errorMessage}`, 'error');
    } finally {
      loading = false;
    }
  }

  function handleSearch(query: string) {
    searchQuery = query;
    fetchPatrons();
  }

  onMount(() => {
    fetchPatrons();
  });

  function openAddModal() {
    selectedPatron = null;
    addPatronModalOpen = true;
  }

  function openEditModal(patron: Patron) {
    selectedPatron = patron;
    addPatronModalOpen = true;
  }

  function handlePatronSaved() {
    toasts.add('Patron saved successfully', 'success');
    fetchPatrons();
  }

  function openDeleteModal(patron: Patron) {
    selectedPatron = patron;
    deleteConfirmationOpen = true;
  }

  async function handleDeleteConfirmed() {
    if (!selectedPatron) return;

    try {
      await apiClient.deletePatron(selectedPatron.patronId);
      toasts.add(`Deleted ${selectedPatron.name} from the library`, 'success');
      fetchPatrons();
    } catch (e) {
      const errorMessage = e instanceof Error ? e.message : 'An unknown error occurred';
      toasts.add(`Failed to delete patron: ${errorMessage}`, 'error');
    }
  }

  async function handleBulkUpload(file: File) {
    return await apiClient.bulkAddPatrons(file);
  }

  let filteredPatrons = $derived(patronList.patrons);
</script>

<AddPatronModal
  bind:open={addPatronModalOpen}
  patronId={selectedPatron?.patronId ?? null}
  initialName={selectedPatron?.name ?? ''}
  onPatronCreated={handlePatronSaved}
  onCancel={() => {
    selectedPatron = null;
  }}
/>

<DeleteConfirmationPrompt
  bind:open={deleteConfirmationOpen}
  itemName={selectedPatron?.name ?? 'this item'}
  onConfirm={handleDeleteConfirmed}
/>

<CsvUploadModal
  bind:open={csvUploadModalOpen}
  title="Bulk Add Patrons"
  successMessage={(count) => `Successfully imported ${count} patron${count !== 1 ? 's' : ''}`}
  onUpload={handleBulkUpload}
  onSuccess={() => {
    fetchPatrons();
  }}
  onCancel={() => {
    // Handle cancel if needed
  }}
  exampleCsvHref="/example_patrons.csv"
/>

<!-- Search bar with debounce -->
<div
  class="border-b border-slate-200 bg-slate-50/50 px-6 py-4 dark:border-slate-700 dark:bg-slate-800/50"
>
  <div class="flex items-center justify-between gap-4">
    <div class="flex-1">
      <SearchBar bind:searchQuery placeholder="Search patrons by name..." onSearch={handleSearch} />
    </div>
    <div class="flex items-center gap-2">
      <Button onclick={openAddModal} color="primary" size="sm">
        <PlusOutline class="mr-2 h-3.5 w-3.5" />
        Add Patron
      </Button>
      <Button color="alternative" size="sm">
        Actions
        <ChevronDownOutline class="ml-2 h-3 w-3" />
      </Button>
      <Dropdown simple class="w-44 divide-y divide-gray-100">
        <DropdownItem onclick={() => (csvUploadModalOpen = true)}>Bulk Add</DropdownItem>
      </Dropdown>
    </div>
  </div>
</div>

<Debounce value={searchQuery} onTrigger={handleSearch} delay={300} {lastValueRef} {cancelKey} />

<!-- Main table -->
<div class="relative overflow-hidden bg-white shadow-md sm:rounded-lg dark:bg-slate-800">
  {#if loading && patronList.patrons.length === 0}
    <div class="p-8 text-center text-slate-500 dark:text-slate-400">Loading patrons...</div>
  {:else if error}
    <div class="p-8 text-center text-rose-500">{error}</div>
  {:else}
    <Table shadow hoverable={true} class="w-full">
      <TableHead>
        <TableHeadCell class="px-4 py-3" scope="col">Patron Name</TableHeadCell>
        <TableHeadCell class="px-4 py-3" scope="col">Action</TableHeadCell>
      </TableHead>

      <TableBody class="divide-y">
        {#if filteredPatrons.length === 0}
          <TableBodyRow>
            <TableBodyCell
              colspan={2}
              class="px-4 py-12 text-center text-slate-500 dark:text-slate-400"
            >
              No patrons found.
            </TableBodyCell>
          </TableBodyRow>
        {:else}
          {#each filteredPatrons as patron (patron.patronId)}
            <TableBodyRow>
              <TableBodyCell
                class="px-4 py-3 text-lg font-medium text-slate-900 dark:text-slate-100"
              >
                {patron.name}
              </TableBodyCell>
              <TableBodyCell class="px-4 py-3">
                <div class="flex gap-2">
                  <Button size="sm" color="yellow" onclick={() => openEditModal(patron)}>
                    Edit
                  </Button>
                  <Button size="sm" color="red" onclick={() => openDeleteModal(patron)}>
                    Delete
                  </Button>
                </div>
              </TableBodyCell>
            </TableBodyRow>
          {/each}
        {/if}
      </TableBody>
    </Table>
  {/if}
</div>

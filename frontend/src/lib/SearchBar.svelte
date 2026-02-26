<script lang="ts">
  import { Search } from "flowbite-svelte";

  export let searchQuery: string = '';
  export let placeholder: string = "Search...";
  export let onSearch: (query: string) => void = () => {};

  import Debounce from './snippets/debounce.svelte';

  let cancelKey = 0;
  const lastValueRef = { v: searchQuery };
</script>

<!-- Render the extracted debounce logic as an invisible helper component -->
<Debounce value={searchQuery} onTrigger={onSearch} minLength={3} {lastValueRef} {cancelKey} />

<div class="flex flex-col md:flex-row md:items-center space-y-4 md:space-y-0 md:space-x-4">
  <div class="relative flex-grow">
    <Search {placeholder} bind:value={searchQuery}  />
  </div>
</div>

<script lang="ts">
  import { Search, Button } from "flowbite-svelte";
  import { SearchOutline } from 'flowbite-svelte-icons';
  
  export let searchQuery: string = '';
  export let placeholder: string = "Search...";
</script>

<div class="flex flex-col md:flex-row md:items-center space-y-4 md:space-y-0 md:space-x-4">
  <div class="relative flex-grow">
    <Search {placeholder} bind:value={searchQuery}>
      <Button color="primary" class="me-1">Search</Button>
    </Search>
  </div>
</div>

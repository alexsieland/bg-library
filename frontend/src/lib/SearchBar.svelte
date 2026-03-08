<script lang="ts">
  import { Search } from 'flowbite-svelte';

  export let searchQuery: string = '';
  export let placeholder: string = 'Search...';
  export let onSearch: (query: string) => void = () => {};

  import Debounce from './snippets/debounce.svelte';

  let cancelKey = 0;
  const lastValueRef = { v: searchQuery };
  let searchInput: HTMLInputElement;

  export let searchInputElement: HTMLInputElement | undefined = undefined;
  $: searchInputElement = searchInput;

  function handleWindowKeydown(event: KeyboardEvent) {
    // Alt+S: focus the search bar
    if (event.altKey && event.key === 's') {
      event.preventDefault();
      if (searchInput) searchInput.focus();
      return;
    }

    // If we're already focusing an input/textarea/editable element, don't do anything
    const activeElement = document.activeElement;
    const isEditing =
      activeElement &&
      (activeElement.tagName === 'INPUT' ||
        activeElement.tagName === 'TEXTAREA' ||
        (activeElement as HTMLElement).isContentEditable);

    if (isEditing) return;

    // Check if it's a single printable character (length 1) and not a modifier key
    // Skip if Alt is held — Alt+B should go to barcode, not search
    if (event.key.length === 1 && !event.ctrlKey && !event.metaKey && !event.altKey) {
      if (searchInput) {
        searchInput.focus();
      }
    }
  }
  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter') {
      cancelKey++;
      lastValueRef.v = searchQuery;
      onSearch(searchQuery);
    }
  }

  function handleSearchClick() {
    cancelKey++;
    lastValueRef.v = searchQuery;
    onSearch(searchQuery);
  }
</script>

<svelte:window on:keydown={handleWindowKeydown} />

<!-- Render the extracted debounce logic as an invisible helper component -->
<Debounce
  value={searchQuery}
  onTrigger={onSearch}
  delay={300}
  minLength={0}
  {lastValueRef}
  {cancelKey}
/>

<div class="flex flex-col space-y-4 md:flex-row md:items-center md:space-y-0 md:space-x-4">
  <div class="relative flex-grow">
    <Search
      {placeholder}
      bind:value={searchQuery}
      bind:elementRef={searchInput}
      onkeydown={handleKeydown}
    />
    <button
      type="button"
      class="absolute inset-y-0 end-0 flex cursor-pointer items-center pe-3"
      onclick={handleSearchClick}
      aria-label="Search"
    >
      <svg
        aria-hidden="true"
        class="h-5 w-5 text-gray-500 dark:text-gray-400"
        fill="none"
        viewBox="0 0 20 20"
        xmlns="http://www.w3.org/2000/svg"
      >
        <path
          d="m19 19-4-4m0-7A7 7 0 1 1 1 8a7 7 0 0 1 14 0Z"
          stroke="currentColor"
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
        ></path>
      </svg>
    </button>
  </div>
</div>

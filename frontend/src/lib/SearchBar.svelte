<script lang="ts">
  import { Search } from "flowbite-svelte";

  export let searchQuery: string = '';
  export let placeholder: string = "Search...";
  export let onSearch: (query: string) => void = () => {};

  import Debounce from './snippets/debounce.svelte';

  let cancelKey = 0;
  const lastValueRef = { v: searchQuery };
  let searchInput: HTMLInputElement;

  function handleWindowKeydown(event: KeyboardEvent) {
    // If we're already focusing an input/textarea/editable element, don't do anything
    const activeElement = document.activeElement;
    const isEditing = activeElement && (
      activeElement.tagName === 'INPUT' || 
      activeElement.tagName === 'TEXTAREA' || 
      (activeElement as HTMLElement).isContentEditable
    );
    
    if (isEditing) return;

    // Check if it's a single printable character (length 1) and not a modifier key
    if (event.key.length === 1 && !event.ctrlKey && !event.metaKey && !event.altKey) {
      if (searchInput) {
        searchInput.focus();
        // Since we're focusing, the character might not be automatically added if we're too late 
        // in the event cycle, but usually it is.
        // Actually, if we focus now, the current keydown event will still result in the character
        // being entered into the now-focused input by the browser.
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
<Debounce value={searchQuery} onTrigger={onSearch} delay={300} minLength={0} {lastValueRef} {cancelKey} />

<div class="flex flex-col md:flex-row md:items-center space-y-4 md:space-y-0 md:space-x-4">
  <div class="relative flex-grow">
    <Search 
      {placeholder} 
      bind:value={searchQuery} 
      bind:elementRef={searchInput} 
      onkeydown={handleKeydown}
    />
    <button 
      type="button"
      class="absolute inset-y-0 end-0 flex items-center pe-3 cursor-pointer"
      onclick={handleSearchClick}
      aria-label="Search"
    >
      <svg
        aria-hidden="true"
        class="text-gray-500 dark:text-gray-400 w-5 h-5"
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
        />
      </svg>
    </button>
  </div>
</div>

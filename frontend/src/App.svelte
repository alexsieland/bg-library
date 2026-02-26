<script lang="ts">
  import { onMount } from 'svelte';
  import AppNavbar from './lib/AppNavbar.svelte';
  import SearchBar from './lib/SearchBar.svelte';
  import CheckOutTable from './lib/CheckOutTable.svelte';
  import type { components } from './generated/library-api';
  import { getBackendUrl } from './lib/config';

  let searchQuery = '';
  let showOnlyAvailable = false;

  let games: components["schemas"]["GameList"] = { games: [] };
  let error: string | null = null;
  let loading = true;

  async function fetchGames() {
    loading = true;
    error = null;
    try {
      const url = new URL('/api/v1/library/games', getBackendUrl());
      // We don't want to filter by checkedOut status as per requirements
      if (searchQuery) {
        url.searchParams.append('title', searchQuery);
      }
      
      const response = await fetch(url.toString());
      if (!response.ok) {
        throw new Error(`Failed to fetch games: ${response.statusText}`);
      }
      games = await response.json();
    } catch (e) {
      error = e instanceof Error ? e.message : 'An unknown error occurred';
      console.error('Error fetching games:', e);
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    fetchGames();
  });

  $: if (searchQuery !== undefined) {
    fetchGames();
  }

  $: filteredGames = {
    games: games.games.filter(gameStatus => {
      const matchesAvailable = showOnlyAvailable ? !gameStatus.patron : true;
      return matchesAvailable;
    })
  };
</script>

<div class="min-h-screen bg-slate-50 dark:bg-slate-900 transition-colors">
  <AppNavbar activeTab="checkout" />

  <main class="container mx-auto px-4 py-8 space-y-8">
    <div class="space-y-2">
      <h1 class="text-3xl text-center font-bold text-slate-900 dark:text-slate-100 ">Checkout Games</h1>
    </div>

    <SearchBar bind:searchQuery placeholder="Search games..." />

    <div class="bg-white dark:bg-slate-800 rounded-xl shadow-lg border border-slate-200 dark:border-slate-700 overflow-hidden">
      {#if loading && games.games.length === 0}
        <div class="p-8 text-center text-slate-500 dark:text-slate-400">Loading games...</div>
      {:else if error}
        <div class="p-8 text-center text-red-500">{error}</div>
      {:else}
        <CheckOutTable games={filteredGames} />
      {/if}
    </div>
  </main>
</div>

<style>
  :global(body) {
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
  }
</style>

<script lang="ts">
  import AppNavbar from './lib/AppNavbar.svelte';
  import SearchBar from './lib/SearchBar.svelte';
  import CheckOutTable from './lib/CheckOutTable.svelte';

  let searchQuery = '';
</script>

<div class="min-h-screen bg-slate-50 dark:bg-slate-900 transition-colors">
  <AppNavbar activeTab="checkout" />

  <main class="container mx-auto px-4 py-8 space-y-8">
    <div class="space-y-2">
      <h1 class="text-3xl text-center font-bold text-slate-900 dark:text-slate-100 ">Checkout Games</h1>
    </div>

    <SearchBar bind:searchQuery placeholder="Search games..." />

    <div class="bg-white dark:bg-slate-800 rounded-xl shadow-lg border border-slate-200 dark:border-slate-700 overflow-hidden">
      <CheckOutTable {searchQuery} />
    </div>
  </main>
</div>

<style>
  :global(body) {
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
  }
</style>

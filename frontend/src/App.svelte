<script lang="ts">
  import AppNavbar from './lib/AppNavbar.svelte';
  import CheckOutTable from './lib/CheckOutTable.svelte';
</script>

<div class="min-h-screen bg-slate-50 dark:bg-slate-900 transition-colors">
  <AppNavbar activeTab="checkout" />

  <main class="container mx-auto px-4 py-8 space-y-8">
    <div class="space-y-2">
      <h1 class="text-3xl text-center font-bold text-slate-900 dark:text-slate-100 ">Checkout Games</h1>
    </div>

    <div class="bg-white dark:bg-slate-800 rounded-xl shadow-lg border border-slate-200 dark:border-slate-700 overflow-hidden">
      <CheckOutTable />
    </div>
  </main>
</div>

<style>
  :global(body) {
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
  }
</style>

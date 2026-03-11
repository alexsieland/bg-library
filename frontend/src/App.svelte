<script lang="ts">
  import AppNavbar from './lib/AppNavbar.svelte';
  import CheckOutTable from './lib/CheckOutTable.svelte';
  import CheckInTable from './lib/CheckInTable.svelte';
  import AdminView from './lib/AdminView.svelte';
  import ToastContainer from './lib/ToastContainer.svelte';
  import PlayToWinTable from './lib/PlayToWinTable.svelte';

  export let activeTab: 'checkout' | 'checkin' | 'ptw' | 'admin' = 'checkout';

  function handleTabChange(tab: 'checkout' | 'checkin' | 'ptw' | 'admin') {
    activeTab = tab;
  }
</script>

<div class="min-h-screen bg-slate-50 transition-colors dark:bg-slate-900">
  <AppNavbar {activeTab} onTabChange={handleTabChange} />

  <main class="container mx-auto space-y-8 px-4 py-8">
    <div class="space-y-2">
      <h1 class="text-center text-3xl font-bold text-slate-900 dark:text-slate-100">
        {#if activeTab === 'checkout'}
          Checkout Games
        {:else if activeTab === 'checkin'}
          Check In Games
        {:else if activeTab === 'ptw'}
          Play To Win
        {:else}
          Admin
        {/if}
      </h1>
    </div>

    <div
      class="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-lg dark:border-slate-700 dark:bg-slate-800"
    >
      {#if activeTab === 'checkout'}
        <CheckOutTable />
      {:else if activeTab === 'checkin'}
        <CheckInTable />
      {:else if activeTab === 'ptw'}
        <PlayToWinTable />
      {:else}
        <AdminView />
      {/if}
    </div>
  </main>
</div>

<ToastContainer />

<style>
  :global(body) {
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
  }
</style>

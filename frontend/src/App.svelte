<script lang="ts">
  import AppNavbar from './lib/AppNavbar.svelte';
  import SearchBar from './lib/SearchBar.svelte';
  import GameTable from './lib/GameTable.svelte';

  let searchQuery = '';
  let showOnlyAvailable = false;

  // Dummy data representing the "Game Discovery and Search" integration test data
  const dummyGames = [
    { gameId: '1', title: 'Catan' },
    { gameId: '2', title: 'Catan: Seafarers' },
    { gameId: '3', title: 'Gloomhaven', patronName: 'Alice Smith' },
    { gameId: '4', title: 'Everdell' },
    { gameId: '5', title: 'Bärenpark' },
    { gameId: '6', title: 'Root', patronName: 'Bob Smith' },
    { gameId: '7', title: 'Spirit Island' },
    { gameId: '8', title: 'Wingspan' },
    { gameId: '9', title: 'Terraforming Mars', patronName: 'Charlie Brown' },
    { gameId: '10', title: 'Azul' }
  ];

  $: filteredGames = dummyGames.filter(game => {
    const matchesSearch = game.title.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesAvailable = showOnlyAvailable ? !game.patronName : true;
    return matchesSearch && matchesAvailable;
  });
</script>

<div class="min-h-screen bg-slate-50 dark:bg-slate-900 transition-colors">
  <AppNavbar activeTab="checkout" />

  <main class="container mx-auto px-4 py-8 space-y-8">
    <div class="space-y-2">
      <h1 class="text-3xl text-center font-bold text-slate-900 dark:text-slate-100 ">Checkout Games</h1>
    </div>

    <SearchBar bind:searchQuery placeholder="Search games..." />

    <div class="bg-white dark:bg-slate-800 rounded-xl shadow-lg border border-slate-200 dark:border-slate-700 overflow-hidden">
      <GameTable games={filteredGames} />
    </div>
  </main>
</div>

<style>
  :global(body) {
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
  }
</style>

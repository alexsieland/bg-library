<script lang="ts">
  import { Tabs, TabItem } from 'flowbite-svelte';
  import AdminGamesTab from './AdminGamesTab.svelte';
  import AdminPatronsTab from './AdminPatronsTab.svelte';
  import AdminPlayToWinTab from './AdminPlayToWinTab.svelte';
</script>

<Tabs style="underline" class="px-6 pt-4" data-testid="admin-view-tabs">
  <TabItem open title="Games">
    <AdminGamesTab />
  </TabItem>
  <TabItem title="Patrons">
    <AdminPatronsTab />
  </TabItem>
  <TabItem title="Play To Win">
    <AdminPlayToWinTab />
  </TabItem>
</Tabs>

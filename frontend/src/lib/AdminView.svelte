<script lang="ts">
  import { Tabs, TabItem } from 'flowbite-svelte';
  import AdminGamesTab from './AdminGamesTab.svelte';
  import AdminPatronsTab from './AdminPatronsTab.svelte';
</script>

<Tabs style="underline" class="px-6 pt-4" data-testid="admin-view-tabs">
  <TabItem open title="Games">
    <AdminGamesTab />
  </TabItem>
  <TabItem title="Patrons">
    <AdminPatronsTab />
  </TabItem>
  <TabItem title="Play-to-Win">
    <div class="flex h-full flex-col items-center justify-center">
      <p class="text-gray-500">Coming soon...</p>
    </div>
  </TabItem>
</Tabs>

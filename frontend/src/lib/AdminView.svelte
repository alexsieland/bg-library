<script lang="ts">
  import { Tabs, TabItem } from 'flowbite-svelte';
  import AdminGamesTab from './AdminGamesTab.svelte';
  import AdminPatronsTab from './AdminPatronsTab.svelte';
</script>

<Tabs style="underline" class="px-6 pt-4">
  <TabItem open title="Games">
    <AdminGamesTab />
  </TabItem>
  <TabItem title="Patrons">
    <AdminPatronsTab />
  </TabItem>
</Tabs>

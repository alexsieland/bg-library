<script lang="ts">
  import { Navbar, NavBrand, NavHamburger, NavUl, NavLi } from 'flowbite-svelte';
  export let activeTab: 'checkout' | 'checkin' | 'admin' = 'checkout';
</script>

<Navbar fluid={true} class="bg-neutral-dark-bg text-white shadow-md">
  <NavBrand href="/">
    <span class="self-center whitespace-nowrap text-2xl font-bold tracking-tight text-white">BG Library</span>
  </NavBrand>
  <NavHamburger />
  <NavUl {activeTab} class="text-white">
    <NavLi href="#" active={activeTab === 'checkout'}>Check Out</NavLi>
    <NavLi href="#" active={activeTab === 'checkin'}>Check In</NavLi>
    <NavLi href="#" active={activeTab === 'admin'}>Admin</NavLi>
  </NavUl>
</Navbar>

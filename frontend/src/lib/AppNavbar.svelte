<script lang="ts">
  import { Navbar, NavBrand, NavHamburger, NavUl, NavLi } from 'flowbite-svelte';
  export let activeTab: 'checkout' | 'checkin' | 'admin' = 'checkout';
</script>

<Navbar fluid={true} class="bg-slate-800 text-white shadow-md">
  <NavBrand href="/">
    <span class="self-center whitespace-nowrap text-2xl font-bold tracking-tight text-white">BG Library</span>
  </NavBrand>
  <NavHamburger />
  <NavUl {activeTab} class="text-white">
    <NavLi href="#" active={activeTab === 'checkout'} activeClass="text-white bg-blue-700 md:bg-transparent md:text-blue-300">Check Out</NavLi>
    <NavLi href="#" active={activeTab === 'checkin'} activeClass="text-white bg-blue-700 md:bg-transparent md:text-blue-300">Check In</NavLi>
    <NavLi href="#" active={activeTab === 'admin'} activeClass="text-white bg-blue-700 md:bg-transparent md:text-blue-300">Admin</NavLi>
  </NavUl>
</Navbar>

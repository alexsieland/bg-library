<script lang="ts">
  import { Navbar, NavBrand, NavHamburger, NavUl, NavLi } from 'flowbite-svelte';
  export let activeTab: 'checkout' | 'checkin' | 'admin' = 'checkout';
  export let onTabChange: (tab: 'checkout' | 'checkin' | 'admin') => void = () => {};
</script>

<Navbar fluid={true} class="bg-slate-800 text-white shadow-md">
  <NavBrand href="/">
    <span class="self-center text-2xl font-bold tracking-tight whitespace-nowrap text-white"
      >BG Library</span
    >
  </NavBrand>
  <NavHamburger />
  <NavUl class="text-white">
    <NavLi
      href="#"
      activeClass="text-blue-300 bg-blue-700 md:bg-transparent md:text-blue-300"
      nonActiveClass="text-white hover:text-blue-200"
      onclick={(e) => {
        e.preventDefault();
        onTabChange('checkout');
      }}
      class="cursor-pointer {activeTab === 'checkout' ? 'text-blue-300' : 'text-white'}"
    >
      Check Out
    </NavLi>
    <NavLi
      href="#"
      activeClass="text-blue-300 bg-blue-700 md:bg-transparent md:text-blue-300"
      nonActiveClass="text-white hover:text-blue-200"
      onclick={(e) => {
        e.preventDefault();
        onTabChange('checkin');
      }}
      class="cursor-pointer {activeTab === 'checkin' ? 'text-blue-300' : 'text-white'}"
    >
      Check In
    </NavLi>
    <NavLi
      href="#"
      activeClass="text-blue-300 bg-blue-700 md:bg-transparent md:text-blue-300"
      nonActiveClass="text-white hover:text-blue-200"
      onclick={(e) => {
        e.preventDefault();
        onTabChange('admin');
      }}
      class="cursor-pointer {activeTab === 'admin' ? 'text-blue-300' : 'text-white'}"
    >
      Admin
    </NavLi>
  </NavUl>
</Navbar>

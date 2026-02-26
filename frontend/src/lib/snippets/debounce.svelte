<script lang="ts">
  import { onDestroy } from 'svelte';
  export let value: string;
  export let onTrigger: (v: string) => void;
  export let delay: number = 300;
  export let lastValueRef: { v: string };
  export let cancelKey: number = 0;

  let timer: ReturnType<typeof setTimeout>;

  $: {
    // include cancelKey as a dependency so changing it clears any pending timer
    void cancelKey;

    clearTimeout(timer);
    if (value === lastValueRef.v) {
      // no change since last trigger
    } else {
      timer = setTimeout(() => {
        lastValueRef.v = value;
        onTrigger(value);
      }, delay);
    }
  }

  onDestroy(() => clearTimeout(timer));
</script>

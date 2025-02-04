<script lang="ts">
  export let title: string;
  export let items: string[];
  export let selectedItem: string;
  export let onSelect: (item: string) => void;
</script>

<div class="w-48 flex-shrink-0 bg-gray-800 p-4 text-white">
  <h2 class="mb-4 text-3xl font-bold">{title}</h2>
  <ul class="space-y-2">
    {#each items as item}
      <li>
        <button
          class="w-full rounded px-2 py-1 text-lg text-left hover:bg-gray-700"
          class:bg-gray-700={selectedItem === item}
          on:click={() => onSelect(item)}
        >
          {item}
        </button>
      </li>
    {/each}
  </ul>
</div> 
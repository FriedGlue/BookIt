<script lang="ts">
	export let title: string;
	export let items: string[];
	export let selectedItem: string;
	export let onSelect: (item: string) => void;

	let isSidebarOpen = false;

	function toggleSidebar() {
		isSidebarOpen = !isSidebarOpen;
	}

	function handleSelect(item: string) {
		onSelect(item);
		if (window.innerWidth < 1024) { // Close sidebar on selection for mobile
			isSidebarOpen = false;
		}
	}
</script>

<!-- Mobile Toggle Button - Fixed Position -->
<button
	class="fixed left-4 top-20 z-40 rounded bg-gray-800 p-2 text-white shadow-lg hover:bg-gray-700 lg:hidden"
	on:click={toggleSidebar}
	aria-label="Toggle sidebar"
>
	<svg class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
		{#if !isSidebarOpen}
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
		{:else}
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
		{/if}
	</svg>
</button>

<!-- Backdrop for mobile -->
{#if isSidebarOpen}
	<div 
		class="fixed inset-0 z-30 bg-black bg-opacity-50 transition-opacity lg:hidden"
		on:click={toggleSidebar}
	></div>
{/if}

<!-- Sidebar -->
<div
	class={`fixed inset-y-0 left-0 z-40 w-64 transform bg-gray-800 p-4 text-white transition-transform duration-300 ease-in-out lg:static lg:translate-x-0 ${
		isSidebarOpen ? 'translate-x-0' : '-translate-x-full'
	}`}
>
	<div class="flex items-center justify-between lg:justify-center">
		<h2 class="text-3xl font-bold">{title}</h2>
		<!-- Close button for mobile -->
		<button
			class="rounded p-2 hover:bg-gray-700 lg:hidden"
			on:click={toggleSidebar}
			aria-label="Close sidebar"
		>
			<svg class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
			</svg>
		</button>
	</div>

	<ul class="mt-8 space-y-2">
		{#each items as item}
			<li>
				<button
					class="w-full rounded px-4 py-2 text-left text-lg transition-colors duration-200 hover:bg-gray-700"
					class:bg-gray-700={selectedItem === item}
					on:click={() => handleSelect(item)}
				>
					{item}
				</button>
			</li>
		{/each}
	</ul>
</div>

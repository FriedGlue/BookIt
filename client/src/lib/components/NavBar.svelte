<script lang="ts">
	import { AuthService } from '$lib/services/authService';
	import { goto } from '$app/navigation';

	let searchQuery = '';
	let searchResults: any[] = [];
	let isSearching = false;
	let showSearchResults = false;
	let isAddingToList = false;
	let toBeReadList: any[] = [];

	const authService = new AuthService();

	export let authenticated: boolean;

	// ---- Move "searchBooks" to call your local route:
	async function handleSearch() {
		if (!searchQuery.trim()) {
			searchResults = [];
			showSearchResults = false;
			return;
		}

		try {
			isSearching = true;
			const res = await fetch(`/api/books/search?q=${encodeURIComponent(searchQuery)}`);
			if (!res.ok) {
				console.error('Error searching books:', await res.text());
				searchResults = [];
				return;
			}
			searchResults = await res.json();
			showSearchResults = true;
		} catch (error) {
			console.error('Error searching books:', error);
			searchResults = [];
		} finally {
			isSearching = false;
		}
	}

	// ---- Move "addToList" to call your local route:
	async function addToList(bookId: string) {
		if (isAddingToList) return;
		try {
			isAddingToList = true;
			const res = await fetch('/api/books/add', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({ bookId, listType: 'toBeRead' })
			});

			if (!res.ok) {
				console.error('Error adding book to list:', await res.text());
				return;
			}

			showSearchResults = false;
			searchQuery = '';
		} catch (error) {
			console.error('Error adding book to list:', error);
		} finally {
			isAddingToList = false;
		}
	}

	async function handleLogout() {
		authService.logout();
		goto('/login');
	}

	function handleClickOutside(event: MouseEvent) {
		const target = event.target as HTMLElement;
		if (!target.closest('.search-container')) {
			showSearchResults = false;
		}
	}
</script>

<nav class="flex flex-wrap items-center justify-between bg-blue-500 p-6">
	<div class="flex items-center space-x-8">
		<div class="flex flex-shrink-0 items-center text-white">
			<span class="text-6xl font-semibold tracking-tight">BookIt</span>
		</div>
		<div class="flex items-center space-x-4">
			<a href="/" class="block text-teal-200 hover:text-white lg:inline-block"> Home </a>
			<a href="/lists" class="block text-teal-200 hover:text-white lg:inline-block"> Lists </a>
		</div>
	</div>

	<div class="search-container relative w-1/3 min-w-[300px]">
		<input
			type="text"
			bind:value={searchQuery}
			on:input={handleSearch}
			placeholder="Search books..."
			class="w-full rounded-full px-4 py-2 text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-400"
		/>
		{#if showSearchResults && searchResults.length > 0}
			<div class="absolute z-50 mt-2 max-h-96 w-full overflow-y-auto rounded-lg bg-white shadow-xl">
				{#each searchResults as book (book.bookId)}
					<button
						type="button"
						class="flex w-full cursor-pointer items-center space-x-4 p-4 text-left hover:bg-gray-100"
						on:click={() => addToList(book.bookId)}
						disabled={isAddingToList}
						on:keydown={(e) => {
							if (e.key === 'Enter' || e.key === ' ') {
								e.preventDefault();
								e.currentTarget.click();
							}
						}}
					>
						{#if isAddingToList}
							<div class="flex h-16 w-12 items-center justify-center">
								<div class="h-6 w-6 animate-spin rounded-full border-b-2 border-blue-500"></div>
							</div>
						{:else if book.thumbnail}
							<img src={book.thumbnail} alt={book.title} class="h-16 w-12 object-cover" />
						{/if}
						<div>
							<h3 class="font-medium text-gray-900">{book.title}</h3>
							<p class="text-sm text-gray-600">
								{book.authors ? book.authors[0] : 'Unknown Author'}
							</p>
						</div>
					</button>
				{/each}
			</div>
		{/if}
	</div>

	<div class="flex items-center space-x-4">
		{#if authenticated}
			<button
				on:click={handleLogout}
				class="inline-block rounded-full border-2 border-blue-500 bg-white px-4 py-2 text-lg font-semibold text-blue-500"
			>
				Log Out
			</button>
			<a
				href="/profile"
				class="inline-block rounded-full border-2 border-blue-500 bg-white px-4 py-2 text-lg font-semibold text-blue-500"
			>
				View Profile
			</a>
		{:else}
			<a
				href="/login"
				class="inline-block rounded-full border-2 border-blue-500 bg-white px-4 py-2 text-lg font-semibold text-blue-500"
			>
				Sign In
			</a>
			<a
				href="/signup"
				class="inline-block rounded-full border-2 border-blue-500 bg-blue-500 px-4 py-2 text-lg font-semibold text-white"
			>
				Sign Up
			</a>
		{/if}
	</div>
</nav>

<svelte:window on:click={handleClickOutside} />

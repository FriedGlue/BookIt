<script lang="ts">
	import { AuthService } from '$lib/services/authService';
	import { goto, invalidateAll } from '$app/navigation';
	import { onMount } from 'svelte';
	import { isAuthenticated } from '$lib/stores/authStore';
	import { fade } from 'svelte/transition';

	let searchQuery = '';
	let searchResults: any[] = [];
	let isSearching = false;
	let showSearchResults = false;
	let isAddingToList = false;
	let toBeReadList: any[] = [];
	let isNavigating = false;
	let isAuthLoading = true;

	const authService = new AuthService();

	onMount(async () => {
		try {
			const isAuth = await authService.isAuthenticated();
			isAuthenticated.set(isAuth);
		} catch (error) {
			console.error('Error checking authentication:', error);
			isAuthenticated.set(false);
		} finally {
			isAuthLoading = false;
		}
	});

	// ---- Move "searchBooks" to call your local route:
	async function handleSearch() {
		if (!searchQuery.trim()) {
			searchResults = [];
			showSearchResults = false;
			return;
		}

		try {
			isSearching = true;
			const res = await fetch(`/api/books/searchByTitle?q=${encodeURIComponent(searchQuery)}`);
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

	async function handleLogout() {
		authService.logout();
		invalidateAll();
		goto('/login');
	}

	function handleClickOutside(event: MouseEvent) {
		const target = event.target as HTMLElement;
		if (!target.closest('.search-container')) {
			showSearchResults = false;
		}
	}

	// Function to fetch book details
	async function fetchBookDetails(bookId: string) {
		const res = await fetch(`/api/books/${bookId}`);
		if (!res.ok) {
			console.error('Error fetching book details:', await res.text());
			return null;
		}
		return await res.json();
	}

	async function preloadImage(src: string): Promise<void> {
		return new Promise((resolve, reject) => {
			const img = new Image();
			img.onload = () => resolve();
			img.onerror = reject;
			img.src = src;
		});
	}

	async function handleBookNavigation(book: any) {
		try {
			isNavigating = true;
			searchQuery = '';
			showSearchResults = false;

			// Preload the book's image if it exists
			if (book.thumbnail) {
				await preloadImage(book.thumbnail);
			}

			const newPath = `/books/${book.bookId}`;

			await goto(newPath, { invalidateAll: true });
		} finally {
			isNavigating = false;
		}
	}
</script>

<nav class="flex flex-col bg-blue-500">
	<!-- Top Row -->
	<div class="flex items-center justify-between p-4">
		<div class="flex flex-shrink-0 items-center text-white">
			<span class="text-6xl font-semibold tracking-tight">BookIt</span>
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
				<div
					class="absolute z-50 mt-2 max-h-96 w-full overflow-y-auto rounded-lg bg-white shadow-xl"
				>
					{#each searchResults as book (book.bookId)}
						<a
							href={`/books/${book.bookId}`}
							class="flex w-full cursor-pointer items-center space-x-4 p-4 text-left hover:bg-gray-100"
							on:click|preventDefault={() => handleBookNavigation(book)}
						>
							{#if book.thumbnail}
								<div class="h-16 w-12">
									<img
										src={book.thumbnail}
										alt={book.title}
										class="h-full w-full object-cover"
										transition:fade={{ duration: 200 }}
									/>
								</div>
							{/if}
							<div>
								<h3 class="font-medium text-gray-900">{book.title}</h3>
								<p class="text-sm text-gray-600">
									{book.authors ? book.authors[0] : 'Unknown Author'}
								</p>
							</div>
						</a>
					{/each}
				</div>
			{/if}
		</div>

		<div class="flex items-center space-x-4">
			{#if !isAuthLoading}
				{#if $isAuthenticated}
					<button
						on:click={handleLogout}
						class="inline-block rounded-full border-2 border-blue-500 bg-white px-4 py-2 text-lg font-semibold text-blue-500"
					>
						Log Out
					</button>
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
			{:else}
				<div class="invisible flex items-center space-x-4">
					<div class="inline-block rounded-full border-2 border-blue-500 bg-white px-4 py-2 text-lg font-semibold text-blue-500">
						Log Out	
					</div>
				</div>
			{/if}
		</div>
	</div>

	<!-- Bottom Row -->
	<div class="flex justify-center bg-blue-800 p-2">
		<div class="flex items-center space-x-8">
			<a href="/" class="block text-teal-200 hover:text-white lg:inline-block"> Home </a>
			<div class="h-4 w-px bg-white"></div>
			<a href="/bookshelves" class="block text-teal-200 hover:text-white lg:inline-block">
				Bookshelves
			</a>
			<div class="h-4 w-px bg-white"></div>
			<a href="/reading-challenges" class="block text-teal-200 hover:text-white lg:inline-block">
				Challenges
			</a>
			<div class="h-4 w-px bg-white"></div>
			<a href="/reading-log" class="block text-teal-200 hover:text-white lg:inline-block">
				Reading Log
			</a>
		</div>
	</div>
</nav>

<svelte:window on:click={handleClickOutside} />

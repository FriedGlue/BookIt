<script lang="ts">
	import type { PageData } from './$types';
	import type { Book } from '$lib/types';
	import { enhance } from '$app/forms';

	export let data: PageData;

	// Use the book directly since it's already a single object
	$: book = data.book || null;

	let showListDropdown = false;
	let isStartReading = false;
	let addingToList: string | null = null;
	let imageLoading = true;
	let lastAddedList: string | null = null;
	
	const defaultLists = [
		{ id: 'toBeRead', name: 'To Be Read' },
		{ id: 'read', name: 'Read' }
	];

	$: customLists = Object.entries(data.customLists || {}).map(([id, _]) => ({
		id,
		name: id
	}));

	$: allLists = [...defaultLists, ...customLists];

	function toggleDropdown() {
		showListDropdown = !showListDropdown;
		if (!showListDropdown) {
			lastAddedList = null;
		}
	}

	function closeDropdown() {
		showListDropdown = false;
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') {
			closeDropdown();
		}
	}

	function handleImageLoad() {
		imageLoading = false;
	}

	$: if (book) {
		imageLoading = true;
	}
</script>

<svelte:window on:keydown={handleKeydown} />

<section class="mx-8 mt-16 px-4 md:mx-16 lg:mx-40">
	{#if book}
		<div class="flex flex-col md:flex-row gap-8">
			<!-- Book Cover -->
			<div class="flex-shrink-0 relative w-64">
				{#if imageLoading}
					<div class="absolute inset-0 bg-gray-100 rounded-lg flex items-center justify-center">
						<svg class="animate-spin h-8 w-8 text-gray-500" viewBox="0 0 24 24">
							<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none"/>
							<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
						</svg>
					</div>
				{/if}
				<img
					src={book.coverImageUrl || 'default-cover.png-image-url'}
					alt={`Cover of ${book.title} by ${book.authors?.join(', ')}`}
					class="w-64 h-auto rounded-lg shadow-lg transition-opacity duration-200"
					class:opacity-0={imageLoading}
					on:load={handleImageLoad}
				/>
			</div>

			<!-- Book Details -->
			<div class="flex flex-col justify-between space-y-6">
				<div>
					<h2 class="text-3xl font-bold">{book.title}</h2>
					{#if book.authors}
						<p class="mt-2 text-xl text-gray-700">by {book.authors.join(', ')}</p>
					{/if}

					<div class="mt-4 space-y-2 text-gray-600">
						{#if book.pageCount}
							<p><span class="font-semibold">Page Count:</span> {book.pageCount}</p>
						{:else if book.totalPages}
							<p><span class="font-semibold">Total Pages:</span> {book.totalPages}</p>
						{/if}
						{#if book.isbn13}
							<p><span class="font-semibold">ISBN:</span> {book.isbn13}</p>
						{/if}
						{#if book.tags}
							<p><span class="font-semibold">Tags:</span> {book.tags.join(', ')}</p>
						{/if}
					</div>

					<!-- Description (if available) -->
					<div class="mt-6">
						<h3 class="text-2xl font-semibold">Description</h3>
						<p class="mt-2 text-gray-600">
							{#if book.description}
								{book.description}
							{:else}
								No description available for this book.
							{/if}
						</p>
					</div>
				</div>

				<!-- Action Buttons -->
				<div class="space-y-4">
					<form 
						action="?/startReading" 
						method="POST" 
						use:enhance={() => {
							isStartReading = true;
							return async ({ result }) => {
								isStartReading = false;
								if (result.type === 'success') {
									// Could add a toast notification here
								}
							};
						}}
					>
						<input type="hidden" name="bookId" value={book.bookId} />
						<button
							type="submit"
							disabled={isStartReading}
							class="inline-block w-full rounded-full bg-green-600 px-6 py-3 text-white transition-all duration-200 hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed relative"
						>
							{#if isStartReading}
								<span class="opacity-0">Start Reading</span>
								<span class="absolute inset-0 flex items-center justify-center">
									<svg class="animate-spin h-5 w-5" viewBox="0 0 24 24">
										<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none"/>
										<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
									</svg>
								</span>
							{:else}
								Start Reading
							{/if}
						</button>
					</form>
					
					<div class="relative">
						<button
							type="button"
							on:click={toggleDropdown}
							aria-expanded={showListDropdown}
							aria-controls="bookshelf-menu"
							class="inline-block w-full rounded-full bg-blue-600 px-6 py-3 text-white transition-all duration-200 hover:bg-blue-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
						>
							{#if lastAddedList}
								Added to {allLists.find(l => l.id === lastAddedList)?.name}
							{:else}
								Add to a Bookshelf
							{/if}
						</button>
						
						{#if showListDropdown}
							<div 
								id="bookshelf-menu"
								role="menu"
								tabindex="0"
								class="absolute z-10 mt-2 w-full rounded-md bg-white shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none"
								on:mouseleave={closeDropdown}
							>
								<div class="py-1">
									{#each allLists as list}
										<form 
											action="?/addToList" 
											method="POST" 
											use:enhance={() => {
												addingToList = list.id;
												return async ({ result }) => {
													addingToList = null;
													if (result.type === 'success') {
														lastAddedList = list.id;
														closeDropdown();
													}
												};
											}}
										>
											<input type="hidden" name="bookId" value={book.bookId} />
											<input type="hidden" name="listType" value={list.id} />
											<button
												type="submit"
												role="menuitem"
												disabled={addingToList === list.id}
												class="block w-full px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 focus:bg-gray-100 focus:outline-none disabled:opacity-50 disabled:cursor-not-allowed relative"
											>
												{#if addingToList === list.id}
													<span class="opacity-0">{list.name}</span>
													<span class="absolute inset-0 flex items-center justify-center">
														<svg class="animate-spin h-4 w-4" viewBox="0 0 24 24">
															<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none"/>
															<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
														</svg>
													</span>
												{:else}
													{list.name}
												{/if}
											</button>
										</form>
									{/each}
								</div>
							</div>
						{/if}
					</div>
				</div>
			</div>
		</div>
	{:else}
		<p class="text-2xl text-gray-500">Loading...</p>
	{/if}
</section>

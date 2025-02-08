<script lang="ts">
	import type { PageData } from './$types';
	import type { Book, ReadingProgress, ToBeReadBook, ReadBook, CustomShelfBook } from '$lib/types';
	import { enhance } from '$app/forms';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';

	export let data: PageData;

	let selectedShelf = 'All';
	let displayBooks: BookWithMeta[] = [];

	// Define a type for our book shelf entries that includes shelf-specific metadata
	type BookWithMeta = Book & {
		_shelfType: string;
		progress: ReadingProgress;
		totalPages: number;
	};

	// Define a type for our books by shelf structure
	type BooksByShelf = {
		[key: string]: BookWithMeta[];
	};

	// Create a list of all available shelf names
	let shelves = ['All', 'To Be Read', 'Read', ...Object.keys(data.bookshelves?.customShelves ?? {})];

	// Helper function to create default progress
	function createDefaultProgress(): ReadingProgress {
		return {
			lastPageRead: 0,
			percentage: 0,
			lastUpdated: new Date().toISOString()
		};
	}

	// Helper function to convert shelf books to BookWithMeta
	function toBookWithMeta(book: ToBeReadBook | ReadBook | CustomShelfBook, shelfType: string): BookWithMeta {
		return {
			...book,
			_shelfType: shelfType,
			progress: createDefaultProgress(),
			totalPages: 0
		};
	}

	const booksByShelf: BooksByShelf = {
		'To Be Read': (data.bookshelves?.toBeRead ?? []).map((book) => toBookWithMeta(book, 'toBeRead')),
		Read: (data.bookshelves?.read ?? []).map((book) => toBookWithMeta(book, 'read')),
		...Object.fromEntries(
			Object.entries(data.bookshelves?.customShelves ?? {}).map(([shelfName, books]) => [
				shelfName,
				books.map((book) => toBookWithMeta(book, shelfName))
			])
		)
	};

	// Update displayed books whenever selected shelf changes
	$: {
		if (selectedShelf === 'All') {
			displayBooks = Object.values(booksByShelf).flat();
		} else {
			const shelfKey =
				selectedShelf === 'To Be Read'
					? 'To Be Read'
					: selectedShelf === 'Read'
						? 'Read'
						: selectedShelf;
			displayBooks = booksByShelf[shelfKey] ?? [];
		}
	}

	function removeBook(bookId: string) {
		displayBooks = displayBooks.filter((book) => book.bookId !== bookId);
		// Also remove from the source shelf to keep data in sync
		const shelfKey =
			selectedShelf === 'To Be Read'
				? 'To Be Read'
				: selectedShelf === 'Read'
					? 'Read'
					: selectedShelf;
		if (shelfKey !== 'All') {
			booksByShelf[shelfKey] = booksByShelf[shelfKey].filter((book) => book.bookId !== bookId);
		}
	}

	onMount(() => {
		const shelfParam = new URL(window.location.href).searchParams.get('shelf');
		if (shelfParam) {
			selectedShelf = shelfParam;
		}
	});

	// ------------------------------
	// New Shelf Modal state & methods
	// ------------------------------
	let showNewShelfModal = false;
	let newShelfName = '';

	// Reset the new shelf form
	function resetNewShelfForm() {
		newShelfName = '';
		showNewShelfModal = false;
	}

	// When the new shelf is successfully created we update the shelves array.
	function addNewShelf(shelfName: string) {
		// Add the new shelf only if it does not already exist.
		if (!shelves.includes(shelfName)) {
			shelves = [...shelves, shelfName];
		}
	}

	// ------------------------------
	// Delete Shelf Modal state & methods
	// ------------------------------
	let showDeleteShelfModal = false;

	// Open delete modal if the current shelf is custom.
	function openDeleteShelfModal() {
		showDeleteShelfModal = true;
	}

	// Reset delete modal
	function resetDeleteShelfModal() {
		showDeleteShelfModal = false;
	}

	// Function called after deletion is successful.
	function deleteShelf() {
		// Remove the custom shelf from shelves.
		shelves = shelves.filter((shelf) => shelf !== selectedShelf);
		// Optionally: Remove the shelf from your booksByShelf
		delete booksByShelf[selectedShelf];
		// Redirect the user back to a default shelf, for example "All"
		selectedShelf = 'All';
		resetDeleteShelfModal();
	}
</script>

<div class="flex min-h-screen">
	<Sidebar
		title="Your Shelves"
		items={shelves}
		selectedItem={selectedShelf}
		onSelect={(item) => {
			selectedShelf = item;
			history.replaceState(null, '', `/bookshelves?shelf=${item}`);
		}}
	/>

	<!-- Main Content -->
	<div class="flex-1 bg-gray-100 p-6">
		<div class="mb-6 flex items-center justify-between">
			<h1 class="text-2xl font-bold">
				{selectedShelf}
				<span class="ml-2 text-sm text-gray-500">({displayBooks.length})</span>
			</h1>

			<!-- Show Create New Shelf button for default shelves; otherwise show Delete Shelf -->
			{#if selectedShelf === 'All' || selectedShelf === 'To Be Read' || selectedShelf === 'Read'}
				<button
					class="rounded-full bg-blue-500 px-6 py-2 text-white hover:bg-blue-600"
					on:click={() => (showNewShelfModal = true)}
				>
					Create A New Shelf
				</button>
			{:else}
				<button
					class="rounded-full bg-red-500 px-6 py-2 text-white hover:bg-red-600"
					on:click={openDeleteShelfModal}
				>
					Delete Shelf
				</button>
			{/if}
		</div>

		{#if displayBooks.length === 0}
			<p class="text-gray-600">No books in this shelf.</p>
		{:else}
			<div class="space-y-4">
				{#each displayBooks as book (book.bookId)}
					<div class="flex rounded bg-white p-4 shadow">
						<!-- Thumbnail -->
						<div class="h-48 w-32 flex-shrink-0 bg-gray-200">
							{#if book.thumbnail}
								<img
									src={book.thumbnail}
									alt="Book cover for {book.title}"
									class="h-full w-full rounded-md object-cover"
								/>
							{:else}
								<div class="flex h-full w-full items-center justify-center text-gray-500">
									Cover Not Found
								</div>
							{/if}
						</div>

						<!-- Book info -->
						<div class="ml-4 flex w-full flex-col justify-between">
							<div>
								<p class="text-xl font-semibold">{book.title ?? 'Unknown'}</p>
								<p class="text-lg text-gray-600">
									{book.authors?.join(', ') ?? 'Unknown'}
								</p>
							</div>

							<form
								method="post"
								action="?/removeFromShelf"
								class="mt-2"
								use:enhance
								on:submit|preventDefault={() => removeBook(book.bookId)}
							>
								<input type="hidden" name="bookId" value={book.bookId} />
								<input type="hidden" name="shelfType" value={book._shelfType} />
								<button
									type="submit"
									class="h-10 w-36 rounded bg-red-500 px-3 py-1 text-lg text-white hover:bg-red-600"
								>
									Remove
								</button>
							</form>
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>
</div>

<!-- New Shelf Modal -->
{#if showNewShelfModal}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50">
		<div class="w-full max-w-md rounded-lg bg-white p-8">
			<h2 class="mb-6 text-2xl font-bold">Create New Shelf</h2>
			<form
				method="POST"
				action="?/createBookshelf"
				use:enhance={() => {
					return async ({ result, update }) => {
						// When the new shelf is successfully created,
						// update the shelves array and close the modal.
						if (result.type === 'success') {
							addNewShelf(newShelfName);
							resetNewShelfForm();
						}
						await update();
					};
				}}
				class="space-y-4"
			>
				<div>
					<label for="newShelfName" class="mb-2 block text-sm font-medium text-gray-700">
						Shelf Name
					</label>
					<input
						type="text"
						id="newShelfName"
						name="newShelfName"
						bind:value={newShelfName}
						class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
						required
					/>
				</div>
				<div class="flex justify-end space-x-4">
					<button
						type="button"
						class="rounded bg-gray-200 px-4 py-2 text-gray-700 hover:bg-gray-300"
						on:click={resetNewShelfForm}
					>
						Cancel
					</button>
					<button
						type="submit"
						class="rounded bg-blue-500 px-4 py-2 text-white hover:bg-blue-600"
					>
						Create Shelf
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<!-- Delete Shelf Confirmation Modal -->
{#if showDeleteShelfModal}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50">
		<div class="w-full max-w-md rounded-lg bg-white p-8">
			<h2 class="mb-6 text-2xl font-bold">Delete Shelf</h2>
			<p class="mb-6 text-gray-700">
				Are you sure you want to delete the custom shelf "{selectedShelf}"? This action cannot be undone.
			</p>
			<div class="flex justify-end space-x-4">
				<button
					type="button"
					class="rounded bg-gray-200 px-4 py-2 text-gray-700 hover:bg-gray-300"
					on:click={resetDeleteShelfModal}
				>
					Cancel
				</button>
				<form
					method="POST"
					action="?/deleteBookshelf"
					use:enhance={() => {
						return async ({ result, update }) => {
							// On success, remove the shelf from shelves and update the view.
							if (result.type === 'success') {
								deleteShelf();
							}
							await update();
						};
					}}
				>
					<input type="hidden" name="shelfName" value={selectedShelf} />
					<button
						type="submit"
						class="rounded bg-red-500 px-4 py-2 text-white hover:bg-red-600"
					>
						Delete Shelf
					</button>
				</form>
			</div>
		</div>
	</div>
{/if}

<script lang="ts">
	import type { PageData } from './$types';
	import type { Book, ReadingProgress, DisplayBook } from '$lib/types';
	import { enhance } from '$app/forms';
	import Sidebar from '$lib/components/Sidebar.svelte';

	export let data: PageData;

	let selectedList = 'All';
	let displayBooks: Book[] = [];

	// Define a type for our book list entries
	type BookWithMeta = Book & {
		_listType: string;
		progress: ReadingProgress;
	};

	// Define a type for our books by list structure
	type BooksByList = {
		[key: string]: BookWithMeta[];
	};

	// Create a list of all available list names
	const lists = ['All', 'To Be Read', 'Read', ...Object.keys(data.customLists ?? {})];

	// Helper function to create default progress
	function createDefaultProgress(): ReadingProgress {
		return {
			lastPageRead: 0,
			percentage: 0,
			lastUpdated: new Date().toISOString()
		};
	}

	// Create a normalized map of all books by list type
	const booksByList: BooksByList = {
		'To Be Read': (data.toBeReadList ?? []).map(book => ({
			...book,
			_listType: 'toBeRead',
			progress: createDefaultProgress()
		})),
		'Read': (data.readList ?? []).map(book => ({
			...book,
			_listType: 'read',
			progress: createDefaultProgress()
		})),
		...Object.fromEntries(
			Object.entries(data.customLists ?? {}).map(([listName, books]) => [
				listName,
				books.map(book => ({
					...book,
					_listType: listName,
					progress: createDefaultProgress()
				}))
			])
		)
	};

	// Update displayed books whenever selected list changes
	$: {
		if (selectedList === 'All') {
			displayBooks = Object.values(booksByList).flat();
		} else {
			const listKey = selectedList === 'To Be Read' ? 'To Be Read' : 
										selectedList === 'Read' ? 'Read' : selectedList;
			displayBooks = booksByList[listKey] ?? [];
		}
	}

	function removeBook(bookId: string) {
		displayBooks = displayBooks.filter(book => book.bookId !== bookId);
		// Also remove from the source list to keep data in sync
		const listKey = selectedList === 'To Be Read' ? 'To Be Read' : 
									 selectedList === 'Read' ? 'Read' : selectedList;
		if (listKey !== 'All') {
			booksByList[listKey] = booksByList[listKey].filter(book => book.bookId !== bookId);
		}
	}
</script>

<div class="flex min-h-screen">
	<Sidebar 
		title="Your Lists"
		items={lists}
		selectedItem={selectedList}
		onSelect={(item) => selectedList = item}
	/>

	<!-- Main Content -->
	<div class="flex-1 bg-gray-100 p-6">
		<h1 class="mb-6 text-2xl font-bold">
			{selectedList}
			<span class="ml-2 text-sm text-gray-500">({displayBooks.length})</span>
		</h1>

		{#if displayBooks.length === 0}
			<p class="text-gray-600">No books in this list.</p>
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
								action="?/removeFromList" 
								class="mt-2" 
								use:enhance 
								on:submit|preventDefault={() => removeBook(book.bookId)}
							>
								<input type="hidden" name="bookId" value={book.bookId} />
								<input type="hidden" name="listType" value={book._listType} />
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
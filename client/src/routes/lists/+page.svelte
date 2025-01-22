<script lang="ts">
	export let data: {
		toBeReadList: Book[];
		readList: Book[];
		customLists: Record<string, Book[]>;
	};

	import { onMount } from 'svelte';
	import { BookService } from '$lib/services/bookService';
	import type { Book } from '$lib/types';

	// We'll store the selected list name locally
	let selectedList = 'All'; // default to "All"
	let lists: string[] = []; // array of list names for the sidebar

	// We'll flatten "All" books in a single array for "All" view
	let allBooks: Book[] = [];
	let displayBooks: Book[] = []; // the books to display in main content

	const bookService = new BookService();

	// On component mount (or after SSR), let's build the list of listNames
	onMount(() => {
		// Basic built-in lists
		lists = ['All', 'toBeRead', 'read'];

		// Add any custom list names
		const customKeys = Object.keys(data.customLists ?? {});
		lists.push(...customKeys);

		// Build a combined array for "All"
		allBooks = [
			...(data.toBeReadList || []).map(b => ({ ...b, _listType: 'toBeRead' })),
			...(data.readList || []).map(b => ({ ...b, _listType: 'read' })),
			// flatten each custom list
			...customKeys.flatMap(key =>
				(data.customLists[key] || []).map(b => ({
					...b,
					_listType: key
				}))
			)
		];

		console.log(allBooks);	

		updateDisplayBooks();
	});

	function selectList(listName: string) {
		selectedList = listName;
		updateDisplayBooks();
	}

	function updateDisplayBooks() {
		switch (selectedList) {
			case 'All':
				displayBooks = allBooks;
				break;
			case 'toBeRead':
				displayBooks = data.toBeReadList?.map(b => ({ ...b, _listType: 'toBeRead' })) || [];
				break; 
			case 'read':
				displayBooks = data.readList?.map(b => ({ ...b, _listType: 'read' })) || [];
				break;
			default:
				// Custom list case
				displayBooks = data.customLists[selectedList]?.map(b => ({
					...b,
					_listType: selectedList
				})) || [];
		}

		// Filter out any null/undefined books
		displayBooks = displayBooks.filter(book => book && book.bookId);
	}

	async function removeBook(bookId: string, listType: string) {
		try {
			await bookService.removeFromList(bookId, listType);
			// Remove from local arrays
			if (listType === 'toBeRead') {
				data.toBeReadList = data.toBeReadList.filter(b => b.bookId !== bookId);
			} else if (listType === 'read') {
				data.readList = data.readList.filter(b => b.bookId !== bookId);
			} else {
				// custom list
				data.customLists[listType] = data.customLists[listType]?.filter(b => b.bookId !== bookId);
			}
			// Update the combined "all" list as well
			allBooks = allBooks.filter(b => b.bookId !== bookId || b._listType !== listType);
			
			updateDisplayBooks();
		} catch (error) {
			console.error('Failed to remove book:', error);
			alert('Error removing book from list');
		}
	}
</script>

<!-- Layout: side-by-side. The left side is a sidebar with the list names; the right side is the main content. -->
<div class="flex min-h-screen">
	<!-- Sidebar -->
	<div class="w-64 bg-gray-800 text-white p-4">
		<h2 class="text-lg font-bold mb-4">Your Lists</h2>
		<ul class="space-y-2">
			{#each lists as listName}
				<li>
					<button
						class="w-full text-left px-2 py-1 rounded hover:bg-gray-700"
						class:selected={selectedList === listName}
						on:click={() => selectList(listName)}>
						{listName === 'toBeRead' ? 'To Be Read' : listName}
					</button>
				</li>
			{/each}
		</ul>
	</div>

	<!-- Main Content -->
	<div class="flex-1 bg-gray-100 p-6">
		<h1 class="text-2xl font-bold mb-6">
			{selectedList === 'toBeRead' ? 'To Be Read' : selectedList} 
			<span class="text-gray-500 text-sm ml-2">({displayBooks.length})</span>
		</h1>

		<!-- Vertical table of books -->
		{#if displayBooks.length === 0}
			<p class="text-gray-600">No books in this list.</p>
		{:else}
			<div class="space-y-4">
				{#each displayBooks as book (book.bookId)}
					<div class="flex p-4 bg-white rounded shadow">
						<!-- Thumbnail -->
						<div class="w-32 h-48 flex-shrink-0  bg-gray-200">
							{#if book.thumbnail}
								<img 
									src={book.thumbnail} 
									alt="Thumbnail" 
									class="w-full h-full object-cover rounded-md" 
								/>
							{/if}
						</div>

						<!-- Book info -->
						<div class="ml-4 flex flex-col justify-between">
							<div>
								<p class="text-xl font-semibold">{book.title ?? 'Unknown'}</p>
								<p class="text-lg text-gray-600">{book.authors?.join(', ') ?? 'Unknown'}</p>
							</div>
							<!-- Remove button -->
							<button
								class="mt-2 px-3 py-1 bg-red-500 text-white text-lg rounded h-10 w-36 hover:bg-red-600"
								on:click={() => removeBook(book.bookId, book._listType!)}>
								Remove
							</button>
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>
</div>
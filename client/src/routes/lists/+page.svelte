<script lang="ts">
	import type { PageData } from './$types';
	import type { Book, ReadingProgress, DisplayBook } from '$lib/types';

	export let data: PageData;

	// Assuming you have the correct type for customLists
	type CustomListsType = Record<string, DisplayBook[]>; // Adjust if necessary

	// Ensure data is typed correctly
	const customLists: CustomListsType = data.customLists ?? {};

	let selectedList = 'All';     // "All" by default
	let lists: string[] = [];     // name of each list in the sidebar
	let allBooks: Book[] = [];    // combined "All" books
	let displayBooks: Book[] = []; // the currently displayed books

	// Build the lists and flatten everything into an "All" list:
	lists = ['All', 'toBeRead', 'read'];
	if (data.customLists) {
		lists.push(...Object.keys(data.customLists));
	}

	allBooks = [
		...(data.toBeReadList ?? []).map((b) => ({
			...b,
			_listType: 'toBeRead',
			progress: {
				lastPageRead: 0,
				percentage: 0,
				lastUpdated: new Date().toISOString(),
			} as ReadingProgress,
		})),
		...(data.readList ?? []).map((b) => ({
			...b,
			_listType: 'read',
			progress: {
				lastPageRead: 0,
				percentage: 0,
				lastUpdated: new Date().toISOString(),
			} as ReadingProgress,
		})),
		...(Object.entries(customLists).flatMap(([listName, books]) => // Use Object.entries to ensure correct typing
			books.map((b) => ({
				...b,
				_listType: listName,
				progress: {
					lastPageRead: 0,
					percentage: 0,
					lastUpdated: new Date().toISOString(),
				} as ReadingProgress,
			}))
		))
	];

	// Whenever selectedList changes, we update displayBooks
	$: updateDisplayBooks();

	function updateDisplayBooks() {
		switch (selectedList) {
			case 'All':
				displayBooks = allBooks;
				break;
			case 'toBeRead':
				displayBooks = (data.toBeReadList ?? []).map((b) => ({ ...b, _listType: 'toBeRead' }));
				break;
			case 'read':
				displayBooks = (data.readList ?? []).map((b) => ({ ...b, _listType: 'read' }));
				break;
			default:
				// custom list
				const customList = data.customLists?.[selectedList] ?? [];
				displayBooks = customList.map((b) => ({
					...b,
					_listType: selectedList,
					progress: {
						lastPageRead: 0,
						percentage: 0,
						lastUpdated: new Date().toISOString(),
					} as ReadingProgress
				}));
		}
	}

	function selectList(listName: string) {
		selectedList = listName;
	}
</script>

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
						on:click={() => selectList(listName)}
					>
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
						<div class="w-32 h-48 flex-shrink-0 bg-gray-200">
							{#if book.thumbnail}
								<img
									src={book.thumbnail}
									alt="Thumbnail"
									class="w-full h-full object-cover rounded-md"
								/>
							{:else}
								<div class="w-full h-full flex items-center justify-center text-gray-500">
									No Cover
								</div>
							{/if}
						</div>

						<!-- Book info -->
						<div class="ml-4 flex flex-col justify-between w-full">
							<div>
								<p class="text-xl font-semibold">{book.title ?? 'Unknown'}</p>
								<p class="text-lg text-gray-600">
									{book.authors?.join(', ') ?? 'Unknown'}
								</p>
							</div>

							<!-- Remove form (calls removeFromList action) -->
							<form method="post" action="?/removeFromList" class="mt-2">
								<!-- Hidden fields that our action expects -->
								<input type="hidden" name="bookId" value={book.bookId} />
								<input type="hidden" name="listType" value={book._listType} />

								<button
									type="submit"
									class="px-3 py-1 bg-red-500 text-white text-lg rounded h-10 w-36 hover:bg-red-600"
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

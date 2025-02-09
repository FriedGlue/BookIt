<script lang="ts">
	import type { PageData } from './$types';
	import type { Book, ReadingProgress } from '$lib/types';
	import { enhance } from '$app/forms';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import { onMount } from 'svelte';

	export let data: PageData;

	let selectedList = 'All';
	let displayBooks: Book[] = [];

	let showNewListModal: boolean = false;
	let newListName: string = '';
	let showDeleteListModal: boolean = false;
	let selectedListToDelete: string = '';

	// Define a type for our book list entries
	type BookWithMeta = Book & {
		_listType: string;
		progress: ReadingProgress;
	};

	// Define a type for our books by list structure
	type BooksByList = {
		'To Be Read': BookWithMeta[];
		'Read': BookWithMeta[];
		[key: string]: BookWithMeta[];
	};

	// Make shelves reactive so it updates when data changes
	$: shelves = ['All', 'To Be Read', 'Read', ...Object.keys(data.customLists ?? {})];

	// Helper function to create default progress
	function createDefaultProgress(): ReadingProgress {
		return {
			lastPageRead: 0,
			percentage: 0,
			lastUpdated: new Date().toISOString()
		};
	}

	// Make booksByList reactive so it updates when data changes
	$: booksByList = {
		'To Be Read': ((data.toBeReadList ?? []).map((book) => ({
			...book,
			_listType: 'toBeRead',
			progress: createDefaultProgress(),
			totalPages: 0
		})) || []),
		Read: ((data.readList ?? []).map((book) => ({
			...book,
			_listType: 'read',
			progress: createDefaultProgress(),
			totalPages: 0
		})) || []),
		...Object.fromEntries(
			Object.entries(data.customLists ?? {}).map(([listName, books]) => [
				listName,
				(books ?? []).map((book) => ({
					...book,
					_listType: listName,
					progress: createDefaultProgress(),
					totalPages: 0
				}))
			])
		)
	};

	// Update displayed books whenever selected list or booksByList changes
	$: {
		if (selectedList === 'All') {
			displayBooks = Object.values(booksByList).flat();
		} else {
			const listKey: keyof BooksByList =
				selectedList === 'To Be Read'
					? 'To Be Read'
					: selectedList === 'Read'
						? 'Read'
						: selectedList;
			displayBooks = booksByList[listKey] ?? [];
		}
	}

	// Update the handler to only call update()
	const handleSubmit = () => {
		return async ({ update }: { update: () => Promise<void> }) => {
			await update();
			// Close modals
			showNewListModal = false;
			showDeleteListModal = false;
			selectedListToDelete = '';
			newListName = '';
		};
	};

	const handleDelete = () => {
		return async ({ update }: { update: () => Promise<void> }) => {
			await update();
			// Close modals
			showDeleteListModal = false;
			selectedListToDelete = '';
			selectedList = 'All';
		};
	};

	onMount(() => {
		const listParam = new URL(window.location.href).searchParams.get('list');
		if (listParam) {
			selectedList = listParam;
		}
	});
</script>

<div class="flex min-h-screen">
	<Sidebar
		title="Your Shelves"
		items={shelves}
		selectedItem={selectedList}
		onSelect={(item) => {
			selectedList = item;
			history.replaceState(null, '', `/bookshelves?shelf=${item}`);
		}}
	/>

	<!-- Main Content -->
	<div class="flex-1 bg-gray-100 p-6">
		<div class="mb-6 flex items-center justify-between">
			<h1 class="mb-6 text-2xl font-bold">
				{selectedList}
				<span class="ml-2 text-sm text-gray-500">({displayBooks.length})</span>
			</h1>
			{#if selectedList === 'All' || selectedList === 'To Be Read' || selectedList === 'Read'}
				<button
					class="rounded-full bg-blue-500 px-6 py-2 text-white hover:bg-blue-600"
					on:click={() => (showNewListModal = true)}
				>
					Create A New Bookshelf 
				</button>
			{:else}
				<button
					class="rounded-full bg-red-500 px-6 py-2 text-white hover:bg-red-600"
					on:click={() => {
						selectedListToDelete = selectedList;
						showDeleteListModal = true;
					}}
				>
					Delete Bookshelf
				</button>
			{/if}
		</div>

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
								use:enhance={handleSubmit}
								class="mt-2"
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

{#if showNewListModal}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50">
		<div class="w-96 rounded-lg bg-white p-6">
			<h2 class="mb-4 text-xl font-bold">Create New Bookshelf</h2>
			<form
				method="post"
				action="?/createCustomBookshelf"
				use:enhance={handleSubmit}
			>
				<input
					type="text"
					name="listName"
					bind:value={newListName}
					placeholder="Enter list name"
					class="mb-4 w-full rounded border p-2"
				/>
				<div class="flex justify-end space-x-2">
					<button
						type="button"
						class="rounded bg-gray-300 px-4 py-2 hover:bg-gray-400"
						on:click={() => {
							showNewListModal = false;
							newListName = '';
						}}
					>
						Cancel
					</button>
					<button type="submit" class="rounded bg-blue-500 px-4 py-2 text-white hover:bg-blue-600">
						Create
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

{#if showDeleteListModal}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50">
		<div class="w-96 rounded-lg bg-white p-6">
			<h2 class="mb-4 text-xl font-bold">Delete Bookshelf</h2>
			<p class="mb-4">Are you sure you want to delete "{selectedListToDelete}"?</p>
			<form
				method="post"
				action="?/deleteCustomBookshelf"
				use:enhance={handleDelete}
			>
				<input type="hidden" name="listName" value={selectedListToDelete} />
				<div class="flex justify-end space-x-2">
					<button
						type="button"
						class="rounded bg-gray-300 px-4 py-2 hover:bg-gray-400"
						on:click={() => {
							showDeleteListModal = false;
							selectedListToDelete = '';
						}}
					>
						Cancel
					</button>
					<button type="submit" class="rounded bg-red-500 px-4 py-2 text-white hover:bg-red-600">
						Delete
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<script lang="ts">
	// Import any types you need and the enhance action for forms
	import type { PageData } from './$types';
	import { enhance } from '$app/forms';
	import type { Book } from '$lib/types';
	import ReadingChallenges from '$lib/components/ReadingChallenges.svelte';
	import { onMount } from 'svelte';

	export let data: PageData;

	let modalVisible = false;
	let showProgressTypeMenu = false;
	let progressType: 'pages' | 'percentage' = 'pages';
	let percentComplete: number;
	let menuButtonRef: HTMLButtonElement;
	let menuRef: HTMLDivElement;

	// The selectedBook here is the raw Book object from the API.
	let selectedBook: Book | null = null;
	let newPageCount: number | '' = '';

	// Handle clicks outside the menu
	function handleClickOutside(event: MouseEvent) {
		if (showProgressTypeMenu && menuRef && menuButtonRef) {
			const target = event.target as Node;
			if (!menuRef.contains(target) && !menuButtonRef.contains(target)) {
				showProgressTypeMenu = false;
			}
		}
	}

	onMount(() => {
		window.addEventListener('click', handleClickOutside);
		return () => {
			window.removeEventListener('click', handleClickOutside);
		};
	});

	function getProgressTypeKey(bookId: string) {
		return `preferredProgressType_${bookId}`;
	}

	// Function to update progress type and save preference
	function updateProgressType(type: 'pages' | 'percentage') {
		if (selectedBook?.bookId) {
			progressType = type;
			localStorage.setItem(getProgressTypeKey(selectedBook.bookId), type);
		}
		showProgressTypeMenu = false;
	}

	// Open the modal – note that we pass the raw Book object.
	function openModal(book: Book) {
		selectedBook = book;
		// Reset to default first
		progressType = 'pages';
		// Then load book-specific progress type preference
		if (book.bookId) {
			const savedProgressType = localStorage.getItem(getProgressTypeKey(book.bookId));
			if (savedProgressType === 'pages' || savedProgressType === 'percentage') {
				progressType = savedProgressType;
			}
		}
		// Initialize with current progress
		if (book.progress) {
			newPageCount = book.progress.lastPageRead;
			percentComplete = book.progress.percentage || (book.progress.lastPageRead / book.totalPages) * 100;
		} else {
			newPageCount = '';
			percentComplete = 0;
		}
		modalVisible = true;
	}

	function closeModal() {
		modalVisible = false;
		selectedBook = null;
		newPageCount = '';
		progressType = 'pages'; // Reset to default when closing
	}

	// Add this calculation before the template
	const booksPerMonth = 2.1; // Replace with actual calculation based on your data
</script>

<!-- Outer container preserving your original layout classes -->
<div class="flex flex-grow flex-col">
	<main class="flex-grow">
		<!-- Current Reads Section -->
		<section class="mx-2 mt-8 flex flex-col items-start px-2 md:mx-16 lg:mx-40">
			{#if !data.profile || !data.profile.currentlyReading || data.profile.currentlyReading.length === 0}
				<div class="flex w-full items-center justify-center py-16">
					<p class="text-4xl text-gray-500">No Books In Progress... Get To Reading!</p>
				</div>
			{:else}
				<div class="grid w-full grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
					{#each data.profile.currentlyReading as reading}
						{#if reading.Book}
							<div
								class="flex w-full transform flex-col overflow-hidden rounded-lg shadow-lg duration-300 hover:scale-105"
							>
								<div
									class="flex h-48 w-full items-center justify-center bg-gray-300 sm:h-64 md:h-72 lg:h-64"
								>
									{#if reading.Book.bookId}
										<img
											src={reading.Book.thumbnail}
											alt={`Cover of ${reading.Book.title} by ${reading.Book.authors?.[0] ?? 'Unknown Author'}`}
											class="max-h-full max-w-full object-contain"
											loading="lazy"
										/>
									{:else}
										<div class="text-gray-500">No Cover Available</div>
									{/if}
								</div>
								<div class="flex flex-grow flex-col bg-white p-6">
									<h2
										class="mb-2 line-clamp-2 h-14 overflow-hidden text-xl font-semibold text-gray-800"
										title={reading.Book.title}
									>
										{reading.Book.title}
									</h2>
									<p class="mb-4 line-clamp-1 h-6 text-gray-600">
										{reading.Book.authors?.[0] ?? 'Unknown Author'}
									</p>

									<!-- Progress Bar -->
									<div class="h-4 w-full overflow-hidden rounded-full bg-gray-200">
										<div
											class="h-full bg-green-500 transition-all duration-300"
											style="width: {reading.Book.progress?.percentage}%;"
										></div>
									</div>

									<div class="mt-2 flex items-center justify-between">
										<span class="text-sm text-gray-700">
											{Math.round(reading.Book.progress?.percentage ?? 0)}%
										</span>

										<div class="space-x-2">

									<a href={`/books/${reading.Book.bookId}`} class="w-3/4">
										<button
											type="button"
											class="h-8 w-full text-center text-blue-500 hover:text-blue-700 transition-all duration-200"
										>
											Details
										</button>
									</a>
										</div>
									</div>
									<!-- Trigger modal to update progress -->
									<button
										on:click={() => openModal(reading.Book)}
										class="mt-4 h-8 w-full rounded-full bg-gray-300 hover:bg-blue-500"
									>
										Update Progress
									</button>
								</div>
							</div>
						{/if}
					{/each}
				</div>
			{/if}
		</section>

		<!-- Divider -->
		<hr class="my-16 border-gray-300" />

		<!---Reading Challenge-->
		<section class="mx-2 mt-8 flex flex-col items-center py-4 md:mx-16 lg:mx-40">
			{#if data.profile?.challenges && data.profile.challenges.length > 0}
				<ReadingChallenges challenges={data.profile.challenges} />
			{:else}
				<div class="flex w-full items-center justify-center py-16">
					<p class="text-4xl text-gray-500">No Reading Challenges Set</p>
				</div>
			{/if}
		</section>

		<!-- Divider -->
		<hr class="my-8 border-gray-300" />

		<!-- To Be Read Section -->
		<section class="mx-2 mt-8 flex flex-col items-start py-4 md:mx-16 lg:mx-40">
			<div class="mb-4 w-full text-left sm:mb-8">
				<h1 class="text-2xl font-bold text-gray-600 sm:text-4xl md:text-5xl lg:text-4xl">
					To Be Read ({data.profile?.lists?.toBeRead?.length ?? 0})
				</h1>
				<a
					href="/lists?list=To Be Read"
					class="mt-2 text-lg font-semibold text-blue-500 hover:text-blue-700">View All</a
				>
			</div>

			{#if !data.profile?.lists?.toBeRead?.length}
				<div class="flex w-full items-center justify-center py-16">
					<p class="text-2xl text-gray-500">Add a book via search to see it in your list</p>
				</div>
			{:else}
				<!-- Grid Container -->
				<div class="grid grid-cols-2 gap-2 sm:grid-cols-3 sm:gap-4 lg:grid-cols-5">
					{#each (data.profile.lists.toBeRead.slice(0, 4) || []).reverse() as book}
						<div class="flex flex-col">
							<div
								class="group relative w-full transform rounded-lg bg-gray-300 transition-shadow duration-300 hover:scale-105 hover:shadow-2xl"
							>
								<img
									src={book.thumbnail || 'default-cover-image-url'}
									alt="Book Cover"
									loading="lazy"
									decoding="async"
									style="opacity: 0; transition: opacity 0.3s"
									on:load={(e) => ((e.currentTarget as HTMLImageElement).style.opacity = '1')}
									class="h-48 w-full rounded-lg sm:h-64 md:h-80 lg:h-64"
								/>
								<div
									class="absolute inset-0 flex flex-col items-center justify-center gap-2 rounded-lg bg-black/50 opacity-0 transition-opacity duration-200 group-hover:opacity-100"
								>
									<a href={`/books/${book.bookId}`} class="w-3/4">
										<button
											type="button"
											class="h-8 w-full rounded-full bg-white text-center text-gray-800 transition-all duration-200 hover:bg-blue-500 hover:text-white"
										>
											Details
										</button>
									</a>

									<!-- "Start" reading form -->
									<form method="post" action="?/startReading" class="w-3/4" use:enhance>
										<input type="hidden" name="bookId" value={book.bookId} />
										<input type="hidden" name="openLibraryId" value={book.openLibraryId || ''} />
										<input type="hidden" name="listName" value="toBeRead" />
										<button
											type="submit"
											class="h-8 w-full rounded-full bg-white text-gray-800 transition-all duration-200 hover:bg-blue-500 hover:text-white"
										>
											Start
										</button>
									</form>

									<!-- "Remove" from list form -->
									<form method="post" action="?/removeFromList" class="w-3/4" use:enhance>
										<input type="hidden" name="bookId" value={book.bookId} />
										<input type="hidden" name="openLibraryId" value={book.openLibraryId || ''} />
										<input type="hidden" name="listType" value="toBeRead" />
										<button
											type="submit"
											class="h-8 w-full rounded-full bg-white text-gray-800 transition-all duration-200 hover:bg-blue-500 hover:text-white"
										>
											Remove
										</button>
									</form>
								</div>
							</div>
						</div>
					{/each}

					{#if data.profile.lists.toBeRead.length >= 5}
						<div class="flex items-center justify-center text-6xl text-gray-600">...</div>
					{/if}
				</div>
			{/if}
		</section>

		<!-- Read Section -->
		<section class="mx-2 mt-8 flex flex-col items-start px-2 md:mx-16 lg:mx-40">
			<div class="mb-4 w-full text-left sm:mb-8">
				<h1 class="text-2xl font-bold text-gray-600 sm:text-4xl md:text-5xl lg:text-4xl">
					Read ({data.profile?.lists?.read?.length ?? 0})
				</h1>
				<a
					href="/lists?list=Read"
					class="mt-2 text-lg font-semibold text-blue-500 hover:text-blue-700">View All</a
				>
			</div>

			{#if !data.profile?.lists?.read?.length}
				<div class="flex w-full items-center justify-center py-16">
					<p class="text-2xl text-gray-500">Add a book via search to see it in your list</p>
				</div>
			{:else}
				<!-- Grid Container -->
				<div class="grid grid-cols-2 gap-2 sm:grid-cols-3 sm:gap-4 lg:grid-cols-5">
					{#each (data.profile.lists.read.slice(0, 4) || []).reverse() as book}
						<div class="flex flex-col">
							<div
								class="group relative w-full transform rounded-lg bg-gray-300 transition-shadow duration-300 hover:scale-105 hover:shadow-2xl"
							>
								<img
									src={book.thumbnail || 'default-cover-image-url'}
									alt="Book Cover"
									loading="lazy"
									decoding="async"
									style="opacity: 0; transition: opacity 0.3s"
									on:load={(e) => ((e.currentTarget as HTMLImageElement).style.opacity = '1')}
									class="h-48 w-full rounded-lg sm:h-64 md:h-80 lg:h-64"
								/>
								<div
									class="absolute inset-0 flex flex-col items-center justify-center gap-2 rounded-lg bg-black/50 opacity-0 transition-opacity duration-200 group-hover:opacity-100"
								>
									<a href={`/books/${book.bookId}`} class="w-3/4">
										<button
											type="button"
											class="h-8 w-full rounded-full bg-white text-center text-gray-800 transition-all duration-200 hover:bg-blue-500 hover:text-white"
										>
											Details
										</button>
									</a>

									<form method="post" action="?/startReading" class="w-3/4" use:enhance>
										<input type="hidden" name="bookId" value={book.bookId} />
										<input type="hidden" name="openLibraryId" value={book.openLibraryId || ''} />
										<input type="hidden" name="listName" value="read" />
										<button
											type="submit"
											class="h-8 w-full rounded-full bg-white text-gray-800 transition-all duration-200 hover:bg-blue-500 hover:text-white"
										>
											Start
										</button>
									</form>

									<form method="post" action="?/removeFromList" class="w-3/4" use:enhance>
										<input type="hidden" name="bookId" value={book.bookId} />
										<input type="hidden" name="openLibraryId" value={book.openLibraryId || ''} />
										<input type="hidden" name="listType" value="read" />
										<button
											type="submit"
											class="h-8 w-full rounded-full bg-white text-gray-800 transition-all duration-200 hover:bg-blue-500 hover:text-white"
										>
											Remove
										</button>
									</form>
								</div>
							</div>
						</div>
					{/each}

					{#if data.profile.lists.read.length >= 5}
						<div class="flex items-center justify-center text-6xl text-gray-600">...</div>
					{/if}
				</div>
			{/if}
		</section>
	</main>

	<!-- Bottom spacing if needed -->
	<div class="mb-32"></div>
</div>

<!-- Modal (Update Progress) -->
{#if modalVisible && selectedBook}
	<div class="relative z-10" aria-labelledby="modal-title" role="dialog" aria-modal="true">
		<!-- Backdrop -->
		<div class="fixed inset-0 bg-gray-500/75 transition-opacity" aria-hidden="true"></div>

		<div class="fixed inset-0 z-10 w-screen overflow-y-auto">
			<div class="flex min-h-full items-end justify-center p-2 text-center sm:items-center sm:p-0">
				<div
					class="relative transform overflow-hidden rounded-lg bg-white py-4 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg"
				>
					<div class="bg-white px-4 pb-8 pt-8 sm:p-8 sm:pb-8">
						<div class="sm:flex sm:items-start">
							<div
								class="mx-auto flex h-12 w-12 shrink-0 items-center justify-center rounded-full bg-blue-100 sm:mx-0 sm:h-10 sm:w-10"
							>
								<svg
									class="h-6 w-6 text-blue-600"
									fill="none"
									viewBox="0 0 24 24"
									stroke-width="1.5"
									stroke="currentColor"
								>
									<path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
								</svg>
							</div>
							<div class="mt-3 text-center sm:ml-4 sm:mt-0 sm:text-left">
								<h3 class="text-base font-semibold text-gray-900" id="modal-title">
									Update Progress for {selectedBook.title}
								</h3>
								<div class="mt-8">
									<div class="flex items-center justify-between gap-4 text-sm text-gray-600">
										<div>
											Current Page: {selectedBook.progress?.lastPageRead || 0} / {selectedBook.totalPages ||
												'Unknown'}
										</div>
										<div>
											Last Updated:
											{selectedBook.progress?.lastUpdated
												? new Date(selectedBook.progress.lastUpdated).toLocaleDateString('en-US', {
														month: 'numeric',
														day: 'numeric',
														year: 'numeric'
													})
												: 'Never'}
										</div>
									</div>
								</div>
								<div class="mt-8">
									<div class="relative">
										<button
											type="button"
											bind:this={menuButtonRef}
											class="mb-4 inline-flex w-full items-center justify-between rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500"
											on:click|stopPropagation={() => showProgressTypeMenu = !showProgressTypeMenu}
										>
											{progressType === 'pages' ? 'Update by Page Count' : 'Update by Percentage'}
											<svg class="ml-2 h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
												<path fill-rule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clip-rule="evenodd" />
											</svg>
										</button>
										
										{#if showProgressTypeMenu}
											<div 
												bind:this={menuRef}
												class="absolute z-10 mt-1 w-full rounded-md bg-white shadow-lg"
											>
												<div class="py-1">
													<button
														class="block w-full px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-100"
														on:click={() => updateProgressType('pages')}
													>
														Update by Page Count
													</button>
													<button
														class="block w-full px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-100"
														on:click={() => updateProgressType('percentage')}
													>
														Update by Percentage
													</button>
												</div>
											</div>
										{/if}
									</div>

									{#if progressType === 'pages'}
										<label for="newPageCount" class="block text-sm font-medium text-gray-700">
											New Page Count
										</label>
										<input
											id="newPageCount"
											type="number"
											min={selectedBook.progress?.lastPageRead || 0}
											max={selectedBook.totalPages}
											bind:value={newPageCount}
											class="mt-4 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
											placeholder="Enter new page count"
											required
										/>
									{:else}
										<label for="percentComplete" class="block text-sm font-medium text-gray-700">
											Percentage Complete
										</label>
										<input
											id="percentComplete"
											type="number"
											min="0"
											max="100"
											bind:value={percentComplete}
											class="mt-4 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
											placeholder="Enter percentage (0-100)"
											required
										/>
									{/if}
								</div>
							</div>
						</div>
					</div>

					<!-- Footer with action buttons -->
					<div class="bg-gray-50 px-4 py-4 sm:px-8 sm:py-6">
						<div class="flex flex-col gap-2 sm:flex-row-reverse sm:gap-3">
							<!-- Form for 'updateProgress' -->
							<form
								method="post"
								action="?/updateProgress"
								class="w-full sm:ml-3 sm:w-auto"
								use:enhance
								on:submit|preventDefault={() => closeModal()}
							>
								<input type="hidden" name="bookId" value={selectedBook.bookId} />
								<input type="hidden" name="openLibraryId" value={selectedBook.openLibraryId || ''} />
								<input type="hidden" name="newPageCount" value={progressType === 'pages' ? newPageCount : Math.round((percentComplete / 100) * selectedBook.totalPages)} />
								<button
									type="submit"
									class="w-full rounded-md bg-blue-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-blue-500 sm:w-auto sm:px-5 sm:py-3"
								>
									Update
								</button>
							</form>

							<!-- Form for 'finishReading' -->
							<form
								method="post"
								action="?/finishBook"
								class="mt-3 w-full sm:ml-3 sm:mt-0 sm:w-auto"
								use:enhance
								on:submit|preventDefault={() => closeModal()}
							>
								<input type="hidden" name="bookId" value={selectedBook.bookId} />
								<input type="hidden" name="openLibraryId" value={selectedBook.openLibraryId || ''} />
								<button
									type="submit"
									class="w-full rounded-md bg-green-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-green-500 sm:w-auto sm:px-5 sm:py-3"
								>
									Finish Book
								</button>
							</form>

							<!-- Form for removing from Currently Reading -->
							<form
								method="post"
								action="?/removeFromCurrentlyReading"
								class="mt-3 w-full sm:ml-3 sm:mt-0 sm:w-auto"
								use:enhance
								on:submit|preventDefault={() => closeModal()}
							>
								<input type="hidden" name="bookId" value={selectedBook.bookId} />
								<input type="hidden" name="openLibraryId" value={selectedBook.openLibraryId || ''} />
								<button
									type="submit"
									class="w-full rounded-md bg-red-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-red-500 sm:w-auto sm:px-5 sm:py-3"
								>
									Remove
								</button>
							</form>

							<!-- Cancel button -->
							<button
								type="button"
								class="w-full rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:w-auto sm:px-5 sm:py-3"
								on:click={closeModal}
							>
								Cancel
							</button>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
{/if}
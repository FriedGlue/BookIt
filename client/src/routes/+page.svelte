<script lang="ts">
	import type { PageData } from './$types';
	import type { DisplayBook } from '$lib/types';
	import { enhance } from '$app/forms';

	// The 'data' prop comes from your load() function on the page.
	export let data: PageData;

	// Local state for controlling the "Update Progress" modal
	let modalVisible = false;
	let selectedBook: {
		bookId: string;
		title: string;
		author: string;
		thumbnail: string;
		progress: number;
		currentPage: number;
		totalPages: number;
		lastUpdated?: string;
	} | null = null;

	let newPageCount: number | '' = '';

	function openModal(book: DisplayBook) {
		selectedBook = book;
		newPageCount = '';
		modalVisible = true;
	}

	function closeModal() {
		modalVisible = false;
		selectedBook = null;
		newPageCount = '';
	}

</script>

<!-- Outer container preserving your original layout classes -->
<div class="flex flex-grow flex-col">
	<main class="flex-grow">
		<!-- Current Reads Section -->
		<section class="mx-8 mt-16 flex flex-col items-start px-4 md:mx-16 lg:mx-40">
			<div class="mb-8 w-full text-left">
				<h1 class="text-4xl font-bold text-gray-800 md:text-5xl lg:text-6xl">Current Reads</h1>
			</div>

			{#if data.books.length === 0}
				<div class="flex w-full items-center justify-center py-16">
					<p class="text-2xl text-gray-500">No current reads</p>
				</div>
			{:else}
				<div class="grid w-full grid-cols-1 gap-8 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
					{#each data.books as book}
						<div
							class="flex w-full transform flex-col overflow-hidden rounded-lg shadow-lg duration-300 hover:scale-105"
						>
							<div class="flex h-64 w-full items-center justify-center bg-gray-300 md:h-72 lg:h-64">
								{#if book.bookId}
									<img
										src={book.thumbnail}
										alt={`Cover of ${book.title} by ${book.author}`}
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
									title={book.title}
								>
									{book.title}
								</h2>
								<p class="mb-4 line-clamp-1 h-6 text-gray-600">{book.author}</p>

								<!-- Progress Bar -->
								<div class="h-4 w-full overflow-hidden rounded-full bg-gray-200">
									<div
										class="h-full bg-green-500 transition-all duration-300"
										style="width: {book.progress}%;"
									></div>
								</div>

								<div class="mt-2 flex items-center justify-between">
									<span class="text-sm text-gray-700">
										{Math.round(book.progress)}%
									</span>
									<div class="space-x-2">
										<a href={`/books/${book.bookId}`} class="text-blue-500 hover:text-blue-700">Details</a>
									</div>
								</div>
								<!-- Progress percentage and button row -->

								<!-- Trigger modal to update progress -->
								<button
									on:click={() => openModal(book)}
									class="mt-4 h-8 w-full rounded-full bg-gray-300 hover:bg-blue-500"
								>
									Update Progress
								</button>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</section>

		<!-- Example Divider -->
		<hr class="my-16 border-gray-300" />

		<!-- To Be Read Section -->
		<section class="mx-8 mt-16 flex flex-col items-start px-4 md:mx-16 lg:mx-40">
			<div class="mb-8 w-full text-left">
				<h1 class="text-4xl font-bold text-gray-600 md:text-5xl lg:text-4xl">
					To Be Read ({data.toBeReadList?.length || 0})
				</h1>
				<button class="mt-2 text-lg font-semibold text-blue-500">View All</button>
			</div>

			<!-- Grid Container -->
			<div class="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-5">
				{#each (data.toBeReadList?.slice(0, 4) || []).reverse() as book}
					<div class="flex flex-col">
						<div
							class="group relative w-full transform rounded-lg
                     bg-gray-300 transition-shadow duration-300 hover:scale-105 hover:shadow-2xl"
						>
							<img
								src={book.thumbnail || 'default-cover-image-url'}
								alt="Book Cover"
								loading="lazy"
								decoding="async"
								style="opacity: 0; transition: opacity 0.3s"
								on:load={(e) => ((e.currentTarget as HTMLImageElement).style.opacity = '1')}
								class="h-64 w-full rounded-lg sm:h-72 md:h-80 lg:h-64"
							/>
							<div
								class="absolute inset-0 flex flex-col items-center justify-center gap-2
                       rounded-lg bg-black/50 opacity-0 transition-opacity duration-200 group-hover:opacity-100"
							>
								<!-- "Details" button (example SSR form or could be a simple link) -->
								<a href={`/books/${book.bookId}`} class="w-3/4">
									<button
										type="button"
										class="h-8 w-full text-center rounded-full bg-white text-gray-800
                           transition-all duration-200 hover:bg-blue-500 hover:text-white"
									>
										Details
									</button>
								</a>

								<!-- "Start" reading form -->
								<form method="post" action="?/startReading" class="w-3/4" use:enhance>
									<input type="hidden" name="bookId" value={book.bookId} />
									<input type="hidden" name="listName" value="toBeRead" />
									<button
										type="submit"
										class="h-8 w-full rounded-full bg-white text-gray-800
                           transition-all duration-200 hover:bg-blue-500 hover:text-white"
									>
										Start
									</button>
								</form>

								<!-- "Remove" from list form -->
								<form method="post" action="?/removeFromList" class="w-3/4" use:enhance>
									<input type="hidden" name="bookId" value={book.bookId} />
									<input type="hidden" name="listType" value="toBeRead" />
									<button
										type="submit"
										class="h-8 w-full rounded-full bg-white text-gray-800
                           transition-all duration-200 hover:bg-blue-500 hover:text-white"
									>
										Remove
									</button>
								</form>
							</div>
						</div>
					</div>
				{/each}

				{#if (data.toBeReadList?.length || 0) >= 5}
					<div class="flex items-center justify-center text-6xl text-gray-600">...</div>
				{/if}
			</div>
		</section>

		<!-- Example Divider -->
		<hr class="my-16 border-gray-300" />

		<!-- Read Section -->
		<section class="mx-8 mt-16 flex flex-col items-start px-4 md:mx-16 lg:mx-40">
			<div class="mb-8 w-full text-left">
				<h1 class="text-4xl font-bold text-gray-600 md:text-5xl lg:text-4xl">
					Read ({data.readList?.length || 0})
				</h1>
				<button class="mt-2 text-lg font-semibold text-blue-500">View All</button>
			</div>

			<!-- Grid Container -->
			<div class="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-5">
				{#each (data.readList?.slice(0, 4) || []).reverse() as book}
					<div class="flex flex-col">
						<div
							class="group relative w-full transform rounded-lg
                     bg-gray-300 transition-shadow duration-300 hover:scale-105 hover:shadow-2xl"
						>
							<img
								src={book.thumbnail || 'default-cover-image-url'}
								alt="Book Cover"
								loading="lazy"
								decoding="async"
								style="opacity: 0; transition: opacity 0.3s"
								on:load={(e) => ((e.currentTarget as HTMLImageElement).style.opacity = '1')}
								class="h-64 w-full rounded-lg sm:h-72 md:h-80 lg:h-64"
							/>
							<div
								class="absolute inset-0 flex flex-col items-center justify-center gap-2
                       rounded-lg bg-black/50 opacity-0 transition-opacity duration-200 group-hover:opacity-100"
							>
								<!-- "Details" button (example SSR form or could be a simple link) -->
								<a href={`/books/${book.bookId}`} class="w-3/4">
									<button
										type="button"
										class="h-8 w-full text-center rounded-full bg-white text-gray-800
                           transition-all duration-200 hover:bg-blue-500 hover:text-white"
									>
										Details
									</button>
								</a>

								<!-- "Start" reading form -->
								<form method="post" action="?/startReading" class="w-3/4" use:enhance>
									<input type="hidden" name="bookId" value={book.bookId} />
									<input type="hidden" name="listName" value="read" />
									<button
										type="submit"
										class="h-8 w-full rounded-full bg-white text-gray-800
                           transition-all duration-200 hover:bg-blue-500 hover:text-white"
									>
										Start
									</button>
								</form>

								<!-- "Remove" from list form -->
								<form method="post" action="?/removeFromList" class="w-3/4">
									<input type="hidden" name="bookId" value={book.bookId} />
									<input type="hidden" name="listType" value="read" />
									<button
										type="submit"
										class="h-8 w-full rounded-full bg-white text-gray-800
                           transition-all duration-200 hover:bg-blue-500 hover:text-white"
									>
										Remove
									</button>
								</form>
							</div>
						</div>
					</div>
				{/each}

				{#if (data.readList?.length || 0) >= 5}
					<div class="flex items-center justify-center text-6xl text-gray-600">...</div>
				{/if}
			</div>
		</section>

		<!-- Example Divider -->
		<hr class="my-16 border-gray-300" />

		<!-- Custom Lists Section -->
		{#each Object.entries(data.customLists) as [listName, books]}
			<section class="mx-8 mt-16 flex flex-col items-start px-4 md:mx-16 lg:mx-40">
				<div class="mb-8 w-full text-left">
					<h1 class="text-4xl font-bold text-gray-600 md:text-5xl lg:text-4xl">
						{listName} ({books.length})
					</h1>
					<button class="mt-2 text-lg font-semibold text-blue-500">View All</button>
				</div>

				<!-- Grid Container -->
				<div class="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-5">
					{#each books.slice(0, 4).reverse() as book}
						<div
							class="flex transform flex-col rounded-lg bg-gray-300 shadow-lg
                     transition-shadow duration-300 hover:scale-105 hover:shadow-2xl"
						>
							<div class="relative w-full">
								<img
									src={book.thumbnail || 'default-cover-image-url'}
									alt="Book Cover"
									loading="lazy"
									decoding="async"
									style="opacity: 0; transition: opacity 0.3s"
									on:load={(e) => ((e.currentTarget as HTMLImageElement).style.opacity = '1')}
									class="h-64 w-full rounded-lg sm:h-72 md:h-80 lg:h-64"
								/>
							</div>
							<!-- If you want "Details"/"Remove" forms for custom lists, 
                   you could add them here in a similar pattern -->
						</div>
					{/each}

					{#if books.length > 4}
						<div class="flex items-center justify-center text-xl text-gray-600">...</div>
					{/if}
				</div>
			</section>
		{/each}
	</main>

	<!-- Add bottom spacing if needed -->
	<div class="mb-32"></div>
</div>

<!-- Modal (Update Progress) -->
{#if modalVisible && selectedBook}
	<div class="relative z-10" aria-labelledby="modal-title" role="dialog" aria-modal="true">
		<!-- Backdrop -->
		<div class="fixed inset-0 bg-gray-500/75 transition-opacity" aria-hidden="true"></div>

		<div class="fixed inset-0 z-10 w-screen overflow-y-auto">
			<div class="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
				<div
					class="relative transform overflow-hidden rounded-lg bg-white py-4 text-left
                 shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg"
				>
					<div class="bg-white px-4 pb-8 pt-8 sm:p-8 sm:pb-8">
						<div class="sm:flex sm:items-start">
							<div
								class="mx-auto flex h-12 w-12 shrink-0 items-center justify-center
                       rounded-full bg-blue-100 sm:mx-0 sm:h-10 sm:w-10"
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
											Current Page: {selectedBook.currentPage || 0} /
											{selectedBook.totalPages || 'Unknown'}
										</div>
										<div>
											Last Updated:
											{selectedBook.lastUpdated
												? new Date(selectedBook.lastUpdated).toLocaleDateString('en-US', {
														month: 'numeric',
														day: 'numeric',
														year: 'numeric'
													})
												: 'Never'}
										</div>
									</div>
								</div>
								<div class="mt-8">
									<label for="newPageCount" class="block text-sm font-medium text-gray-700">
										New Page Count
									</label>
									<input
										id="newPageCount"
										type="number"
										min={selectedBook.currentPage || 0}
										bind:value={newPageCount}
										class="mt-4 block w-full rounded-md border-gray-300
                           shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
										placeholder="Enter new page count"
										required
									/>
								</div>
							</div>
						</div>
					</div>

					<!-- Footer with "Update," "Finish Book," "Remove," "Cancel" -->
					<div class="bg-gray-50 px-8 py-6 sm:flex sm:flex-row-reverse sm:px-8">
						<!-- Form for 'updateProgress' -->
						<form
							method="post"
							action="?/updateProgress"
							class="inline-flex w-full justify-center sm:ml-3 sm:w-auto"
							use:enhance
							on:submit|preventDefault={() => closeModal()}
						>
							<input type="hidden" name="bookId" value={selectedBook.bookId} />
							<input type="hidden" name="newPageCount" value={newPageCount} />

							<button
								type="submit"
								class="inline-flex w-full justify-center rounded-md bg-blue-600
                       px-5 py-3 text-sm font-semibold text-white shadow-sm
                       hover:bg-blue-500 sm:w-auto"
							>
								Update
							</button>
						</form>

						<!-- Form for 'finishReading' -->
						<form
							method="post"
							action="?/finishBook"
							class="mt-3 inline-flex w-full justify-center sm:ml-3 sm:mt-0 sm:w-auto"
							use:enhance
							on:submit|preventDefault={() => closeModal()}
						>
							<input type="hidden" name="bookId" value={selectedBook.bookId} />
							<button
								type="submit"
								class="inline-flex w-full justify-center rounded-md bg-green-600
                       px-5 py-3 text-sm font-semibold text-white shadow-sm
                       hover:bg-green-500 sm:w-auto"
							>
								Finish Book
							</button>
						</form>

						<!-- Form for removing from Currently Reading -->
						<form
							method="post"
							action="?/removeFromCurrentlyReading"
							class="mt-3 inline-flex w-full justify-center sm:ml-3 sm:mt-0 sm:w-auto"
							use:enhance
							on:submit|preventDefault={() => closeModal()}
						>
							<input type="hidden" name="bookId" value={selectedBook?.bookId} />
							<button
								type="submit"
								class="inline-flex w-full justify-center rounded-md bg-red-600 px-5 py-3 text-sm font-semibold text-white shadow-sm hover:bg-red-500 sm:w-auto"
							>
								Remove
							</button>
						</form>

						<!-- Cancel button closes the modal locally -->
						<button
							type="button"
							class="mt-3 inline-flex w-full justify-center rounded-md bg-white
                     px-5 py-3 text-sm font-semibold text-gray-900 shadow-sm
                     ring-1 ring-inset ring-gray-300 hover:bg-gray-50
                     sm:ml-3 sm:mt-0 sm:w-auto"
							on:click={closeModal}
						>
							Cancel
						</button>
					</div>
				</div>
			</div>
		</div>
	</div>
{/if}

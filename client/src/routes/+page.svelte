<script lang="ts">
	// Import any types you need and the enhance action for forms
	import type { PageData } from './$types';
	import { enhance } from '$app/forms';
	import type { Book } from '$lib/types';

	export let data: PageData;

	let modalVisible = false;

	// The selectedBook here is the raw Book object from the API.
	let selectedBook: Book | null = null;
	let newPageCount: number | '' = '';

	// Open the modal – note that we pass the raw Book object.
	function openModal(book: Book) {
		selectedBook = book;
		newPageCount = '';
		modalVisible = true;
	}

	function closeModal() {
		modalVisible = false;
		selectedBook = null;
		newPageCount = '';
	}

	// Add this calculation before the template
	const booksPerMonth = 2.1; // Replace with actual calculation based on your data
</script>

<!-- Outer container preserving your original layout classes -->
<div class="flex flex-grow flex-col">
	<main class="flex-grow">
		<!-- Current Reads Section -->
		<section class="mx-8 mt-16 flex flex-col items-start px-4 md:mx-16 lg:mx-40">
			<div class="mb-8 w-full text-left">
				<h1 class="text-4xl font-bold text-gray-800 md:text-5xl lg:text-6xl">Currently Reading</h1>
			</div>

			{#if !data.profile || !data.profile.currentlyReading || data.profile.currentlyReading.length === 0}
				<div class="flex w-full items-center justify-center py-16">
					<p class="text-4xl text-gray-500">Nothing... Get To Reading!</p>
				</div>
			{:else}
				<div class="grid w-full grid-cols-1 gap-8 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
					{#each data.profile.currentlyReading as reading}
						{#if reading.Book}
							<div class="flex w-full transform flex-col overflow-hidden rounded-lg shadow-lg duration-300 hover:scale-105">
								<div class="flex h-64 w-full items-center justify-center bg-gray-300 md:h-72 lg:h-64">
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
									<p class="mb-4 line-clamp-1 h-6 text-gray-600">{reading.Book.authors?.[0] ?? 'Unknown Author'}</p>

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
											<a href={`/books/${reading.Book.bookId}`} class="text-blue-500 hover:text-blue-700">
												Details
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
		<section class="mx-8 mt-16 flex flex-col items-center py-4 md:mx-16 lg:mx-40">

			<div class="flex flex-col items-center">
				<div class="mb-8 w-full text-left">
					<h1 class="text-2xl font-bold text-gray-600 md:text-3xl lg:text-4xl">2025 Reading Challenge</h1>
					<p class="text-gray-600 my-4">
						Read 100 books in 2025.
					</p>
				<div class="mt-4 w-full">
					<!-- Progress Bar Container -->
					<div class="h-8 w-full overflow-hidden border-2 border-gray-300 rounded-full bg-gray-200">
						<div
							class="h-full bg-blue-500 transition-all duration-300"
							style="width: 25%;"
						></div>
					</div>

					<!-- Stats Below Progress Bar -->
					<div class="mt-2 flex items-center justify-between text-sm text-gray-600">
						<span>25 books read</span>
						<span>75 books remaining</span>
					</div>

					<!-- Additional Stats -->
					<div class="mt-4 grid grid-cols-3 gap-4">
						<div class="rounded-lg bg-gray-100 p-4 text-center">
							<p class="text-2xl font-bold text-blue-500">25%</p>
							<p class="text-xs text-gray-600">Complete</p>
						</div>
						<div class="rounded-lg bg-gray-100 p-4 text-center">
							<p class="text-2xl font-bold text-blue-500">2.1</p>
							<p class="text-xs text-gray-600">Books/Month</p>
						</div>
						<div class="rounded-lg bg-gray-100 p-4 text-center">
							<p class="text-2xl font-bold text-blue-500">8.3</p>
							<p class="text-xs text-gray-600">Books/Month Needed</p>
						</div>
					</div>

					<!-- Reading Challenge Status -->
					<div class="mt-6 flex items-center gap-4">
						{#if booksPerMonth >= 8.3}
							<div class="flex items-center gap-2">
								<div class="h-3 w-3 rounded-full bg-green-500"></div>
								<span class="text-sm text-gray-600">On Track</span>
							</div>
						{:else if booksPerMonth >= 6}
							<div class="flex items-center gap-2">
								<div class="h-3 w-3 rounded-full bg-yellow-500"></div>
								<span class="text-sm text-gray-600">Slightly Behind</span>
							</div>
						{:else}
							<div class="flex items-center gap-2">
								<div class="h-3 w-3 rounded-full bg-red-500"></div>
								<span class="text-sm text-gray-600">Behind Schedule</span>
							</div>
						{/if}
						</div>
					</div>
				</div>

				<!-- February Reading Challenge -->

				<div class="mb-8 w-full text-left py-16">
					<h1 class="text-2xl font-bold text-gray-600 md:text-3xl lg:text-4xl">February Reading Challenge</h1>
					<p class="text-gray-600 my-4">
						Read 6 books in February.
					</p>
				<div class="mt-4 w-full">
					<!-- Progress Bar Container -->
					<div class="h-8 w-full overflow-hidden border-2 border-gray-300 rounded-full bg-gray-200">
						<div
							class="h-full bg-blue-500 transition-all duration-300"
							style="width: 80%;"
						></div>
					</div>

					<!-- Stats Below Progress Bar -->
					<div class="mt-2 flex items-center justify-between text-sm text-gray-600">
						<span>2 books read</span>
						<span>4 books remaining</span>
					</div>

					<!-- Additional Stats -->
					<div class="mt-4 grid grid-cols-3 gap-4">
						<div class="rounded-lg bg-gray-100 p-4 text-center">
							<p class="text-2xl font-bold text-blue-500">80%</p>
							<p class="text-xs text-gray-600">Complete</p>
						</div>
						<div class="rounded-lg bg-gray-100 p-4 text-center">
							<p class="text-2xl font-bold text-blue-500">1.5</p>
							<p class="text-xs text-gray-600">Books/week</p>
						</div>
						<div class="rounded-lg bg-gray-100 p-4 text-center">
							<p class="text-2xl font-bold text-blue-500">1.5</p>
							<p class="text-xs text-gray-600">Books/Week Needed</p>
						</div>
					</div>

					<!-- Reading Challenge Status -->
					<div class="mt-6 flex items-center gap-4">
						{#if booksPerMonth >= 2.1}
							<div class="flex items-center gap-2">
								<div class="h-3 w-3 rounded-full bg-green-500"></div>
								<span class="text-sm text-gray-600">On Track</span>
							</div>
						{:else if booksPerMonth >= 2.1}
							<div class="flex items-center gap-2">
								<div class="h-3 w-3 rounded-full bg-yellow-500"></div>
								<span class="text-sm text-gray-600">Slightly Behind</span>
							</div>
						{:else}
							<div class="flex items-center gap-2">
								<div class="h-3 w-3 rounded-full bg-red-500"></div>
								<span class="text-sm text-gray-600">Behind Schedule</span>
							</div>
						{/if}
						</div>
					</div>

				<!-- Pages Challenge -->

				<div class="mb-8 w-full text-left pt-16">
					<h1 class="text-2xl font-bold text-gray-600 md:text-3xl lg:text-4xl">2025 Pages Challenge</h1>
					<p class="text-gray-600 my-4">
						Read 10,000 pages in 2025.
					</p>
				<div class="mt-4 w-full">
					<!-- Progress Bar Container -->
					<div class="h-8 w-full overflow-hidden border-2 border-gray-300 rounded-full bg-gray-200">
						<div
							class="h-full bg-blue-500 transition-all duration-300"
							style="width: 80%;"
						></div>
					</div>

					<!-- Stats Below Progress Bar -->
					<div class="mt-2 flex items-center justify-between text-sm text-gray-600">
						<span>2,000 pages read</span>
						<span>8,000 pages remaining</span>
					</div>

					<!-- Additional Stats -->
					<div class="mt-4 grid grid-cols-3 gap-4">
						<div class="rounded-lg bg-gray-100 p-4 text-center">
							<p class="text-2xl font-bold text-blue-500">20%</p>
							<p class="text-xs text-gray-600">Complete</p>
						</div>
						<div class="rounded-lg bg-gray-100 p-4 text-center">
							<p class="text-2xl font-bold text-blue-500">200</p>
							<p class="text-xs text-gray-600">Pages/Day</p>
						</div>
						<div class="rounded-lg bg-gray-100 p-4 text-center">
							<p class="text-2xl font-bold text-blue-500">166.7</p>
							<p class="text-xs text-gray-600">Pages/Day Needed</p>
						</div>
					</div>

					<!-- Reading Challenge Status -->
					<div class="mt-6 flex items-center gap-4">
						{#if booksPerMonth >= 5}
							<div class="flex items-center gap-2">
								<div class="h-3 w-3 rounded-full bg-green-500"></div>
								<span class="text-sm text-gray-600">On Track</span>
							</div>
						{:else if booksPerMonth >= 1}
							<div class="flex items-center gap-2">
								<div class="h-3 w-3 rounded-full bg-yellow-500"></div>
								<span class="text-sm text-gray-600">Slightly Behind</span>
							</div>
						{:else}
							<div class="flex items-center gap-2">
								<div class="h-3 w-3 rounded-full bg-red-500"></div>
								<span class="text-sm text-gray-600">Behind Schedule</span>
							</div>
						{/if}
						</div>
					</div>

				</div>
			</div>
		</section>

		<!-- Divider -->
		<hr class="my-8 border-gray-300" />

		<!-- To Be Read Section -->
		{#if data.profile && data.profile.lists && data.profile.lists.toBeRead}
			<section class="mx-8 mt-16 flex flex-col items-start py-4 md:mx-16 lg:mx-40">
				<div class="mb-8 w-full text-left">
					<h1 class="text-4xl font-bold text-gray-600 md:text-5xl lg:text-4xl">
						To Be Read ({data.profile.lists.toBeRead.length})
					</h1>
					<a href="/lists?list=To Be Read" class="mt-2 text-lg font-semibold text-blue-500 hover:text-blue-700">View All</a>
				</div>

				<!-- Grid Container -->
				<div class="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-5">
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
									class="h-64 w-full rounded-lg sm:h-72 md:h-80 lg:h-64"
								/>
								<div
									class="absolute inset-0 flex flex-col items-center justify-center gap-2 rounded-lg bg-black/50 opacity-0 transition-opacity duration-200 group-hover:opacity-100"
								>
									<a href={`/books/${book.bookId}`} class="w-3/4">
										<button
											type="button"
											class="h-8 w-full text-center rounded-full bg-white text-gray-800 transition-all duration-200 hover:bg-blue-500 hover:text-white"
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
											class="h-8 w-full rounded-full bg-white text-gray-800 transition-all duration-200 hover:bg-blue-500 hover:text-white"
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
			</section>
		{/if}

		<!-- Divider -->
		<hr class="my-16 border-gray-300" />

		<!-- Read Section (if available) -->
		{#if data.profile && data.profile.lists && data.profile.lists.read}
			<section class="mx-8 mt-16 flex flex-col items-start px-4 md:mx-16 lg:mx-40">
				<div class="mb-8 w-full text-left">
					<h1 class="text-4xl font-bold text-gray-600 md:text-5xl lg:text-4xl">
						Read ({data.profile.lists.read.length})
					</h1>
					<a href="/lists?list=Read" class="mt-2 text-lg font-semibold text-blue-500 hover:text-blue-700">View All</a>
				</div>

				<!-- Grid Container -->
				<div class="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-5">
					{#each (data.profile.lists.read.slice(0, 4) || []).reverse() as book}
						<div class="flex flex-col">
							<!-- Similar markup as the "To Be Read" section -->
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
									class="h-64 w-full rounded-lg sm:h-72 md:h-80 lg:h-64"
								/>
								<div
									class="absolute inset-0 flex flex-col items-center justify-center gap-2 rounded-lg bg-black/50 opacity-0 transition-opacity duration-200 group-hover:opacity-100"
								>
									<a href={`/books/${book.bookId}`} class="w-3/4">
										<button
											type="button"
											class="h-8 w-full text-center rounded-full bg-white text-gray-800 transition-all duration-200 hover:bg-blue-500 hover:text-white"
										>
											Details
										</button>
									</a>

									<form method="post" action="?/startReading" class="w-3/4" use:enhance>
										<input type="hidden" name="bookId" value={book.bookId} />
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
			</section>
		{/if}

		<!-- Divider -->
		<hr class="my-16 border-gray-300" />

		<!-- Custom Lists Section -->
		{#if data.profile && data.profile.lists && data.profile.lists.customLists}
			{#each Object.entries(data.profile.lists.customLists) as [listName, books]}
				<section class="mx-8 mt-16 flex flex-col items-start px-4 md:mx-16 lg:mx-40">
					<div class="mb-8 w-full text-left">
						<h1 class="text-4xl font-bold text-gray-600 md:text-5xl lg:text-4xl">
							{listName} ({books.length})
						</h1>
						<a href="/lists?list={listName}" class="mt-2 text-lg font-semibold text-blue-500 hover:text-blue-700">View All</a>
					</div>

					<!-- Grid Container -->
					<div class="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-5">
						{#each books.slice(0, 4).reverse() as book}
							<div
								class="flex transform flex-col rounded-lg bg-gray-300 shadow-lg transition-shadow duration-300 hover:scale-105 hover:shadow-2xl"
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
								<!-- Optionally add "Details" or "Remove" forms here -->
							</div>
						{/each}

						{#if books.length > 4}
							<div class="flex items-center justify-center text-xl text-gray-600">...</div>
						{/if}
					</div>
				</section>
			{/each}
		{/if}
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
			<div class="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
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
											Current Page: {selectedBook.progress?.lastPageRead || 0} / {selectedBook.totalPages || 'Unknown'}
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
									<label for="newPageCount" class="block text-sm font-medium text-gray-700">
										New Page Count
									</label>
									<input
										id="newPageCount"
										type="number"
										min={selectedBook.progress?.lastPageRead || 0}
										bind:value={newPageCount}
										class="mt-4 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
										placeholder="Enter new page count"
										required
									/>
								</div>
							</div>
						</div>
					</div>

					<!-- Footer with action buttons -->
					<div class="bg-gray-50 px-8 py-6 sm:flex sm:flex-row-reverse">
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
								class="inline-flex w-full justify-center rounded-md bg-blue-600 px-5 py-3 text-sm font-semibold text-white shadow-sm hover:bg-blue-500 sm:w-auto"
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
								class="inline-flex w-full justify-center rounded-md bg-green-600 px-5 py-3 text-sm font-semibold text-white shadow-sm hover:bg-green-500 sm:w-auto"
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
							<input type="hidden" name="bookId" value={selectedBook.bookId} />
							<button
								type="submit"
								class="inline-flex w-full justify-center rounded-md bg-red-600 px-5 py-3 text-sm font-semibold text-white shadow-sm hover:bg-red-500 sm:w-auto"
							>
								Remove
							</button>
						</form>

						<!-- Cancel button -->
						<button
							type="button"
							class="mt-3 inline-flex w-full justify-center rounded-md bg-white px-5 py-3 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:ml-3 sm:mt-0 sm:w-auto"
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

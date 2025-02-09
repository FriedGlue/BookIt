<!-- client/src/routes/reading-log/+page.svelte -->
<script lang="ts">
	import type { PageData } from './$types';
	import type { ReadingLogItem } from '$lib/types';
	import { enhance } from '$app/forms';
	import Sidebar from '$lib/components/Sidebar.svelte';

	export let data: PageData;

	let error: string | false = false;
	let isDeleting = false;
	let isUpdating = false;

	// In case the API returns an empty array
	$: readingLog = data.readingLog ?? [];

	const options = ['Calendar', 'Charts & Graphs', 'List'];
	let selectedList = 'Calendar';

	// -- MODAL STATE FOR CREATING/EDITING ENTRIES --
	let showEntryForm = false;
	let isEditingEntry = false;
	// Form fields for a reading log entry
	let entryId = '';
	let entryDate = new Date().toISOString().split('T')[0];
	let bookThumbnail = '';
	let title = '';
	let pagesRead = '';
	let notes = '';

	// Open the form for a new (past) entry.
	function openNewEntryForm() {
		isEditingEntry = false;
		entryId = '';
		entryDate = new Date().toISOString().split('T')[0];
		bookThumbnail = '';
		title = '';
		pagesRead = '';
		notes = '';
		showEntryForm = true;
	}

	// Open the form for editing an existing entry.
	function openEditEntryForm(entry: ReadingLogItem) {
		isEditingEntry = true;
		entryId = entry._id;
		entryDate = new Date(entry.date).toISOString().split('T')[0];
		title = entry.title;
		pagesRead = entry.pagesRead?.toString() ?? '';
		notes = entry.notes ?? '';
		showEntryForm = true;
	}

	// Helper function to generate calendar days
	function generateCalendarDays() {
		const days = [];
		for (let i = 1; i <= 28; i++) {
			// Find reading log entries for this day
			const dayEntries = readingLog.filter((log: ReadingLogItem) => {
				const logDate = new Date(log.date);
				return logDate.getDate() === i;
			});

			days.push({
				day: i,
				entries: dayEntries
			});
		}
		return days;
	}

	$: calendarDays = generateCalendarDays();
</script>

<div class="flex min-h-screen">
	<Sidebar
		title="View By"
		items={options}
		selectedItem={selectedList}
		onSelect={(item) => (selectedList = item)}
	/>

	<section class="container mx-auto p-4">
		{#if error}
			<div
				class="relative mb-4 rounded border border-red-400 bg-red-100 px-4 py-3 text-red-700"
				role="alert"
			>
				<span class="block sm:inline">{error}</span>
			</div>
		{/if}

		{#if readingLog.length === 0}
			<p class="text-gray-600">No reading log entries found.</p>
		{/if}

		{#if readingLog.length > 0 && selectedList === 'Calendar'}
			<div class="my-4 flex h-full w-full flex-col justify-center gap-4">
				<div class="my-4 flex items-center justify-center gap-4">
					<button class="mb-4 text-2xl font-bold" aria-label="Previous month">
						<svg
							xmlns="http://www.w3.org/2000/svg"
							class="h-6 w-6"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M15 19l-7-7 7-7"
							/>
						</svg>
					</button>
					<h1 class="mb-4 flex-grow text-center text-6xl font-bold">
						{new Date().toLocaleDateString('en-US', { month: 'long', year: 'numeric' })}
					</h1>
					<button class="mb-4 text-2xl font-bold" aria-label="Next month">
						<svg
							xmlns="http://www.w3.org/2000/svg"
							class="h-6 w-6"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M9 5l7 7-7 7"
							/>
						</svg>
					</button>
				</div>

				<div class="calendar my-32 grid grid-cols-7 gap-2 rounded-lg p-4">
					<!-- Days of the week -->
					<div class="text-center font-bold">Sun</div>
					<div class="text-center font-bold">Mon</div>
					<div class="text-center font-bold">Tue</div>
					<div class="text-center font-bold">Wed</div>
					<div class="text-center font-bold">Thu</div>
					<div class="text-center font-bold">Fri</div>
					<div class="text-center font-bold">Sat</div>

					<!-- Pad the first week (assuming February starts on a Thursday) -->
					{#each Array(6) as _}
						<div class="h-24 border bg-gray-100 p-2"></div>
					{/each}

					<!-- Calendar days -->
					{#each calendarDays as day}
						<div class="min-h-52 min-w-32 p-2">
							<div class="mb-2 font-bold">{day.day}</div>
							<div class="flex flex-col gap-2">
								{#each [...new Set(day.entries.map((e: ReadingLogItem) => e.bookThumbnail))] as thumbnail (thumbnail)}
									<img
										src={thumbnail as string}
										alt={`Cover of ${day.entries.find((e: ReadingLogItem) => e.bookThumbnail === thumbnail)?.title} by ${day.entries.find((e: ReadingLogItem) => e.bookThumbnail === thumbnail)?.author}`}
										class="h-32 w-auto rounded object-contain transition-transform hover:scale-150"
										title={day.entries.find((e: ReadingLogItem) => e.bookThumbnail === thumbnail)
											?.title}
									/>
								{/each}
							</div>
						</div>
					{/each}
				</div>
			</div>
		{/if}

		{#if readingLog.length > 0 && selectedList === 'List'}
			<div class="mb-6 flex justify-end">
				<button
                	class="rounded-full bg-blue-500 px-6 py-2 text-white hover:bg-blue-600"
					on:click={openNewEntryForm}
				>
					Add Past Entry
				</button>
			</div>
			<table class="min-w-full border border-gray-200 bg-white">
				<thead>
					<tr class="bg-gray-100">
						<th class="border px-4 py-2">Date</th>
						<th class="border px-4 py-2">Cover</th>
						<th class="border px-4 py-2">Title</th>
						<th class="border px-4 py-2">Pages Read</th>
						<th class="border px-4 py-2">Notes</th>
						<th class="border px-4 py-2">Actions</th>
					</tr>
				</thead>
				<tbody>
					{#each readingLog as log}
						<tr class="text-center">
							<td class="border px-4 py-2">
								{new Date(log.date).toLocaleDateString()}
							</td>
							<td class="border px-4 py-2">
								{#if log.bookThumbnail}
									<img
										src={log.bookThumbnail}
										alt="Book cover"
										class="h-52 w-full rounded-md object-contain"
									/>
								{/if}
							</td>
							<td class="border px-4 py-2 text-xl">{log.title}</td>
							<td class="text-md border px-4 py-2">{log.pagesRead}</td>
							<td class="text-md border px-4 py-2">{log.notes}</td>
							<td class="border px-4 py-2 space-y-2">
								<!-- Edit Button -->
								<button
									class="mb-2 w-full rounded-md bg-blue-500 px-3 py-2 text-white hover:bg-blue-400"
									on:click={() => openEditEntryForm(log)}
								>
									Edit Entry
								</button>
								<!-- Delete Form -->
								<form
									method="POST"
									action="?/removeFromReadingLog"
									use:enhance={() => {
										isDeleting = true;
										return async ({ update }) => {
											await update();
											isDeleting = false;
										};
									}}
								>
									<input type="hidden" name="readingLogEntryId" value={log._id} />
									<button
										type="submit"
										class="w-full rounded-md bg-red-600 px-3 py-2 text-white hover:bg-red-500 disabled:cursor-not-allowed disabled:opacity-50"
										disabled={isDeleting}
									>
										{#if isDeleting}
											Deleting...
										{:else}
											Delete Entry
										{/if}
									</button>
								</form>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		{/if}

		{#if selectedList === 'Charts & Graphs'}
			<div class="mb-6 flex justify-end">
				<button
					class="rounded-full bg-blue-500 px-6 py-2 text-white hover:bg-blue-600"
					on:click={openNewEntryForm}
				>
					Add Past Entry
				</button>
			</div>

			<h1 class="mb-6 text-2xl font-bold">Coming Soon</h1>
		{/if}
	</section>
</div>

<!-- Modal for Adding / Editing Reading Log Entry -->
{#if showEntryForm}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50">
		<div class="w-full max-w-2xl rounded-lg bg-white p-8">
			<h2 class="mb-6 text-2xl font-bold">
				{isEditingEntry ? 'Edit Reading Log Entry' : 'Add Past Reading Log Entry'}
			</h2>
			<form
				method="POST"
				action={isEditingEntry ? '?/update' : '?/create'}
				use:enhance={() => {
					return async ({ update }) => {
						// After a successful update/creation, close the modal.
						showEntryForm = false;
						// Optionally update the local readingLog list
						await update();
					};
				}}
				class="space-y-4"
			>
				{#if isEditingEntry}
					<input type="hidden" name="readingLogEntryId" value={entryId} />
				{/if}
				<div>
					<label for="entryDate" class="mb-2 block text-sm font-medium text-gray-700">Date</label>
					<input
						type="date"
						id="entryDate"
						name="date"
						bind:value={entryDate}
						class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
						required
					/>
				</div>
				<div>
					<label for="title" class="mb-2 block text-sm font-medium text-gray-700">Title</label>
					<input
						type="text"
						id="title"
						name="title"
						bind:value={title}
						class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
						required
					/>
				</div>
				<div>
					<label for="pagesRead" class="mb-2 block text-sm font-medium text-gray-700">Pages Read</label>
					<input
						type="number"
						id="pagesRead"
						name="pagesRead"
						bind:value={pagesRead}
						min="1"
						class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
						required
					/>
				</div>
				<div>
					<label for="notes" class="mb-2 block text-sm font-medium text-gray-700">Notes</label>
					<textarea
						id="notes"
						name="notes"
						bind:value={notes}
						class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
						rows="3"
					></textarea>
				</div>
			<div class="mt-6 flex justify-end space-x-4">
					<button
						type="button"
						class="rounded-full bg-gray-200 px-6 py-2 text-gray-700 hover:bg-gray-300"
						on:click={() => (showEntryForm = false)}
					>
						Cancel
					</button>
					<button
						type="submit"
						class="rounded-full bg-blue-500 px-6 py-2 text-white hover:bg-blue-600"
					>
						{isEditingEntry ? 'Update Entry' : 'Add Entry'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
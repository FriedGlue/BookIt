<!-- client/src/routes/reading-log/+page.svelte -->
<script lang="ts">
	import type { PageData } from './$types';
	import type { ReadingLogItem } from '$lib/types';
	import { enhance } from '$app/forms';
	import Sidebar from '$lib/components/Sidebar.svelte';

	export let data: PageData;

	let error = false;
	let isDeleting = false;
	let isUpdating = false;

	// In case the API returns an empty array
	$: readingLog = data.readingLog ?? [];

	const options = ['Calendar', 'Charts & Graphs', 'List', ];
	let selectedList = 'Calendar';

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
							</td><td class="border px-4 py-2 text-xl">{log.title}</td>
							<td class="text-md border px-4 py-2">{log.pagesRead}</td>
							<td class="text-md border px-4 py-2">{log.notes}</td>
							<td class="border px-4 py-2">
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
									class="mt-3 inline-flex w-full justify-center sm:ml-3 sm:mt-0 sm:w-auto"
								>
									<input type="hidden" name="readingLogEntryId" value={log._id} />
									<button
										type="submit"
										class="inline-flex w-full justify-center rounded-md bg-red-600
                    px-5 py-3 text-sm font-semibold text-white shadow-sm
                    hover:bg-red-500 disabled:cursor-not-allowed disabled:opacity-50 sm:w-auto"
										disabled={isDeleting}
									>
										{#if isDeleting}
											Deleting...
										{:else}
											Delete Entry
										{/if}
									</button>
								</form>
							</td></tr
						>
					{/each}
				</tbody>
			</table>
		{/if}
	</section>
</div>

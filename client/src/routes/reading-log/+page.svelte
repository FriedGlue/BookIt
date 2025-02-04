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

  const options = ['Calendar', 'List'];
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
		onSelect={(item) => selectedList = item}
	/>

<section class="container mx-auto p-4">
  
  {#if error}
    <div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative mb-4" role="alert">
      <span class="block sm:inline">{error}</span>
    </div>
  {/if}

  {#if readingLog.length === 0 }
    <p class="text-gray-600">No reading log entries found.</p>
  {/if}

  {#if readingLog.length > 0 && selectedList === 'Calendar'}
  <div class="flex flex-col justify-center w-full h-full my-4 gap-4">
    <div class="flex items-center justify-center my-4 gap-4">
      <button class="text-2xl font-bold mb-4" aria-label="Previous month">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
        </svg>
      </button>
      <h1 class="text-6xl font-bold mb-4 flex-grow text-center">
        {new Date().toLocaleDateString('en-US', { month: 'long', year: 'numeric' })}
      </h1>
      <button class="text-2xl font-bold mb-4" aria-label="Next month">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
      </button>
    </div>

    <div class="calendar grid my-32 grid-cols-7 gap-2 rounded-lg p-4">
      <!-- Days of the week -->
      <div class="font-bold text-center">Sun</div>
      <div class="font-bold text-center">Mon</div>
      <div class="font-bold text-center">Tue</div>
      <div class="font-bold text-center">Wed</div>
      <div class="font-bold text-center">Thu</div>
      <div class="font-bold text-center">Fri</div>
      <div class="font-bold text-center">Sat</div>

      <!-- Pad the first week (assuming February starts on a Thursday) -->
      {#each Array(6) as _}
        <div class="border p-2 h-24 bg-gray-100"></div>
      {/each}

      <!-- Calendar days -->
      {#each calendarDays as day}
        <div class="p-2 min-h-52 min-w-32">
          <div class="font-bold mb-2">{day.day}</div>
          <div class="flex flex-col gap-2">
            {#each [...new Set(day.entries.map((e: ReadingLogItem) => e.bookThumbnail))] as thumbnail (thumbnail)}
              <img
                src={thumbnail as string}
                alt={`Cover of ${day.entries.find((e: ReadingLogItem) => e.bookThumbnail === thumbnail)?.title} by ${day.entries.find((e: ReadingLogItem) => e.bookThumbnail === thumbnail)?.author}`}
                class="h-32 w-auto object-contain rounded hover:scale-150 transition-transform"
                title={day.entries.find((e: ReadingLogItem) => e.bookThumbnail === thumbnail)?.title}
              />
            {/each}
          </div>
        </div>
      {/each}
    </div>
  </div>
  {/if}

  {#if readingLog.length > 0 && selectedList === 'List'}
    <table class="min-w-full bg-white border border-gray-200">
      <thead>
        <tr class="bg-gray-100">
          <th class="py-2 px-4 border">Date</th>
          <th class="py-2 px-4 border">Cover</th>
          <th class="py-2 px-4 border">Title</th>
          <th class="py-2 px-4 border">Pages Read</th>
          <th class="py-2 px-4 border">Notes</th>
          <th class="py-2 px-4 border">Actions</th>
        </tr>
      </thead>
      <tbody>
        {#each readingLog as log}
          <tr class="text-center">
            <td class="py-2 px-4 border">
              {new Date(log.date).toLocaleDateString()}
            </td>
            <td class="py-2 px-4 border">
              {#if log.bookThumbnail}
                <img
                  src={log.bookThumbnail}
                  alt="Book cover"
				  class="h-52 w-full rounded-md object-contain"
                />
              {/if}
            <td class="py-2 px-4 border text-xl">{log.title}</td>
            <td class="py-2 px-4 border text-md">{log.pagesRead}</td>
            <td class="py-2 px-4 border text-md">{log.notes}</td>
            <td class="py-2 px-4 border">
              <form
                method="POST"
                action="?/removeFromReadingLog"
                use:enhance={() =>{
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
                    hover:bg-red-500 sm:w-auto disabled:opacity-50 disabled:cursor-not-allowed"
                  disabled={isDeleting}
                >
                  {#if isDeleting}
                    Deleting...
                  {:else}
                    Delete Entry
                  {/if}
                </button>
              </form>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</section>
</div>
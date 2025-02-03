<!-- client/src/routes/reading-log/+page.svelte -->
<script lang="ts">
  import type { PageData } from './$types';
  import type { ReadingLogItem } from '$lib/types';
  import { enhance } from '$app/forms';

  export let data: PageData;

  let error = false;
  let isDeleting = false;
  let isUpdating = false;

  // In case the API returns an empty array
  $: readingLog = data.readingLog ?? [];
</script>

<section class="container mx-auto p-4">
  <h1 class="text-3xl font-bold mb-6">Reading Log</h1>
  
  {#if error}
    <div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative mb-4" role="alert">
      <span class="block sm:inline">{error}</span>
    </div>
  {/if}

  {#if readingLog.length > 0}
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
  {:else}
    <p class="text-gray-600">No reading log entries found.</p>
  {/if}
</section>
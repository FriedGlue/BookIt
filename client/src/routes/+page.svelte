<script lang="ts">
  import { onMount } from 'svelte';
  import { BookService } from '$lib/services/bookService';
  import type { DisplayBook, ToBeReadItem, ReadItem } from '$lib/types';

  let books: DisplayBook[] = [];
  let modalVisible = false;
  let selectedBook: DisplayBook | null = null;
  let newPageCount: number | '' = '';
  let toBeReadList: ToBeReadItem[] = [];
  let readList: ReadItem[] = [];
  let customLists: Record<string, any[]> = {};
  let isStartingBook = false;
  let isFinishingBook = false;
  let isRemovingFromList = false;

  const bookService = new BookService();

  onMount(() => {
    // Fetch profile
    (async () => {
        try {
            const profile = await bookService.getProfile();
            console.log('Raw profile data:', profile);
            console.log('Currently reading books:', profile.currentlyReading);
            
            books = (profile.currentlyReading || []).map(item => {
                console.log('Mapping book item:', item);
                console.log('Book title:', item.Book?.title);
                return {
                    bookId: item.Book.bookId,
                    title: item.Book?.title || 'Unknown',
                    author: item.Book?.authors?.[0] || 'Unknown Author',
                    thumbnail: item.Book?.thumbnail || '',
                    progress: item.Book?.progress?.percentage || 0,
                    totalPages: item.Book?.totalPages || 1,
                    currentPage: item.Book?.progress?.lastPageRead || 0,
                    lastUpdated: item.Book?.progress?.lastUpdated || new Date().toISOString()
                };
            });
            console.log('Mapped books:', books);

            toBeReadList = profile.lists?.toBeRead || [];
            readList = profile.lists?.read || [];
            customLists = profile.lists?.customLists || {};
        } catch (error) {
            console.error('Error fetching profile:', error);
        }
    })();
  });

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

  async function submitUpdate() {
      if (!selectedBook || newPageCount === '') return;

      try {
          await bookService.updateBookProgress(selectedBook.bookId, Number(newPageCount));
          
          // Update local state
          const updatedPage = Number(newPageCount);
          const total = selectedBook.totalPages;
          const newProgress = Math.floor((updatedPage / total) * 100);

          selectedBook.progress = newProgress;
          selectedBook.currentPage = updatedPage;
          selectedBook.lastUpdated = new Date().toISOString();
          const index = books.findIndex(b => b.bookId === selectedBook!.bookId);
          if(index !== -1) {
              books[index] = { ...selectedBook };
          }

          closeModal();
      } catch (error) {
          console.error('Error updating book progress:', error);
      }
  }

  async function RemoveFromList(bookId: string, listType: string) {
      try {
          isRemovingFromList = true;
          await bookService.removeFromList(bookId, listType);
          books = books.filter(b => b.bookId !== selectedBook!.bookId);
          closeModal();
      } catch (error) {
          console.error('Error deleting book:', error);
      } finally {
          isRemovingFromList = false;
      }

      // Update local state
      const profile = await bookService.getProfile();
      if (listType === 'toBeRead') {
          toBeReadList = profile.lists?.toBeRead || [];
      } else if (listType === 'read') {
          readList = profile.lists?.read || [];
      }   
    }
  
  async function RemoveFromCurrentlyReading(bookId: string) {
      try {
          isRemovingFromList = true;
          await bookService.removeFromCurrentlyReading(bookId);
          books = books.filter(b => b.bookId !== selectedBook!.bookId);
          closeModal();
      } catch (error) {
          console.error('Error deleting book:', error);
      } finally {
          isRemovingFromList = false;
      }

      // Update local state
      const profile = await bookService.getProfile();
      books = profile.currentlyReading.map(item => ({
          bookId: item.Book.bookId,
          title: item.Book.title || 'Untitled',
          author: item.Book.authors ? item.Book.authors[0] : 'Unknown Author',
          thumbnail: item.Book.thumbnail || '',
          progress: item.Book.progress ? item.Book.progress.percentage : 0,
          totalPages: item.Book.totalPages || 1,
          currentPage: item.Book.progress ? item.Book.progress.lastPageRead : 0,
          lastUpdated: item.Book.progress ? item.Book.progress.lastUpdated : new Date().toISOString()
      }));
  }

  async function startReading(bookId: string, listName: string) {
    if (isStartingBook) return;
    
    try {
        isStartingBook = true;
        await bookService.startReading(bookId, listName);
        
        // Update local state
        const profile = await bookService.getProfile();
        
        // Update currently reading list
        books = profile.currentlyReading.map(item => ({
            bookId: item.Book.bookId,
            title: item.Book.title || 'Untitled',
            author: item.Book.authors ? item.Book.authors[0] : 'Unknown Author',
            thumbnail: item.Book.thumbnail || '',
            progress: item.Book.progress ? item.Book.progress.percentage : 0,
            totalPages: item.Book.totalPages || 1,
            currentPage: item.Book.progress ? item.Book.progress.lastPageRead : 0,
            lastUpdated: item.Book.progress ? item.Book.progress.lastUpdated : new Date().toISOString()
        }));

        // Update source list
        if (listName === 'toBeRead') {
            toBeReadList = profile.lists?.toBeRead || [];
        } else if (listName === 'read') {
            readList = profile.lists?.read || [];
        } else {
            customLists = profile.lists?.customLists || {};
        }
    } catch (error) {
        console.error('Error starting book:', error);
    } finally {
        isStartingBook = false;
    }
  }

  async function finishBook() {
    if (!selectedBook || isFinishingBook) return;

    try {
        isFinishingBook = true;
        await bookService.finishReading(selectedBook.bookId);
        
        // Update local state
        const profile = await bookService.getProfile();
        
        // Update currently reading list
        books = (profile.currentlyReading || []).map(item => ({
            bookId: item.Book.bookId,
            title: item.Book.title || 'Untitled',
            author: item.Book.authors ? item.Book.authors[0] : 'Unknown Author',
            thumbnail: item.Book.thumbnail || '',
            progress: item.Book.progress ? item.Book.progress.percentage : 0,
            totalPages: item.Book.totalPages || 1,
            currentPage: item.Book.progress ? item.Book.progress.lastPageRead : 0,
            lastUpdated: item.Book.progress ? item.Book.progress.lastUpdated : new Date().toISOString()
        }));

        // Update read list
        readList = profile.lists?.read || [];
        
        closeModal();
    } catch (error) {
        console.error('Error finishing book:', error);
    } finally {
        isFinishingBook = false;
    }
  }
</script>

<div class="flex flex-col flex-grow">
  <main class="flex-grow">
    <section class="mt-16 mx-8 md:mx-16 lg:mx-40 flex flex-col items-start px-4">
      <div class="w-full text-left mb-8">
          <h1 class="text-4xl md:text-5xl lg:text-6xl font-bold text-gray-800">Current Reads</h1>
      </div>
      {#if books.length === 0}
        <div class="w-full flex justify-center items-center py-16">
          <p class="text-2xl text-gray-500">No current reads</p>
        </div>
      {:else}
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-8 w-full">
          {#each books as book}
              <div class="flex flex-col rounded-lg overflow-hidden w-full duration-300 transform hover:scale-105 shadow-lg">
                  <div class="w-full h-64 md:h-72 lg:h-64 bg-gray-300 flex items-center justify-center">
                      {#if book.bookId}
                          <img
                              src={book.thumbnail}
                              alt={`Cover of ${book.title} by ${book.author}`}
                              class="max-w-full max-h-full object-contain"
                              loading="lazy"
                          />
                      {:else}
                          <div class="text-gray-500">No Cover Available</div>
                      {/if}
                  </div>
                  <div class="flex flex-col flex-grow p-6 bg-white">
                      <h2 
                          class="text-xl font-semibold text-gray-800 mb-2 h-14 line-clamp-2 overflow-hidden"
                          title={book.title}
                      >
                          {book.title}
                      </h2>
                      <p class="text-gray-600 mb-4 h-6 line-clamp-1">{book.author}</p>
                      <div class="w-full bg-gray-200 rounded-full h-4 overflow-hidden">
                          <div
                              class="h-full bg-green-500 transition-all duration-300"
                              style="width: {book.progress}%;"
                          ></div>
                      </div>
                      <div class="mt-2 flex justify-between items-center">
                          <span class="text-sm text-gray-700">{Math.round(book.progress)}%</span>
                          <div class="space-x-2">
                              <button class="text-blue-500 hover:text-blue-700">Details</button>
                          </div>
                      </div>
                      <button 
                          on:click={() => openModal(book)} 
                          class="w-full h-8 mt-4 bg-gray-300 rounded-full hover:bg-blue-500"
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
    <section class="mt-16 mx-8 md:mx-16 lg:mx-40 flex flex-col items-start px-4">
        <div class="w-full text-left mb-8">
            <h1 class="text-4xl md:text-5xl lg:text-4xl font-bold text-gray-600">
                To Be Read ({toBeReadList?.length || 0})
            </h1>
            <button class="text-lg mt-2 font-semibold text-blue-500">View All</button>
        </div>
        
        <!-- Grid Container -->
        <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-4">
            {#each (toBeReadList?.slice(0, 4) || []).reverse() as book}
                <div class="flex flex-col">
                    <div class="relative w-full bg-gray-300 rounded-lg hover:shadow-2xl transition-shadow duration-300 transform hover:scale-105 group">
                        <img
                            src={book.thumbnail || 'default-cover-image-url'}
                            alt="Book Cover"
                            loading="lazy"
                            decoding="async"
                            on:load={(e) => (e.currentTarget as HTMLImageElement).style.opacity = '1'}
                            style="opacity: 0; transition: opacity 0.3s"
                            class="w-full h-64 sm:h-72 md:h-80 lg:h-64 rounded-lg"
                        />
                        <div class="absolute inset-0 flex flex-col items-center justify-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity duration-200 bg-black/50 rounded-lg">
                            <button 
                                class="w-3/4 h-8 bg-white text-gray-800 rounded-full hover:bg-blue-500 hover:text-white disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200"
                                on:click={() => startReading(book.bookId, 'toBeRead')}
                                disabled={isStartingBook}
                            >
                                {isStartingBook ? 'Loading...' : 'Details'}
                            </button>
                            <button 
                                class="w-3/4 h-8 bg-white text-gray-800 rounded-full hover:bg-blue-500 hover:text-white disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200"
                                on:click={() => startReading(book.bookId, 'toBeRead')}
                                disabled={isStartingBook}
                            >
                                {isStartingBook ? 'Starting...' : 'Start'}
                            </button>
                            <button 
                                class="w-3/4 h-8 bg-white text-gray-800 rounded-full hover:bg-blue-500 hover:text-white disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200"
                                on:click={() => RemoveFromList(book.bookId, 'toBeRead')}
                                disabled={isRemovingFromList}
                            >
                                {isRemovingFromList? 'Removing...' : 'Remove'}
                            </button>
                        </div>
                    </div>
                </div>
            {/each}
            {#if (toBeReadList?.length || 0) >= 5}
                <div class="flex items-center justify-center text-gray-600 text-6xl">...</div>
            {/if}
        </div>
    </section>

    <!-- Read Section -->
    <section class="mt-16 mx-8 md:mx-16 lg:mx-40 flex flex-col items-start px-4">
        <div class="w-full text-left mb-8">
            <h1 class="text-4xl md:text-5xl lg:text-4xl font-bold text-gray-600">
                Read ({readList?.length})
            </h1>
            <button class="text-lg mt-2 font-semibold text-blue-500">View All</button>
        </div>
        
        <!-- Grid Container -->
        <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-4">
            {#each (readList?.slice(0, 4) || []).reverse() as book}
                <div class="flex flex-col">
                    <div class="relative w-full bg-gray-300 rounded-lg hover:shadow-2xl transition-shadow duration-300 transform hover:scale-105 group">
                        <img
                            src={book.thumbnail || 'default-cover-image-url'}
                            alt="Book Cover"
                            loading="lazy"
                            decoding="async"
                            on:load={(e) => (e.currentTarget as HTMLImageElement).style.opacity = '1'}
                            style="opacity: 0; transition: opacity 0.3s"
                            class="w-full h-64 sm:h-72 md:h-80 lg:h-64 rounded-lg"
                        />
                        <div class="absolute inset-0 flex flex-col items-center justify-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity duration-200 bg-black/50 rounded-lg">
                            <button 
                                class="w-3/4 h-8 bg-white text-gray-800 rounded-full hover:bg-blue-500 hover:text-white disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200"
                                on:click={() => startReading(book.bookId, 'read')}
                                disabled={isStartingBook}
                            >
                                {isStartingBook ? 'Loading...' : 'Details'}
                            </button>
                            <button 
                                class="w-3/4 h-8 bg-white text-gray-800 rounded-full hover:bg-blue-500 hover:text-white disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200"
                                on:click={() => startReading(book.bookId, 'read')}
                                disabled={isStartingBook}
                            >
                                {isStartingBook ? 'Starting...' : 'Start'}
                            </button>
                            <button 
                                class="w-3/4 h-8 bg-white text-gray-800 rounded-full hover:bg-blue-500 hover:text-white disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200"
                                on:click={() => RemoveFromList(book.bookId, 'read')}
                                disabled={isRemovingFromList}
                            >
                                {isRemovingFromList? 'Removing...' : 'Remove'}
                            </button>
                        </div>
                    </div>
                </div>
            {/each}
            {#if (toBeReadList?.length || 0) >= 5}
                <div class="flex items-center justify-center text-gray-600 text-6xl">...</div>
            {/if}
        </div>
    </section>

    <!-- Custom Lists Section -->
    {#each Object.entries(customLists) as [listName, books]}
        <section class="mt-16 mx-8 md:mx-16 lg:mx-40 flex flex-col items-start px-4">
            <div class="w-full text-left mb-8">
                <h1 class="text-4xl md:text-5xl lg:text-4xl font-bold text-gray-600">
                    {listName} ({books.length})
                </h1>
                <button class="text-lg mt-2 font-semibold text-blue-500">View All</button>
            </div>
            
            <!-- Grid Container -->
            <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-4">
                {#each books.slice(0, 4).reverse() as book}
                    <div class="flex flex-col rounded-lg bg-gray-300 shadow-lg hover:shadow-2xl transition-shadow duration-300 transform hover:scale-105">
                        <div class="relative w-full">
                            <img
                                src={book.thumbnail || 'default-cover-image-url'}
                                alt="Book Cover"
                                loading="lazy"
                                decoding="async"
                                on:load={(e) => (e.currentTarget as HTMLImageElement).style.opacity = '1'}
                                style="opacity: 0; transition: opacity 0.3s"
                                class="w-full h-64 sm:h-72 md:h-80 lg:h-64 rounded-lg"
                            />
                        </div>
                    </div>
                {/each}
                {#if books.length > 4}
                    <div class="flex items-center justify-center text-gray-600 text-xl">...</div>
                {/if}
            </div>
        </section>
    {/each}
  </main>
  <div class="mb-32"></div>
</div>

<!-- Modal Markup -->
{#if modalVisible}
  <div class="relative z-10" aria-labelledby="modal-title" role="dialog" aria-modal="true">
    <!-- Backdrop -->
    <div class="fixed inset-0 bg-gray-500/75 transition-opacity" aria-hidden="true"></div>

    <div class="fixed inset-0 z-10 w-screen overflow-y-auto">
      <div class="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
        <div class="relative transform overflow-hidden rounded-lg bg-white py-4 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg">
          <div class="bg-white px-4 pb-8 pt-8 sm:p-8 sm:pb-8">
            <div class="sm:flex sm:items-start">
              <div class="mx-auto flex h-12 w-12 shrink-0 items-center justify-center rounded-full bg-blue-100 sm:mx-0 sm:h-10 sm:w-10">
                <svg class="h-6 w-6 text-blue-600" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
                </svg>
              </div>
              <div class="mt-3 text-center sm:ml-4 sm:mt-0 sm:text-left">
                <h3 class="text-base font-semibold text-gray-900" id="modal-title">
                  Update Progress for {selectedBook?.title}
                </h3>
                <div class="mt-8">
                  <div class="flex justify-between items-center gap-4 text-sm text-gray-600">
                    <div>
                      Current Page: {selectedBook?.currentPage || 0} / {selectedBook?.totalPages || 'Unknown'}
                    </div>
                    <div>
                      Last Updated: {selectedBook?.lastUpdated ? new Date(selectedBook.lastUpdated).toLocaleDateString('en-US', {month: 'numeric', day: 'numeric', year: 'numeric'}) : 'Never'}
                    </div>
                  </div>
                </div>
                <div class="mt-8">
                  <label for="newPageCount" class="block text-sm font-medium text-gray-700">New Page Count</label>
                  <input
                    id="newPageCount"
                    type="number"
                    bind:value={newPageCount}
                    min={selectedBook?.currentPage || 0}
                    class="mt-4 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
                    placeholder="Enter new page count"
                  />
                </div>
              </div>
            </div>
          </div>
          <div class="bg-gray-50 px-8 py-6 sm:flex sm:flex-row-reverse sm:px-8">
            <button 
              type="button" 
              class="inline-flex w-full justify-center rounded-md bg-blue-600 px-5 py-3 text-sm font-semibold text-white shadow-sm hover:bg-blue-500 sm:ml-3 sm:w-auto"
              on:click={submitUpdate}
              disabled={newPageCount === ''}
            >
              Update
            </button>
            <button 
              type="button" 
              class="mt-3 inline-flex w-full justify-center rounded-md bg-green-600 px-5 py-3 text-sm font-semibold text-white shadow-sm hover:bg-green-500 sm:ml-3 sm:mt-0 sm:w-auto"
              on:click={finishBook}
              disabled={isFinishingBook}
            >
              {isFinishingBook ? 'Finishing...' : 'Finish Book'}
            </button>
            <button 
              type="button" 
              class="mt-3 inline-flex w-full justify-center rounded-md bg-red-600 px-5 py-3 text-sm font-semibold text-white shadow-sm hover:bg-red-500 sm:ml-3 sm:mt-0 sm:w-auto"
              on:click={() => selectedBook && RemoveFromCurrentlyReading(selectedBook.bookId)}
            >
              Remove
            </button>
            <button 
              type="button" 
              class="mt-3 inline-flex w-full justify-center rounded-md bg-white px-5 py-3 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:mt-0 sm:w-auto"
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

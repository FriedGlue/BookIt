<script lang="ts">
  import { onMount } from 'svelte';
  import { BookService } from '$lib/services/bookService';
  import { AuthService } from '$lib/services/authService';
  import type { DisplayBook, ToBeReadItem, ReadItem } from '$lib/types';
	import { token } from '$lib/stores/authStore';

  let books: DisplayBook[] = [];
  let modalVisible = false;
  let selectedBook: DisplayBook | null = null;
  let newPageCount: number | '' = '';
  let toBeReadList: ToBeReadItem[] = [];
  let readList: ReadItem[] = [];
  let customLists: Record<string, any[]> = {};
  let isStartingBook = false;
  let isFinishingBook = false;
  let searchQuery = '';
  let searchResults: any[] = [];
  let isSearching = false;
  let showSearchResults = false;
  let isAddingToList = false;
  
  const bookService = new BookService();
  const authService = new AuthService();

  onMount(() => {
    // Fetch profile
    (async () => {
        try {
            const profile = await bookService.getProfile();
            books = (profile.currentlyReading || []).map(item => ({
                bookId: item.Book.bookId,
                title: item.Book.title || 'Untitled',
                author: item.Book.authors ? item.Book.authors[0] : 'Unknown Author',
                thumbnail: item.Book.thumbnail || '',
                progress: item.Book.progress ? item.Book.progress.percentage : 0,
                totalPages: item.Book.totalPages || 1
            }));

            toBeReadList = profile.lists?.toBeRead || [];
            readList = profile.lists?.read || [];
            customLists = profile.lists?.customLists || {};
        } catch (error) {
            console.error('Error fetching profile:', error);
        }
    })();

    // Add click listener
    document.addEventListener('click', handleClickOutside);
    return () => document.removeEventListener('click', handleClickOutside);
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
          const index = books.findIndex(b => b.bookId === selectedBook!.bookId);
          if(index !== -1) {
              books[index] = { ...selectedBook };
          }

          closeModal();
      } catch (error) {
          console.error('Error updating book progress:', error);
      }
  }

  async function deleteBook() {
      if (!selectedBook) return;

      try {
          await bookService.deleteBook(selectedBook.bookId);
          books = books.filter(b => b.bookId !== selectedBook!.bookId);
          closeModal();
      } catch (error) {
          console.error('Error deleting book:', error);
      }
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
            totalPages: item.Book.totalPages || 1
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
            totalPages: item.Book.totalPages || 1
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

  function handleLogout() {
    authService.logout();
  }

  async function handleSearch() {
    if (!searchQuery.trim()) {
        searchResults = [];
        showSearchResults = false;
        return;
    }

    try {
        isSearching = true;
        searchResults = await bookService.searchBooks(searchQuery);
        showSearchResults = true;
    } catch (error) {
        console.error('Error searching books:', error);
        searchResults = [];
    } finally {
        isSearching = false;
    }
  }

  function handleClickOutside(event: MouseEvent) {
    const target = event.target as HTMLElement;
    if (!target.closest('.search-container')) {
        showSearchResults = false;
    }
  }

</script>

<div class="flex flex-col min-h-screen">
  <nav class="flex flex-wrap items-center justify-between bg-blue-500 p-6">
	<div class="mr-6 flex flex-shrink-0 items-center text-white">
		<span class="text-6xl font-semibold tracking-tight">BookIt</span>
	</div>
	<div class="block w-full flex-grow lg:flex lg:w-auto lg:items-center">
		<div class="text-lg lg:flex-grow flex items-center">
			<a
				href="#responsive-header"
				class="mr-4 mt-4 block text-teal-200 hover:text-white lg:mt-0 lg:inline-block"
			>
				Home
			</a>
			<a
				href="#responsive-header"
				class="mr-4 mt-4 block text-teal-200 hover:text-white lg:mt-0 lg:inline-block"
			>
				About
			</a>
			<div class="relative flex-grow max-w-xl ml-4">
				<input
					type="text"
					bind:value={searchQuery}
					on:input={() => handleSearch()}
					placeholder="Search books..."
					class="w-full px-4 py-2 text-gray-900 rounded-full focus:outline-none focus:ring-2 focus:ring-blue-400"
				/>
				{#if showSearchResults && searchResults.length > 0}
					<div class="absolute z-50 w-full mt-2 bg-white rounded-lg shadow-xl max-h-96 overflow-y-auto">
						{#each searchResults as book (book.bookId)}
							<button 
								type="button"
								class="w-full p-4 hover:bg-gray-100 cursor-pointer flex items-center space-x-4 text-left"
								on:click={async () => {
									if (isAddingToList) return;
									try {
										isAddingToList = true;
										await bookService.addToList(book.bookId, 'toBeRead');
										const profile = await bookService.getProfile();
										toBeReadList = profile.lists?.toBeRead || [];
										showSearchResults = false;
										searchQuery = '';
									} catch (error) {
										console.error('Error adding book to list:', error);
									} finally {
										isAddingToList = false;
									}
								}}
								disabled={isAddingToList}
								on:keydown={(e) => {
									if (e.key === 'Enter' || e.key === ' ') {
										e.preventDefault();
										e.currentTarget.click();
									}
								}}
							>
								{#if isAddingToList}
									<div class="w-12 h-16 flex items-center justify-center">
										<div class="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-500"></div>
									</div>
								{:else if book.thumbnail}
									<img src={book.thumbnail} alt={book.title} class="w-12 h-16 object-cover" />
								{/if}
								<div>
									<h3 class="text-gray-900 font-medium">{book.title}</h3>
									<p class="text-gray-600 text-sm">{book.authors ? book.authors[0] : 'Unknown Author'}</p>
								</div>
							</button>
						{/each}
					</div>
				{/if}
			</div>
		</div>
		<div>
			{#if $token}
				<button
					on:click={handleLogout}
					class="mr-2 inline-block rounded-full border-2 border-blue-500 bg-white px-4 py-2 text-lg font-semibold text-blue-500"
				>
					Log Out
				</button>
				<a
					href="#profile"
					class="mr-2 inline-block rounded-full border-2 border-blue-500 bg-white px-4 py-2 text-lg font-semibold text-blue-500"
				>
					View Profile
				</a>
			{:else}
				<a
					href="/login"
					class="mr-2 inline-block rounded-full border-2 border-blue-500 bg-white px-4 py-2 text-lg font-semibold text-blue-500"
				>
					Sign In
				</a>
				<a
					href="/signup"
					class="inline-block rounded-full border-2 border-blue-500 bg-blue-500 px-4 py-2 text-lg font-semibold text-white"
				>
					Sign Up
				</a>
			{/if}
		</div>
	</div>
</nav>

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
                  <div class="relative w-full bg-gray-300 rounded-lg hover:shadow-2xl transition-shadow duration-300 transform hover:scale-105">
                      <img
                          src={book.thumbnail || 'default-cover-image-url'}
                          alt="Book Cover"
                          class="w-full h-64 sm:h-72 md:h-80 lg:h-64 rounded-lg"
                      />
                  </div>
                  <button 
                      class="w-full h-8 mt-4 bg-gray-100 rounded-full hover:bg-blue-500 hover:text-white disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200"
                      on:click={() => startReading(book.bookId, 'toBeRead')}
                      disabled={isStartingBook}
                  >
                      {isStartingBook ? 'Starting...' : 'Start Reading'}
                  </button>
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
          {#each (readList?.slice(-4) || []).reverse() as book}
              <div class="flex flex-col">
                  <div class="relative rounded-lg bg-gray-300 shadow-lg hover:shadow-2xl transition-shadow duration-300 transform hover:scale-105 w-full">
                      <img
                          src={book.thumbnail || 'default-cover-image-url'}
                          alt="Book Cover"
                          class="w-full h-64 sm:h-72 md:h-80 lg:h-64 rounded-lg"
                      />
                  </div>
              </div>
          {/each}
          {#if (readList?.length || 0) > 4}
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

<footer class="bg-gray-800 text-white mt-32">
  <div class="max-w-7xl mx-auto py-12 px-4 sm:px-6 lg:px-8">
    <div class="grid grid-cols-1 md:grid-cols-3 gap-8">
      <div>
        <h3 class="text-2xl font-bold mb-4">BookIt</h3>
        <p class="text-gray-400">Track your reading journey, one page at a time.</p>
      </div>
      <div>
        <h4 class="text-lg font-semibold mb-4">Quick Links</h4>
        <ul class="space-y-2">
          <li><a href="/" class="text-gray-400 hover:text-white">Home</a></li>
          <li><a href="/about" class="text-gray-400 hover:text-white">About</a></li>
          <li><a href="/profile" class="text-gray-400 hover:text-white">Profile</a></li>
        </ul>
      </div>
      <div>
        <h4 class="text-lg font-semibold mb-4">Contact</h4>
        <ul class="space-y-2">
          <li class="text-gray-400">Email: support@bookit.com</li>
          <li class="text-gray-400">Follow us on Twitter @BookItApp</li>
        </ul>
      </div>
    </div>
    <div class="mt-8 pt-8 border-t border-gray-700 text-center">
      <p class="text-gray-400">&copy; {new Date().getFullYear()} BookIt. All rights reserved.</p>
    </div>
  </div>
</footer>
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
                  <label for="newPageCount" class="block text-sm font-medium text-gray-700">New Page Count</label>
                  <input
                    id="newPageCount"
                    type="number"
                    bind:value={newPageCount}
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
              on:click={deleteBook}
            >
              Delete
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

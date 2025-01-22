<script lang="ts">
  import { BookService } from '$lib/services/bookService';
  import { AuthService } from '$lib/services/authService';
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';

  let searchQuery = '';
  let searchResults: any[] = [];
  let isSearching = false;
  let showSearchResults = false;
  let isAddingToList = false;
  let toBeReadList: any[] = [];

  const bookService = new BookService();
  const authService = new AuthService();

  let authenticated = false;

  onMount(() => {
    authenticated = document.cookie.includes('token=');
  });

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

  async function handleLogout() {
    authService.logout();
    goto('/login');
  }

  function handleClickOutside(event: MouseEvent) {
    const target = event.target as HTMLElement;
    if (!target.closest('.search-container')) {
        showSearchResults = false;
    }
  }
</script>

<nav class="flex flex-wrap items-center justify-between bg-blue-500 p-6">
  <div class="flex items-center space-x-8">
    <div class="flex flex-shrink-0 items-center text-white">
      <span class="text-6xl font-semibold tracking-tight">BookIt</span>
    </div>
    <div class="flex items-center space-x-4">
      <a
        href="/"
        class="block text-teal-200 hover:text-white lg:inline-block"
      >
        Home
      </a>
      <a
        href="/lists"
        class="block text-teal-200 hover:text-white lg:inline-block"
      >
        Lists
      </a>
    </div>
  </div>

  <div class="relative w-1/3 min-w-[300px] search-container">
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

  <div class="flex items-center space-x-4">
    {#if authenticated}
      <button
        on:click={handleLogout}
        class="inline-block rounded-full border-2 border-blue-500 bg-white px-4 py-2 text-lg font-semibold text-blue-500"
      >
        Log Out
      </button>
      <a
        href="/profile"
        class="inline-block rounded-full border-2 border-blue-500 bg-white px-4 py-2 text-lg font-semibold text-blue-500"
      >
        View Profile
      </a>
    {:else}
      <a
        href="/login"
        class="inline-block rounded-full border-2 border-blue-500 bg-white px-4 py-2 text-lg font-semibold text-blue-500"
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
</nav>

<svelte:window on:click={handleClickOutside} /> 
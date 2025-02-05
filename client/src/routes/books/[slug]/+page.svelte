<script lang="ts">
    import type { PageData } from './$types';
    import type { Book } from '$lib/types';

    export let data: PageData;

    $: book = data.book ? (data.book[0] as Book) : null;
    console.log('Data:', data);

    async function addToBeRead() {
        const response = await fetch(`/api/books/add`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ bookId: book?.bookId, listType: 'toBeRead' })
        });
        if (!response.ok) {
            throw new Error('Failed to add book to list');
        }
        return await response.json();
    }
</script>

<section class="mx-8 mt-16 flex flex-col items-start px-4 md:mx-16 lg:mx-40">
	{#if book}
		<div>
			<div>
				<img 
					src={book.coverImageUrl || 'default-cover-image-url'} 
					alt={`Cover of ${book.title} by ${book.authors?.join(', ')}`} 
					class="max-w-xs shadow-lg"
				/>
				<h2 class="mt-4 text-2xl font-bold">{book.title}</h2>
				<p class="text-gray-600">{book.authors?.join(', ')}</p>
				{#if book.pageCount}
					<p class="text-sm text-gray-500">{book.pageCount} pages</p>
				{/if}
			</div>
		</div>
	{:else}
		<p class="text-2xl text-gray-500">Loading...</p>
	{/if}

	{#if book}
		<button 
			on:click={addToBeRead}
			type="submit"
			class="mt-4 h-8 w-full rounded-full bg-white text-gray-800
						   transition-all duration-200 hover:bg-blue-500 hover:text-white"
		>
			Add Book to to be read list
		</button>
	{/if}
</section>
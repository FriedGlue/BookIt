<script lang="ts">
    import type { PageData } from './$types';
    import type { Book } from '$lib/types';

    export let data: PageData;

    let book = data.book ? (data.book[0] as Book) : null;
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

<div>
    <p>{book?.title}</p>
</div>

<section class="mx-8 mt-16 flex flex-col items-start px-4 md:mx-16 lg:mx-40">
	{#if data.book}
		<div>
			<div>
				<img src={book?.coverImageUrl || 'default-cover-image-url'} alt={`Cover of ${book?.title} by ${book?.authors}`} />
				<h2>{book?.title}</h2>
				<p>{book?.authors}</p>
			</div>
		</div>
	{:else}
		<p class="text-2xl text-gray-500">Loading...</p>
	{/if}
</section>

    <button 
        on:click={() => {
            addToBeRead();
        }}
        type="submit"
        class="h-8 w-full rounded-full bg-white text-gray-800
                           transition-all duration-200 hover:bg-blue-500 hover:text-white"
    >Add Book to to be read list</button>
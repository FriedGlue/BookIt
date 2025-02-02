import type { Book } from '$lib/types';
import type { PageServerLoad } from './$types';

export const load = (async ({ fetch, params }) => {
	let book: Book | null = null;

	try {
		const res = await fetch(`/api/books/searchByBookId?q=${encodeURIComponent(params.slug)}`);

		if (!res.ok) {
			console.error('Error searching books:', await res.text());
			return {
				book: null
			};
		}

		book = await res.json();

		console.log('book', book);

	} catch (error) {
		console.error('Error searching books:', error);
	}

	if (!book) {
		console.log('No book found');
		return {
			book: null
		};
	}

	return {
		book: book
	};
}) satisfies PageServerLoad;

// move add books down here
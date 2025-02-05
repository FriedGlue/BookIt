import type { Book } from '$lib/types';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch, params }) => {
	try {
		console.log('Fetching book with slug:', params.slug);
		const res = await fetch(`/api/books/searchByBookId?q=${encodeURIComponent(params.slug)}`);

		if (!res.ok) {
			console.error('Error searching books:', await res.text());
			return {
				book: null
			};
		}

		const books = await res.json();
		console.log('Raw API response:', books);

		if (!books || !Array.isArray(books) || books.length === 0) {
			console.log('No books found');
			return { book: null };
		}

		// Take the first book from the array
		const book = books[0];
		console.log('Selected book:', book);

		if (!book || !book.bookId) {
			console.log('Invalid book data');
			return { book: null };
		}

		return {
			book: [book]
		};
	} catch (error) {
		console.error('Error searching books:', error);
		return { book: null };
	}
};

// move add books down here
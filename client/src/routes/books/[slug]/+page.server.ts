import type { PageServerLoad, Actions } from './$types';
import { BookService } from '$lib/services/bookService';
import { PUBLIC_API_BASE_URL } from '$env/static/public';
import type { Profile } from '$lib/types';

export const load: PageServerLoad = async ({ fetch, params, cookies }) => {
	try {
		console.log('Fetching book with slug:', params.slug);
		const res = await fetch(`/api/books/searchByBookId?q=${encodeURIComponent(params.slug)}`);

		if (!res.ok) {
			console.error('Error searching books:', await res.text());
			return {
				book: null,
				customLists: {}
			};
		}

		const books = await res.json();
		console.log('Raw API response:', books);

		if (!books || !Array.isArray(books) || books.length === 0) {
			console.log('No books found');
			return { book: null, customLists: {} };
		}

		// Take the first book from the array
		const book = books[0];
		console.log('Selected book:', book);

		if (!book?.bookId) {
			console.log('Invalid book data');
			return { book: null, customLists: {} };
		}

		// Fetch user's custom lists
		const token = cookies.get('idToken');
		let customLists = {};

		if (token) {
			const profileResponse = await fetch(`${PUBLIC_API_BASE_URL}/profile`, {
				headers: {
					Authorization: `Bearer ${token}`,
					'Content-Type': 'application/json'
				}
			});

			if (profileResponse.ok) {
				const profile: Profile = await profileResponse.json();
				customLists = profile.lists?.customLists || {};
			}
		}

		return {
			book: book,
			customLists
		};
	} catch (error) {
		console.error('Error searching books:', error);
		return { book: null, customLists: {} };
	}
};

export const actions: Actions = {
	addToList: async ({ request, cookies }) => {
		const token = cookies.get('idToken');
		if (!token) return { error: 'No token' };

		const formData = await request.formData();
		const bookId = formData.get('bookId')?.toString();
		const listType = formData.get('listType')?.toString();

		if (!bookId || !listType) {
			return { error: 'Missing required fields' };
		}

		try {
			const bookService = new BookService(token);
			await bookService.addToList(bookId, listType);
			return { success: true };
		} catch (error) {
			console.error('Failed to add book to list:', error);
			return { error: 'Failed to add book to list' };
		}
	},

	startReading: async ({ request, cookies }) => {
		const token = cookies.get('idToken');
		if (!token) return { error: 'No token' };

		const formData = await request.formData();
		const bookId = formData.get('bookId')?.toString();

		if (!bookId) {
			return { error: 'Missing bookId' };
		}

		try {
			const bookService = new BookService(token);
			await bookService.addToCurrentlyReading(bookId);
			return { success: true };
		} catch (error) {
			console.error('Failed to start reading book:', error);
			return { error: 'Failed to start reading book' };
		}
	}
};

// move add books down here

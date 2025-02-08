import type { PageServerLoad, Actions } from './$types';
import { PUBLIC_API_BASE_URL } from '$env/static/public';
import { BookService } from '$lib/services/bookService';
import type { Profile, UserBookshelves } from '$lib/types';

export const load = (async ({ fetch, cookies }) => {
	try {
		const token = cookies.get('idToken');

		if (!token) {
			console.log('No token found');
			return {
				bookshelves: {
					toBeRead: [],
					read: [],
					customShelves: {}
				} as UserBookshelves
			};
		}

		console.log('Fetching profile...');
		const response = await fetch(`${PUBLIC_API_BASE_URL}/profile`, {
			headers: {
				Authorization: `Bearer ${token}`,
				'Content-Type': 'application/json'
			}
		});

		if (!response.ok) {
			console.error('Profile fetch failed:', response.status, response.statusText);
			throw new Error(`Failed to fetch profile: ${response.status}`);
		}

		const profile: Profile = await response.json();

		return {
			bookshelves: profile.bookshelves ?? {
				toBeRead: [],
				read: [],
				customShelves: {}
			}
		};
	} catch (error) {
		console.error('Error loading bookshelves:', error);
		// Return fallback data
		return {
			bookshelves: {
				toBeRead: [],
				read: [],
				customShelves: {}
			} as UserBookshelves
		};
	}
}) satisfies PageServerLoad;

export const actions: Actions = {
	updateProgress: async ({ request, cookies }) => {
		const token = cookies.get('idToken');
		if (!token) return { error: 'No token' };

		const formData = await request.formData();
		const bookId = formData.get('bookId')?.toString();
		const newPageCount = formData.get('newPageCount')?.toString();

		if (!bookId || !newPageCount) {
			return { error: 'Missing form data' };
		}

		try {
			const bookService = new BookService(token);
			await bookService.updateBookProgress(bookId, Number(newPageCount));
			return { success: true };
		} catch (err) {
			console.error('Failed to update progress:', err);
			return { error: 'Failed to update progress' };
		}
	},

	removeFromShelf: async ({ request, cookies }) => {
		const token = cookies.get('idToken');
		if (!token) return { error: 'No token' };

		const formData = await request.formData();
		const bookId = formData.get('bookId')?.toString();
		const shelfType = formData.get('shelfType')?.toString();

		if (!bookId || !shelfType) {
			return { error: 'Missing form data' };
		}

		try {
			const bookService = new BookService(token);
			await bookService.removeFromShelf(bookId, shelfType);
			return { success: true };
		} catch (err) {
			console.error('Failed to remove from shelf:', err);
			return { error: 'Failed to remove from shelf' };
		}
	},

	createBookshelf: async ({ request, cookies }) => {
		const token = cookies.get('idToken');
		if (!token) return { error: 'No token' };

		const formData = await request.formData();
		const shelfName = formData.get('shelfName')?.toString();

		if (!shelfName) {
			return { error: 'Missing form data' };
		}

		try {
			const bookService = new BookService(token);
			await bookService.createBookshelf(shelfName);
			return { success: true };
		} catch (err) {
			console.error('Failed to create bookshelf:', err);
			return { error: 'Failed to create bookshelf' };
		}
	},

	deleteBookshelf: async ({ request, cookies }) => {
		const token = cookies.get('idToken');
		if (!token) return { error: 'No token' };

		const formData = await request.formData();
		const shelfName = formData.get('shelfName')?.toString();

		if (!shelfName) {
			return { error: 'Missing form data' };
		}

		try {
			const bookService = new BookService(token);
			await bookService.deleteBookshelf(shelfName);
			return { success: true };
		} catch (err) {
			console.error('Failed to delete bookshelf:', err);
			return { error: 'Failed to delete bookshelf' };
		}
	}
};

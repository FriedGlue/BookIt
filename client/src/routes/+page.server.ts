// +page.server.ts
import type { Actions, PageServerLoad } from './$types';
import { BookService } from '$lib/services/bookService'; // <--- use your correct import path
import type { Profile, CurrentlyReadingItem } from '$lib/types';

export const load = (async ({ cookies }) => {
	try {
		const token = cookies.get('idToken');
		if (!token) {
			console.log('No token found');
			return {
				books: [],
				toBeReadList: [],
				readList: [],
				customLists: {}
			};
		}

		// Use the BookService to fetch the profile
		const bookService = new BookService(token);
		const profile: Profile = await bookService.getProfile();

		// Transform data as needed
		return {
			books: (profile.currentlyReading || []).map((item: CurrentlyReadingItem) => ({
				bookId: item.Book.bookId,
				title: item.Book.title ?? 'Untitled',
				author: item.Book.authors ? item.Book.authors[0] : 'Unknown Author',
				thumbnail: item.Book.thumbnail ?? '',
				progress: item.Book.progress?.percentage ?? 0,
				totalPages: item.Book.totalPages ?? 1,
				currentPage: item.Book.progress?.lastPageRead ?? 0,
				lastUpdated: item.Book.progress?.lastUpdated ?? new Date().toISOString()
			})),
			toBeReadList: profile.lists?.toBeRead || [],
			readList: profile.lists?.read || [],
			customLists: profile.lists?.customLists || {}
		};
	} catch (error) {
		console.error('Error loading profile:', error);
		return {
			books: [],
			toBeReadList: [],
			readList: [],
			customLists: {}
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

	removeFromList: async ({ request, cookies }) => {
		const token = cookies.get('idToken');
		if (!token) return { error: 'No token' };

		const formData = await request.formData();
		const bookId = formData.get('bookId')?.toString();
		const listType = formData.get('listType')?.toString();

		if (!bookId || !listType) {
			return { error: 'Missing form data' };
		}

		try {
			const bookService = new BookService(token);
			await bookService.removeFromList(bookId, listType);
			return { success: true };
		} catch (err) {
			console.error('Failed to remove from list:', err);
			return { error: 'Failed to remove from list' };
		}
	},

	removeFromCurrentlyReading: async ({ request, cookies }) => {
		const token = cookies.get('idToken');
		if (!token) return { error: 'No token' };

		const formData = await request.formData();
		const bookId = formData.get('bookId')?.toString();
		if (!bookId) return { error: 'Missing bookId' };

		try {
			const bookService = new BookService(token);
			await bookService.removeFromCurrentlyReading(bookId);
			return { success: true };
		} catch (err) {
			console.error('Failed to remove from currently reading:', err);
			return { error: 'Failed to remove from currently reading' };
		}
	},

	startReading: async ({ request, cookies }) => {
		const token = cookies.get('idToken');
		if (!token) return { error: 'No token' };

		const formData = await request.formData();
		const bookId = formData.get('bookId')?.toString();
		const listName = formData.get('listName')?.toString();

		if (!bookId || !listName) {
			return { error: 'Missing form data' };
		}

		try {
			const bookService = new BookService(token);
			await bookService.startReading(bookId, listName);
			return { success: true };
		} catch (err) {
			console.error('Failed to start reading book:', err);
			return { error: 'Failed to start reading book' };
		}
	},

	finishBook: async ({ request, cookies }) => {
		const token = cookies.get('idToken');
		if (!token) return { error: 'No token' };

		const formData = await request.formData();
		const bookId = formData.get('bookId')?.toString();
		if (!bookId) return { error: 'Missing bookId' };

		try {
			const bookService = new BookService(token);
			await bookService.finishReading(bookId);
			return { success: true };
		} catch (err) {
			console.error('Failed to finish book:', err);
			return { error: 'Failed to finish book' };
		}
	}
};

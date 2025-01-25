// +page.server.ts
import type { PageServerLoad, Actions } from './$types';
import { BookService } from '$lib/services/bookService'; // <--- use your correct import path
import type { Profile, CurrentlyReadingItem } from '$lib/types';

export const load: PageServerLoad = async ({ fetch }) => {
	let profile: Profile | null = null;

	try {
		const res = await fetch(`/api/profile`);

		if (!res.ok) {
			console.error('Error searching books:', await res.text());
			return {
				books: [],
				toBeReadList: [],
				readList: [],
				customLists: {}
			};
		}

		profile = await res.json();

	} catch (error) {
		console.error('Error searching books:', error);
	}

	if (!profile) {
		console.log('No profile found');
		return {
			books: [],
			toBeReadList: [],
			readList: [],
			customLists: {}
		};
	}

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
};

export const actions: Actions = {

	startReading: async ({ request }) => {
		const formData = await request.formData();
		const bookId = formData.get('bookId')?.toString();
		const listName = formData.get('listName')?.toString();

		if (!bookId || !listName) {
			return { error: 'Missing form data' };
		}

		const response = await fetch('/api/books/add', {
			method: 'POST',
			body: JSON.stringify({ bookId, listName })
		});
		if (!response.ok) {
			throw new Error('Failed to fetch profile');
		}
		return await response.json();

	},

	viewDetails: async ({ request }) => {
		const formData = await request.formData();
		const bookId = formData.get('bookId')?.toString();
		if (!bookId) return { error: 'Missing bookId' };
		return { redirect: `/books/${bookId}` };
	},

	getProfile: async ({ cookies, fetch }) => {
		const token = cookies.get('idToken');
		const response = await fetch('/api/profile', {
			headers: {
				Authorization: `Bearer ${token}`
			}
		});
		if (!response.ok) {
			throw new Error('Failed to fetch profile');
		}
		return await response.json();
	},

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

	finishReading: async () => {
		const response = await fetch('/api/currentlyReading/finishReading', {
		});
		if (!response.ok) {
			throw new Error('Failed to finish reading');
		}
		return await response.json();
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

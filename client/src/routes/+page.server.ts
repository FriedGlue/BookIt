// +page.server.ts
import type { PageServerLoad, Actions } from './$types';
import { BookService } from '$lib/services/bookService';
import type { Profile } from '$lib/types';

export const load: PageServerLoad = async ({ fetch }) => {
	let profile: Profile | null = null;

	try {
		const response = await fetch('/api/profile');

		if (!response.ok) {
			return { 
				profile: {
					currentlyReading: [],
					challenges: [],
					lists: {
						toBeRead: [],
						read: [],
						customLists: {}
					}
				} 
			};
		}

		profile = await response.json();

		// Ensure lists are initialized
		if (profile) {
			if (!profile.lists) {
				profile.lists = {
					toBeRead: [],
					read: [],
					customLists: {}
				};
			}

			if (!profile.lists.toBeRead) {
				profile.lists.toBeRead = [];
			}

			if (!profile.lists.read) {
				profile.lists.read = [];
			}

			if (!profile.lists.customLists) {
				profile.lists.customLists = {};
			}

			if (!profile.currentlyReading) {
				profile.currentlyReading = [];
			}

			if (!profile.challenges) {
				profile.challenges = [];
			}

			// Load reading challenges
			try {
				const challenges = profile.challenges;
				if (challenges && challenges.length > 0) {
					profile.challenges = challenges;
				}
			} catch (error) {
				console.error('Error loading reading challenges:', error);
			}
		}

		return { profile };
	} catch (error) {
		console.error('Error loading profile:', error);
		return { 
			profile: {
				currentlyReading: [],
				challenges: [],
				lists: {
					toBeRead: [],
					read: [],
					customLists: {}
				}
			} 
		};
	}
};

export const actions: Actions = {
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
			console.error('Failed to start reading:', err);
			return { error: 'Failed to start reading' };
		}
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
			throw Error('Failed to remove from list', { cause: err });
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
			throw Error('Failed to remove from currently reading', { cause: err });
		}
	},

	finishReading: async () => {
		const response = await fetch('/api/currentlyReading/finishReading', {});
		if (!response.ok) {
			throw new Error('Failed to finish reading', { cause: response.statusText });
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

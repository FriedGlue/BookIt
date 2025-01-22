import type { PageServerLoad, Actions } from './$types';
import { PUBLIC_API_BASE_URL } from '$env/static/public';
import { BookService } from '$lib/services/bookService';
import type { Profile } from '$lib/types';

export const load = (async ({ fetch, cookies }) => {
	try {
        const token = cookies.get('idToken');

		if (!token) {
			console.log('No token found');
			return {
				toBeReadList: [],
				readList: [],
				customLists: {}
			};
		}

		console.log('Fetching profile...');
		const response = await fetch(`${PUBLIC_API_BASE_URL}/profile`, {
			headers: {
				'Authorization': `Bearer ${token}`,
				'Content-Type': 'application/json'
			}
		});

		if (!response.ok) {
			console.error('Profile fetch failed:', response.status, response.statusText);
			throw new Error(`Failed to fetch profile: ${response.status}`);
		}

		const profile: Profile = await response.json();

		return {
			toBeReadList: profile.lists?.toBeRead,
			readList: profile.lists?.read,
			customLists: profile.lists?.customLists
		};
	} catch (error) {
		console.error('Error loading lists:', error);
		// Return fallback data or handle as you prefer
		return {
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
};

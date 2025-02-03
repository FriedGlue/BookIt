// +page.server.ts
import { PUBLIC_API_BASE_URL } from '$env/static/public';
import type { PageServerLoad, Actions } from './$types';
import { BookService } from '$lib/services/bookService'; // <--- use your correct import path
import type { Profile } from '$lib/types';

export const load: PageServerLoad = async ({ fetch, cookies }) => {
  let profile: Profile | null = null;

  const token = cookies.get('idToken');
  if (!token) {
    return { profile: null };
  }

  console.log('Fetching reading log...');
  const response = await fetch(`${PUBLIC_API_BASE_URL}/profile`, {
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });

  if (!response.ok) {
    console.error('Failed to load profile for reading log', response.statusText);
    return { profile: null };
  }

  profile = await response.json();

  return { profile };
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

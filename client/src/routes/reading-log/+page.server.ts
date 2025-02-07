// client/src/routes/reading-log/+page.server.ts
import type { Actions, PageServerLoad } from './$types';
import { PUBLIC_API_BASE_URL } from '$env/static/public';
import { ReadingLogService } from '$lib/services/readingLogServer';

export const load: PageServerLoad = async ({ fetch, cookies }) => {
	const token = cookies.get('idToken');
	if (!token) {
		// Optionally, you can redirect the user to login if not authenticated.
		return { readingLog: [] };
	}

	console.log('Fetching reading log...');
	const response = await fetch(`${PUBLIC_API_BASE_URL}/reading-log`, {
		headers: {
			Authorization: `Bearer ${token}`,
			'Content-Type': 'application/json'
		}
	});

	if (!response.ok) {
		console.error('Failed to load profile for reading log', response.statusText);
		return { readingLog: [] };
	}

	const readingLog = await response.json();

	return { readingLog: readingLog || [] };
};

export const actions: Actions = {
	updateReadingLogEntry: async ({ request, cookies }) => {
		const token = cookies.get('idToken');
		if (!token) return { error: 'No token' };

		const formData = await request.formData();
		const readingLogEntryId = formData.get('bookId')?.toString();
		const pageCount = formData.get('newPageCount')?.toString();
		const notes = formData.get('notes')?.toString();

		if (!readingLogEntryId) {
			return { error: 'Missing reading log id' };
		}

		if (!pageCount || !notes) {
			return { error: 'Missing form data' };
		}

		try {
			const readingLogService = new ReadingLogService(token);
			await readingLogService.updateBookProgress(readingLogEntryId, Number(pageCount), notes);
			return { success: true };
		} catch (err) {
			console.error('Failed to update progress:', err);
			return { error: 'Failed to update progress' };
		}
	},

	removeFromReadingLog: async ({ request, cookies }) => {
		const token = cookies.get('idToken');
		if (!token) return { error: 'No token' };

		const formData = await request.formData();
		const readingLogEntryId = formData.get('readingLogEntryId')?.toString();

		if (!readingLogEntryId) {
			return { error: 'Missing reading log id' };
		}

		try {
			const readingLogService = new ReadingLogService(token);
			await readingLogService.removeFromList(readingLogEntryId);
			return { success: true };
		} catch (err) {
			console.error('Failed to remove entry:', err);
			return { error: 'Failed to remove entry' };
		}
	}
};

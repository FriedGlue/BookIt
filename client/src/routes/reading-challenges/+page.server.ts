import type { PageServerLoad, Actions } from './$types';
import { ReadingChallengeService } from '$lib/services/readingChallengeService';
import type { Profile } from '$lib/types';
import { fail } from '@sveltejs/kit';

export const load: PageServerLoad = async ({ fetch, cookies }) => {
	let profile: Profile | null = null;

	const readingChallengeService = new ReadingChallengeService(cookies.get('idToken') || '');

	try {
		const response = await fetch('/api/profile', {
			headers: {
				cookie: cookies.toString()
			}
		});

		if (!response.ok) {
			return { profile: null };
		}

		profile = await response.json();

		// Load reading challenges
		if (profile) {
			try {
				const challenges = await readingChallengeService.getUserChallenges();
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
		return { profile: null };
	}
};

export const actions: Actions = {
	create: async ({ request, cookies }) => {
		const token = cookies.get('idToken');
		if (!token) return { error: 'No token' };

		const formData = await request.formData();
		const name = formData.get('name') as string;
		const type = formData.get('type') as 'BOOKS' | 'PAGES';
		const timeframe = formData.get('timeframe') as 'YEAR' | 'MONTH' | 'WEEK';
		const startDateRaw = formData.get('startDate') as string;
		const endDateRaw = formData.get('endDate') as string;
		const target = parseInt(formData.get('target') as string);

		if (!name || !type || !timeframe || !startDateRaw || !endDateRaw || !target) {
			return fail(400, { error: 'Missing required fields' });
		}

		// Convert dates to RFC3339 format
		const startDate = new Date(startDateRaw).toISOString();
		const endDate = new Date(endDateRaw).toISOString();

		try {
			const readingChallengeService = new ReadingChallengeService(token);
			await readingChallengeService.createChallenge({
				name,
				type,
				timeframe,
				startDate,
				endDate,
				target
			});
			return { success: true };
		} catch (error) {
			console.error('Error creating challenge:', error);
			return fail(500, { error: 'Failed to create challenge' });
		}
	},

	delete: async ({ request, cookies }) => {
		const token = cookies.get('idToken');
		if (!token) return { error: 'No token' };

		const readingChallengeService = new ReadingChallengeService(token);

		const formData = await request.formData();
		const id = formData.get('id') as string;

		if (!id) {
			return fail(400, { error: 'Challenge ID is required' });
		}

		try {
			await readingChallengeService.deleteChallenge(id);
			return { success: true };
		} catch (error) {
			console.error('Error deleting challenge:', error);
			return fail(500, { error: 'Failed to delete challenge' });
		}
	}
};

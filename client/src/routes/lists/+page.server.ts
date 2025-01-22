import type { PageServerLoad } from './$types';
import { PUBLIC_API_BASE_URL } from '$env/static/public';
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
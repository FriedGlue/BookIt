import type { PageServerLoad } from './$types';
import { PUBLIC_API_BASE_URL } from '$env/static/public';
import type { CurrentlyReadingItem, Profile } from '$lib/types';

export const load = (async ({ fetch, cookies }) => {
    try {
        const token = cookies.get('idToken');
		const response = await fetch(`${PUBLIC_API_BASE_URL}/profile`, {
			headers: {
				'Authorization': `Bearer ${token}`,
				'Content-Type': 'application/json'
			}
		});

        if (!response.ok) {
            console.error('Profile fetch failed:', response.status, response.statusText);
            throw new Error('Failed to fetch profile');
        }

        const profile: Profile = await response.json();

        return {
            books: (profile.currentlyReading || []).map((item: CurrentlyReadingItem) => ({
                bookId: item.Book.bookId,
                title: item.Book.title ?? 'Untitled',
                author: item.Book.authors ? item.Book.authors[0] : 'Unknown Author',
                thumbnail: item.Book.thumbnail ?? '',
                progress: item.Book.progress ? item.Book.progress.percentage : 0,
                totalPages: item.Book.totalPages ?? 1
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
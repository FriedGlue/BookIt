import { PUBLIC_API_BASE_URL } from '$env/static/public';
import type { Cookies } from '@sveltejs/kit';

export async function GET({ cookies }: { cookies: Cookies }) {
	// Check if user is authenticated
	const token = cookies.get('idToken');

	if (!token) {
		return new Response('Unauthorized', { status: 401 });
	}

	const response = await fetch(`${PUBLIC_API_BASE_URL}/profile`, {
		method: 'GET',
		credentials: 'include',
		headers: {
			'Content-Type': 'application/json',
			Authorization: `Bearer ${token}`
		}
	});

	if (!response.ok) {
		return { status: 500, body: { error: 'Failed to fetch profile' } };
	}

	const data = await response.json();
	return new Response(JSON.stringify(data), { status: 200 });
}

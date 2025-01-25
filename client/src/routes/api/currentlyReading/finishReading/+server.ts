import { PUBLIC_API_BASE_URL } from '$env/static/public';
import type { Cookies } from '@sveltejs/kit';

export async function POST({ cookies, request }: { cookies: Cookies, request: Request }) {
	// Check if user is authenticated
	const token = cookies.get('idToken');

	if (!token) {
		return new Response('Unauthorized', { status: 401 });
	}

	const formData = await request.formData();
	const bookId = formData.get('bookId')?.toString();

	const response = await fetch(`${PUBLIC_API_BASE_URL}/currentlyReading/finishReading`, {
		method: 'POST',
		credentials: 'include',
		headers: {
			'Content-Type': 'application/json',
			Authorization: `Bearer ${token}`
		},
		body: JSON.stringify({ bookId })
	});

	if (!response.ok) {
		return { status: 500, body: { error: 'Failed to finish reading' } };
	}

	const data = await response.json();
	return new Response(JSON.stringify(data), { status: 200 });
}

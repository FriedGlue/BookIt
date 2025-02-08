// file: src/routes/api/books/add/+server.ts
import type { RequestHandler } from './$types';
import { PUBLIC_API_BASE_URL } from '$env/static/public';

export const POST: RequestHandler = async ({ request, cookies, fetch }) => {
	const token = cookies.get('idToken');

	if (!token) {
		return new Response('Unauthorized', { status: 401 });
	}

	try {
		const { bookId, shelfName } = await request.json();

		if (!bookId || !shelfName) {
			return new Response('Missing bookId or shelfName', { status: 400 });
		}

		const response = await fetch(`${PUBLIC_API_BASE_URL}/bookshelf`, {
			method: 'POST',
			headers: {
				Authorization: `Bearer ${token}`,
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				bookId,
				shelfName
			})
		});

		if (!response.ok) {
			console.error('Error adding book to list:', response.status, response.statusText);
			return new Response('Error adding book to list', { status: response.status });
		}

		return new Response(JSON.stringify({ success: true }), { status: 200 });
	} catch (error) {
		console.error('Add to list route error:', error);
		return new Response('Internal Server Error', { status: 500 });
	}
};

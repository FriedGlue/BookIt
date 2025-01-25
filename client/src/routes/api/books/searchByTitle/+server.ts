// file: src/routes/api/books/search/+server.ts
import type { RequestHandler } from './$types';
import { PUBLIC_API_BASE_URL } from '$env/static/public';

export const GET: RequestHandler = async ({ url, cookies, fetch }) => {
	// Check if user is authenticated
	const token = cookies.get('idToken');

	if (!token) {
		return new Response('Unauthorized', { status: 401 });
	}

	// Grab 'q' (query) from URL query params
	const query = url.searchParams.get('q');
	if (!query) {
		return new Response(JSON.stringify([]), { status: 200 });
	}

	console.log('query', query);

	// Forward request to your external API
	try {
		const response = await fetch(
			`${PUBLIC_API_BASE_URL}/books/search?q=${encodeURIComponent(query)}`,
			{
				method: 'GET',
				headers: {
					Authorization: `Bearer ${token}`,
					'Content-Type': 'application/json'
				}
			}
		);

		if (!response.ok) {
			return new Response('Error searching books', { status: response.status });
		}

		const result = await response.json();
		return new Response(JSON.stringify(result), { status: 200 });
	} catch (err) {
		console.error('Search route error:', err);
		return new Response('Internal Server Error', { status: 500 });
	}
};

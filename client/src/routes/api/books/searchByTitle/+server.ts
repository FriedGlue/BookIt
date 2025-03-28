// file: src/routes/api/books/search/+server.ts
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ url, cookies, fetch }) => {
	// Check if user is authenticated (optional for Open Library, but you may still want it)
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

	try {
		// Fetch results from the Open Library API.
		// Using title search here; you can adjust query parameters as needed.
		const response = await fetch(
			`https://openlibrary.org/search.json?title=${encodeURIComponent(query)}`,
			{
				method: 'GET',
				// The Open Library API does not require these headers, but you may leave them if needed
				headers: {
					'Content-Type': 'application/json',
					'User-Agent': 'BookIt-Dev/1.0 (josh.hayes121@icloud.com)'

				}
			}
		);

		if (!response.ok) {
			return new Response('Error searching books', { status: response.status });
		}

		const result = await response.json();

		// Transform the results so that your UI receives an array of objects
		// with the keys: bookId, title, authors, thumbnail.
		const docs = result.docs || [];

		const transformed = docs.map((doc: any) => {
			return {
				// Strip the leading '/works/' from the key to get a clean ID
				bookId: doc.key.replace('/works/', ''),
				title: doc.title,
				// Use the first author if available; fallback to 'Unknown Author'
				authors: doc.author_name ? doc.author_name : ['Unknown Author'],
				// If the cover_i field exists, build the URL using Open Library covers service.
				thumbnail: doc.cover_i
					? `https://covers.openlibrary.org/b/id/${doc.cover_i}-M.jpg`
					: null
			};
		});

		return new Response(JSON.stringify(transformed), { status: 200 });
	} catch (err) {
		console.error('Search route error:', err);
		return new Response('Internal Server Error', { status: 500 });
	}
};

import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ cookies }) => {
	const token = cookies.get('idToken');
	
	if (!token) {
		return new Response('No token found', { status: 401 });
	}

	return new Response(token);
};

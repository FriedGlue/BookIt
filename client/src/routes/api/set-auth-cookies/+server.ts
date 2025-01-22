import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

export const POST: RequestHandler = async ({ request, cookies }) => {
	const { idToken, refreshToken } = await request.json();

	// Set the cookies
	cookies.set('idToken', idToken, {
		path: '/',
		maxAge: 3600,
		sameSite: 'none',
		httpOnly: true,
		secure: true
	});

	cookies.set('refreshToken', refreshToken, {
		path: '/',
		maxAge: 86400, // 24 hours
		sameSite: 'none',
		httpOnly: true,
		secure: true
	});

	return json({ success: true });
};

export const DELETE: RequestHandler = async ({ cookies }) => {
	cookies.delete('idToken', { path: '/', sameSite: 'none', secure: true });
	cookies.delete('refreshToken', { path: '/', sameSite: 'none', secure: true });
	return json({ success: true });
};

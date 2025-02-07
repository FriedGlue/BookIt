import type { Handle } from '@sveltejs/kit';
import { PUBLIC_API_BASE_URL } from '$env/static/public';
import { jwtDecode } from 'jwt-decode';
import type { DecodedToken } from '$lib/types';

export const handle: Handle = async ({ event, resolve }) => {
	const token = event.cookies.get('token');
	const refreshToken = event.cookies.get('refreshToken');
	if (token && refreshToken) {
		try {
			const decoded = jwtDecode<DecodedToken>(token);
			const expirationTime = decoded.exp * 1000;
			const now = Date.now();

			if (expirationTime - now < 5 * 60 * 1000) {
				console.log('Token is about to expire, refreshing...');
				const res = await fetch(`${PUBLIC_API_BASE_URL}/auth/refresh`, {
					method: 'POST',
					headers: {
						Authorization: `Bearer ${refreshToken}`
					}
				});
				if (res.ok) {
					const { IdToken } = await res.json();
					event.cookies.set('token', IdToken, {
						path: '/',
						maxAge: 3600
					});
					event.locals.token = IdToken;
				}
			} else {
				event.locals.token = token;
			}
		} catch (err) {
			console.log('Failed to decode token:', err);
			event.cookies.delete('token', { path: '/' });
			event.cookies.delete('refreshToken', { path: '/' });
			event.locals.token = undefined;
		}
	}
	return await resolve(event);
};

import { jwtDecode }  from 'jwt-decode';
import type { Cookies } from '@sveltejs/kit';
import type { DecodedToken } from '$lib/types';

export async function GET({ cookies }: { cookies: Cookies }) {
	// Check if user is authenticated by looking for the token
	const token = cookies.get('idToken');

	if (!token) {
		return new Response(JSON.stringify({ authenticated: false, error: 'Unauthorized' }), {
			status: 401,
			headers: { 'Content-Type': 'application/json' }
		});
	}

	try {
		// Decode the JWT token
		const decoded = jwtDecode<DecodedToken>(token);
		const expirationTime = decoded.exp * 1000; // Convert seconds to milliseconds
		const now = Date.now();

		// if current time is greater than or equal to expiration time, the token has expired
		if (now >= expirationTime) {
			console.log('Token is expired');
			return new Response(JSON.stringify({ authenticated: false, error: 'Token expired' }), {
				status: 401,
				headers: { 'Content-Type': 'application/json' }
			});
		}
	} catch (err) {
		console.log('Error decoding token:', err);
		return new Response(JSON.stringify({ authenticated: false, error: 'Error decoding token' }), {
			status: 401,
			headers: { 'Content-Type': 'application/json' }
		});
	}

	// If we reach here, the token is valid and not expired.
	return new Response(JSON.stringify({ authenticated: true }), {
		status: 200,
		headers: { 'Content-Type': 'application/json' }
	});
}

import type { Handle } from '@sveltejs/kit';
import { jwtDecode } from 'jwt-decode';
import { PUBLIC_API_BASE_URL } from '$env/static/public';

interface DecodedToken {
    exp: number;
    sub: string;
    email: string;
}

interface Locals {
    token?: string;
}

export const handle: Handle = async ({ event, resolve }) => {
    // Get tokens from cookies
    const token = event.cookies.get('token');
    const refreshToken = event.cookies.get('refreshToken');
    
    if (token && refreshToken) {
        try {
            // Decode token to check expiration
            const decoded = jwtDecode(token) as DecodedToken;
            const expirationTime = decoded.exp * 1000; // Convert to milliseconds
            const currentTime = Date.now();
            const timeUntilExpiry = expirationTime - currentTime;
            
            // If token will expire in less than 5 minutes, refresh it
            if (timeUntilExpiry < 300000) { // 5 minutes in milliseconds
                const response = await fetch(`${PUBLIC_API_BASE_URL}/auth/refresh`, {
                    method: 'POST',
                    headers: {
                        'Authorization': `Bearer ${refreshToken}`
                    }
                });

                if (response.ok) {
                    const { IdToken } = await response.json();
                    // Set the new token in the cookie
                    event.cookies.set('token', IdToken, {
                        path: '/',
                        maxAge: 3600,
                        sameSite: 'strict'
                    });
                    // Update the token in locals
                    event.locals.token = IdToken;
                }
            } else {
                // Token is still valid
                event.locals.token = token;
            }
        } catch {
            // If token is invalid or decoding fails, clear both tokens
            event.cookies.delete('token', { path: '/' });
            event.cookies.delete('refreshToken', { path: '/' });
            event.locals.token = undefined;
        }
    }

    const response = await resolve(event);
    return response;
};

export function isLoggedIn(locals: Locals): boolean {
    return !!locals.token;
} 
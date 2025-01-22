import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

export const POST: RequestHandler = async ({ request, cookies }) => {
    const { token, refreshToken } = await request.json();

    // Set the cookies
    cookies.set('token', token, {
        path: '/',
        maxAge: 3600,
        sameSite: 'strict',
        httpOnly: true
    });

    cookies.set('refreshToken', refreshToken, {
        path: '/',
        maxAge: 86400, // 24 hours
        sameSite: 'strict',
        httpOnly: true
    });

    return json({ success: true });
}; 
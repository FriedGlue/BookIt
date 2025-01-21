import type { Handle } from '@sveltejs/kit';

export const handle: Handle = async ({ event, resolve }) => {
    // Get token from cookie
    const token = event.cookies.get('token');
    
    // Add token to event.locals
    event.locals.token = token;

    const response = await resolve(event);
    return response;
}; 
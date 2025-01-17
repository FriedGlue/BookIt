import { writable } from 'svelte/store';
import { browser } from '$app/environment';

// Create a writable store for the token
export const token = writable<string>(
    browser ? localStorage.getItem('token') || '' : ''
);

// Subscribe to changes and update localStorage
if (browser) {
    token.subscribe((value) => {
        if (value) {
            localStorage.setItem('token', value);
        } else {
            localStorage.removeItem('token');
        }
    });
}

export function setToken(newToken: string) {
    token.set(newToken);
}

export function clearToken() {
    token.set('');
} 
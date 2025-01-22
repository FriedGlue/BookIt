import { writable } from 'svelte/store';
import { browser } from '$app/environment';

// Create writable stores for both tokens
export const idToken = writable<string>(
    browser ? localStorage.getItem('idToken') ?? '' : ''
);

export const refreshToken = writable<string>(
    browser ? localStorage.getItem('refreshToken') ?? '' : ''
);

// Subscribe to changes and update localStorage
if (browser) {
    idToken.subscribe((value) => {
        if (value) {
            localStorage.setItem('idToken', value);
        } else {
            localStorage.removeItem('idToken');
        }
    });

    refreshToken.subscribe((value) => {
        if (value) {
            localStorage.setItem('refreshToken', value);
        } else {
            localStorage.removeItem('refreshToken');
        }
    });
}

export function setTokens(newIdToken: string, newRefreshToken?: string) {
    idToken.set(newIdToken);
    if (newRefreshToken) {
        refreshToken.set(newRefreshToken);
    }
}

export function clearTokens() {
    idToken.set('');
    refreshToken.set('');
} 
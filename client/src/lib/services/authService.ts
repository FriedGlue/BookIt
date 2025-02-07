import { PUBLIC_API_BASE_URL } from '$env/static/public';
import { isAuthenticated } from '$lib/stores/authStore';
import { goto } from '$app/navigation';

export class AuthService {
	async login(username: string, password: string): Promise<void> {
		const response = await fetch(`${PUBLIC_API_BASE_URL}/auth/signin`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ username, password }),
			credentials: 'omit'
		});

		if (!response.ok) {
			throw new Error('Login failed');
		}

		const data = await response.json();

		console.log(data);

		// Set cookies via a server endpoint
		const setCookiesResponse = await fetch('/api/set-auth-cookies', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				idToken: data.IdToken,
				refreshToken: data.RefreshToken
			}),
		});

		if (!setCookiesResponse.ok) {
			throw new Error('Failed to set authentication cookies');
		}

		await goto('/');
		isAuthenticated.set(true);
	}

	async refreshToken(): Promise<void> {
		const response = await fetch(`${PUBLIC_API_BASE_URL}/auth/refresh`, {
			method: 'POST',
			headers: {
				Authorization: `Bearer ${this.refreshToken}`
			}
		});
		if (!response.ok) {
			throw new Error('Refresh token failed');
		}
		const data = await response.json();

		return data.IdToken;
	}

	logout(): void {
		// Clear cookies via server endpoint
		fetch('/api/set-auth-cookies', {
			method: 'DELETE'
		}).catch(console.error);

		// Only call signout endpoint if not in local development
		if (!PUBLIC_API_BASE_URL.includes('amazonaws')) {
			fetch(`${PUBLIC_API_BASE_URL}/auth/signout`, {
				method: 'POST',
			}).catch(console.error);
		}

		isAuthenticated.set(false);
	}

	async isAuthenticated(): Promise<boolean> {
		const response = await fetch(`/api/isAuthenticated`, {
			method: 'GET',
		});

		if (!response.ok) {
			return false;
		}

		const data = await response.json();

		return data.authenticated;
	}

	async signup(username: string, email: string, password: string): Promise<void> {
		const response = await fetch(`${PUBLIC_API_BASE_URL}/auth/signup`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ username, email, password }),
			credentials: 'omit'
		});

		if (!response.ok) {
			throw new Error('Signup failed');
		}

		await goto('/signup/confirm');
	}

	async confirm(username: string, code: string): Promise<void> {
		const response = await fetch(`${PUBLIC_API_BASE_URL}/auth/confirm`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ username, code }),
			credentials: 'omit'
		});

		if (!response.ok) {
			throw new Error('Confirmation failed');
		}

		await goto('/login');
	}

	async getToken(): Promise<string> {
		const response = await fetch('/api/getToken', {
			method: 'GET',
		});
		return response.text();
	}
}

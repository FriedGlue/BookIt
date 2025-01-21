import { setToken, clearToken } from '$lib/stores/authStore';
import { PUBLIC_API_BASE_URL } from '$env/static/public';

export class AuthService {
    async login(username: string, password: string): Promise<void> {
        const response = await fetch(`${PUBLIC_API_BASE_URL}/auth/signin`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ username, password })
        });

        if (!response.ok) {
            throw new Error('Login failed');
        }

        const data = await response.json();
        setToken(data.IdToken);

        // Set cookie for server-side rendering
        document.cookie = `token=${data.IdToken}; path=/; max-age=3600; SameSite=Strict`;
    }

    logout(): void {
        clearToken();
        // Clear cookie
        document.cookie = 'token=; path=/; expires=Thu, 01 Jan 1970 00:00:01 GMT;';
    }
}
import { PUBLIC_API_BASE_URL } from '$env/static/public';

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
        
        // Set cookies via a server endpoint
        await fetch('/api/set-auth-cookies', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                token: data.IdToken,
                refreshToken: data.RefreshToken
            })
        });
    }

    async refreshToken(): Promise<void> {
        const response = await fetch(`${PUBLIC_API_BASE_URL}/auth/refresh`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${this.refreshToken}`
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
        if (!PUBLIC_API_BASE_URL.includes('localhost')) {
            fetch(`${PUBLIC_API_BASE_URL}/auth/signout`, {
                method: 'POST',
                credentials: 'include'
            }).catch(console.error);
        }
    }
}
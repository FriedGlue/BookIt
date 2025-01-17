import type { Profile } from '$lib/types';
import { token } from '$lib/stores/authStore';
import { get } from 'svelte/store';
import { PUBLIC_API_BASE_URL } from '$env/static/public';

export class BookService {
    private getHeaders() {
        const currentToken = get(token);
        return {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${currentToken}`
        };
    }

    async getProfile(): Promise<Profile> {
        const response = await fetch(`${PUBLIC_API_BASE_URL}/profile`, {
            method: "GET",
            headers: this.getHeaders()
        });

        if (!response.ok) {
            throw new Error('Failed to fetch profile');
        }

        return await response.json();
    }

    async updateBookProgress(bookId: string, currentPage: number): Promise<void> {
        const response = await fetch(`${PUBLIC_API_BASE_URL}/currently-reading`, {
            method: "PUT",
            headers: this.getHeaders(),
            body: JSON.stringify({ bookId, currentPage })
        });

        if (!response.ok) {
            throw new Error('Failed to update book progress');
        }
    }

    async deleteBook(bookId: string): Promise<void> {
        const response = await fetch(`${PUBLIC_API_BASE_URL}/currently-reading?bookId=${bookId}`, {
            method: "DELETE",
            headers: this.getHeaders()
        });

        if (!response.ok) {
            throw new Error('Failed to delete book');
        }
    }
} 
// src/lib/services/BookService.ts
import { PUBLIC_API_BASE_URL } from '$env/static/public';
import type { Profile } from '$lib/types';

interface SearchResult {
	bookId: string;
	title: string;
	authors?: string[];
	thumbnail?: string;
}

export class BookService {
	private token: string;

	constructor(token: string) {
		this.token = token;
	}

	private getOptions(method: string, body?: unknown) {
		return {
			method,
			credentials: 'include' as RequestCredentials,
			headers: {
				'Content-Type': 'application/json',
				Authorization: `Bearer ${this.token}` // Attach your token as a Bearer token
			},
			body: body ? JSON.stringify(body) : undefined
		};
	}

	async getProfile(): Promise<Profile> {
		const response = await fetch(`${PUBLIC_API_BASE_URL}/profile`, this.getOptions('GET'));
		if (!response.ok) {
			throw new Error('Failed to fetch profile');
		}
		return await response.json();
	}

	async updateBookProgress(bookId: string, currentPage: number): Promise<void> {
		const response = await fetch(
			`${PUBLIC_API_BASE_URL}/currently-reading`, 
			this.getOptions('PUT', { bookId, currentPage })
		);
		if (!response.ok) {
			throw new Error('Failed to update book progress');
		}
	}

	async removeFromList(bookId: string, listType: string): Promise<void> {
		const response = await fetch(
			`${PUBLIC_API_BASE_URL}/list?listType=${listType}&bookId=${bookId}`,
			this.getOptions('DELETE')
		);
		if (!response.ok) {
			throw new Error('Failed to remove book from list');
		}
	}

	async removeFromCurrentlyReading(bookId: string): Promise<void> {
		const response = await fetch(
			`${PUBLIC_API_BASE_URL}/currently-reading?bookId=${bookId}`,
			this.getOptions('DELETE')
		);
		if (!response.ok) {
			throw new Error('Failed to remove book from currently reading');
		}
	}

	async startReading(bookId: string, listName: string): Promise<void> {
		const response = await fetch(
			`${PUBLIC_API_BASE_URL}/currently-reading/start-reading`,
			this.getOptions('POST', { bookId, listName })
		);
		if (!response.ok) {
			throw new Error('Failed to start reading book');
		}
	}

	async finishReading(bookId: string): Promise<void> {
		const response = await fetch(
			`${PUBLIC_API_BASE_URL}/currently-reading/finish-reading`,
			this.getOptions('POST', { bookId })
		);
		if (!response.ok) {
			throw new Error('Failed to finish reading book');
		}
	}

	async searchBooks(query: string): Promise<SearchResult[]> {
		const response = await fetch(
			`${PUBLIC_API_BASE_URL}/books/search?q=${encodeURIComponent(query)}`,
			this.getOptions('GET')
		);
		if (!response.ok) {
			throw new Error('Failed to search books');
		}
		return await response.json();
	}

	async addToList(bookId: string, listType: string): Promise<void> {
		const response = await fetch(
			`${PUBLIC_API_BASE_URL}/list`,
			this.getOptions('POST', { bookId, listType })
		);
		if (!response.ok) {
			throw new Error('Failed to add book to list');
		}
	}
}

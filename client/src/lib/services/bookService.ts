// src/lib/services/BookService.ts
import { PUBLIC_API_BASE_URL } from '$env/static/public';
import type { Book } from '$lib/types';

interface SearchResult {
	bookId: string;
	title: string;
	authors?: string[];
	thumbnail?: string;
}

export class BookService {
	private readonly token: string;

	constructor(token: string) {
		this.token = token;
	}

	private getOptions(method: string, body?: unknown) {
		return {
			method,
			headers: {
				'Content-Type': 'application/json',
				Authorization: `Bearer ${this.token}`
			},
			body: body ? JSON.stringify(body) : undefined
		};
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

	async removeFromShelf(bookId: string, shelfType: string): Promise<void> {
		const response = await fetch(
			`${PUBLIC_API_BASE_URL}/bookshelves?shelfType=${shelfType}&bookId=${bookId}`,
			this.getOptions('DELETE')
		);
		if (!response.ok) {
			throw new Error('Failed to remove book from shelf');
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

	async startReading(bookId: string, shelfName: string): Promise<void> {
		const response = await fetch(
			`${PUBLIC_API_BASE_URL}/currently-reading/start-reading`,
			this.getOptions('POST', { bookId, shelfName })
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

	async addToShelf(bookId: string, shelfType: string): Promise<void> {
		const response = await fetch(
			`${PUBLIC_API_BASE_URL}/bookshelves`,
			this.getOptions('POST', { bookId, shelfType })
		);
		if (!response.ok) {
			throw new Error('Failed to add book to shelf');
		}
	}

	async getBook(bookId: string): Promise<Book> {
		const response = await fetch(
			`${PUBLIC_API_BASE_URL}/books/query?bookId=${bookId}`,
			this.getOptions('GET')
		);
		return await response.json();
	}

	async createBookshelf(shelfName: string): Promise<void> {
		const response = await fetch(
			`${PUBLIC_API_BASE_URL}/bookshelves`,
			this.getOptions('POST', { shelfName })
		);
		if (!response.ok) {
			throw new Error('Failed to create bookshelf');
		}
	}

	async deleteBookshelf(shelfName: string): Promise<void> {
		const response = await fetch(
			`${PUBLIC_API_BASE_URL}/bookshelves?shelfName=${shelfName}`,
			this.getOptions('DELETE')
		);
		if (!response.ok) {
			throw new Error('Failed to delete bookshelf');
		}
	}
}

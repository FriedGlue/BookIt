// src/lib/services/ReadingLogServer.ts
import { PUBLIC_API_BASE_URL } from '$env/static/public';
import type { ReadingLogItem } from '$lib/types';

export class ReadingLogService{
	private readonly token: string;

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

	async getReadingList(): Promise<ReadingLogItem> {
		const response = await fetch(
			`${PUBLIC_API_BASE_URL}/reading-log`,
			this.getOptions('GET')
		);
		return await response.json();
	}

	async updateBookProgress(readingLogItemId: string, pagesRead: number, notes: string): Promise<void> {
		const response = await fetch(
			`${PUBLIC_API_BASE_URL}/currently-reading`,
			this.getOptions('PUT', { readingLogItemId, pagesRead, notes })
		);
		if (!response.ok) {
			throw new Error('Failed to update book progress');
		}
	}

	async removeFromList(readingLogItemId: string): Promise<void> {
		const response = await fetch(
			`${PUBLIC_API_BASE_URL}/reading-log?readingLogId=${readingLogItemId}`,
			this.getOptions('DELETE')
		);
		if (!response.ok) {
			throw new Error('Failed to remove book from list');
		}
	}
}

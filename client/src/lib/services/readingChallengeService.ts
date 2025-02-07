import type { ReadingChallenge } from '$lib/types';
import { PUBLIC_API_BASE_URL } from '$env/static/public';

export class ReadingChallengeService {
	private token: string;

	constructor(token: string) {
		this.token = token;
	}

	private getOptions(method: string, body?: unknown) {
		return {
			method,
			headers: {
				'Content-Type': 'application/json',
				Authorization: `Bearer ${this.token}` // Attach your token as a Bearer token
			},
			body: body ? JSON.stringify(body) : undefined
		};
	}

	async createChallenge(
		challenge: Omit<ReadingChallenge, 'id' | 'userId' | 'createdAt' | 'updatedAt' | 'progress'>
	): Promise<ReadingChallenge> {
		const response = await fetch(
			`${PUBLIC_API_BASE_URL}/challenges`,
			this.getOptions('POST', challenge)
		);

		if (!response.ok) {
			throw new Error('Failed to create challenge');
		}

		return response.json();
	}

	async getUserChallenges(): Promise<ReadingChallenge[]> {
		const response = await fetch(`${PUBLIC_API_BASE_URL}/challenges`, this.getOptions('GET'));

		console.log(response);

		if (!response.ok) {
			throw new Error('Failed to fetch challenges');
		}

		return response.json();
	}

	async updateChallenge(id: string, current: number): Promise<ReadingChallenge> {
		const response = await fetch(
			`${PUBLIC_API_BASE_URL}/challenges/${id}`,
			this.getOptions('PUT', { current })
		);

		if (!response.ok) {
			throw new Error('Failed to update challenge');
		}

		return response.json();
	}

	async deleteChallenge(id: string): Promise<void> {
		const response = await fetch(
			`${PUBLIC_API_BASE_URL}/challenges/${id}`,
			this.getOptions('DELETE')
		);

		if (!response.ok) {
			throw new Error('Failed to delete challenge');
		}
	}
}

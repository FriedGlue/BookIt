// src/lib/services/BookService.ts
import { PUBLIC_API_BASE_URL } from '$env/static/public';
import type { Book } from '$lib/types';

interface SearchResult {
	bookId: string;
	title: string;
	authors?: string[];
	thumbnail?: string;
	source: string;
	openLibraryId?: string;
}

export class BookService {
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
		try {
			// First ensure the book exists in our database
			const internalBookId = await this.ensureBookExists(bookId);
			
			// Now start reading the book with the internal ID
			const response = await fetch(
				`${PUBLIC_API_BASE_URL}/currently-reading/start-reading`,
				this.getOptions('POST', { bookId: internalBookId || bookId, listName })
			);
			
			if (!response.ok) {
				throw new Error('Failed to start reading book');
			}
		} catch (error) {
			console.error('Error starting to read book:', error);
			throw new Error('Failed to start reading book');
		}
	}

	async addToCurrentlyReading(bookId: string): Promise<void> {
		try {
			// First ensure the book exists in our database
			const internalBookId = await this.ensureBookExists(bookId);
			
			// Now add the book to currently reading with the internal ID
			const response = await fetch(
				`${PUBLIC_API_BASE_URL}/currently-reading`,
				this.getOptions('POST', { bookId: internalBookId || bookId })
			);
			
			if (!response.ok) {
				throw new Error('Failed to add book to currently reading');
			}
		} catch (error) {
			console.error('Error adding book to currently reading:', error);
			throw new Error('Failed to add book to currently reading');
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
			`${PUBLIC_API_BASE_URL}/books/combined-search?q=${encodeURIComponent(query)}`,
			this.getOptions('GET')
		);
		if (!response.ok) {
			throw new Error('Failed to search books');
		}
		return await response.json();
	}

	async addToList(bookId: string, listType: string): Promise<void> {
		try {
			// First ensure the book exists in our database
			const internalBookId = await this.ensureBookExists(bookId);
			
			// Now add the book to the list with the internal ID
			const response = await fetch(
				`${PUBLIC_API_BASE_URL}/list`,
				this.getOptions('POST', { bookId: internalBookId || bookId, listType })
			);
			
			if (!response.ok) {
				throw new Error('Failed to add book to list');
			}
		} catch (error) {
			console.error('Error adding book to list:', error);
			throw new Error('Failed to add book to list');
		}
	}

	async getBook(bookId: string): Promise<Book> {
		const response = await fetch(
			`${PUBLIC_API_BASE_URL}/books/${bookId}`,
			this.getOptions('GET')
		);
		return await response.json();
	}

	async createCustomBookshelf(listName: string): Promise<void> {
		const response = await fetch(
			`${PUBLIC_API_BASE_URL}/list?listName=${encodeURIComponent(listName)}`,
			this.getOptions('POST')
		);
		if (!response.ok) {
			console.error('Failed to create custom bookshelf', response);
			throw new Error('Failed to create custom bookshelf');
		}
	}

	async deleteCustomBookshelf(listName: string): Promise<void> {
		const response = await fetch(
			`${PUBLIC_API_BASE_URL}/list?listName=${encodeURIComponent(listName)}`,
			this.getOptions('DELETE')
		);
		if (!response.ok) {
			console.error('Failed to delete custom bookshelf', response);
			throw new Error('Failed to delete custom bookshelf');
		}
	}

	// Helper method to ensure a book exists in our database
	async ensureBookExists(bookId: string): Promise<string> {
		console.log(`[DEBUG] ensureBookExists called for bookId: ${bookId}`);
		
		if (!bookId) {
			console.error('[DEBUG] Invalid bookId - empty or undefined');
			return ''; // Return empty string to signal error
		}
		
		try {
			// Try to get the book from our database
			console.log(`[DEBUG] Attempting to retrieve book from database: ${bookId}`);
			
			// Check if this is a UUID format
			const isUuid = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(bookId);
			
			// Check if this is an Open Library ID format
			const isOpenLibraryId = bookId.startsWith('OL') && (bookId.endsWith('W') || bookId.endsWith('M'));
			
			// For UUID format or what appears to be our internal ID format, try direct book lookup
			if (isUuid || (!isOpenLibraryId && bookId.length > 8)) {
				try {
					const book = await this.getBook(bookId);
					console.log(`[DEBUG] Book ${bookId} found in database by BookID`);
					return bookId;
				} catch (error) {
					console.log(`[DEBUG] Book not found in database by BookID: ${error}`);
				}
			}
			
			// For Open Library ID format, try searching by OpenLibrary ID
			if (isOpenLibraryId) {
				// If it's an OL ID, try to find it directly by openLibraryId
				try {
					const response = await fetch(
						`${PUBLIC_API_BASE_URL}/books/search?openLibraryId=${encodeURIComponent(bookId)}`,
						this.getOptions('GET')
					);
					
					if (response.ok) {
						const books = await response.json();
						console.log(`[DEBUG] Book search by OpenLibraryId response:`, books);
						
						// If we found books and have a valid bookId in the first result
						if (Array.isArray(books) && books.length > 0 && books[0].bookId) {
							console.log(`[DEBUG] Book ${bookId} found in database as ${books[0].bookId}`);
							return books[0].bookId;
						}
						
						console.log(`[DEBUG] Book ${bookId} not found by OpenLibraryId, will save it`);
					} else {
						console.log(`[DEBUG] Book ${bookId} not found by OpenLibraryId (${response.status}), will save it`);
					}
				} catch (error) {
					console.log(`[DEBUG] Error searching for book by OpenLibraryId: ${error}`);
				}
				
				// If we have an OpenLibrary ID, save it to our database
				// We need to save the book from external source
				console.log(`[DEBUG] Book not found in database, saving from external source: ${bookId}`);
				const url = `${PUBLIC_API_BASE_URL}/books/save-external-book`;
				const options = this.getOptions('POST', { bookId });
				
				console.log(`[DEBUG] Making request to: ${url}`);
				console.log(`[DEBUG] With options:`, options);
				
				const response = await fetch(url, options);
				
				console.log(`[DEBUG] Response status:`, response.status);
				
				if (!response.ok) {
					const errorText = await response.text();
					console.error(`[DEBUG] Error response from save-external-book:`, errorText);
					throw new Error(`Failed to save external book: ${response.status} - ${errorText}`);
				}
				
				const responseData = await response.json();
				console.log(`[DEBUG] Successfully saved book:`, responseData);
				
				// Verify that we have a valid bookId in the response
				const savedBookId = responseData.bookId;
				if (!savedBookId) {
					console.error('[DEBUG] Saved book response missing bookId:', responseData);
					return bookId; // Fall back to original ID
				}
				
				if (savedBookId !== bookId) {
					console.log(`[DEBUG] Note: Original bookId ${bookId} was saved with internal ID ${savedBookId}`);
				}
				
				return savedBookId;
			}
			
			// If we reach here, just return the original ID
			return bookId;
			
		} catch (fetchError: any) {
			console.error(`[DEBUG] Error ensuring book exists:`, fetchError);
			// Return the original ID on error - safer than returning nothing
			return bookId;
		}
	}
}


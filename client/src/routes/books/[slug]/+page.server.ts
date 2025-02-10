// file: src/routes/[slug]/+page.server.ts
import type { PageServerLoad, Actions } from './$types';
import { BookService } from '$lib/services/bookService';
import { PUBLIC_API_BASE_URL } from '$env/static/public';
import type { Profile } from '$lib/types';

export const load: PageServerLoad = async ({ fetch, params, cookies }) => {
  // First, try to get the custom lists from the profile (if a token is present)
  const token = cookies.get('idToken');
  let customLists = {};
  if (token) {
    const profileResponse = await fetch(`${PUBLIC_API_BASE_URL}/profile`, {
      headers: {
        Authorization: `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    });
    if (profileResponse.ok) {
      const profile: Profile = await profileResponse.json();
      customLists = profile.lists?.customLists || {};
    }
  }

  try {
    console.log('Fetching book with slug:', params.slug);

    // First, attempt to get the book from your database via your API
    const dbRes = await fetch(`/api/books/searchByOpenLibraryId?q=${encodeURIComponent(params.slug)}`);
    let book = null;

	console.log('dbRes', dbRes);

    if (dbRes.ok) {
      const books = await dbRes.json();
      console.log('Raw API response from DB:', books);

      if (Array.isArray(books) && books.length > 0) {
        // Use the first result if available
        book = books[0];
      }
    } else {
      console.error('Error searching DB for book:', await dbRes.text());
    }

    // If the book was not found in your DB, fetch it from Open Library.
    if (!book) {
      console.log('No book found in DB, falling back to Open Library');
      const olRes = await fetch(`https://openlibrary.org/works/${params.slug}.json`);
      if (olRes.ok) {
        const olBook = await olRes.json();
        console.log('Raw Open Library response:', olBook);

        // Transform the Open Library data into the shape your page expects.
        book = {
          // Use the provided slug as the bookId.
          bookId: params.slug,
          title: olBook.title,
          // For authors, note that the works endpoint returns an array of objects like { author: { key: '/authors/OL12345A' } }
          // You might want to fetch the names separately. Here we simply extract the author keys.
          authors: olBook.authors
            ? olBook.authors.map((a: any) => a.author.key)
            : ['Unknown Author'],
          // Use the first cover ID if available to construct a cover image URL.
          coverImageUrl: olBook.covers && olBook.covers.length > 0
            ? `https://covers.openlibrary.org/b/id/${olBook.covers[0]}-L.jpg`
            : null,
          // For description, sometimes it can be a string or an object with a "value" key.
          description: olBook.description
            ? typeof olBook.description === 'string'
              ? olBook.description
              : olBook.description.value
            : null,
          // You can add additional fields if needed, such as page count or tags.
        };
      } else {
        console.error('Open Library request failed:', await olRes.text());
      }
    }

    return {
      book,
      customLists
    };
  } catch (error) {
    console.error('Error loading book:', error);
    return { book: null, customLists };
  }
};

export const actions: Actions = {
  addToList: async ({ request, cookies }) => {
    const token = cookies.get('idToken');
    if (!token) return { error: 'No token' };

    const formData = await request.formData();
    const bookId = formData.get('bookId')?.toString();
    const listType = formData.get('listType')?.toString();

    if (!bookId || !listType) {
      return { error: 'Missing required fields' };
    }

    try {
      const bookService = new BookService(token);
      await bookService.addToList(bookId, listType);
      return { success: true };
    } catch (error) {
      console.error('Failed to add book to list:', error);
      return { error: 'Failed to add book to list' };
    }
  },

  startReading: async ({ request, cookies }) => {
    const token = cookies.get('idToken');
    if (!token) return { error: 'No token' };

    const formData = await request.formData();
    const bookId = formData.get('bookId')?.toString();

    if (!bookId) {
      return { error: 'Missing bookId' };
    }

    try {
      const bookService = new BookService(token);
      await bookService.addToCurrentlyReading(bookId);
      return { success: true };
    } catch (error) {
      console.error('Failed to start reading book:', error);
      return { error: 'Failed to start reading book' };
    }
  }
};

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
    let bookId = params.slug;
    let book = null;
    let notFound = false;
    let error = null;

    // Validate if this is a reasonable ID format before making requests
    const isOpenLibraryId = bookId.startsWith('OL') && (bookId.endsWith('W') || bookId.endsWith('M'));
    const isUuid = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(bookId);
    
    if (!isOpenLibraryId && !isUuid) {
      console.warn('Invalid book ID format:', bookId);
      notFound = true;
      error = "Invalid book ID format. Book IDs must be either OpenLibrary IDs (starting with 'OL' and ending with 'W' or 'M') or valid UUIDs.";
    } else {
      // First, attempt to get the book from your database via your API
      try {
        const dbRes = await fetch(`/api/books/searchByOpenLibraryId?q=${encodeURIComponent(params.slug)}`);
        
        if (dbRes.ok) {
          const books = await dbRes.json();
          console.log('Raw API response from DB:', books);

          if (Array.isArray(books) && books.length > 0) {
            // Use the first result if available
            book = books[0];
          } else {
            // Empty array means no books found
            notFound = true;
          }
        } else {
          console.error('Error searching DB for book:', await dbRes.text());
          // API error, but we'll still try OpenLibrary
        }
      } catch (dbError) {
        console.error('Exception searching database:', dbError);
        // Continue to OpenLibrary if there was an error with the DB
      }

      // If the book was not found in your DB, fetch it from Open Library.
      if (!book && isOpenLibraryId) {
        console.log('No book found in DB, falling back to Open Library');
        try {
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
              openLibraryId: params.slug,
              // Default to 0 for totalPages until we know the actual value
              totalPages: olBook.number_of_pages || 0,
            };
          } else {
            const errorText = await olRes.text();
            console.error('Open Library request failed:', errorText);
            notFound = true;
            error = `Book not found in OpenLibrary: ${errorText}`;
          }
        } catch (olError) {
          console.error('Exception fetching from OpenLibrary:', olError);
          notFound = true;
          error = `Error fetching from OpenLibrary: ${olError instanceof Error ? olError.message : String(olError)}`;
        }
      } else if (!book) {
        notFound = true;
        error = "Book not found in database";
      }
    }

    return {
      book,
      customLists,
      notFound,
      error
    };
  } catch (error) {
    console.error('Error loading book:', error);
    return { 
      book: null, 
      customLists,
      notFound: true,
      error: `Error loading book: ${error instanceof Error ? error.message : String(error)}`
    };
  }
};

export const actions: Actions = {
  addToList: async ({ request, cookies }) => {
    const token = cookies.get('idToken');
    if (!token) return { error: 'No token' };

    const formData = await request.formData();
    const bookId = formData.get('bookId')?.toString();
    const openLibraryId = formData.get('openLibraryId')?.toString();
    const listType = formData.get('listType')?.toString();

    if (!bookId || !listType) {
      return { error: 'Missing required fields' };
    }

    try {
      const bookService = new BookService(token);
      // Use OpenLibraryId if it exists and bookId appears to be an OL ID
      const idToUse = (bookId.startsWith('OL') && openLibraryId) ? openLibraryId : bookId;
      console.log(`Adding book to list ${listType}. Using ID: ${idToUse}`);
      await bookService.addToList(idToUse, listType);
      return { success: true };
    } catch (error) {
      console.error('Failed to add book to list:', error);
      return { error: String(error) };
    }
  },

  startReading: async ({ request, cookies }) => {
    const token = cookies.get('idToken');
    if (!token) return { error: 'No token' };

    const formData = await request.formData();
    const bookId = formData.get('bookId')?.toString();
    const openLibraryId = formData.get('openLibraryId')?.toString();

    if (!bookId) {
      return { error: 'Missing bookId' };
    }

    try {
      const bookService = new BookService(token);
      // Use OpenLibraryId if it exists and bookId appears to be an OL ID
      const idToUse = (bookId.startsWith('OL') && openLibraryId) ? openLibraryId : bookId;
      console.log(`Starting reading book. Using ID: ${idToUse}`);
      await bookService.addToCurrentlyReading(idToUse);
      return { success: true };
    } catch (error) {
      console.error('Failed to start reading book:', error);
      return { error: String(error) };
    }
  }
};

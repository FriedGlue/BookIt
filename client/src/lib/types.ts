export interface ReadingProgress {
      lastPageRead: number;
      percentage: number;
      lastUpdated: string;
      notes?: string;
  }

export interface Book {
      bookId: string;
      isbn?: string;
      title?: string;
      authors?: string[];
      thumbnail?: string;
      totalPages?: number;
      progress?: ReadingProgress;
  }

export interface CurrentlyReadingItem {
      Book: Book;
      startedDate?: string;
  }

export interface Profile {
      _id: string;
      profileInformation: {
          username?: string;
          email?: string;
      };
      currentlyReading: CurrentlyReadingItem[];
      lists?: {
          toBeRead?: ToBeReadItem[];
          read?: ReadItem[];
          customLists?: Record<string, DisplayBook[]>;
      };
  }

export interface DisplayBook {
      title: string;
      author: string;
      bookId: string;
      thumbnail: string;
      progress: number;
      totalPages: number;
  }

export interface ToBeReadItem {
      bookId: string;
      thumbnail?: string;
      addedDate?: string;
      order?: number;
  }

export interface ReadItem {
      bookId: string;
      completedDate?: string;
      rating?: number;
      order?: number;
      review?: string;
      thumbnail?: string;
  }
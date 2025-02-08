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
	totalPages: number;
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
	bookshelves?: UserBookshelves;
	readingLog?: ReadingLogItem[];
	challenges?: ReadingChallenge[];
}

export interface UserBookshelves {
	toBeRead?: ToBeReadBook[];
	read?: ReadBook[];
	customShelves?: Record<string, CustomShelfBook[]>;
}

export interface ToBeReadBook {
	bookId: string;
	thumbnail?: string;
	addedDate?: string;
	order?: number;
	title?: string;
	authors?: string[];
}

export interface ReadBook {
	bookId: string;
	thumbnail?: string;
	completedDate?: string;
	rating?: number;
	order?: number;
	review?: string;
	title?: string;
	authors?: string[];
}

export interface CustomShelfBook {
	bookId: string;
	thumbnail?: string;
	addedDate?: string;
	order?: number;
	title?: string;
	authors?: string[];
}

export interface ReadingLogItem {
	_id: string;
	bookId: string;
	title: string;
	date: string;
	bookThumbnail?: string;
	pagesRead?: number;
	notes?: string;
}

export interface DecodedToken {
	exp: number;
	sub: string;
	email: string;
}

interface Rate {
	required: number;
	currentPace: number;
	scheduleDiff: number;
	unit: string;
	status: 'AHEAD' | 'BEHIND' | 'ON_TRACK';
}

interface Progress {
	current: number;
	percentage: number;
	rate: Rate;
}

export interface ReadingChallenge {
	id: string;
	userId: string;
	name: string;
	type: 'BOOKS' | 'PAGES';
	timeframe: 'YEAR' | 'MONTH' | 'WEEK';
	startDate: string;
	endDate: string;
	target: number;
	progress: Progress;
	createdAt: string;
	updatedAt: string;
}

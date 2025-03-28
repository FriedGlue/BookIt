export interface ReadingProgress {
	lastPageRead: number;
	percentage: number;
	lastUpdated: string;
	notes?: string;
}

export interface Book {
	bookId: string;
	isbn13?: string;
	title?: string;
	titleLowercase?: string;
	authors?: string[];
	pageCount?: number;
	coverImageUrl?: string;
	thumbnail?: string;
	tags?: string[];
	progress?: ReadingProgress;
	_listType?: string;
	totalPages: number;
	openLibraryId?: string;
	description?: string;
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
	challenges?: ReadingChallenge[];
}

export interface DisplayBook {
	title: string;
	author: string;
	bookId: string;
	thumbnail: string;
	progress: number;
	totalPages: number;
	currentPage: number;
	lastUpdated: string;
	openLibraryId?: string;
}

export interface ToBeReadItem {
	bookId: string;
	thumbnail?: string;
	addedDate?: string;
	order?: number;
	openLibraryId?: string;
}

export interface ReadItem {
	bookId: string;
	completedDate?: string;
	rating?: number;
	order?: number;
	review?: string;
	thumbnail?: string;
	title?: string;
	authors?: string[];
	openLibraryId?: string;
}

export interface ReadingLogItem {
	_id: string;
	date: string;
	bookId: string;
	bookThumbnail: string;
	pagesRead: number;
	notes: string;
}

export interface DecodedToken {
	exp: number;
	sub: string;
	email: string;
}

interface Progress {
	current: number;
	percentage: number;
	rate: Rate;
}

interface Rate {
	current: number;
	required: number;
	currentPace: number;
	unit: string;
	status: 'AHEAD' | 'BEHIND' | 'ON_TRACK';
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

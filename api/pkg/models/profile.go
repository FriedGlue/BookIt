package models

type Profile struct {
	ID                 string                 `json:"_id"`
	ProfileInformation ProfileInformation     `json:"profileInformation"`
	CurrentlyReading   []CurrentlyReadingItem `json:"currentlyReading,omitempty"`
	Lists              UserLists              `json:"lists,omitempty"`
	ReadingLog         []ReadingLogItem       `json:"readingLog,omitempty"`
}

type ProfileInformation struct {
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
}

type CurrentlyReadingItem struct {
	Book        Book   `json:"Book"`
	StartedDate string `json:"startedDate,omitempty"`
}

type Book struct {
	BookID     string          `json:"bookId"`
	ISBN       string          `json:"isbn,omitempty"`
	Title      string          `json:"title,omitempty"`
	Authors    []string        `json:"authors,omitempty"`
	Thumbnail  string          `json:"thumbnail,omitempty"`
	TotalPages int             `json:"totalPages,omitempty"`
	Progress   ReadingProgress `json:"progress,omitempty"`
}

type ReadingProgress struct {
	LastPageRead int     `json:"lastPageRead"`
	Percentage   float64 `json:"percentage"`
	LastUpdated  string  `json:"lastUpdated"`
	Notes        string  `json:"notes,omitempty"`
}

type UserLists struct {
	ToBeRead    []ToBeReadItem              `json:"toBeRead,omitempty"`
	Read        []ReadItem                  `json:"read,omitempty"`
	CustomLists map[string][]CustomListItem `json:"customLists,omitempty"`
}

type ToBeReadItem struct {
	BookID    string   `json:"bookId"`
	Thumbnail string   `json:"thumbnail,omitempty"`
	AddedDate string   `json:"addedDate,omitempty"`
	Order     int      `json:"order,omitempty"`
	Title     string   `json:"title,omitempty"`
	Authors   []string `json:"authors,omitempty"`
}

type ReadItem struct {
	BookID        string   `json:"bookId"`
	Thumbnail     string   `json:"thumbnail,omitempty"`
	CompletedDate string   `json:"completedDate,omitempty"`
	Rating        int      `json:"rating,omitempty"`
	Order         int      `json:"order,omitempty"`
	Review        string   `json:"review,omitempty"`
	Title         string   `json:"title,omitempty"`
	Authors       []string `json:"authors,omitempty"`
}

type CustomListItem struct {
	BookID    string   `json:"bookId"`
	Thumbnail string   `json:"thumbnail,omitempty"`
	AddedDate string   `json:"addedDate,omitempty"`
	Order     int      `json:"order,omitempty"`
	Title     string   `json:"title,omitempty"`
	Authors   []string `json:"authors,omitempty"`
}

type ReadingLogItem struct {
	Id            string `json:"_id"`
	BookID        string `json:"bookId"`
	Title         string `json:"title"`
	Date          string `json:"date"`
	BookThumbnail string `json:"bookThumbnail,omitempty"`
	PagesRead     int    `json:"pagesRead,omitempty"`
	Notes         string `json:"notes,omitempty"`
}

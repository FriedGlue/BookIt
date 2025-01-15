package models

type Profile struct {
	ID                 string                 `json:"_id"` // Partition Key = userId
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
	Book Book `json:"Book"`
}

type Book struct {
	BookID     string          `json:"bookId"`
	ISBN       string          `json:"isbn,omitempty"`
	Title      string          `json:"title,omitempty"`
	CoverImage string          `json:"coverImage,omitempty"`
	Progress   ReadingProgress `json:"progress,omitempty"`
}

type ReadingProgress struct {
	LastPageRead int     `json:"lastPageRead,omitempty"`
	Percentage   float64 `json:"percentage,omitempty"`
	StartedDate  string  `json:"startedDate,omitempty"`
	Notes        string  `json:"notes,omitempty"`
	LastUpdated  string  `json:"lastUpdated,omitempty"`
}

type UserLists struct {
	ToBeRead    []ToBeReadItem              `json:"toBeRead,omitempty"`
	Read        []ReadItem                  `json:"read,omitempty"`
	CustomLists map[string][]CustomListItem `json:"customLists,omitempty"`
}

type ToBeReadItem struct {
	BookID    string `json:"bookId"`
	Thumbnail string `json:"thumbnail,omitempty"`
	AddedDate string `json:"addedDate,omitempty"`
	Order     int    `json:"order,omitempty"`
}

type ReadItem struct {
	BookID        string `json:"bookId"`
	CompletedDate string `json:"completedDate,omitempty"`
	Rating        int    `json:"rating,omitempty"`
	Order         int    `json:"order,omitempty"`
	Review        string `json:"review,omitempty"`
}

type CustomListItem struct {
	BookID    string `json:"bookId"`
	Thumbnail string `json:"thumbnail,omitempty"`
	AddedDate string `json:"addedDate,omitempty"`
	Order     int    `json:"order,omitempty"`
}

type ReadingLogItem struct {
	Date             string `json:"date"`
	BookID           string `json:"bookId"`
	BookThumbnail    string `json:"bookThumbnail,omitempty"`
	PagesRead        int    `json:"pagesRead,omitempty"`
	TimeSpentMinutes int    `json:"timeSpentMinutes,omitempty"`
	Notes            string `json:"notes,omitempty"`
}

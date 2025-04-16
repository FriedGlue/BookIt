// pkg/service/reading_list_svc.go
package service

import (
	"context"
	"errors"
	"time"

	"github.com/FriedGlue/BookIt/api/pkg/models"
	"github.com/FriedGlue/BookIt/api/pkg/usecase"
)

type readingListSvc struct {
	repo     usecase.ProfileRepo
	bookRepo usecase.BookRepo
}

// NewReadingListService creates a new reading list service that handles both standard and custom lists
func NewReadingListService(r usecase.ProfileRepo, book usecase.BookRepo) usecase.ReadingListService {
	return &readingListSvc{repo: r, bookRepo: book}
}

func (s *readingListSvc) GetLists(ctx context.Context, userID string) (map[string][]models.ReadItem, error) {
	profile, err := s.repo.LoadProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	lists := make(map[string][]models.ReadItem)

	// Convert ToBeRead
	for _, item := range profile.Lists.ToBeRead {
		lists["to_be_read"] = append(lists["to_be_read"], models.ReadItem{
			BookID:        item.BookID,
			Title:         item.Title,
			Authors:       item.Authors,
			Thumbnail:     item.Thumbnail,
			CompletedDate: item.AddedDate,
		})
	}

	// Convert Read
	for _, item := range profile.Lists.Read {
		lists["read"] = append(lists["read"], models.ReadItem{
			BookID:        item.BookID,
			Title:         item.Title,
			Authors:       item.Authors,
			Thumbnail:     item.Thumbnail,
			Rating:        item.Rating,
			Review:        item.Review,
			CompletedDate: item.CompletedDate,
		})
	}

	// Add custom lists
	customLists, err := s.GetCustomLists(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Merge custom lists into all lists
	for name, items := range customLists {
		lists[name] = items
	}

	return lists, nil
}

func (s *readingListSvc) GetToBeRead(ctx context.Context, userID string) ([]models.ReadItem, error) {
	profile, err := s.repo.LoadProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	var items []models.ReadItem
	for _, item := range profile.Lists.ToBeRead {
		items = append(items, models.ReadItem{
			BookID:        item.BookID,
			Title:         item.Title,
			Authors:       item.Authors,
			Thumbnail:     item.Thumbnail,
			CompletedDate: item.AddedDate,
		})
	}
	return items, nil
}

func (s *readingListSvc) GetRead(ctx context.Context, userID string) ([]models.ReadItem, error) {
	profile, err := s.repo.LoadProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	var items []models.ReadItem
	for _, item := range profile.Lists.Read {
		items = append(items, models.ReadItem{
			BookID:        item.BookID,
			Title:         item.Title,
			Authors:       item.Authors,
			Thumbnail:     item.Thumbnail,
			Rating:        item.Rating,
			Review:        item.Review,
			CompletedDate: item.CompletedDate,
		})
	}
	return items, nil
}

func (s *readingListSvc) AddToBeRead(ctx context.Context, userID string, book models.Book) error {
	profile, err := s.repo.LoadProfile(ctx, userID)
	if err != nil {
		return err
	}

	item := models.ToBeReadItem{
		BookID:    book.BookID,
		Title:     book.Title,
		Authors:   book.Authors,
		Thumbnail: book.Thumbnail,
		AddedDate: time.Now().Format(time.RFC3339),
		Order:     len(profile.Lists.ToBeRead),
	}

	profile.Lists.ToBeRead = append(profile.Lists.ToBeRead, item)
	return s.repo.SaveProfile(ctx, profile)
}

func (s *readingListSvc) AddToRead(ctx context.Context, userID string, book models.Book, rating int, review string) error {
	profile, err := s.repo.LoadProfile(ctx, userID)
	if err != nil {
		return err
	}

	item := models.ReadItem{
		BookID:        book.BookID,
		Title:         book.Title,
		Authors:       book.Authors,
		Thumbnail:     book.Thumbnail,
		Rating:        rating,
		Review:        review,
		CompletedDate: time.Now().Format(time.RFC3339),
	}

	profile.Lists.Read = append(profile.Lists.Read, item)
	return s.repo.SaveProfile(ctx, profile)
}

// Custom list methods
func (s *readingListSvc) GetCustomLists(ctx context.Context, userID string) (map[string][]models.ReadItem, error) {
	profile, err := s.repo.LoadProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	customLists := make(map[string][]models.ReadItem)
	for name, items := range profile.Lists.CustomLists {
		for _, item := range items {
			customLists[name] = append(customLists[name], models.ReadItem{
				BookID:        item.BookID,
				Title:         item.Title,
				Authors:       item.Authors,
				Thumbnail:     item.Thumbnail,
				CompletedDate: item.AddedDate,
			})
		}
	}
	return customLists, nil
}

func (s *readingListSvc) CreateCustomList(ctx context.Context, userID, listName string) error {
	profile, err := s.repo.LoadProfile(ctx, userID)
	if err != nil {
		return err
	}

	// Check if list already exists
	if _, exists := profile.Lists.CustomLists[listName]; exists {
		return errors.New("list already exists")
	}

	// Initialize empty custom list
	if profile.Lists.CustomLists == nil {
		profile.Lists.CustomLists = make(map[string][]models.CustomListItem)
	}
	profile.Lists.CustomLists[listName] = []models.CustomListItem{}

	return s.repo.SaveProfile(ctx, profile)
}

func (s *readingListSvc) AddToCustomList(ctx context.Context, userID, listName string, book models.Book) error {
	profile, err := s.repo.LoadProfile(ctx, userID)
	if err != nil {
		return err
	}

	// Check if list exists
	if _, exists := profile.Lists.CustomLists[listName]; !exists {
		return errors.New("list does not exist")
	}

	// Add book to custom list
	item := models.CustomListItem{
		BookID:    book.BookID,
		Title:     book.Title,
		Authors:   book.Authors,
		Thumbnail: book.Thumbnail,
		AddedDate: time.Now().Format(time.RFC3339),
		Order:     len(profile.Lists.CustomLists[listName]),
	}

	profile.Lists.CustomLists[listName] = append(profile.Lists.CustomLists[listName], item)
	return s.repo.SaveProfile(ctx, profile)
}

func (s *readingListSvc) DeleteCustomList(ctx context.Context, userID, listName string) error {
	profile, err := s.repo.LoadProfile(ctx, userID)
	if err != nil {
		return err
	}

	// Check if list exists
	if _, exists := profile.Lists.CustomLists[listName]; !exists {
		return errors.New("list does not exist")
	}

	// Remove list
	delete(profile.Lists.CustomLists, listName)
	return s.repo.SaveProfile(ctx, profile)
}

func (s *readingListSvc) RemoveFromList(ctx context.Context, userID, listName, bookID string) error {
	profile, err := s.repo.LoadProfile(ctx, userID)
	if err != nil {
		return err
	}

	// Handle based on list type
	switch listName {
	case "to_be_read":
		for i, item := range profile.Lists.ToBeRead {
			if item.BookID == bookID {
				profile.Lists.ToBeRead = append(profile.Lists.ToBeRead[:i], profile.Lists.ToBeRead[i+1:]...)
				break
			}
		}
	case "read":
		for i, item := range profile.Lists.Read {
			if item.BookID == bookID {
				profile.Lists.Read = append(profile.Lists.Read[:i], profile.Lists.Read[i+1:]...)
				break
			}
		}
	default:
		// Assume it's a custom list
		list, exists := profile.Lists.CustomLists[listName]
		if !exists {
			return errors.New("list does not exist")
		}

		found := false
		for i, item := range list {
			if item.BookID == bookID {
				profile.Lists.CustomLists[listName] = append(list[:i], list[i+1:]...)
				found = true
				break
			}
		}

		if !found {
			return errors.New("book not found in list")
		}
	}

	return s.repo.SaveProfile(ctx, profile)
}

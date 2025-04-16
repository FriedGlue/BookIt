// pkg/service/readinglog_svc.go
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/FriedGlue/BookIt/api/pkg/models"
	"github.com/FriedGlue/BookIt/api/pkg/usecase"
)

type readingLogSvc struct {
	repo usecase.ProfileRepo
}

func NewReadingLogService(r usecase.ProfileRepo) usecase.ReadingLogService {
	return &readingLogSvc{repo: r}
}

func (s *readingLogSvc) List(ctx context.Context, userID string) ([]models.ReadingLogItem, error) {
	p, err := s.repo.LoadProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	return p.ReadingLog, nil
}

func (s *readingLogSvc) Create(ctx context.Context, userID string, entry models.ReadingLogItem) error {
	p, err := s.repo.LoadProfile(ctx, userID)
	if err != nil {
		return err
	}
	entry.Id = fmt.Sprintf("%d", time.Now().UnixNano())
	entry.Date = time.Now().Format(time.RFC3339)
	p.ReadingLog = append(p.ReadingLog, entry)
	return s.repo.SaveProfile(ctx, p)
}

func (s *readingLogSvc) Update(ctx context.Context, userID, entryID string, pagesRead int, notes string) error {
	p, err := s.repo.LoadProfile(ctx, userID)
	if err != nil {
		return err
	}
	for i := range p.ReadingLog {
		if p.ReadingLog[i].Id == entryID {
			p.ReadingLog[i].PagesRead = pagesRead
			p.ReadingLog[i].Notes = notes
			break
		}
	}
	return s.repo.SaveProfile(ctx, p)
}

func (s *readingLogSvc) Delete(ctx context.Context, userID, entryID string) error {
	p, err := s.repo.LoadProfile(ctx, userID)
	if err != nil {
		return err
	}
	for i := range p.ReadingLog {
		if p.ReadingLog[i].Id == entryID {
			p.ReadingLog = append(p.ReadingLog[:i], p.ReadingLog[i+1:]...)
			break
		}
	}
	return s.repo.SaveProfile(ctx, p)
}

// pkg/service/profile_svc.go
package service

import (
	"github.com/FriedGlue/BookIt/api/pkg/usecase"
)

// NewProfileService creates a simple profile service for profile management
func NewProfileService(repo usecase.ProfileRepo) usecase.ProfileService {
	return usecase.NewProfileService(repo)
}

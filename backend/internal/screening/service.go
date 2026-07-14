package screening

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/domain"
)

var ErrNotFound = errors.New("screening not found")

type Repository interface {
	List(context.Context) ([]domain.Screening, error)
	FindByID(context.Context, bson.ObjectID) (domain.Screening, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context) ([]domain.Screening, error) {
	screenings, err := s.repository.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list screenings: %w", err)
	}

	return screenings, nil
}

func (s *Service) FindByID(ctx context.Context, id bson.ObjectID) (domain.Screening, error) {
	screening, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return domain.Screening{}, fmt.Errorf("find screening: %w", err)
	}

	return screening, nil
}

package places

import (
	"context"
	"errors"

	"github.com/nguyentanvinh/recommendations-be/backend/internal/platform/handler"
	"github.com/nguyentanvinh/recommendations-be/backend/internal/platform/model"
)

// Service holds the rules that are true of a place regardless of how it was
// asked for: what makes an input valid, what a missing row means, and how a
// page is bounded. The handler owns HTTP, the repository owns SQL, and
// neither owns any of this.
type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// List takes params by pointer so the caller sees the normalised values: the
// handler reports the page actually served, which is not the page requested
// when limit was omitted.
func (s *Service) List(ctx context.Context, params *ListParams) (*model.CursorReponse[Place], error) {
	if err := params.Normalise(); err != nil {
		return nil, err
	}

	rows, total, err := s.repo.List(ctx, *params)
	if err != nil {
		return nil, err
	}

	return model.NewCursorReponse(rows, params.Limit, total, func(p Place) int64 { return p.ID }), nil
}

func (s *Service) Get(ctx context.Context, id int64) (*Place, error) {
	place, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, asAPIError(err)
	}
	return place, nil
}

func (s *Service) Create(ctx context.Context, in *PlaceInput) (*Place, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, in)
}

func (s *Service) Update(ctx context.Context, id int64, in *PlaceInput) (*Place, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}
	place, err := s.repo.Update(ctx, id, in)
	if err != nil {
		return nil, asAPIError(err)
	}
	return place, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return asAPIError(err)
	}
	return nil
}

// asAPIError turns the repository's ErrNotFound into a 404; anything else
// keeps its own classification and falls through to INTERNAL.
func asAPIError(err error) error {
	if errors.Is(err, ErrNotFound) {
		return handler.Error(handler.ECNotFound, err)
	}
	return err
}

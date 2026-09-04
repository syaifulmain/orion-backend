package corpus

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrInvalidSplit = errors.New("split must be one of: train, eval, or test")
	ErrQueryTooLong = errors.New("search query (q) exceeds maximum length of 200 characters")
)

type Service interface {
	GetCorpusList(ctx context.Context, params FilterParams) ([]Corpus, int64, FilterParams, error)
	GetCorpusByID(ctx context.Context, id string) (*Corpus, error)
	ExportCorpusStream(ctx context.Context, params FilterParams, writeLine func(record *Corpus) error) error
	ValidateFilter(params FilterParams) (FilterParams, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) ValidateFilter(params FilterParams) (FilterParams, error) {
	// Normalize & Validate Q
	params.Q = strings.TrimSpace(params.Q)
	if len(params.Q) > 200 {
		return params, ErrQueryTooLong
	}

	// Normalize & Validate Split
	params.Split = strings.ToLower(strings.TrimSpace(params.Split))
	if params.Split != "" {
		if params.Split != "train" && params.Split != "eval" && params.Split != "test" {
			return params, ErrInvalidSplit
		}
	}

	// Normalize Source & License
	params.Source = strings.TrimSpace(params.Source)
	params.License = strings.TrimSpace(params.License)

	return params, nil
}

func (s *service) GetCorpusList(ctx context.Context, params FilterParams) ([]Corpus, int64, FilterParams, error) {
	validated, err := s.ValidateFilter(params)
	if err != nil {
		return nil, 0, params, err
	}

	// Pagination defaults
	if validated.Limit <= 0 {
		validated.Limit = 25
	} else if validated.Limit > 1000 {
		validated.Limit = 1000
	}

	if validated.Offset < 0 {
		validated.Offset = 0
	}

	records, total, err := s.repo.Find(ctx, validated)
	if err != nil {
		return nil, 0, validated, err
	}

	return records, total, validated, nil
}

func (s *service) ExportCorpusStream(ctx context.Context, params FilterParams, writeLine func(record *Corpus) error) error {
	validated, err := s.ValidateFilter(params)
	if err != nil {
		return err
	}

	cursor, err := s.repo.FindStream(ctx, validated)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var record Corpus
		if err := cursor.Decode(&record); err != nil {
			return err
		}
		if err := writeLine(&record); err != nil {
			return err
		}
	}

	return cursor.Err()
}

func (s *service) GetCorpusByID(ctx context.Context, id string) (*Corpus, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}
	return s.repo.FindByID(ctx, id)
}

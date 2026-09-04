package corpus

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	ErrNotFound = errors.New("corpus record not found")
)

type Repository interface {
	Find(ctx context.Context, params FilterParams) ([]Corpus, int64, error)
	FindStream(ctx context.Context, params FilterParams) (*mongo.Cursor, error)
	FindByID(ctx context.Context, id string) (*Corpus, error)
}

type repository struct {
	collection *mongo.Collection
}

func NewRepository(db *mongo.Database) Repository {
	return &repository{
		collection: db.Collection("corpus"),
	}
}

func (r *repository) buildFilter(params FilterParams) bson.M {
	filter := bson.M{}

	// Search text (case-insensitive) on text and source
	if params.Q != "" {
		escapedQ := regexp.QuoteMeta(params.Q)
		filter["$or"] = bson.A{
			bson.M{"text": bson.M{"$regex": escapedQ, "$options": "i"}},
			bson.M{"source": bson.M{"$regex": escapedQ, "$options": "i"}},
		}
	}

	// Split subset filter: train, eval, or test
	if params.Split != "" {
		filter["split"] = params.Split
	}

	// Source filter: exact match or prefix
	if params.Source != "" {
		escapedSource := regexp.QuoteMeta(params.Source)
		filter["source"] = bson.M{
			"$regex":   fmt.Sprintf("^%s(/|$)", escapedSource),
			"$options": "i",
		}
	}

	// License filter: case-insensitive exact match
	if params.License != "" {
		escapedLicense := regexp.QuoteMeta(params.License)
		filter["license"] = bson.M{
			"$regex":   fmt.Sprintf("^%s$", escapedLicense),
			"$options": "i",
		}
	}

	return filter
}

func (r *repository) Find(ctx context.Context, params FilterParams) ([]Corpus, int64, error) {
	filter := r.buildFilter(params)

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpts := options.Find().
		SetLimit(int64(params.Limit)).
		SetSkip(int64(params.Offset))

	cursor, err := r.collection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var records []Corpus
	if err := cursor.All(ctx, &records); err != nil {
		return nil, 0, err
	}

	if records == nil {
		records = []Corpus{}
	}

	return records, total, nil
}

func (r *repository) FindStream(ctx context.Context, params FilterParams) (*mongo.Cursor, error) {
	filter := r.buildFilter(params)
	return r.collection.Find(ctx, filter)
}

func (r *repository) FindByID(ctx context.Context, id string) (*Corpus, error) {
	var c Corpus
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&c)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

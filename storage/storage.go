package storage

import (
	"context"
	"errors"
	"github.com/awakari/interests/model/interest"
	"io"
	"time"
)

type (

	// Storage represents the interests storage
	Storage interface {
		io.Closer

		// Create an interest with the specified account and data.
		// Returns a created interest id if successful.
		Create(ctx context.Context, id, groupId, userId string, sd interest.Data) (err error)

		// Read the interest.Data by the interest.Interest id.
		Read(ctx context.Context, id, groupId, userId string) (sd interest.Data, ownerGroupId, ownerUserId string, err error)

		// Update updates the interest.Data
		Update(ctx context.Context, id, groupId, userId string, sd interest.Data) (prev interest.Data, err error)

		// UpdateFollowers updates the followers count
		UpdateFollowers(ctx context.Context, id string, count int64) (err error)

		UpdateResultTime(ctx context.Context, id string, last time.Time) (err error)

		SetEnabledBatch(ctx context.Context, ids []string, enabled bool) (n int64, err error)

		// Delete removes the interest.Interest specified by its unique id.
		// Returns the interest.Data if deleted, error otherwise.
		Delete(ctx context.Context, id, groupId, userId string) (sd interest.Data, err error)

		// Search returns all interest ids matching the query.
		Search(ctx context.Context, q interest.Query, cursor interest.Cursor) (ids []string, err error)

		// SearchByCondition finds all interests those match the specified condition id and feeds these to the
		// specified consumer func.
		SearchByCondition(ctx context.Context, q interest.QueryByCondition, cursor string) (page []interest.ConditionMatch, err error)

		Count(ctx context.Context) (count int64, err error)
		CountUsersUnique(ctx context.Context) (count int64, err error)
	}
)

var (

	// ErrNotFound indicates the interest is missing in the storage and can not be read/updated/deleted.
	ErrNotFound = errors.New("interest was not found")

	// ErrConflict indicates the interest id is already in use.
	ErrConflict = errors.New("interest id is already in use")

	// ErrInternal indicates the internal storage failure happened.
	ErrInternal = errors.New("internal interest storage failure")
)

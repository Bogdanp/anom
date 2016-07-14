package anom

import (
	"errors"
	"reflect"
	"time"

	"github.com/qedus/nds"
	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
)

var (
	// ErrMissingKey is the error that is returned when a Model lacks a
	// key.  This can be returned if you attempt to Delete a Model that
	// hasn't been saved yet or if you attempt to Get a Model that
	// doesn't have a Key.
	ErrMissingKey = errors.New("anom: model does not have a Key")
)

const (
	// StateActive is the state of Model instances that have been persisted.
	StateActive = "active"
	// StateDeleted is the state of Model instances that have been deleted.
	StateDeleted = "deleted"
)

// Model is the interface that all Model structs automatically
// implement when they embed Meta.
type Model interface {
	meta() *Meta
}

// Meta is the struct that all Model structs must embed.  It extends
// entities with metadata about when they were created, last updated
// and deleted as well as their current state and their Key.
//
//   type User struct {
//       Meta
//       Username string
//   }
type Meta struct {
	Key       *datastore.Key `json:"-" datastore:"-"`
	Parent    *datastore.Key `json:"-" datastore:"-"`
	State     string         `json:"-"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt time.Time      `json:"deletedAt"`
}

func (m *Meta) meta() *Meta {
	return m
}

// Option is the type of functional Model options.
type Option func(Model)

// getKind returns the datastore kind of a given Model.
func getKind(m Model) string {
	return reflect.TypeOf(m).Elem().Name()
}

// WithKey is an Option for assigning a datastore Key to a Model's Meta.
func WithKey(k *datastore.Key) Option {
	return func(m Model) {
		meta := m.meta()
		meta.Key = k
	}
}

// WithParent is an Option for assigning a datastore Key to a Model's
// Meta as that Model's parent.
//
//   p := &Post{Content: "Hello"}
//   Put(ctx, p, WithParent(u.Key))
func WithParent(k *datastore.Key) Option {
	return func(m Model) {
		meta := m.meta()
		meta.Parent = k
	}
}

// WithStringID is an Option for assigning a datastore Key with the
// given string id to a Model's Meta.
func WithStringID(ctx context.Context, id string) Option {
	return func(m Model) {
		meta := m.meta()
		meta.Key = datastore.NewKey(ctx, getKind(m), id, 0, nil)
	}
}

// WithIntID is an Option for assigning a datastore Key with the given
// int64 id to a Model's Meta.
func WithIntID(ctx context.Context, id int64) Option {
	return func(m Model) {
		meta := m.meta()
		meta.Key = datastore.NewKey(ctx, getKind(m), "", id, nil)
	}
}

// Query returns a new datastore query for the given kind that will
// ignore deleted entities.  Note that querying does not hydrate Meta
// Keys and Parents so you will have to do that manually.
func Query(kind string) *datastore.Query {
	return datastore.NewQuery(kind).
		Filter("State=", StateActive)
}

// Get retrieves a Model from datastore by its id.
func Get(ctx context.Context, m Model, options ...Option) error {
	meta := m.meta()
	for _, option := range options {
		option(m)
	}

	if meta.Key == nil {
		return ErrMissingKey
	}

	return nds.Get(ctx, meta.Key, m)
}

// Put stores a Model to datastore.
func Put(ctx context.Context, m Model, options ...Option) error {
	meta := m.meta()
	for _, option := range options {
		option(m)
	}

	kind := getKind(m)
	if meta.Key == nil {
		meta.Key = datastore.NewIncompleteKey(ctx, kind, meta.Parent)
	}

	if meta.State == "" {
		meta.State = StateActive
	}

	meta.UpdatedAt = time.Now()
	if meta.CreatedAt.IsZero() {
		meta.CreatedAt = time.Now()
	}

	key, err := nds.Put(ctx, meta.Key, m)
	if err != nil {
		return err
	}

	meta.Key = key
	return nil
}

// Delete soft deletes a Model from datastore.
func Delete(ctx context.Context, m Model) error {
	meta := m.meta()
	if meta.Key == nil {
		return ErrMissingKey
	}

	meta.State = StateDeleted
	meta.DeletedAt = time.Now()
	return Put(ctx, m)
}

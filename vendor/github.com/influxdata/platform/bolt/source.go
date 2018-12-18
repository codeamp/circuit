package bolt

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/coreos/bbolt"
	"github.com/influxdata/platform"
)

var (
	sourceBucket = []byte("sourcesv1")
)

// DefaultSource is the default source.
var DefaultSource = platform.Source{
	Default: true,
	Name:    "autogen",
	Type:    platform.SelfSourceType,
}

const (
	// DefaultSourceID it the default source identifier
	DefaultSourceID = "020f755c3c082000"
	// DefaultSourceOrganizationID is the default source's organization identifier
	DefaultSourceOrganizationID = "50616e67652c206c"
)

func init() {
	if err := DefaultSource.ID.DecodeFromString(DefaultSourceID); err != nil {
		panic(fmt.Sprintf("failed to decode default source id: %v", err))
	}

	if err := DefaultSource.OrganizationID.DecodeFromString(DefaultSourceOrganizationID); err != nil {
		panic(fmt.Sprintf("failed to decode default source organization id: %v", err))
	}
}

func (c *Client) initializeSources(ctx context.Context, tx *bolt.Tx) error {
	if _, err := tx.CreateBucketIfNotExists(sourceBucket); err != nil {
		return err
	}

	_, err := c.findSourceByID(ctx, tx, DefaultSource.ID)
	if err != nil && err != platform.ErrSourceNotFound {
		return err
	}

	if err == platform.ErrSourceNotFound {
		if err := c.putSource(ctx, tx, &DefaultSource); err != nil {
			return err
		}
	}

	return nil
}

// DefaultSource retrieves the default source.
func (c *Client) DefaultSource(ctx context.Context) (*platform.Source, error) {
	var s *platform.Source

	err := c.db.View(func(tx *bolt.Tx) error {
		// TODO(desa): make this faster by putting the default source in an index.
		srcs, err := c.findSources(ctx, tx, platform.FindOptions{})
		if err != nil {
			return err
		}
		for _, src := range srcs {
			if src.Default {
				s = src
				return nil
			}
		}
		return fmt.Errorf("no default source found")
	})

	if err != nil {
		return nil, err
	}

	return s, nil
}

// FindSourceByID retrieves a source by id.
func (c *Client) FindSourceByID(ctx context.Context, id platform.ID) (*platform.Source, error) {
	var s *platform.Source

	err := c.db.View(func(tx *bolt.Tx) error {
		src, err := c.findSourceByID(ctx, tx, id)
		if err != nil {
			return err
		}
		s = src
		return nil
	})

	if err != nil {
		return nil, err
	}

	return s, nil
}

func (c *Client) findSourceByID(ctx context.Context, tx *bolt.Tx, id platform.ID) (*platform.Source, error) {
	encodedID, err := id.Encode()
	if err != nil {
		return nil, err
	}

	v := tx.Bucket(sourceBucket).Get(encodedID)
	if len(v) == 0 {
		return nil, platform.ErrSourceNotFound
	}

	var s platform.Source
	if err := json.Unmarshal(v, &s); err != nil {
		return nil, err
	}

	return &s, nil
}

// FindSources retrives all sources that match an arbitrary source filter.
// Filters using ID, or OrganizationID and source Name should be efficient.
// Other filters will do a linear scan across all sources searching for a match.
func (c *Client) FindSources(ctx context.Context, opt platform.FindOptions) ([]*platform.Source, int, error) {
	ss := []*platform.Source{}
	err := c.db.View(func(tx *bolt.Tx) error {
		srcs, err := c.findSources(ctx, tx, opt)
		if err != nil {
			return err
		}
		ss = srcs
		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	return ss, len(ss), nil
}

func (c *Client) findSources(ctx context.Context, tx *bolt.Tx, opt platform.FindOptions) ([]*platform.Source, error) {
	ss := []*platform.Source{}

	err := c.forEachSource(ctx, tx, func(s *platform.Source) bool {
		ss = append(ss, s)
		return true
	})

	if err != nil {
		return nil, err
	}

	return ss, nil
}

// CreateSource creates a platform source and sets s.ID.
func (c *Client) CreateSource(ctx context.Context, s *platform.Source) error {
	return c.db.Update(func(tx *bolt.Tx) error {
		s.ID = c.IDGenerator.ID()

		// Generating an organization id if it missing or invalid
		if !s.OrganizationID.Valid() {
			s.OrganizationID = c.IDGenerator.ID()
		}

		return c.putSource(ctx, tx, s)
	})
}

// PutSource will put a source without setting an ID.
func (c *Client) PutSource(ctx context.Context, s *platform.Source) error {
	return c.db.Update(func(tx *bolt.Tx) error {
		return c.putSource(ctx, tx, s)
	})
}

func (c *Client) putSource(ctx context.Context, tx *bolt.Tx, s *platform.Source) error {
	v, err := json.Marshal(s)
	if err != nil {
		return err
	}

	encodedID, err := s.ID.Encode()
	if err != nil {
		return err
	}

	if err := tx.Bucket(sourceBucket).Put(encodedID, v); err != nil {
		return err
	}

	return nil
}

// forEachSource will iterate through all sources while fn returns true.
func (c *Client) forEachSource(ctx context.Context, tx *bolt.Tx, fn func(*platform.Source) bool) error {
	cur := tx.Bucket(sourceBucket).Cursor()
	for k, v := cur.First(); k != nil; k, v = cur.Next() {
		s := &platform.Source{}
		if err := json.Unmarshal(v, s); err != nil {
			return err
		}
		if !fn(s) {
			break
		}
	}

	return nil
}

// UpdateSource updates a source according the parameters set on upd.
func (c *Client) UpdateSource(ctx context.Context, id platform.ID, upd platform.SourceUpdate) (*platform.Source, error) {
	var s *platform.Source
	err := c.db.Update(func(tx *bolt.Tx) error {
		src, err := c.updateSource(ctx, tx, id, upd)
		if err != nil {
			return err
		}
		s = src
		return nil
	})

	return s, err
}

func (c *Client) updateSource(ctx context.Context, tx *bolt.Tx, id platform.ID, upd platform.SourceUpdate) (*platform.Source, error) {
	s, err := c.findSourceByID(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	if err := upd.Apply(s); err != nil {
		return nil, err
	}

	if err := c.putSource(ctx, tx, s); err != nil {
		return nil, err
	}

	return s, nil
}

// DeleteSource deletes a source and prunes it from the index.
func (c *Client) DeleteSource(ctx context.Context, id platform.ID) error {
	return c.db.Update(func(tx *bolt.Tx) error {
		return c.deleteSource(ctx, tx, id)
	})
}

func (c *Client) deleteSource(ctx context.Context, tx *bolt.Tx, id platform.ID) error {
	if id == DefaultSource.ID {
		return fmt.Errorf("cannot delete autogen source")
	}
	_, err := c.findSourceByID(ctx, tx, id)
	if err != nil {
		return err
	}

	encodedID, err := id.Encode()
	if err != nil {
		return err
	}

	return tx.Bucket(sourceBucket).Delete(encodedID)
}

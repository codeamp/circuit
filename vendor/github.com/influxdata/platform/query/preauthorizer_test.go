package query_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/platform"
	"github.com/influxdata/platform/kit/errors"
	"github.com/influxdata/platform/mock"
	"github.com/influxdata/platform/query"
	_ "github.com/influxdata/platform/query/builtin"
)

func newBucketServiceWithOneBucket(bucket platform.Bucket) platform.BucketService {
	bs := mock.NewBucketService()
	bs.FindBucketFn = func(ctx context.Context, bucketFilter platform.BucketFilter) (*platform.Bucket, error) {
		if *bucketFilter.Name == bucket.Name {
			return &bucket, nil
		}

		return nil, errors.New("Unknown bucket")
	}

	return bs
}

func TestPreAuthorizer_PreAuthorize(t *testing.T) {
	// TODO(adam) add this test back when BucketsAccessed is restored for the from function
	// https://github.com/influxdata/flux/issues/114
	t.Skip("https://github.com/influxdata/flux/issues/114")
	ctx := context.Background()
	now := time.Now().UTC()

	q := `from(bucket:"my_bucket") |> range(start:-2h) |> yield()`
	spec, err := flux.Compile(ctx, q, now)
	if err != nil {
		t.Fatalf("Error compiling query: %v", err)
	}

	// Try to pre-authorize with bucket service with no buckets
	// and no authorization
	auth := &platform.Authorization{Status: platform.Active}
	emptyBucketService := mock.NewBucketService()
	preAuthorizer := query.NewPreAuthorizer(emptyBucketService)

	err = preAuthorizer.PreAuthorize(ctx, spec, auth)
	if diagnostic := cmp.Diff("Bucket service returned nil bucket", err.Error()); diagnostic != "" {
		t.Errorf("Authorize message mismatch: -want/+got:\n%v", diagnostic)
	}

	// Try to authorize with a bucket service that knows about one bucket
	// (still no authorization)
	id, _ := platform.IDFromString("deadbeefdeadbeef")
	bucketService := newBucketServiceWithOneBucket(platform.Bucket{
		Name: "my_bucket",
		ID:   *id,
	})

	preAuthorizer = query.NewPreAuthorizer(bucketService)
	err = preAuthorizer.PreAuthorize(ctx, spec, auth)
	if diagnostic := cmp.Diff(`No read permission for bucket: "my_bucket"`, err.Error()); diagnostic != "" {
		t.Errorf("Authorize message mismatch: -want/+got:\n%v", diagnostic)
	}

	// Try to authorize with read permission on bucket
	auth = &platform.Authorization{
		Status:      platform.Active,
		Permissions: []platform.Permission{platform.ReadBucketPermission(*id)},
	}

	err = preAuthorizer.PreAuthorize(ctx, spec, auth)
	if err != nil {
		t.Errorf("Expected successful authorization, but got error: \"%v\"", err.Error())
	}
}

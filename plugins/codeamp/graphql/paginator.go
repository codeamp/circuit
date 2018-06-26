package graphql_resolver

import (
	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
)

// FeatureListResolver resolver for Feature
type FeatureListResolver struct {
	DBFeatureListResolver *db_resolver.FeatureListResolver
}

func (r FeatureListResolver) Entries() ([]*FeatureResolver, error) {
	features, err := r.DBFeatureListResolver.Entries()
	if err != nil {
		return []*FeatureResolver{}, err
	}

	results := make([]*FeatureResolver, len(features))
	for _, feature := range features {
		results = append(results, &FeatureResolver{
			DBFeatureResolver: feature,
		})
	}

	return results, nil
}

func (r FeatureListResolver) Page() (int32, error) {
	return r.DBFeatureListResolver.Page()
}

func (r FeatureListResolver) NextCursor() (string, error) {
	return r.DBFeatureListResolver.NextCursor()
}

func (r FeatureListResolver) Count() (int32, error) {
	return r.DBFeatureListResolver.Count()
}

// ReleaseListResolver
type ReleaseListResolver struct {
	DBReleaseListResolver *db_resolver.ReleaseListResolver
}

func (r ReleaseListResolver) Entries() ([]*ReleaseResolver, error) {
	releases, err := r.DBReleaseListResolver.Entries()
	if err != nil {
		return []*ReleaseResolver{}, err
	}

	results := make([]*ReleaseResolver, len(releases))
	for _, release := range releases {
		results = append(results, &ReleaseResolver{
			DBReleaseResolver: release,
		})
	}

	return results, nil
}

func (r ReleaseListResolver) Page() (int32, error) {
	return r.DBReleaseListResolver.Page()
}

func (r ReleaseListResolver) NextCursor() (string, error) {
	return r.DBReleaseListResolver.NextCursor()
}

func (r ReleaseListResolver) Count() (int32, error) {
	return r.DBReleaseListResolver.Count()
}

// ServiceListResolver
type ServiceListResolver struct {
	DBServiceListResolver *db_resolver.ServiceListResolver
}

func (r ServiceListResolver) Entries() ([]*ServiceResolver, error) {
	features, err := r.DBServiceListResolver.Entries()
	if err != nil {
		return []*ServiceResolver{}, err
	}

	results := make([]*ServiceResolver, len(features))
	for _, feature := range features {
		results = append(results, &ServiceResolver{
			DBServiceResolver: feature,
		})
	}

	return results, nil
}

func (r ServiceListResolver) Page() (int32, error) {
	return r.DBServiceListResolver.Page()
}

func (r ServiceListResolver) NextCursor() (string, error) {
	return r.DBServiceListResolver.NextCursor()
}

func (r ServiceListResolver) Count() (int32, error) {
	return r.DBServiceListResolver.Count()
}

// SecretListResolver
type SecretListResolver struct {
	DBSecretListResolver *db_resolver.SecretListResolver
}

func (r SecretListResolver) Entries() ([]*SecretResolver, error) {
	features, err := r.DBSecretListResolver.Entries()
	if err != nil {
		return []*SecretResolver{}, err
	}

	results := make([]*SecretResolver, len(features))
	for _, feature := range features {
		results = append(results, &SecretResolver{
			DBSecretResolver: feature,
		})
	}

	return results, nil
}

func (r SecretListResolver) Page() (int32, error) {
	return r.DBSecretListResolver.Page()
}

func (r SecretListResolver) NextCursor() (string, error) {
	return r.DBSecretListResolver.NextCursor()
}

func (r SecretListResolver) Count() (int32, error) {
	return r.DBSecretListResolver.Count()
}

// ProjectListResolver
type ProjectListResolver struct {
	DBProjectListResolver *db_resolver.ProjectListResolver
}

func (r ProjectListResolver) Entries() ([]*ProjectResolver, error) {
	features, err := r.DBProjectListResolver.Entries()
	if err != nil {
		return []*ProjectResolver{}, err
	}

	results := make([]*ProjectResolver, len(features))
	for _, feature := range features {
		results = append(results, &ProjectResolver{
			DBProjectResolver: feature,
		})
	}

	return results, nil
}

func (r ProjectListResolver) Page() (int32, error) {
	return r.DBProjectListResolver.Page()
}

func (r ProjectListResolver) NextCursor() (string, error) {
	return r.DBProjectListResolver.NextCursor()
}

func (r ProjectListResolver) Count() (int32, error) {
	return r.DBProjectListResolver.Count()
}

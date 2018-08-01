package graphql_resolver

import (
	"encoding/json"

	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	graphql "github.com/graph-gophers/graphql-go"
)

// ReleaseExtensionResolver resolver for ReleaseExtension
type ReleaseExtensionResolver struct {
	DBReleaseExtensionResolver *db_resolver.ReleaseExtensionResolver
}

// ID
func (r *ReleaseExtensionResolver) ID() graphql.ID {
	return graphql.ID(r.DBReleaseExtensionResolver.ReleaseExtension.Model.ID.String())
}

// Release
func (r *ReleaseExtensionResolver) Release() (*ReleaseResolver, error) {
	resolver, err := r.DBReleaseExtensionResolver.Release()
	return &ReleaseResolver{DBReleaseResolver: resolver}, err
}

// ProjectExtension
func (r *ReleaseExtensionResolver) Extension() (*ProjectExtensionResolver, error) {
	resolver, err := r.DBReleaseExtensionResolver.Extension()
	return &ProjectExtensionResolver{DBProjectExtensionResolver: resolver}, err
}

// ServicesSignature
func (r *ReleaseExtensionResolver) ServicesSignature() string {
	return r.DBReleaseExtensionResolver.ReleaseExtension.ServicesSignature
}

// SecretsSignature
func (r *ReleaseExtensionResolver) SecretsSignature() string {
	return r.DBReleaseExtensionResolver.ReleaseExtension.SecretsSignature
}

// State
func (r *ReleaseExtensionResolver) State() string {
	return string(r.DBReleaseExtensionResolver.ReleaseExtension.State)
}

// Type
func (r *ReleaseExtensionResolver) Type() string {
	return string(r.DBReleaseExtensionResolver.ReleaseExtension.Type)
}

// StateMessage
func (r *ReleaseExtensionResolver) StateMessage() string {
	return r.DBReleaseExtensionResolver.ReleaseExtension.StateMessage
}

// Artifacts
func (r *ReleaseExtensionResolver) Artifacts() model.JSON {
	return model.JSON{r.DBReleaseExtensionResolver.ReleaseExtension.Artifacts.RawMessage}
}

// Started
func (r *ReleaseExtensionResolver) Started() graphql.Time {
	return graphql.Time{Time: r.DBReleaseExtensionResolver.ReleaseExtension.Started}
}

// Finished
func (r *ReleaseExtensionResolver) Finished() graphql.Time {
	return graphql.Time{Time: r.DBReleaseExtensionResolver.ReleaseExtension.Finished}
}

// Created
func (r *ReleaseExtensionResolver) Created() graphql.Time {
	return graphql.Time{Time: r.DBReleaseExtensionResolver.ReleaseExtension.Model.CreatedAt}
}

func (r *ReleaseExtensionResolver) MarshalJSON() ([]byte, error) {
	return json.Marshal(&r.DBReleaseExtensionResolver.ReleaseExtension)
}

func (r *ReleaseExtensionResolver) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.DBReleaseExtensionResolver.ReleaseExtension)
}

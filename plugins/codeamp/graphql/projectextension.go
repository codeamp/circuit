package graphql_resolver

import (
	"encoding/json"

	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	graphql "github.com/graph-gophers/graphql-go"
)

// ProjectExtensionResolver resolver for ProjectExtension
type ProjectExtensionResolver struct {
	DBProjectExtensionResolver *db_resolver.ProjectExtensionResolver
}

// ID
func (r *ProjectExtensionResolver) ID() graphql.ID {
	return graphql.ID(r.DBProjectExtensionResolver.ProjectExtension.Model.ID.String())
}

// Project
func (r *ProjectExtensionResolver) Project() *ProjectResolver {
	return &ProjectResolver{DBProjectResolver: r.DBProjectExtensionResolver.Project()}
}

// Extension
func (r *ProjectExtensionResolver) Extension() *ExtensionResolver {
	return &ExtensionResolver{DBExtensionResolver: r.DBProjectExtensionResolver.Extension()}
}

// Artifacts
func (r *ProjectExtensionResolver) Artifacts() model.JSON {
	return model.JSON{r.DBProjectExtensionResolver.ProjectExtension.Artifacts.RawMessage}
}

// Config
func (r *ProjectExtensionResolver) Config() model.JSON {
	return model.JSON{r.DBProjectExtensionResolver.ProjectExtension.Config.RawMessage}
}

// CustomConfig
func (r *ProjectExtensionResolver) CustomConfig() model.JSON {
	return model.JSON{r.DBProjectExtensionResolver.ProjectExtension.CustomConfig.RawMessage}
}

// State
func (r *ProjectExtensionResolver) State() string {
	return string(r.DBProjectExtensionResolver.ProjectExtension.State)
}

// StateMessage
func (r *ProjectExtensionResolver) StateMessage() string {
	return r.DBProjectExtensionResolver.ProjectExtension.StateMessage
}

// Environment
func (r *ProjectExtensionResolver) Environment() (*EnvironmentResolver, error) {
	resolver, err := r.DBProjectExtensionResolver.Environment()
	return &EnvironmentResolver{DBEnvironmentResolver: resolver}, err
}

// Created
func (r *ProjectExtensionResolver) Created() graphql.Time {
	return graphql.Time{Time: r.DBProjectExtensionResolver.ProjectExtension.Model.CreatedAt}
}

func (r *ProjectExtensionResolver) MarshalJSON() ([]byte, error) {
	return json.Marshal(&r.DBProjectExtensionResolver.ProjectExtension)
}

func (r *ProjectExtensionResolver) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.DBProjectExtensionResolver.ProjectExtension)
}

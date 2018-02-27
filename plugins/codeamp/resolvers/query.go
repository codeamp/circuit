package codeamp_resolvers

import (
	graphql "github.com/neelance/graphql-go"
)

// User User
func (r *Resolver) User(args *struct {
	ID *graphql.ID
}) *UserResolver {
	return nil
}

// Users Users
func (r *Resolver) Users() []*UserResolver {
	return nil
}

// Project Project
func (r *Resolver) Project(args *struct {
	ID            *graphql.ID
	Slug          *string
	Name          *string
	EnvironmentId *string
}) *ProjectResolver {
	return nil
}

// Projects Projects
func (r *Resolver) Projects() []*ProjectResolver {
	return nil
}

// Features Features
func (r *Resolver) Features() []*FeatureResolver {
	return nil
}

// Services Services
func (r *Resolver) Services() []*ServiceResolver {
	return nil
}

// ServiceSpecs Service specs
func (r *Resolver) ServiceSpecs() []*ServiceSpecResolver {
	return nil
}

// Releases Releases
func (r *Resolver) Releases() []*ReleaseResolver {
	return nil
}

// Environments Environments
func (r *Resolver) Environments() []*EnvironmentResolver {
	return nil
}

// EnvironmentVariables Environment variables
func (r *Resolver) EnvironmentVariables() []*EnvironmentVariableResolver {
	return nil
}

// ExtensionSpecs Extension spec
func (r *Resolver) ExtensionSpecs() []*ExtensionSpecResolver {
	return nil
}

// Extensions Extensions
func (r *Resolver) Extensions() []*ExtensionResolver {
	return nil
}

// ReleaseExtensions Release extensions
func (r *Resolver) ReleaseExtensions() []*ReleaseExtensionResolver {
	return nil
}

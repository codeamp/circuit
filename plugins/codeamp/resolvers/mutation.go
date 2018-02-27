package codeamp_resolvers

import (
	graphql "github.com/neelance/graphql-go"
)

// CreateProject Create project
func (r *Resolver) CreateProject(args *struct {
	Project *ProjectInput
}) *ProjectResolver {
	return nil
}

// UpdateProject Update project
func (r *Resolver) UpdateProject(args *struct {
	Project *ProjectInput
}) *ProjectResolver {
	return nil
}

// CreateRelease Create release
func (r *Resolver) CreateRelease(args *struct {
	Release *ReleaseInput
}) *ReleaseResolver {
	return nil
}

// RollbackRelease Rollback release
func (r *Resolver) RollbackRelease(args *struct {
	ReleaseId graphql.ID
}) *ReleaseResolver {
	return nil
}

// CreateService Create service
func (r *Resolver) CreateService(args *struct {
	Service *ServiceInput
}) *ServiceResolver {
	return nil
}

// UpdateService Update Service
func (r *Resolver) UpdateService(args *struct {
	Service *ServiceInput
}) *ServiceResolver {
	return nil
}

// DeleteService Delete service
func (r *Resolver) DeleteService(args *struct {
	Service *ServiceInput
}) *ServiceResolver {
	return nil
}

// CreateServiceSpec Create service spec
func (r *Resolver) CreateServiceSpec(args *struct {
	ServiceSpec *ServiceSpecInput
}) *ServiceSpecResolver {
	return nil
}

// UpdateServiceSpec Update service spec
func (r *Resolver) UpdateServiceSpec(args *struct {
	ServiceSpec *ServiceSpecInput
}) *ServiceSpecResolver {
	return nil
}

// DeleteServiceSpec Delete service spec
func (r *Resolver) DeleteServiceSpec(args *struct {
	ServiceSpec *ServiceSpecInput
}) *ServiceSpecResolver {
	return nil
}

// CreateEnvironment Create environment
func (r *Resolver) CreateEnvironment(args *struct {
	Environment *EnvironmentInput
}) *EnvironmentResolver {
	return nil
}

// UpdateEnvironment Update environment
func (r *Resolver) UpdateEnvironment(args *struct {
	Environment *EnvironmentInput
}) *EnvironmentResolver {
	return nil
}

// DeleteEnvironment Delete environment
func (r *Resolver) DeleteEnvironment(args *struct {
	Environment *EnvironmentInput
}) *EnvironmentResolver {
	return nil
}

// CreateEnvironmentVariable Create environment variable
func (r *Resolver) CreateEnvironmentVariable(args *struct {
	EnvironmentVariable *EnvironmentVariableInput
}) *EnvironmentVariableResolver {
	return nil
}

// UpdateEnvironmentVariable Update environment variable
func (r *Resolver) UpdateEnvironmentVariable(args *struct {
	EnvironmentVariable *EnvironmentVariableInput
}) *EnvironmentVariableResolver {
	return nil
}

// DeleteEnvironmentVariable Delete environment variable
func (r *Resolver) DeleteEnvironmentVariable(args *struct {
	EnvironmentVariable *EnvironmentVariableInput
}) *EnvironmentVariableResolver {
	return nil
}

// CreateExtensionSpec Create extension spec
func (r *Resolver) CreateExtensionSpec(args *struct {
	ExtensionSpec *ExtensionSpecInput
}) *ExtensionSpecResolver {
	return nil
}

// UpdateExtensionSpec Update extension spec
func (r *Resolver) UpdateExtensionSpec(args *struct {
	ExtensionSpec *ExtensionSpecInput
}) *ExtensionSpecResolver {
	return nil
}

// DeleteExtensionSpec Delete extension spec
func (r *Resolver) DeleteExtensionSpec(args *struct {
	ExtensionSpec *ExtensionSpecInput
}) *ExtensionSpecResolver {
	return nil
}

// CreateExtension Create extension
func (r *Resolver) CreateExtension(args *struct {
	Extension *ExtensionInput
}) *ExtensionResolver {
	return nil
}

// UpdateExtension Update extension
func (r *Resolver) UpdateExtension(args *struct {
	Extension *ExtensionInput
}) *ExtensionResolver {
	return nil
}

// DeleteExtension Delete extesion
func (r *Resolver) DeleteExtension(args *struct {
	Extension *ExtensionInput
}) *ExtensionResolver {
	return nil
}

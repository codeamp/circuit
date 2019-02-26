package graphql_resolver

import (
	"context"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/transistor"
	graphql "github.com/graph-gophers/graphql-go"
)

// CreateProject Create project
func (r *Resolver) CreateProject(ctx context.Context, args *struct {
	Project *model.ProjectInput
}) (*ProjectResolver, error) {
	mut := ProjectResolverMutation{r.DB}
	return mut.CreateProject(ctx, args)
}

// UpdateProject Update project
func (r *Resolver) UpdateProject(ctx context.Context, args *struct {
	Project *model.ProjectInput
}) (*ProjectResolver, error) {
	mut := ProjectResolverMutation{r.DB}
	return mut.UpdateProject(ctx, args)
}

// StopRelease
func (r *Resolver) StopRelease(ctx context.Context, args *struct{ ID graphql.ID }) (*ReleaseResolver, error) {
	mut := ReleaseResolverMutation{r.DB, r.Events}
	return mut.StopRelease(ctx, args)
}

// CreateRelease
func (r *Resolver) CreateRelease(ctx context.Context, args *struct{ Release *model.ReleaseInput }) (*ReleaseResolver, error) {
	mut := ReleaseResolverMutation{r.DB, r.Events}
	return mut.CreateRelease(ctx, args)
}

// CreateService Create service
func (r *Resolver) CreateService(args *struct{ Service *model.ServiceInput }) (*ServiceResolver, error) {
	mut := ServiceResolverMutation{r.DB}
	return mut.CreateService(args)
}

// UpdateService Update Service
func (r *Resolver) UpdateService(args *struct{ Service *model.ServiceInput }) (*ServiceResolver, error) {
	mut := ServiceResolverMutation{r.DB}
	return mut.UpdateService(args)
}

// DeleteService Delete service
func (r *Resolver) DeleteService(args *struct{ Service *model.ServiceInput }) (*ServiceResolver, error) {
	mut := ServiceResolverMutation{r.DB}
	return mut.DeleteService(args)
}

// ImportServices Import services
func (r *Resolver) ImportServices(args *struct{ Services *model.ImportServicesInput }) ([]*ServiceResolver, error) {
	mut := ServiceResolverMutation{r.DB}
	return mut.ImportServices(args)
}

func (r *Resolver) CreateServiceSpec(args *struct{ ServiceSpec *model.ServiceSpecInput }) (*ServiceSpecResolver, error) {
	mut := ServiceSpecResolverMutation{r.DB}
	return mut.CreateServiceSpec(args)
}

func (r *Resolver) UpdateServiceSpec(args *struct{ ServiceSpec *model.ServiceSpecInput }) (*ServiceSpecResolver, error) {
	mut := ServiceSpecResolverMutation{r.DB}
	return mut.UpdateServiceSpec(args)
}

func (r *Resolver) DeleteServiceSpec(args *struct{ ServiceSpec *model.ServiceSpecInput }) (*ServiceSpecResolver, error) {
	mut := ServiceSpecResolverMutation{r.DB}
	return mut.DeleteServiceSpec(args)
}

func (r *Resolver) CreateEnvironment(ctx context.Context, args *struct{ Environment *model.EnvironmentInput }) (*EnvironmentResolver, error) {
	mut := EnvironmentResolverMutation{r.DB}
	return mut.CreateEnvironment(ctx, args)
}

func (r *Resolver) UpdateEnvironment(ctx context.Context, args *struct{ Environment *model.EnvironmentInput }) (*EnvironmentResolver, error) {
	mut := EnvironmentResolverMutation{r.DB}
	return mut.UpdateEnvironment(ctx, args)
}

func (r *Resolver) DeleteEnvironment(ctx context.Context, args *struct{ Environment *model.EnvironmentInput }) (*EnvironmentResolver, error) {
	mut := EnvironmentResolverMutation{r.DB}
	return mut.DeleteEnvironment(ctx, args)
}

func (r *Resolver) ImportSecrets(ctx context.Context, args *struct{ Secrets *model.ImportSecretsInput }) ([]*SecretResolver, error) {
	mut := SecretResolverMutation{r.DB}
	return mut.ImportSecrets(ctx, args)
}

func (r *Resolver) CreateSecret(ctx context.Context, args *struct{ Secret *model.SecretInput }) (*SecretResolver, error) {
	mut := SecretResolverMutation{r.DB}
	return mut.CreateSecret(ctx, args)
}

func (r *Resolver) UpdateSecret(ctx context.Context, args *struct{ Secret *model.SecretInput }) (*SecretResolver, error) {
	mut := SecretResolverMutation{r.DB}
	return mut.UpdateSecret(ctx, args)
}

func (r *Resolver) DeleteSecret(ctx context.Context, args *struct{ Secret *model.SecretInput }) (*SecretResolver, error) {
	mut := SecretResolverMutation{r.DB}
	return mut.DeleteSecret(ctx, args)
}

func (r *Resolver) CreateExtension(args *struct{ Extension *model.ExtensionInput }) (*ExtensionResolver, error) {
	mut := ExtensionResolverMutation{r.DB}
	return mut.CreateExtension(args)
}

func (r *Resolver) UpdateExtension(args *struct{ Extension *model.ExtensionInput }) (*ExtensionResolver, error) {
	mut := ExtensionResolverMutation{r.DB}
	return mut.UpdateExtension(args)
}

func (r *Resolver) DeleteExtension(args *struct{ Extension *model.ExtensionInput }) (*ExtensionResolver, error) {
	mut := ExtensionResolverMutation{r.DB}
	return mut.DeleteExtension(args)
}

func (r *Resolver) CreateProjectExtension(ctx context.Context, args *struct{ ProjectExtension *model.ProjectExtensionInput }) (*ProjectExtensionResolver, error) {
	mut := ProjectExtensionResolverMutation{r.DB, r.Events}
	return mut.CreateProjectExtension(ctx, args)
}

func (r *Resolver) UpdateProjectExtension(args *struct{ ProjectExtension *model.ProjectExtensionInput }) (*ProjectExtensionResolver, error) {
	mut := ProjectExtensionResolverMutation{r.DB, r.Events}
	return mut.UpdateProjectExtension(args)
}

func (r *Resolver) DeleteProjectExtension(args *struct{ ProjectExtension *model.ProjectExtensionInput }) (*ProjectExtensionResolver, error) {
	mut := ProjectExtensionResolverMutation{r.DB, r.Events}
	return mut.DeleteProjectExtension(args)
}

// UpdateUserPermissions
func (r *Resolver) UpdateUserPermissions(ctx context.Context, args *struct{ UserPermissions *model.UserPermissionsInput }) ([]string, error) {
	mut := UserResolverMutation{r.DB}
	return mut.UpdateUserPermissions(ctx, args)
}

// UpdateProjectEnvironments
func (r *Resolver) UpdateProjectEnvironments(ctx context.Context, args *struct {
	ProjectEnvironments *model.ProjectEnvironmentsInput
}) ([]*EnvironmentResolver, error) {
	mut := ProjectResolverMutation{r.DB}
	return mut.UpdateProjectEnvironments(ctx, args)
}

// GetGitCommits
func (r *Resolver) GetGitCommits(ctx context.Context, args *struct {
	ProjectID     graphql.ID
	EnvironmentID graphql.ID
	New           *bool
}) (bool, error) {
	if args.New != nil && *args.New {
		var err error
		project := model.Project{}
		env := model.Environment{}
		projectSettings := model.ProjectSettings{}
		latestFeature := model.Feature{}
		hash := ""

		if err = r.DB.Where("id = ?", args.ProjectID).First(&project).Error; err != nil {
			return false, err
		}

		if err = r.DB.Where("id = ?", args.EnvironmentID).First(&env).Error; err != nil {
			return false, err
		}

		if err = r.DB.Where("project_id = ? AND environment_id = ?", project.Model.ID, env.Model.ID).First(&projectSettings).Error; err != nil {
			return false, err
		}

		if err = r.DB.Where("project_id = ?", project.Model.ID).Order("created_at DESC").First(&latestFeature).Error; err == nil {
			hash = latestFeature.Hash
		}

		payload := plugins.GitSync{
			Project: plugins.Project{
				ID:         project.Model.ID.String(),
				Repository: project.Repository,
			},
			Git: plugins.Git{
				Url:           project.GitUrl,
				Protocol:      project.GitProtocol,
				Branch:        projectSettings.GitBranch,
				RsaPrivateKey: project.RsaPrivateKey,
				RsaPublicKey:  project.RsaPublicKey,
			},
			From: hash,
		}

		r.Events <- transistor.NewEvent(plugins.GetEventName("gitsync"), transistor.GetAction("create"), payload)
		return true, nil
	}
	return true, nil
}

func (r *Resolver) BookmarkProject(ctx context.Context, args *struct{ ID graphql.ID }) (bool, error) {
	mut := ProjectResolverMutation{r.DB}
	return mut.BookmarkProject(ctx, args)
}

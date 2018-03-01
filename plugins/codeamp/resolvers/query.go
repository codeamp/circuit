package codeamp_resolvers

import (
	"context"
	"encoding/json"
	"fmt"

	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
)

// User
func (r *Resolver) User(ctx context.Context, args *struct {
	ID *graphql.ID
}) (*UserResolver, error) {
	var userID string
	var err error
	var user User

	if args.ID != nil {
		userID = string(*args.ID)
	} else {
		claims := ctx.Value("jwt").(Claims)
		userID = claims.UserID
	}

	if _, err = CheckAuth(ctx, []string{fmt.Sprintf("user/%s", userID)}); err != nil {
		return nil, err
	}

	if err = r.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}

	return &UserResolver{DB: r.DB, User: user}, nil
}

// Users
func (r *Resolver) Users(ctx context.Context) ([]*UserResolver, error) {
	if _, err := CheckAuth(ctx, []string{"admin"}); err != nil {
		return nil, err
	}

	var rows []User
	var results []*UserResolver

	r.DB.Order("created_at desc").Find(&rows)

	for _, user := range rows {
		results = append(results, &UserResolver{DB: r.DB, User: user})
	}

	return results, nil
}

// Project
func (r *Resolver) Project(ctx context.Context, args *struct {
	ID            *graphql.ID
	Slug          *string
	Name          *string
	EnvironmentID *string
}) (*ProjectResolver, error) {
	if _, err := CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var project Project
	var environment Environment
	var query *gorm.DB

	// get environment
	if r.DB.Where("id = ?", *args.EnvironmentID).Find(&environment).RecordNotFound() {
		log.InfoWithFields("Environment doesn't exist.", log.Fields{
			"args": args,
		})
		return nil, fmt.Errorf("Environment doesn't exist.")
	}

	if args.ID != nil {
		query = r.DB.Where("id = ?", *args.ID)
	} else if args.Slug != nil {
		query = r.DB.Where("slug = ?", *args.Slug)
	} else if args.Name != nil {
		query = r.DB.Where("name = ?", *args.Name)
	} else {
		return nil, fmt.Errorf("Missing argument id or slug")
	}

	if err := query.First(&project).Error; err != nil {
		return nil, err
	}

	return &ProjectResolver{DB: r.DB, Project: project, Environment: environment}, nil
}

// Projects Projects
func (r *Resolver) Projects(ctx context.Context) ([]*ProjectResolver, error) {
	if _, err := CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []Project
	var results []*ProjectResolver

	r.DB.Find(&rows)
	for _, project := range rows {
		results = append(results, &ProjectResolver{DB: r.DB, Project: project})
	}

	return results, nil
}

func (r *Resolver) Features(ctx context.Context) ([]*FeatureResolver, error) {
	if _, err := CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []Feature
	var results []*FeatureResolver

	r.DB.Order("created_at desc").Find(&rows)
	for _, feature := range rows {
		results = append(results, &FeatureResolver{DB: r.DB, Feature: feature})
	}

	return results, nil
}

func (r *Resolver) Services(ctx context.Context) ([]*ServiceResolver, error) {
	if _, err := CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []Service
	var results []*ServiceResolver

	r.DB.Order("created_at desc").Find(&rows)
	for _, service := range rows {
		results = append(results, &ServiceResolver{DB: r.DB, Service: service})
	}

	return results, nil
}

func (r *Resolver) ServiceSpecs(ctx context.Context) ([]*ServiceSpecResolver, error) {
	if _, err := CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []ServiceSpec
	var results []*ServiceSpecResolver

	r.DB.Order("created_at desc").Find(&rows)
	for _, serviceSpec := range rows {
		results = append(results, &ServiceSpecResolver{DB: r.DB, ServiceSpec: serviceSpec})
	}

	return results, nil
}

func (r *Resolver) Releases(ctx context.Context) ([]*ReleaseResolver, error) {
	if _, err := CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []Release
	var results []*ReleaseResolver

	r.DB.Order("created_at desc").Find(&rows)
	for _, release := range rows {
		results = append(results, &ReleaseResolver{DB: r.DB, Release: release})
	}

	return results, nil
}

func (r *Resolver) Environments(ctx context.Context) ([]*EnvironmentResolver, error) {
	if _, err := CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []Environment
	var results []*EnvironmentResolver

	r.DB.Order("created_at desc").Find(&rows)
	for _, env := range rows {
		results = append(results, &EnvironmentResolver{DB: r.DB, Environment: env})
	}

	return results, nil
}

func (r *Resolver) Secrets(ctx context.Context) ([]*SecretResolver, error) {
	if _, err := CheckAuth(ctx, []string{"admin"}); err != nil {
		return nil, err
	}

	var rows []Secret
	var results []*SecretResolver

	r.DB.Where("scope != ?", "project").Order("created_at desc").Find(&rows)
	for _, secret := range rows {
		var secretValue SecretValue
		r.DB.Where("secret_id = ?", secret.Model.ID).Order("created_at desc").First(&secretValue)
		results = append(results, &SecretResolver{DB: r.DB, Secret: secret, SecretValue: secretValue})
	}

	return results, nil
}

func (r *Resolver) ExtensionSpecs(ctx context.Context) ([]*ExtensionSpecResolver, error) {
	if _, err := CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []ExtensionSpec
	var results []*ExtensionSpecResolver

	r.DB.Order("created_at desc").Find(&rows)
	for _, extensionSpec := range rows {
		results = append(results, &ExtensionSpecResolver{DB: r.DB, ExtensionSpec: extensionSpec})
	}

	return results, nil
}

func (r *Resolver) Extensions(ctx context.Context) ([]*ExtensionResolver, error) {
	if _, err := CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []Extension
	var results []*ExtensionResolver

	r.DB.Order("created_at desc").Find(&rows)
	for _, extension := range rows {
		results = append(results, &ExtensionResolver{DB: r.DB, Extension: extension})
	}

	return results, nil
}

func (r *Resolver) ReleaseExtensions(ctx context.Context) ([]*ReleaseExtensionResolver, error) {
	if _, err := CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []ReleaseExtension
	var results []*ReleaseExtensionResolver

	r.DB.Order("created_at desc").Find(&rows)
	for _, releaseExtension := range rows {
		results = append(results, &ReleaseExtensionResolver{DB: r.DB, ReleaseExtension: releaseExtension})
	}

	return results, nil
}

// Permissions
func (r *Resolver) Permissions(ctx context.Context) (JSON, error) {
	var rows []UserPermission
	var results = make(map[string]bool)

	r.DB.Unscoped().Select("DISTINCT(value)").Find(&rows)

	for _, userPermission := range rows {
		if _, err := CheckAuth(ctx, []string{userPermission.Value}); err != nil {
			results[userPermission.Value] = false
		} else {
			results[userPermission.Value] = true
		}
	}

	bytes, err := json.Marshal(results)
	if err != nil {
		return JSON{}, err
	}

	return JSON{bytes}, nil
}

package resolvers

import (
	"context"
	"fmt"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/codeamp/circuit/plugins/codeamp/schema/scalar"
	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
)

type ReleaseExtensionResolver struct {
	db               *gorm.DB
	ReleaseExtension models.ReleaseExtension
}

func (r *Resolver) ReleaseExtension(ctx context.Context, args *struct{ ID graphql.ID }) (*ReleaseExtensionResolver, error) {
	releaseExtension := models.ReleaseExtension{}
	if err := r.db.Where("id = ?", args.ID).First(&releaseExtension).Error; err != nil {
		return nil, err
	}

	return &ReleaseExtensionResolver{db: r.db, ReleaseExtension: releaseExtension}, nil
}

func (r *ReleaseExtensionResolver) ID() graphql.ID {
	return graphql.ID(r.ReleaseExtension.Model.ID.String())
}

func (r *ReleaseExtensionResolver) Release(ctx context.Context) (*ReleaseResolver, error) {
	release := models.Release{}
	if r.db.Where("id = ?", r.ReleaseExtension.ReleaseId.String()).Find(&release).RecordNotFound() {
		log.InfoWithFields("extension", log.Fields{
			"release extension": r.ReleaseExtension,
		})
		return &ReleaseResolver{db: r.db, Release: release}, fmt.Errorf("Couldn't find release")
	}
	return &ReleaseResolver{db: r.db, Release: release}, nil
}

func (r *ReleaseExtensionResolver) FeatureHash() string {
	return r.ReleaseExtension.FeatureHash
}

func (r *ReleaseExtensionResolver) ServicesSignature() string {
	return r.ReleaseExtension.ServicesSignature
}

func (r *ReleaseExtensionResolver) SecretsSignature() string {
	return r.ReleaseExtension.SecretsSignature
}

func (r *ReleaseExtensionResolver) Extension(ctx context.Context) (*ExtensionResolver, error) {
	extension := models.Extension{}
	if r.db.Where("id = ?", r.ReleaseExtension.ExtensionId).Find(&extension).RecordNotFound() {
		log.InfoWithFields("extension", log.Fields{
			"release extension": r.ReleaseExtension,
		})
		return &ExtensionResolver{db: r.db, Extension: extension}, fmt.Errorf("Couldn't find extension")
	}

	return &ExtensionResolver{db: r.db, Extension: extension}, nil
}

func (r *ReleaseExtensionResolver) State() string {
	return string(r.ReleaseExtension.State)
}

func (r *ReleaseExtensionResolver) Type() string {
	return string(r.ReleaseExtension.Type)
}

func (r *ReleaseExtensionResolver) StateMessage() string {
	return r.ReleaseExtension.StateMessage
}

func (r *ReleaseExtensionResolver) Artifacts() scalar.Json {
	return scalar.Json{r.ReleaseExtension.Artifacts.RawMessage}
}

func (r *ReleaseExtensionResolver) Finished() *graphql.Time {
	return &graphql.Time{Time: r.ReleaseExtension.Finished}
}

func (r *ReleaseExtensionResolver) Created() graphql.Time {
	return graphql.Time{Time: r.ReleaseExtension.Model.CreatedAt}
}

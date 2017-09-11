package codeamp_schema_resolvers

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"

	"github.com/codeamp/circuit/plugins"
	codeamp_models "github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/codeamp/transistor"
	"github.com/davecgh/go-spew/spew"
	"github.com/extemporalgenome/slug"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
	"golang.org/x/crypto/ssh"
)

type ProjectInput struct {
	GitProtocol string
	GitUrl      string
	Bookmarked  bool
}

func (r *Resolver) Project(ctx context.Context, args *struct {
	ID   *graphql.ID
	Slug *string
	Name *string
}) (*ProjectResolver, error) {
	var project codeamp_models.Project
	var query *gorm.DB

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

	return &ProjectResolver{DB: r.DB, Project: project}, nil
}

func (r *Resolver) CreateProject(args *struct{ Project *ProjectInput }) (*ProjectResolver, error) {
	project := codeamp_models.Project{
		GitProtocol: args.Project.GitProtocol,
		GitUrl:      args.Project.GitUrl,
		Secret:      transistor.RandomString(30),
	}

	res := plugins.GetRegexParams("(?P<host>(git@|https?:\\/\\/)([\\w\\.@]+)(\\/|:))(?P<owner>[\\w,\\-,\\_]+)\\/(?P<repo>[\\w,\\-,\\_]+)(.git){0,1}((\\/){0,1})", args.Project.GitUrl)
	repository := fmt.Sprintf("%s/%s", res["owner"], res["repo"])

	// Check if project already exists with same name
	existingProject := codeamp_models.Project{}
	spew.Dump("existingProject")
	if err := r.DB.Unscoped().Where("repository = ?", repository).First(&existingProject).Error; err != nil {
		log.Println("No record found")
	} else {
		if existingProject != (codeamp_models.Project{}) {
			return nil, fmt.Errorf("This repository already exists. Try again with a different git url.")
		}
	}

	project.Name = repository
	project.Repository = repository
	project.Slug = slug.Slug(repository)

	deletedProject := codeamp_models.Project{}
	if err := r.DB.Unscoped().Where("repository = ?", repository).First(&deletedProject).Error; err != nil {
		project.Model.ID = deletedProject.Model.ID
	}

	// priv *rsa.PrivateKey;
	priv, err := rsa.GenerateKey(rand.Reader, 2014)
	if err != nil {
		return nil, err
	}

	err = priv.Validate()
	if err != nil {
		return nil, err
	}

	// Get der format. priv_der []byte
	priv_der := x509.MarshalPKCS1PrivateKey(priv)

	// pem.Block
	priv_blk := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   priv_der,
	}

	// Public Key generation
	pub, err := ssh.NewPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, err
	}

	project.RsaPrivateKey = string(pem.EncodeToMemory(&priv_blk))
	project.RsaPublicKey = string(ssh.MarshalAuthorizedKey(pub))

	r.DB.Create(&project)

	return &ProjectResolver{DB: r.DB, Project: project}, nil
}

type ProjectResolver struct {
	DB      *gorm.DB
	Project codeamp_models.Project
}

func (r *ProjectResolver) ID() graphql.ID {
	return graphql.ID(r.Project.Model.ID.String())
}

func (r *ProjectResolver) Name() string {
	return r.Project.Name
}

func (r *ProjectResolver) Slug() string {
	return r.Project.Slug
}

func (r *ProjectResolver) Repository() string {
	return r.Project.Repository
}

func (r *ProjectResolver) Secret() string {
	return r.Project.Secret
}

func (r *ProjectResolver) GitUrl() string {
	return r.Project.GitUrl
}

func (r *ProjectResolver) GitProtocol() string {
	return r.Project.GitProtocol
}

func (r *ProjectResolver) RsaPrivateKey() string {
	return r.Project.RsaPrivateKey
}

func (r *ProjectResolver) RsaPublicKey() string {
	return r.Project.RsaPublicKey
}

func (r *ProjectResolver) Releases(ctx context.Context) ([]*ReleaseResolver, error) {
	var rows []codeamp_models.Release
	var results []*ReleaseResolver

	r.DB.Model(r.Project).Related(&rows)

	for _, release := range rows {
		results = append(results, &ReleaseResolver{DB: r.DB, Release: release})
	}

	return results, nil
}

package codeamp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/codeamp/circuit/plugins"
	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/spf13/viper"
	gormigrate "gopkg.in/gormigrate.v1"
)

func (x *CodeAmp) Migrate() {
	db, err := gorm.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s password=%s",
		viper.GetString("plugins.codeamp.postgres.host"),
		viper.GetString("plugins.codeamp.postgres.port"),
		viper.GetString("plugins.codeamp.postgres.user"),
		viper.GetString("plugins.codeamp.postgres.dbname"),
		viper.GetString("plugins.codeamp.postgres.sslmode"),
		viper.GetString("plugins.codeamp.postgres.password"),
	))
	if err != nil {
		log.Fatal(err)
	}

	db.LogMode(false)
	db.Set("gorm:auto_preload", true)

	timeMigrationBegins := time.Now()

	db.AutoMigrate(
		&model.User{},
		&model.UserPermission{},
		&model.Project{},
		&model.ProjectSettings{},
		&model.Release{},
		&model.Feature{},
		&model.Service{},
		&model.ServicePort{},
		&model.ServiceDeploymentStrategy{},
		&model.ServiceHealthProbe{},
		&model.ServiceHealthProbeHttpHeader{},
		&model.ServiceSpec{},
		&model.Extension{},
		&model.ProjectExtension{},
		&model.Secret{},
		&model.SecretValue{},
		&model.ReleaseExtension{},
		&model.Environment{},
		&model.ProjectEnvironment{},
		&model.ProjectBookmark{},
	)

	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		// create users
		{
			ID: "201803021521",
			Migrate: func(tx *gorm.DB) error {
				emails := []string{
					"kilgore@kilgore.trout",
				}

				for _, email := range emails {
					user := model.User{
						Email: email,
					}

					if err := db.Save(&user).Error; err != nil {
						return err
					}

					userPermission := model.UserPermission{
						UserID: user.Model.ID,
						Value:  "admin",
					}

					if err := db.Save(&userPermission).Error; err != nil {
						return err
					}
				}

				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				log.Error("Migration 201803021521 Rollback")
				return nil
			},
		},
		// create environments
		{
			ID: "201803021522",
			Migrate: func(tx *gorm.DB) error {
				environments := []string{
					"development",
					"production",
				}

				for _, name := range environments {
					environment := model.Environment{
						Name:  name,
						Color: "red",
					}
					if err := db.Save(&environment).Error; err != nil {
						return err
					}
				}

				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				log.Error("Migration 201803021522 Rollback")
				return nil
			},
		},
		// create extension secrets
		{
			ID: "201803021530",
			Migrate: func(tx *gorm.DB) error {
				envSecrets := []string{
					"HOSTED_ZONE_ID",
					"HOSTED_ZONE_NAME",
					"AWS_SECRET_KEY",
					"AWS_ACCESS_KEY_ID",
					"DOCKER_ORG",
					"DOCKER_HOST",
					"DOCKER_USER",
					"DOCKER_EMAIL",
					"DOCKER_PASS",
					"ACCESS_LOG_S3_BUCKET",
					"SSL_CERT_ARN",
				}

				fileSecrets := []string{
					"CERTIFICATE_AUTHORITY",
					"CLIENT_KEY",
					"CLIENT_CERTIFICATE",
					"KUBECONFIG",
				}

				var user model.User
				var environments []model.Environment

				if err := db.First(&user).Error; err != nil {
					return err
				}

				if err := db.Find(&environments).Error; err != nil {
					return err
				}
				for _, environment := range environments {
					// ENV
					for _, name := range envSecrets {
						secret := model.Secret{
							Key:           name,
							Type:          "env",
							Scope:         graphql_resolver.GetSecretScope("extension"),
							EnvironmentID: environment.Model.ID,
						}
						if err := db.Save(&secret).Error; err != nil {
							return err
						}

						secretValue := model.SecretValue{
							SecretID: secret.Model.ID,
							UserID:   user.Model.ID,
							Value:    "",
						}
						if err := db.Save(&secretValue).Error; err != nil {
							return err
						}
					}
					// FILE
					for _, name := range fileSecrets {
						secret := model.Secret{
							Key:           name,
							Type:          "file",
							Scope:         graphql_resolver.GetSecretScope("extension"),
							EnvironmentID: environment.Model.ID,
						}
						if err := db.Save(&secret).Error; err != nil {
							return err
						}
					}
				}

				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				log.Error("Migration 201803021530 Rollback")
				return db.Delete(&model.Secret{}).Error
			},
		},
		// create Service Spec
		{
			ID: "201803031530",
			Migrate: func(tx *gorm.DB) error {
				serviceSpec := model.ServiceSpec{
					Name:                   "default",
					CpuRequest:             "500",
					CpuLimit:               "500",
					MemoryRequest:          "500",
					MemoryLimit:            "500",
					TerminationGracePeriod: "300",
				}
				if err := db.Save(&serviceSpec).Error; err != nil {
					return err
				}

				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				log.Error("Migration 201803031530 Rollback")
				return db.Delete(&model.ServiceSpec{}).Error
			},
		},
		// create extensions
		{
			ID: "201803021531",
			Migrate: func(tx *gorm.DB) error {
				var environments []model.Environment
				var config []map[string]interface{}
				var marshalledConfig []byte
				var extension model.Extension

				if err := db.Find(&environments).Error; err != nil {
					return err
				}
				for _, environment := range environments {
					// dockerbuilder
					var dockerOrg model.Secret
					if err := db.Where("key = ? AND environment_id = ?", "DOCKER_ORG", environment.Model.ID).FirstOrInit(&dockerOrg).Error; err != nil {
						return err
					}

					var dockerHost model.Secret
					if err := db.Where("key = ? AND environment_id = ?", "DOCKER_HOST", environment.Model.ID).FirstOrInit(&dockerHost).Error; err != nil {
						return err
					}

					var dockerUser model.Secret
					if err := db.Where("key = ? AND environment_id = ?", "DOCKER_USER", environment.Model.ID).FirstOrInit(&dockerUser).Error; err != nil {
						return err
					}

					var dockerEmail model.Secret
					if err := db.Where("key = ? AND environment_id = ?", "DOCKER_EMAIL", environment.Model.ID).FirstOrInit(&dockerEmail).Error; err != nil {
						return err
					}

					var dockerPass model.Secret
					if err := db.Where("key = ? AND environment_id = ?", "DOCKER_PASS", environment.Model.ID).FirstOrInit(&dockerPass).Error; err != nil {
						return err
					}

					config = []map[string]interface{}{
						{"key": "ORG", "value": dockerOrg.Model.ID.String()},
						{"key": "HOST", "value": dockerHost.Model.ID.String()},
						{"key": "USER", "value": dockerUser.Model.ID.String()},
						{"key": "EMAIL", "value": dockerEmail.Model.ID.String()},
						{"key": "PASSWORD", "value": dockerPass.Model.ID.String()},
					}

					marshalledConfig, err = json.Marshal(config)
					if err != nil {
						log.Error("could not marshal config")
						return err
					}

					extension = model.Extension{
						Type:          plugins.GetType("workflow"),
						Key:           "dockerbuilder",
						Name:          "Docker Builder",
						Component:     "",
						EnvironmentID: environment.Model.ID,
						Config:        postgres.Jsonb{marshalledConfig},
					}

					if err := db.Save(&extension).Error; err != nil {
						return err
					}

					// loadbalancer
					var sslArn model.Secret
					if err := db.Where("key = ? AND environment_id = ?", "SSL_CERT_ARN", environment.Model.ID).FirstOrInit(&sslArn).Error; err != nil {
						return err
					}

					var s3Bucket model.Secret
					if err := db.Where("key = ? AND environment_id = ?", "ACCESS_LOG_S3_BUCKET", environment.Model.ID).FirstOrInit(&s3Bucket).Error; err != nil {
						return err
					}

					var hostedZoneID model.Secret
					if err := db.Where("key = ? AND environment_id = ?", "HOSTED_ZONE_ID", environment.Model.ID).FirstOrInit(&hostedZoneID).Error; err != nil {
						return err
					}

					var hostedZoneName model.Secret
					if err := db.Where("key = ? AND environment_id = ?", "HOSTED_ZONE_NAME", environment.Model.ID).FirstOrInit(&hostedZoneName).Error; err != nil {
						return err
					}

					var awsAccessKeyID model.Secret
					if err := db.Where("key = ? AND environment_id = ?", "AWS_ACCESS_KEY_ID", environment.Model.ID).FirstOrInit(&awsAccessKeyID).Error; err != nil {
						return err
					}

					var awsSecretKey model.Secret
					if err := db.Where("key = ? AND environment_id = ?", "AWS_SECRET_KEY", environment.Model.ID).FirstOrInit(&awsSecretKey).Error; err != nil {
						return err
					}

					var clientCert model.Secret
					if err := db.Where("key = ? AND environment_id = ?", "CLIENT_CERTIFICATE", environment.Model.ID).FirstOrInit(&clientCert).Error; err != nil {
						return err
					}

					var clientKey model.Secret
					if err := db.Where("key = ? AND environment_id = ?", "CLIENT_KEY", environment.Model.ID).FirstOrInit(&clientKey).Error; err != nil {
						return err
					}

					var certificateAuthority model.Secret
					if err := db.Where("key = ? AND environment_id = ?", "CERTIFICATE_AUTHORITY", environment.Model.ID).FirstOrInit(&certificateAuthority).Error; err != nil {
						return err
					}

					config = []map[string]interface{}{
						{"key": "SSL_CERT_ARN", "value": sslArn.Model.ID.String()},
						{"key": "ACCESS_LOG_S3_BUCKET", "value": s3Bucket.Model.ID.String()},
						{"key": "HOSTED_ZONE_ID", "value": hostedZoneID.Model.ID.String()},
						{"key": "HOSTED_ZONE_NAME", "value": hostedZoneName.Model.ID.String()},
						{"key": "AWS_ACCESS_KEY_ID", "value": awsAccessKeyID.Model.ID.String()},
						{"key": "AWS_SECRET_KEY", "value": awsSecretKey.Model.ID.String()},
						{"key": "CLIENT_CERTIFICATE", "value": clientCert.Model.ID.String()},
						{"key": "CLIENT_KEY", "value": clientKey.Model.ID.String()},
						{"key": "CERTIFICATE_AUTHORITY", "value": certificateAuthority.Model.ID.String()},
					}

					marshalledConfig, err = json.Marshal(config)
					if err != nil {
						log.Error("could not marshal config")
						return err
					}

					extension = model.Extension{
						Type:          plugins.GetType("once"),
						Key:           "kubernetes:loadbalancer",
						Name:          "Load Balancer",
						Component:     "LoadBalancer",
						EnvironmentID: environment.Model.ID,
						Config:        postgres.Jsonb{marshalledConfig},
					}

					if err := db.Save(&extension).Error; err != nil {
						return err
					}

					// kubernetes
					var kubeConfig model.Secret
					if err := db.Where("key = ? AND environment_id = ?", "KUBECONFIG", environment.Model.ID).FirstOrInit(&kubeConfig).Error; err != nil {
						return err
					}

					if err := db.Where("key = ? AND environment_id = ?", "CLIENT_CERTIFICATE", environment.Model.ID).FirstOrInit(&clientCert).Error; err != nil {
						return err
					}

					if err := db.Where("key = ? AND environment_id = ?", "CLIENT_KEY", environment.Model.ID).FirstOrInit(&clientKey).Error; err != nil {
						return err
					}

					if err := db.Where("key = ? AND environment_id = ?", "CERTIFICATE_AUTHORITY", environment.Model.ID).FirstOrInit(&certificateAuthority).Error; err != nil {
						return err
					}

					config = []map[string]interface{}{
						{"key": "KUBECONFIG", "value": kubeConfig.Model.ID.String()},
						{"key": "CLIENT_CERTIFICATE", "value": clientCert.Model.ID.String()},
						{"key": "CLIENT_KEY", "value": clientKey.Model.ID.String()},
						{"key": "CERTIFICATE_AUTHORITY", "value": certificateAuthority.Model.ID.String()},
					}

					marshalledConfig, err = json.Marshal(config)
					if err != nil {
						log.Error("could not marshal config")
						return err
					}

					extension = model.Extension{
						Type:          plugins.GetType("deployment"),
						Key:           "kubernetes:deployment",
						Name:          "Kubernetes",
						Component:     "",
						EnvironmentID: environment.Model.ID,
						Config:        postgres.Jsonb{marshalledConfig},
					}

					if err := db.Save(&extension).Error; err != nil {
						return err
					}
				}

				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				log.Error("Migration 201803021531 Rollback")
				return db.Delete(&model.Extension{}).Error
			},
		},
		// create ProjectEnvironments
		{
			ID: "201803081647",
			Migrate: func(tx *gorm.DB) error {

				// create default project permission for projects that don't have it
				projects := []model.Project{}

				if err := db.Find(&projects).Error; err != nil {
					return err
				}

				// give permission to all environments
				// for each project
				envs := []model.Environment{}

				if err := db.Find(&envs).Error; err != nil {
					return err
				}

				for _, env := range envs {
					for _, project := range projects {
						if err := db.FirstOrCreate(&model.ProjectEnvironment{
							EnvironmentID: env.Model.ID,
							ProjectID:     project.Model.ID,
						}).Error; err != nil {
							return err
						}
					}
				}
				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				log.Error("Migration 201803081647 Rollback")
				return db.DropTable(&model.ProjectEnvironment{}).Error
			},
		},
		// add key attribute to environment
		{
			ID: "201803081103",
			Migrate: func(tx *gorm.DB) error {
				var environments []model.Environment
				if err := db.Find(&environments).Error; err != nil {
					return err
				}
				for _, env := range environments {
					if env.Key == "" {
						env.Key = env.Name
					}
					if err := db.Save(&env).Error; err != nil {
						return err
					}
				}
				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				log.Error("Migration 201803081103 Rollback")
				return db.Model(&model.Environment{}).DropColumn("key").Error
			},
		},
		// add is_default attribute to environment
		{
			ID: "201803191507",
			Migrate: func(tx *gorm.DB) error {
				var environments []model.Environment
				if err := db.Find(&environments).Error; err != nil {
					return err
				}
				for _, env := range environments {
					env.IsDefault = true
					if err := db.Save(&env).Error; err != nil {
						return err
					}
				}
				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				log.Error("Migration 201803191507 Rollback")
				return db.Model(&model.Environment{}).DropColumn("is_default").Error
			},
		},
		// migrate ProjectExtension config to customConfig
		{
			ID: "201803271507",
			Migrate: func(tx *gorm.DB) error {

				var projectExtensions []model.ProjectExtension
				if err := db.Find(&projectExtensions).Error; err != nil {
					return err
				}

				for _, projectExtension := range projectExtensions {
					config := make(map[string]interface{})
					err = json.Unmarshal(projectExtension.Config.RawMessage, &config)
					if err != nil {
						log.Error(err.Error())
						return err
					}

					if config["config"] != nil {
						configMarshaled, err := json.Marshal(config["config"].([]interface{}))
						if err != nil {
							log.Error(err)
							return err
						}

						projectExtension.Config = postgres.Jsonb{configMarshaled}
					}

					if config["custom"] != nil {
						customConfigMarshaled, err := json.Marshal(config["custom"].(interface{}))
						if err != nil {
							log.Error(err)
							return err
						}
						projectExtension.CustomConfig = postgres.Jsonb{customConfigMarshaled}
					}

					if err := db.Save(&projectExtension).Error; err != nil {
						return err
					}
				}

				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				log.Error("Migration 201803271507 Rollback")
				return nil
			},
		},
		// Add cacheable property to extensions
		{
			ID: "201810181637",
			Migrate: func(tx *gorm.DB) error {

				extensions := []model.Extension{}
				tx.Find(&extensions)

				for _, ext := range extensions {
					ext.Cacheable = false
					if err := tx.Save(&ext).Error; err != nil {
						return err
					}
				}

				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				log.Error("Migration 201810181637 Rollback")
				return db.Model(&model.Extension{}).DropColumn("cacheable").Error
			},
		},
		{
			ID: "201811080959",
			Migrate: func(tx *gorm.DB) error {
				serviceSpecs := []model.ServiceSpec{}
				tx.Find(&serviceSpecs)

				for _, serviceSpec := range serviceSpecs {
					serviceSpec.IsDefault = false
					if err := tx.Save(&serviceSpec).Error; err != nil {
						return err
					}
				}

				defaultServiceSpec := model.ServiceSpec{
					Name:                   "default",
					CpuLimit:               "1000",
					CpuRequest:             "100",
					MemoryLimit:            "1000",
					MemoryRequest:          "100",
					TerminationGracePeriod: "300",
					IsDefault:              true,
				}

				return tx.Create(&defaultServiceSpec).Error
			},
			Rollback: func(tx *gorm.DB) error {
				log.Error("Migration 201811080959 Rollback")
				return db.Model(&model.ServiceSpec{}).DropColumn("is_default").Error
			},
		},
	})

	if err = m.Migrate(); err != nil {
		log.Fatal(fmt.Sprintf("Could not migrate: %v", err))
	}

	timeTaskDuration := time.Since(timeMigrationBegins)
	log.Info(fmt.Sprintf("Migration Completed Successfully in %v", timeTaskDuration))
	defer db.Close()
}

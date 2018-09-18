package codeamp

import (
	"encoding/json"
	"fmt"

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
					db.Save(&user)

					userPermission := model.UserPermission{
						UserID: user.Model.ID,
						Value:  "admin",
					}
					db.Save(&userPermission)
				}

				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				return db.Delete(&model.Environment{}).Error
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
					db.Save(&environment)
				}

				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				return db.Delete(&model.Environment{}).Error
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
					"CERTIFICATE_AUTHORITY",
					"CLIENT_KEY",
					"CLIENT_CERTIFICATE",
					"KUBECONFIG",
				}

				fileSecrets := []string{
					"CERTIFICATE_AUTHORITY",
					"CLIENT_KEY",
					"CLIENT_CERTIFICATE",
					"KUBECONFIG",
				}

				var user model.User
				var environments []model.Environment

				db.First(&user)

				db.Find(&environments)
				for _, environment := range environments {
					// ENV
					for _, name := range envSecrets {
						secret := model.Secret{
							Key:           name,
							Type:          "env",
							Scope:         graphql_resolver.GetSecretScope("extension"),
							EnvironmentID: environment.Model.ID,
						}
						db.Save(&secret)

						secretValue := model.SecretValue{
							SecretID: secret.Model.ID,
							UserID:   user.Model.ID,
							Value:    "",
						}
						db.Save(&secretValue)
					}
					// FILE
					for _, name := range fileSecrets {
						secret := model.Secret{
							Key:           name,
							Type:          "file",
							Scope:         graphql_resolver.GetSecretScope("extension"),
							EnvironmentID: environment.Model.ID,
						}
						db.Save(&secret)
					}
				}

				return nil
			},
			Rollback: func(tx *gorm.DB) error {
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
				db.Save(&serviceSpec)

				return nil
			},
			Rollback: func(tx *gorm.DB) error {
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

				db.Find(&environments)
				for _, environment := range environments {
					// dockerbuilder
					var dockerOrg model.Secret
					db.Where("key = ? AND environment_id = ?", "DOCKER_ORG", environment.Model.ID).FirstOrInit(&dockerOrg)

					var dockerHost model.Secret
					db.Where("key = ? AND environment_id = ?", "DOCKER_HOST", environment.Model.ID).FirstOrInit(&dockerHost)

					var dockerUser model.Secret
					db.Where("key = ? AND environment_id = ?", "DOCKER_USER", environment.Model.ID).FirstOrInit(&dockerUser)

					var dockerEmail model.Secret
					db.Where("key = ? AND environment_id = ?", "DOCKER_EMAIL", environment.Model.ID).FirstOrInit(&dockerEmail)

					var dockerPass model.Secret
					db.Where("key = ? AND environment_id = ?", "DOCKER_PASS", environment.Model.ID).FirstOrInit(&dockerPass)

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
					}

					extension = model.Extension{
						Type:          plugins.GetType("workflow"),
						Key:           "dockerbuilder",
						Name:          "Docker Builder",
						Component:     "",
						EnvironmentID: environment.Model.ID,
						Config:        postgres.Jsonb{marshalledConfig},
					}

					db.Save(&extension)

					// loadbalancer
					var sslArn model.Secret
					db.Where("key = ? AND environment_id = ?", "SSL_CERT_ARN", environment.Model.ID).FirstOrInit(&sslArn)

					var s3Bucket model.Secret
					db.Where("key = ? AND environment_id = ?", "ACCESS_LOG_S3_BUCKET", environment.Model.ID).FirstOrInit(&s3Bucket)

					var hostedZoneID model.Secret
					db.Where("key = ? AND environment_id = ?", "HOSTED_ZONE_ID", environment.Model.ID).FirstOrInit(&hostedZoneID)

					var hostedZoneName model.Secret
					db.Where("key = ? AND environment_id = ?", "HOSTED_ZONE_NAME", environment.Model.ID).FirstOrInit(&hostedZoneName)

					var awsAccessKeyID model.Secret
					db.Where("key = ? AND environment_id = ?", "AWS_ACCESS_KEY_ID", environment.Model.ID).FirstOrInit(&awsAccessKeyID)

					var awsSecretKey model.Secret
					db.Where("key = ? AND environment_id = ?", "AWS_SECRET_KEY", environment.Model.ID).FirstOrInit(&awsSecretKey)

					var clientCert model.Secret
					db.Where("key = ? AND environment_id = ?", "CLIENT_CERTIFICATE", environment.Model.ID).FirstOrInit(&clientCert)

					var clientKey model.Secret
					db.Where("key = ? AND environment_id = ?", "CLIENT_KEY", environment.Model.ID).FirstOrInit(&clientKey)

					var certificateAuthority model.Secret
					db.Where("key = ? AND environment_id = ?", "CERTIFICATE_AUTHORITY", environment.Model.ID).FirstOrInit(&certificateAuthority)

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
					}

					extension = model.Extension{
						Type:          plugins.GetType("once"),
						Key:           "kubernetes:loadbalancer",
						Name:          "Load Balancer",
						Component:     "LoadBalancer",
						EnvironmentID: environment.Model.ID,
						Config:        postgres.Jsonb{marshalledConfig},
					}

					db.Save(&extension)

					// kubernetes
					var kubeConfig model.Secret
					db.Where("key = ? AND environment_id = ?", "KUBECONFIG", environment.Model.ID).FirstOrInit(&kubeConfig)

					db.Where("key = ? AND environment_id = ?", "CLIENT_CERTIFICATE", environment.Model.ID).FirstOrInit(&clientCert)

					db.Where("key = ? AND environment_id = ?", "CLIENT_KEY", environment.Model.ID).FirstOrInit(&clientKey)

					db.Where("key = ? AND environment_id = ?", "CERTIFICATE_AUTHORITY", environment.Model.ID).FirstOrInit(&certificateAuthority)

					config = []map[string]interface{}{
						{"key": "KUBECONFIG", "value": kubeConfig.Model.ID.String()},
						{"key": "CLIENT_CERTIFICATE", "value": clientCert.Model.ID.String()},
						{"key": "CLIENT_KEY", "value": clientKey.Model.ID.String()},
						{"key": "CERTIFICATE_AUTHORITY", "value": certificateAuthority.Model.ID.String()},
					}

					marshalledConfig, err = json.Marshal(config)
					if err != nil {
						log.Error("could not marshal config")
					}

					extension = model.Extension{
						Type:          plugins.GetType("deployment"),
						Key:           "kubernetes:deployment",
						Name:          "Kubernetes",
						Component:     "",
						EnvironmentID: environment.Model.ID,
						Config:        postgres.Jsonb{marshalledConfig},
					}

					db.Save(&extension)
				}

				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				return db.Delete(&model.Extension{}).Error
			},
		},
		// create ProjectEnvironments
		{
			ID: "201803081647",
			Migrate: func(tx *gorm.DB) error {

				// create default project permission for projects that don't have it
				projects := []model.Project{}

				db.Find(&projects)

				// give permission to all environments
				// for each project
				envs := []model.Environment{}

				db.Find(&envs)

				for _, env := range envs {
					for _, project := range projects {
						db.FirstOrCreate(&model.ProjectEnvironment{
							EnvironmentID: env.Model.ID,
							ProjectID:     project.Model.ID,
						})
					}
				}
				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				return db.DropTable(&model.ProjectEnvironment{}).Error
			},
		},
		// add key attribute to environment
		{
			ID: "201803081103",
			Migrate: func(tx *gorm.DB) error {
				var environments []model.Environment
				db.Find(&environments)
				for _, env := range environments {
					if env.Key == "" {
						env.Key = env.Name
					}
					db.Save(&env)
				}
				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				return db.Model(&model.Environment{}).DropColumn("key").Error
			},
		},
		// add is_default attribute to environment
		{
			ID: "201803191507",
			Migrate: func(tx *gorm.DB) error {
				var environments []model.Environment
				db.Find(&environments)
				for _, env := range environments {
					env.IsDefault = true
					db.Save(&env)
				}
				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				return db.Model(&model.Environment{}).DropColumn("is_default").Error
			},
		},
		// migrate ProjectExtension config to customConfig
		{
			ID: "201803271507",
			Migrate: func(tx *gorm.DB) error {

				var projectExtensions []model.ProjectExtension
				db.Find(&projectExtensions)

				for _, projectExtension := range projectExtensions {
					config := make(map[string]interface{})
					err = json.Unmarshal(projectExtension.Config.RawMessage, &config)
					if err != nil {
						log.Error(err.Error())
					}

					if config["config"] != nil {
						configMarshaled, err := json.Marshal(config["config"].([]interface{}))
						if err != nil {
							log.Error(err)
						}

						projectExtension.Config = postgres.Jsonb{configMarshaled}
					}

					if config["custom"] != nil {
						customConfigMarshaled, err := json.Marshal(config["custom"].(interface{}))
						if err != nil {
							log.Error(err)
						}
						projectExtension.CustomConfig = postgres.Jsonb{customConfigMarshaled}
					}

					db.Save(&projectExtension)
				}

				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				return nil
			},
		},
	})

	if err = m.Migrate(); err != nil {
		log.Fatal(fmt.Sprintf("Could not migrate: %v", err))
	}

	log.Info("Migration did run successfully")

	defer db.Close()
}

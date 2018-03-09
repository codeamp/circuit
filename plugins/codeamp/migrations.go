package codeamp

import (
	"encoding/json"
	"fmt"

	"github.com/codeamp/circuit/plugins"
	resolvers "github.com/codeamp/circuit/plugins/codeamp/resolvers"
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
		&resolvers.User{},
		&resolvers.UserPermission{},
		&resolvers.Project{},
		&resolvers.ProjectSettings{},
		&resolvers.Release{},
		&resolvers.Feature{},
		&resolvers.Service{},
		&resolvers.ServicePort{},
		&resolvers.ServiceSpec{},
		&resolvers.Extension{},
		&resolvers.ProjectExtension{},
		&resolvers.Secret{},
		&resolvers.SecretValue{},
		&resolvers.ReleaseExtension{},
		&resolvers.Environment{},
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
					user := resolvers.User{
						Email: email,
					}
					db.Save(&user)

					userPermission := resolvers.UserPermission{
						UserID: user.Model.ID,
						Value:  "admin",
					}
					db.Save(&userPermission)
				}

				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				return db.Delete(&resolvers.Environment{}).Error
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
					environment := resolvers.Environment{
						Name:  name,
						Color: "red",
					}
					db.Save(&environment)
				}

				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				return db.Delete(&resolvers.Environment{}).Error
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

				var user resolvers.User
				var environments []resolvers.Environment

				db.First(&user)

				db.Find(&environments)
				for _, environment := range environments {
					// ENV
					for _, name := range envSecrets {
						secret := resolvers.Secret{
							Key:           name,
							Type:          "env",
							Scope:         resolvers.GetSecretScope("extension"),
							EnvironmentID: environment.Model.ID,
						}
						db.Save(&secret)

						secretValue := resolvers.SecretValue{
							SecretID: secret.Model.ID,
							UserID:   user.Model.ID,
							Value:    "",
						}
						db.Save(&secretValue)
					}
					// FILE
					for _, name := range fileSecrets {
						secret := resolvers.Secret{
							Key:           name,
							Type:          "file",
							Scope:         resolvers.GetSecretScope("extension"),
							EnvironmentID: environment.Model.ID,
						}
						db.Save(&secret)
					}
				}

				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				return db.Delete(&resolvers.Secret{}).Error
			},
		},
		// create Service Spec
		{
			ID: "201803031530",
			Migrate: func(tx *gorm.DB) error {
				serviceSpec := resolvers.ServiceSpec{
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
				return db.Delete(&resolvers.ServiceSpec{}).Error
			},
		},
		// create extensions
		{
			ID: "201803021531",
			Migrate: func(tx *gorm.DB) error {
				var environments []resolvers.Environment
				var config []map[string]interface{}
				var marshalledConfig []byte
				var extension resolvers.Extension

				db.Find(&environments)
				for _, environment := range environments {
					// dockerbuilder
					var dockerOrg resolvers.Secret
					db.Where("key = ? AND environment_id = ?", "DOCKER_ORG", environment.Model.ID).FirstOrInit(&dockerOrg)

					var dockerHost resolvers.Secret
					db.Where("key = ? AND environment_id = ?", "DOCKER_HOST", environment.Model.ID).FirstOrInit(&dockerHost)

					var dockerUser resolvers.Secret
					db.Where("key = ? AND environment_id = ?", "DOCKER_USER", environment.Model.ID).FirstOrInit(&dockerUser)

					var dockerEmail resolvers.Secret
					db.Where("key = ? AND environment_id = ?", "DOCKER_EMAIL", environment.Model.ID).FirstOrInit(&dockerEmail)

					var dockerPass resolvers.Secret
					db.Where("key = ? AND environment_id = ?", "DOCKER_PASS", environment.Model.ID).FirstOrInit(&dockerPass)

					config = []map[string]interface{}{
						map[string]interface{}{"key": "ORG", "value": dockerOrg.Model.ID.String()},
						map[string]interface{}{"key": "HOST", "value": dockerHost.Model.ID.String()},
						map[string]interface{}{"key": "USER", "value": dockerUser.Model.ID.String()},
						map[string]interface{}{"key": "EMAIL", "value": dockerEmail.Model.ID.String()},
						map[string]interface{}{"key": "PASSWORD", "value": dockerPass.Model.ID.String()},
					}

					marshalledConfig, err = json.Marshal(config)
					if err != nil {
						log.Error("could not marshal config")
					}

					extension = resolvers.Extension{
						Type:          plugins.GetType("workflow"),
						Key:           "dockerbuilder",
						Name:          "Docker Builder",
						Component:     "",
						EnvironmentID: environment.Model.ID,
						Config:        postgres.Jsonb{marshalledConfig},
					}

					db.Save(&extension)

					// loadbalancer
					var sslArn resolvers.Secret
					db.Where("key = ? AND environment_id = ?", "SSL_CERT_ARN", environment.Model.ID).FirstOrInit(&sslArn)

					var s3Bucket resolvers.Secret
					db.Where("key = ? AND environment_id = ?", "ACCESS_LOG_S3_BUCKET", environment.Model.ID).FirstOrInit(&s3Bucket)

					var hostedZoneID resolvers.Secret
					db.Where("key = ? AND environment_id = ?", "HOSTED_ZONE_ID", environment.Model.ID).FirstOrInit(&hostedZoneID)

					var hostedZoneName resolvers.Secret
					db.Where("key = ? AND environment_id = ?", "HOSTED_ZONE_NAME", environment.Model.ID).FirstOrInit(&hostedZoneName)

					var awsAccessKeyID resolvers.Secret
					db.Where("key = ? AND environment_id = ?", "AWS_ACCESS_KEY_ID", environment.Model.ID).FirstOrInit(&awsAccessKeyID)

					var awsSecretKey resolvers.Secret
					db.Where("key = ? AND environment_id = ?", "AWS_SECRET_KEY", environment.Model.ID).FirstOrInit(&awsSecretKey)

					var clientCert resolvers.Secret
					db.Where("key = ? AND environment_id = ?", "CLIENT_CERTIFICATE", environment.Model.ID).FirstOrInit(&clientCert)

					var clientKey resolvers.Secret
					db.Where("key = ? AND environment_id = ?", "CLIENT_KEY", environment.Model.ID).FirstOrInit(&clientKey)

					var certificateAuthority resolvers.Secret
					db.Where("key = ? AND environment_id = ?", "CERTIFICATE_AUTHORITY", environment.Model.ID).FirstOrInit(&certificateAuthority)

					config = []map[string]interface{}{
						map[string]interface{}{"key": "SSL_CERT_ARN", "value": sslArn.Model.ID.String()},
						map[string]interface{}{"key": "ACCESS_LOG_S3_BUCKET", "value": s3Bucket.Model.ID.String()},
						map[string]interface{}{"key": "HOSTED_ZONE_ID", "value": hostedZoneID.Model.ID.String()},
						map[string]interface{}{"key": "HOSTED_ZONE_NAME", "value": hostedZoneName.Model.ID.String()},
						map[string]interface{}{"key": "AWS_ACCESS_KEY_ID", "value": awsAccessKeyID.Model.ID.String()},
						map[string]interface{}{"key": "AWS_SECRET_KEY", "value": awsSecretKey.Model.ID.String()},
						map[string]interface{}{"key": "CLIENT_CERTIFICATE", "value": clientCert.Model.ID.String()},
						map[string]interface{}{"key": "CLIENT_KEY", "value": clientKey.Model.ID.String()},
						map[string]interface{}{"key": "CERTIFICATE_AUTHORITY", "value": certificateAuthority.Model.ID.String()},
					}

					marshalledConfig, err = json.Marshal(config)
					if err != nil {
						log.Error("could not marshal config")
					}

					extension = resolvers.Extension{
						Type:          plugins.GetType("workflow"),
						Key:           "kubernetesloadbalancers",
						Name:          "Load Balancer",
						Component:     "LoadBalancer",
						EnvironmentID: environment.Model.ID,
						Config:        postgres.Jsonb{marshalledConfig},
					}

					db.Save(&extension)

					// kubernetes
					var kubeConfig resolvers.Secret
					db.Where("key = ? AND environment_id = ?", "KUBECONFIG", environment.Model.ID).FirstOrInit(&kubeConfig)

					db.Where("key = ? AND environment_id = ?", "CLIENT_CERTIFICATE", environment.Model.ID).FirstOrInit(&clientCert)

					db.Where("key = ? AND environment_id = ?", "CLIENT_KEY", environment.Model.ID).FirstOrInit(&clientKey)

					db.Where("key = ? AND environment_id = ?", "CERTIFICATE_AUTHORITY", environment.Model.ID).FirstOrInit(&certificateAuthority)

					config = []map[string]interface{}{
						map[string]interface{}{"key": "KUBECONFIG", "value": kubeConfig.Model.ID.String()},
						map[string]interface{}{"key": "CLIENT_CERTIFICATE", "value": clientCert.Model.ID.String()},
						map[string]interface{}{"key": "CLIENT_KEY", "value": clientKey.Model.ID.String()},
						map[string]interface{}{"key": "CERTIFICATE_AUTHORITY", "value": certificateAuthority.Model.ID.String()},
					}

					marshalledConfig, err = json.Marshal(config)
					if err != nil {
						log.Error("could not marshal config")
					}

					extension = resolvers.Extension{
						Type:          plugins.GetType("workflow"),
						Key:           "kubernetesdeployments",
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
				return db.Delete(&resolvers.Extension{}).Error
			},
		},
		// create ProjectPermissions
		{
			ID: "201803081647",
			Migrate: func(tx *gorm.DB) error {

				// create default project permission for projects that don't have it
				projects := []resolvers.Project{}

				db.Find(&projects)

				// give permission to all environments
				// for each project
				envs := []resolvers.Environment{}

				db.Find(&envs)

				for _, env := range envs {
					for _, project := range projects {
						db.FirstOrCreate(&resolvers.ProjectPermission{
							EnvironmentID: env.Model.ID,
							ProjectID:     project.Model.ID,
						})
					}
				}
				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				return db.DropTable(&resolvers.ProjectPermission{}).Error
			},
		},
	})

	if err = m.Migrate(); err != nil {
		log.Fatal("Could not migrate: %v", err)
	}

	log.Info("Migration did run successfully")

	defer db.Close()
}

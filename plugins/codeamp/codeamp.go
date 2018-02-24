package codeamp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/actions"
	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/codeamp/circuit/plugins/codeamp/schema"
	"github.com/codeamp/circuit/plugins/codeamp/schema/resolvers"
	"github.com/codeamp/circuit/plugins/codeamp/utils"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	redis "github.com/go-redis/redis"
	socketio "github.com/googollee/go-socket.io"
	"github.com/gorilla/handlers"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	graphql "github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/relay"
	sioredis "github.com/satyakb/go-socket.io-redis"
	"github.com/spf13/viper"
)

var _schema *graphql.Schema

func init() {
	transistor.RegisterPlugin("codeamp", func() transistor.Plugin { return NewCodeAmp() })
}

type CodeAmp struct {
	ServiceAddress string `mapstructure:"service_address"`
	Events         chan transistor.Event
	Schema         *graphql.Schema
	Actions        *actions.Actions
	SocketIO       *socketio.Server
	Db             *gorm.DB
	Redis          *redis.Client
}

func NewCodeAmp() *CodeAmp {
	return &CodeAmp{}
}

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
		&models.User{},
		&models.UserPermission{},
		&models.Project{},
		&models.ProjectSettings{},
		&models.Release{},
		&models.Feature{},
		&models.Service{},
		&models.ContainerPort{},
		&models.ServiceSpec{},
		&models.ExtensionSpec{},
		&models.Extension{},
		&models.EnvironmentVariable{},
		&models.EnvironmentVariableValue{},
		&models.ReleaseExtension{},
		&models.Environment{},
	)

	hashedPassword, _ := utils.HashPassword("password")

	user := models.User{
		Email:    "admin@codeamp.com",
		Password: hashedPassword,
	}

	db.FirstOrInit(&user, models.User{
		Email: "admin@codeamp.com",
	})
	db.Save(&user)

	userPermission := models.UserPermission{
		UserId: user.Model.ID,
		Value:  "admin",
	}
	db.FirstOrInit(&userPermission, userPermission)
	db.Save(&userPermission)

	developmentEnv := models.Environment{
		Name: "development",
	}
	db.FirstOrInit(&developmentEnv, developmentEnv)
	db.Save(&developmentEnv)

	productionEnv := models.Environment{
		Name: "production",
	}
	db.FirstOrInit(&productionEnv, productionEnv)
	db.Save(&productionEnv)

	// hosted zone id
	// hosted zone name
	hostedZoneId := models.EnvironmentVariable{
		Key:           "HOSTED_ZONE_ID",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
	}
	db.FirstOrInit(&hostedZoneId, hostedZoneId)
	db.Save(&hostedZoneId)

	eValue := models.EnvironmentVariableValue{
		EnvironmentVariableId: hostedZoneId.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	hostedZoneIdDev := models.EnvironmentVariable{
		Key:           "HOSTED_ZONE_ID",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&hostedZoneIdDev, hostedZoneIdDev)
	db.Save(&hostedZoneIdDev)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: hostedZoneIdDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// hosted zone name
	hostedZone := models.EnvironmentVariable{
		Key:           "HOSTED_ZONE_NAME",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
	}
	db.FirstOrInit(&hostedZone, hostedZone)
	db.Save(&hostedZone)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: hostedZone.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	hostedZoneDev := models.EnvironmentVariable{
		Key:           "HOSTED_ZONE_NAME",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&hostedZoneDev, hostedZoneDev)
	db.Save(&hostedZoneDev)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: hostedZoneDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// aws secret key
	awsSecretKey := models.EnvironmentVariable{
		Key:           "AWS_SECRET_KEY",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
		IsSecret:      true,
	}
	db.FirstOrInit(&awsSecretKey, awsSecretKey)
	db.Save(&awsSecretKey)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: awsSecretKey.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	awsSecretKeyDev := models.EnvironmentVariable{
		Key:           "AWS_SECRET_KEY",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
		IsSecret:      true,
	}
	db.FirstOrInit(&awsSecretKeyDev, awsSecretKeyDev)
	db.Save(&awsSecretKeyDev)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: awsSecretKeyDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// aws access key
	awsAccessKey := models.EnvironmentVariable{
		Key:           "AWS_ACCESS_KEY_ID",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
	}
	db.FirstOrInit(&awsAccessKey, awsAccessKey)
	db.Save(&awsAccessKey)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: awsAccessKey.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	awsAccessKeyDev := models.EnvironmentVariable{
		Key:           "AWS_ACCESS_KEY_ID",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&awsAccessKeyDev, awsAccessKeyDev)
	db.Save(&awsAccessKeyDev)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: awsAccessKeyDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// docker org
	dockerOrg := models.EnvironmentVariable{
		Key:           "DOCKER_ORG",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
	}
	db.FirstOrInit(&dockerOrg, dockerOrg)
	db.Save(&dockerOrg)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: dockerOrg.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	dockerOrgDev := models.EnvironmentVariable{
		Key:           "DOCKER_ORG",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&dockerOrgDev, dockerOrgDev)
	db.Save(&dockerOrgDev)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: dockerOrgDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// docker host
	dockerHost := models.EnvironmentVariable{
		Key:           "DOCKER_HOST",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
	}
	db.FirstOrInit(&dockerHost, dockerHost)
	db.Save(&dockerHost)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: dockerHost.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	dockerHostDev := models.EnvironmentVariable{
		Key:           "DOCKER_HOST",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&dockerHostDev, dockerHostDev)
	db.Save(&dockerHostDev)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: dockerHostDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// docker user
	dockerUser := models.EnvironmentVariable{
		Key:           "DOCKER_USER",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
	}
	db.FirstOrInit(&dockerUser, dockerUser)
	db.Save(&dockerUser)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: dockerUser.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// docker user
	dockerUserDev := models.EnvironmentVariable{
		Key:           "DOCKER_USER",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&dockerUserDev, dockerUserDev)
	db.Save(&dockerUserDev)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: dockerUserDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// docker email
	dockerEmail := models.EnvironmentVariable{
		Key:           "DOCKER_EMAIL",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
	}
	db.FirstOrInit(&dockerEmail, dockerEmail)
	db.Save(&dockerEmail)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: dockerEmail.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// docker email
	dockerEmailDev := models.EnvironmentVariable{
		Key:           "DOCKER_EMAIL",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&dockerEmailDev, dockerEmailDev)
	db.Save(&dockerEmailDev)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: dockerEmailDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// docker pass
	dockerPass := models.EnvironmentVariable{
		Key:           "DOCKER_PASS",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
		IsSecret:      true,
	}
	db.FirstOrInit(&dockerPass, dockerPass)
	db.Save(&dockerPass)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: dockerPass.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// docker pass
	dockerPassDev := models.EnvironmentVariable{
		Key:           "DOCKER_PASS",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
		IsSecret:      true,
	}
	db.FirstOrInit(&dockerPassDev, dockerPassDev)
	db.Save(&dockerPassDev)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: dockerPassDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// s3 bucket
	s3Bucket := models.EnvironmentVariable{
		Key:           "S3_BUCKET",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
	}
	db.FirstOrInit(&s3Bucket, s3Bucket)
	db.Save(&s3Bucket)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: s3Bucket.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	s3BucketDev := models.EnvironmentVariable{
		Key:           "S3_BUCKET",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&s3BucketDev, s3BucketDev)
	db.Save(&s3BucketDev)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: s3BucketDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// ssl arn
	sslArn := models.EnvironmentVariable{
		Key:           "SSL_ARN",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
	}
	db.FirstOrInit(&sslArn, sslArn)
	db.Save(&sslArn)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: sslArn.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	sslArnDev := models.EnvironmentVariable{
		Key:           "SSL_ARN",
		Type:          "env",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&sslArnDev, sslArnDev)
	db.Save(&sslArnDev)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: sslArnDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	certificateAuthorityDev := models.EnvironmentVariable{
		Key:           "CERTIFICATE_AUTHORITY",
		Type:          "file",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
		IsSecret:      true,
	}
	db.FirstOrInit(&certificateAuthorityDev, certificateAuthorityDev)
	db.Save(&certificateAuthorityDev)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: certificateAuthorityDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	certificateAuthorityProd := models.EnvironmentVariable{
		Key:           "CERTIFICATE_AUTHORITY",
		Type:          "file",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
		IsSecret:      true,
	}
	db.FirstOrInit(&certificateAuthorityProd, certificateAuthorityProd)
	db.Save(&certificateAuthorityProd)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: certificateAuthorityProd.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	clientKeyDev := models.EnvironmentVariable{
		Key:           "CLIENT_KEY",
		Type:          "file",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
		IsSecret:      true,
	}

	db.FirstOrInit(&clientKeyDev, clientKeyDev)
	db.Save(&clientKeyDev)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: clientKeyDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	clientKeyProd := models.EnvironmentVariable{
		Key:           "CLIENT_KEY",
		Type:          "file",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
		IsSecret:      true,
	}
	db.FirstOrInit(&clientKeyProd, clientKeyProd)
	db.Save(&clientKeyProd)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: clientKeyProd.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	clientCertDev := models.EnvironmentVariable{
		Key:           "CLIENT_CERTIFICATE",
		Type:          "file",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
		IsSecret:      true,
	}

	db.FirstOrInit(&clientCertDev, clientCertDev)
	db.Save(&clientCertDev)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: clientCertDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	clientCertProd := models.EnvironmentVariable{
		Key:           "CLIENT_CERTIFICATE",
		Type:          "file",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
		IsSecret:      true,
	}

	db.FirstOrInit(&clientCertProd, clientCertProd)
	db.Save(&clientCertProd)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: clientCertProd.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// kubeconfig
	kubeConfig := models.EnvironmentVariable{
		Key:           "KUBECONFIG",
		Type:          "file",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
	}
	db.FirstOrInit(&kubeConfig, kubeConfig)
	db.Save(&kubeConfig)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: kubeConfig.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	kubeConfigDev := models.EnvironmentVariable{
		Key:           "KUBECONFIG",
		Type:          "file",
		Scope:         models.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&kubeConfigDev, kubeConfigDev)
	db.Save(&kubeConfigDev)
	eValue = models.EnvironmentVariableValue{
		EnvironmentVariableId: kubeConfigDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	serviceSpec := models.ServiceSpec{
		Name:                   "default",
		CpuRequest:             "500",
		CpuLimit:               "500",
		MemoryRequest:          "500",
		MemoryLimit:            "500",
		TerminationGracePeriod: "300",
	}
	db.FirstOrInit(&serviceSpec, models.ServiceSpec{
		Name: "default",
	})
	db.Save(&serviceSpec)

	// docker builder
	dbConfig := []map[string]interface{}{
		map[string]interface{}{"key": "KUBECONFIG", "value": kubeConfig.Model.ID.String()},
		map[string]interface{}{"key": "ORG", "value": dockerOrg.Model.ID.String()},
		map[string]interface{}{"key": "HOST", "value": dockerHost.Model.ID.String()},
		map[string]interface{}{"key": "USER", "value": dockerUser.Model.ID.String()},
		map[string]interface{}{"key": "EMAIL", "value": dockerEmail.Model.ID.String()},
		map[string]interface{}{"key": "PASSWORD", "value": dockerPass.Model.ID.String()},
	}
	marshalledDbConfig, err := json.Marshal(dbConfig)
	if err != nil {
		log.Info("could not marshal dockerbuilder config")
	}
	extensionSpec := models.ExtensionSpec{
		Type:          plugins.GetType("workflow"),
		Key:           "dockerbuilder",
		Name:          "Docker Builder",
		Component:     "",
		EnvironmentId: productionEnv.Model.ID,
		Config:        postgres.Jsonb{marshalledDbConfig},
	}
	db.FirstOrInit(&extensionSpec, extensionSpec)
	db.Save(&extensionSpec)

	dbConfig = []map[string]interface{}{
		map[string]interface{}{"key": "KUBECONFIG", "value": kubeConfigDev.Model.ID.String()},
		map[string]interface{}{"key": "ORG", "value": dockerOrgDev.Model.ID.String()},
		map[string]interface{}{"key": "HOST", "value": dockerHostDev.Model.ID.String()},
		map[string]interface{}{"key": "USER", "value": dockerUserDev.Model.ID.String()},
		map[string]interface{}{"key": "EMAIL", "value": dockerEmailDev.Model.ID.String()},
		map[string]interface{}{"key": "PASSWORD", "value": dockerPassDev.Model.ID.String()},
	}
	marshalledDbConfig, err = json.Marshal(dbConfig)
	if err != nil {
		log.Info("could not marshal dockerbuilder config")
	}
	extensionSpec = models.ExtensionSpec{
		Type:          plugins.GetType("workflow"),
		Key:           "dockerbuilder",
		Name:          "Docker Builder",
		Component:     "",
		EnvironmentId: developmentEnv.Model.ID,
		Config:        postgres.Jsonb{marshalledDbConfig},
	}

	db.FirstOrInit(&extensionSpec, extensionSpec)
	db.Save(&extensionSpec)

	// load balancer
	lbConfig := []map[string]interface{}{
		map[string]interface{}{"key": "KUBECONFIG", "value": kubeConfig.Model.ID.String()},
		map[string]interface{}{"key": "SSL_CERT_ARN", "value": sslArn.Model.ID.String()},
		map[string]interface{}{"key": "ACCESS_LOG_S3_BUCKET", "value": s3Bucket.Model.ID.String()},
		map[string]interface{}{"key": "HOSTED_ZONE_ID", "value": hostedZoneId.Model.ID.String()},
		map[string]interface{}{"key": "HOSTED_ZONE_NAME", "value": hostedZone.Model.ID.String()},
		map[string]interface{}{"key": "AWS_ACCESS_KEY_ID", "value": awsAccessKey.Model.ID.String()},
		map[string]interface{}{"key": "AWS_SECRET_KEY", "value": awsSecretKey.Model.ID.String()},
		map[string]interface{}{"key": "CLIENT_CERTIFICATE", "value": clientCertProd.Model.ID.String()},
		map[string]interface{}{"key": "CLIENT_KEY", "value": clientKeyProd.Model.ID.String()},
		map[string]interface{}{"key": "CERTIFICATE_AUTHORITY", "value": certificateAuthorityProd.Model.ID.String()},
	}
	marshalledLbConfig, err := json.Marshal(lbConfig)
	if err != nil {
		log.Info("could not marshal loadbalancer config")
	}
	extensionSpec = models.ExtensionSpec{
		Type:          plugins.GetType("once"),
		Key:           "kubernetesloadbalancers",
		Name:          "Load Balancer",
		Component:     "LoadBalancer",
		EnvironmentId: productionEnv.Model.ID,
		Config:        postgres.Jsonb{marshalledLbConfig},
	}

	db.FirstOrInit(&extensionSpec, extensionSpec)
	db.Save(&extensionSpec)

	lbConfig = []map[string]interface{}{
		map[string]interface{}{"key": "KUBECONFIG", "value": kubeConfigDev.Model.ID.String()},
		map[string]interface{}{"key": "SSL_CERT_ARN", "value": sslArnDev.Model.ID.String()},
		map[string]interface{}{"key": "ACCESS_LOG_S3_BUCKET", "value": s3BucketDev.Model.ID.String()},
		map[string]interface{}{"key": "HOSTED_ZONE_ID", "value": hostedZoneIdDev.Model.ID.String()},
		map[string]interface{}{"key": "HOSTED_ZONE_NAME", "value": hostedZoneDev.Model.ID.String()},
		map[string]interface{}{"key": "AWS_ACCESS_KEY_ID", "value": awsAccessKeyDev.Model.ID.String()},
		map[string]interface{}{"key": "AWS_SECRET_KEY", "value": awsSecretKeyDev.Model.ID.String()},
		map[string]interface{}{"key": "CLIENT_CERTIFICATE", "value": clientCertDev.Model.ID.String()},
		map[string]interface{}{"key": "CLIENT_KEY", "value": clientKeyDev.Model.ID.String()},
		map[string]interface{}{"key": "CERTIFICATE_AUTHORITY", "value": certificateAuthorityDev.Model.ID.String()},
	}
	marshalledLbConfig, err = json.Marshal(lbConfig)
	if err != nil {
		log.Info("could not marshal loadbalancer config")
	}
	extensionSpec = models.ExtensionSpec{
		Type:          plugins.GetType("once"),
		Key:           "kubernetesloadbalancers",
		Name:          "Load Balancer",
		Component:     "LoadBalancer",
		EnvironmentId: developmentEnv.Model.ID,
		Config:        postgres.Jsonb{marshalledLbConfig},
	}

	db.FirstOrInit(&extensionSpec, extensionSpec)
	db.Save(&extensionSpec)

	// kubernetes
	kubeConfigSpec := []map[string]interface{}{
		map[string]interface{}{"key": "KUBECONFIG", "value": kubeConfig.Model.ID.String()},
		map[string]interface{}{"key": "CLIENT_CERTIFICATE", "value": clientCertProd.Model.ID.String()},
		map[string]interface{}{"key": "CLIENT_KEY", "value": clientKeyProd.Model.ID.String()},
		map[string]interface{}{"key": "CERTIFICATE_AUTHORITY", "value": certificateAuthorityProd.Model.ID.String()},
	}
	marshalledKubeConfig, err := json.Marshal(kubeConfigSpec)
	if err != nil {
		log.Info("could not marshal kube config")
	}
	extensionSpec = models.ExtensionSpec{
		Type:          plugins.GetType("deployment"),
		Key:           "kubernetesdeployments",
		Name:          "Kubernetes",
		Component:     "",
		EnvironmentId: productionEnv.Model.ID,
		Config:        postgres.Jsonb{marshalledKubeConfig},
	}

	db.FirstOrInit(&extensionSpec, extensionSpec)
	db.Save(&extensionSpec)

	kubeConfigSpec = []map[string]interface{}{
		map[string]interface{}{"key": "KUBECONFIG", "value": kubeConfigDev.Model.ID.String()},
		map[string]interface{}{"key": "CLIENT_CERTIFICATE", "value": clientCertDev.Model.ID.String()},
		map[string]interface{}{"key": "CLIENT_KEY", "value": clientKeyDev.Model.ID.String()},
		map[string]interface{}{"key": "CERTIFICATE_AUTHORITY", "value": certificateAuthorityDev.Model.ID.String()},
	}
	marshalledKubeConfig, err = json.Marshal(kubeConfigSpec)
	if err != nil {
		log.Info("could not marshal kube config")
	}
	extensionSpec = models.ExtensionSpec{
		Type:          plugins.GetType("deployment"),
		Key:           "kubernetesdeployments",
		Name:          "Kubernetes",
		Component:     "",
		EnvironmentId: developmentEnv.Model.ID,
		Config:        postgres.Jsonb{marshalledKubeConfig},
	}

	db.FirstOrInit(&extensionSpec, extensionSpec)
	db.Save(&extensionSpec)

	defer db.Close()
}

//Custom server which basically only contains a socketio variable
//But we need it to enhance it with functions
type socketIOServer struct {
	Server *socketio.Server
}

//Header handling, this is necessary to adjust security and/or header settings in general
//Please keep in mind to adjust that later on in a productive environment!
//Access-Control-Allow-Origin will be set to whoever will call the server
func (s *socketIOServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	origin := r.Header.Get("Origin")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	s.Server.ServeHTTP(w, r)
}

func (x *CodeAmp) Listen() {
	x.SocketIO.On("connection", func(so socketio.Socket) {
		so.Join("general")
	})

	x.SocketIO.On("error", func(so socketio.Socket, err error) {
		log.Println("socket-io error:", err)
	})

	sIOServer := new(socketIOServer)
	sIOServer.Server = x.SocketIO
	http.Handle("/socket.io/", sIOServer)

	_, filename, _, _ := runtime.Caller(0)
	fs := http.FileServer(http.Dir(path.Join(path.Dir(filename), "static/")))
	http.Handle("/", fs)
	http.Handle("/query", utils.CorsMiddleware(utils.AuthMiddleware(&relay.Handler{Schema: x.Schema}, x.Db, x.Redis)))

	log.Info(fmt.Sprintf("running GraphQL server on %v", x.ServiceAddress))
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s", x.ServiceAddress), handlers.LoggingHandler(os.Stdout, http.DefaultServeMux)))
}

func (x *CodeAmp) Start(events chan transistor.Event) error {
	var err error

	x.Events = events

	db, err := gorm.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s password=%s",
		viper.GetString("plugins.codeamp.postgres.host"),
		viper.GetString("plugins.codeamp.postgres.port"),
		viper.GetString("plugins.codeamp.postgres.user"),
		viper.GetString("plugins.codeamp.postgres.dbname"),
		viper.GetString("plugins.codeamp.postgres.sslmode"),
		viper.GetString("plugins.codeamp.postgres.password"),
	))
	//defer x.DB.Close()

	actions := actions.NewActions(events, db)
	resolver := resolvers.NewResolver(events, db, actions)

	s, err := schema.Schema()
	if err != nil {
		log.Fatal(err)
	}

	parsedSchema, err := graphql.ParseSchema(string(s), resolver)
	if err != nil {
		panic(err)
	}

	// Socket-io
	sio, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	split := strings.Split(viper.GetString("redis.server"), ":")
	host, port := split[0], split[1]

	opts := map[string]string{
		"host": host,
		"port": port,
	}
	sio.SetAdaptor(sioredis.Redis(opts))

	redisDb, err := strconv.Atoi(viper.GetString("redis.database"))
	if err != nil {
		log.Fatal(err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis.server"),
		Password: viper.GetString("redis.password"),
		DB:       redisDb,
	})

	if _, err := redisClient.Ping().Result(); err != nil {
		log.Fatal(err)
	}

	x.SocketIO = sio
	x.Actions = actions
	x.Schema = parsedSchema
	x.Redis = redisClient

	// DEBUG
	db.LogMode(false)

	x.Db = db

	go x.Listen()

	return nil
}

func (x *CodeAmp) Stop() {
	log.Info("stopping CodeAmp service")
}

func (x *CodeAmp) Subscribe() []string {
	return []string{
		"plugins.GitPing",
		"plugins.GitCommit",
		"plugins.GitStatus",
		"plugins.HeartBeat",
		"plugins.WebsocketMsg",
		"plugins.Extension:status",
		"plugins.Extension:update",
		"plugins.Extension:complete",
		"plugins.ReleaseExtension:status",
		"plugins.Release:status",
		"plugins.Release:complete",
	}
}

func (x *CodeAmp) Process(e transistor.Event) error {
	log.InfoWithFields("process CodeAmp event", log.Fields{
		"event_name": e.Name,
	})

	if e.Matches("plugins.HeartBeat") {
		heartBeat := e.Payload.(plugins.HeartBeat)
		switch heartBeat.Tick {
		case "minute":
			x.Actions.HeartBeat("minute")
		}
		return nil
	}

	if e.Matches("plugins.GitCommit") {
		payload := e.Payload.(plugins.GitCommit)
		x.Actions.GitCommit(payload)

		var project models.Project
		if x.Db.Where("repository = ?", payload.Repository).Find(&project).RecordNotFound() {
			log.InfoWithFields("project not found", log.Fields{
				"repository": payload.Repository,
			})
		} else {
			x.Events <- transistor.NewEvent(plugins.WebsocketMsg{
				Event: fmt.Sprintf("projects/%s/features", project.Slug),
			}, nil)
		}

	}

	if e.Matches("plugins.WebsocketMsg") {
		payload := e.Payload.(plugins.WebsocketMsg)

		if payload.Channel == "" {
			payload.Channel = "general"
		}

		x.SocketIO.BroadcastTo(payload.Channel, payload.Event, payload.Payload, nil)
	}

	if e.Matches("plugins.Extension:status") {
		payload := e.Payload.(plugins.Extension)
		var extension models.Extension
		var project models.Project

		if x.Db.Where("id = ?", payload.Id).Find(&extension).RecordNotFound() {
			log.InfoWithFields("extension not found", log.Fields{
				"id": payload.Id,
			})
			return fmt.Errorf(fmt.Sprintf("Could not handle Extension status event because Extension not found given payload id: %s.", payload.Id))
		}

		if x.Db.Where("id = ?", extension.ProjectId).Find(&project).RecordNotFound() {
			log.InfoWithFields("project not found", log.Fields{
				"id": extension.ProjectId,
			})
			return fmt.Errorf(fmt.Sprintf("Could not handle Extension status event because Project not found given payload id: %s.", extension.ProjectId))
		}

		// get old artifacts and then merge with the payload.Artifacts
		// this prevents data regressions upon extension failure events
		// e.g. load balancer completes w/ Artifacts, but route53 fails with empty artifacts

		mergedArtifacts := make(map[string]string)
		err := json.Unmarshal(extension.Artifacts.RawMessage, &mergedArtifacts)
		if err != nil {
			log.Info(err.Error())
			return err
		}
		if len(mergedArtifacts) > 0 {
			for key, value := range payload.Artifacts {
				mergedArtifacts[key] = value
			}
		} else {
			mergedArtifacts = payload.Artifacts
		}

		marshalledArtifacts, err := json.Marshal(mergedArtifacts)
		if err != nil {
			log.InfoWithFields(err.Error(), log.Fields{})
			return err
		}

		extension.State = payload.State
		extension.StateMessage = payload.StateMessage
		extension.Artifacts = postgres.Jsonb{marshalledArtifacts}
		x.Db.Save(&extension)

		x.Events <- transistor.NewEvent(plugins.WebsocketMsg{
			Event:   fmt.Sprintf("projects/%s/extensions", project.Slug),
			Payload: extension,
		}, nil)

		if payload.State == plugins.GetState("complete") {
			x.Actions.ExtensionInitCompleted(&extension)
		}
	}

	if e.Matches("plugins.ReleaseExtension:status") {
		payload := e.Payload.(plugins.ReleaseExtension)
		var releaseExtension models.ReleaseExtension
		var release models.Release

		if x.Db.Where("id = ?", payload.Release.Id).Find(&release).RecordNotFound() {
			log.InfoWithFields("release", log.Fields{
				"id": payload.Release.Id,
			})
			return nil
		}

		if x.Db.Where("id = ?", payload.Id).Find(&releaseExtension).RecordNotFound() {
			log.InfoWithFields("release extension not found", log.Fields{
				"id": payload.Id,
			})
			return nil
		}

		mergedArtifacts := make(map[string]string)
		err := json.Unmarshal(releaseExtension.Artifacts.RawMessage, &mergedArtifacts)
		if err != nil {
			log.Info(err.Error())
			return err
		}
		if len(mergedArtifacts) > 0 {
			for key, value := range payload.Artifacts {
				mergedArtifacts[key] = value
			}
		} else {
			mergedArtifacts = payload.Artifacts
		}

		marshalledReArtifacts, err := json.Marshal(mergedArtifacts)
		if err != nil {
			log.InfoWithFields(err.Error(), log.Fields{})
			return err
		}

		releaseExtension.State = payload.State
		releaseExtension.StateMessage = payload.StateMessage
		releaseExtension.Artifacts = postgres.Jsonb{marshalledReArtifacts}
		x.Db.Save(&releaseExtension)

		mergedArtifacts = make(map[string]string)
		err = json.Unmarshal(release.Artifacts.RawMessage, &mergedArtifacts)
		if err != nil {
			log.Info(err.Error())
			return err
		}
		if len(mergedArtifacts) > 0 {
			for key, value := range payload.Artifacts {
				mergedArtifacts[key] = value
			}
		} else {
			mergedArtifacts = payload.Artifacts
		}

		marshalledArtifacts, err := json.Marshal(mergedArtifacts)
		if err != nil {
			log.InfoWithFields(err.Error(), log.Fields{})
			return err
		}
		release.Artifacts = postgres.Jsonb{marshalledArtifacts}
		x.Db.Save(&release)

		if payload.State == plugins.GetState("complete") {
			x.Actions.ReleaseExtensionCompleted(&releaseExtension)
		}

		if payload.State == plugins.GetState("failed") {
			release.State = plugins.GetState("failed")
			release.StateMessage = payload.StateMessage
			x.Db.Save(&release)
		}
	}

	return nil
}

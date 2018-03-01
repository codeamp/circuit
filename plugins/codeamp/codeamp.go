package codeamp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/codeamp/circuit/assets"
	"github.com/codeamp/circuit/plugins"
	resolvers "github.com/codeamp/circuit/plugins/codeamp/resolvers"
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
	uuid "github.com/satori/go.uuid"
	sioredis "github.com/satyakb/go-socket.io-redis"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	transistor.RegisterPlugin("codeamp", func() transistor.Plugin { return NewCodeAmp() })
}

type CodeAmp struct {
	ServiceAddress string `mapstructure:"service_address"`
	Events         chan transistor.Event
	Schema         *graphql.Schema
	SocketIO       *socketio.Server
	DB             *gorm.DB
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

	bytes, err := bcrypt.GenerateFromPassword([]byte("password"), 14)
	if err != nil {
		log.Fatal(err)
	}

	user := resolvers.User{
		Email:    "admin@codeamp.com",
		Password: string(bytes),
	}

	db.FirstOrInit(&user, resolvers.User{
		Email: "admin@codeamp.com",
	})
	db.Save(&user)

	userPermission := resolvers.UserPermission{
		UserID: user.Model.ID,
		Value:  "admin",
	}
	db.FirstOrInit(&userPermission, userPermission)
	db.Save(&userPermission)

	developmentEnv := resolvers.Environment{
		Name: "development",
	}
	db.FirstOrInit(&developmentEnv, developmentEnv)
	db.Save(&developmentEnv)

	productionEnv := resolvers.Environment{
		Name: "production",
	}
	db.FirstOrInit(&productionEnv, productionEnv)
	db.Save(&productionEnv)

	// hosted zone id
	// hosted zone name
	hostedZoneID := resolvers.Secret{
		Key:           "HOSTED_ZONE_ID",
		Type:          "env",
		Scope:         resolvers.GetSecretScope("extension"),
		EnvironmentID: productionEnv.Model.ID,
	}
	db.FirstOrInit(&hostedZoneID, hostedZoneID)
	db.Save(&hostedZoneID)

	eValue := resolvers.SecretValue{
		SecretID: hostedZoneID.Model.ID,
		UserID:   user.Model.ID,
		Value:    "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	hostedZoneIDDev := resolvers.Secret{
		Key:           "HOSTED_ZONE_ID",
		Type:          "env",
		Scope:         resolvers.GetSecretScope("extension"),
		EnvironmentID: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&hostedZoneIDDev, hostedZoneIDDev)
	db.Save(&hostedZoneIDDev)
	eValue = resolvers.SecretValue{
		SecretID: hostedZoneIDDev.Model.ID,
		UserID:   user.Model.ID,
		Value:    "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// hosted zone name
	hostedZone := resolvers.Secret{
		Key:           "HOSTED_ZONE_NAME",
		Type:          "env",
		Scope:         resolvers.GetSecretScope("extension"),
		EnvironmentID: productionEnv.Model.ID,
	}
	db.FirstOrInit(&hostedZone, hostedZone)
	db.Save(&hostedZone)
	eValue = resolvers.SecretValue{
		SecretID: hostedZone.Model.ID,
		UserID:   user.Model.ID,
		Value:    "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	hostedZoneDev := resolvers.Secret{
		Key:           "HOSTED_ZONE_NAME",
		Type:          "env",
		Scope:         resolvers.GetSecretScope("extension"),
		EnvironmentID: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&hostedZoneDev, hostedZoneDev)
	db.Save(&hostedZoneDev)
	eValue = resolvers.SecretValue{
		SecretID: hostedZoneDev.Model.ID,
		UserID:   user.Model.ID,
		Value:    "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// aws secret key
	awsSecretKey := resolvers.Secret{
		Key:           "AWS_SECRET_KEY",
		Type:          "env",
		Scope:         resolvers.GetSecretScope("extension"),
		EnvironmentID: productionEnv.Model.ID,
		IsSecret:      true,
	}
	db.FirstOrInit(&awsSecretKey, awsSecretKey)
	db.Save(&awsSecretKey)
	eValue = resolvers.SecretValue{
		SecretID: awsSecretKey.Model.ID,
		UserID:   user.Model.ID,
		Value:    "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	awsSecretKeyDev := resolvers.Secret{
		Key:           "AWS_SECRET_KEY",
		Type:          "env",
		Scope:         resolvers.GetSecretScope("extension"),
		EnvironmentID: developmentEnv.Model.ID,
		IsSecret:      true,
	}
	db.FirstOrInit(&awsSecretKeyDev, awsSecretKeyDev)
	db.Save(&awsSecretKeyDev)
	eValue = resolvers.SecretValue{
		SecretID: awsSecretKeyDev.Model.ID,
		UserID:   user.Model.ID,
		Value:    "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// aws access key
	awsAccessKey := resolvers.Secret{
		Key:           "AWS_ACCESS_KEY_ID",
		Type:          "env",
		Scope:         resolvers.GetSecretScope("extension"),
		EnvironmentID: productionEnv.Model.ID,
	}
	db.FirstOrInit(&awsAccessKey, awsAccessKey)
	db.Save(&awsAccessKey)
	eValue = resolvers.SecretValue{
		SecretID: awsAccessKey.Model.ID,
		UserID:   user.Model.ID,
		Value:    "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	awsAccessKeyDev := resolvers.Secret{
		Key:           "AWS_ACCESS_KEY_ID",
		Type:          "env",
		Scope:         resolvers.GetSecretScope("extension"),
		EnvironmentID: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&awsAccessKeyDev, awsAccessKeyDev)
	db.Save(&awsAccessKeyDev)
	eValue = resolvers.SecretValue{
		SecretID: awsAccessKeyDev.Model.ID,
		UserID:   user.Model.ID,
		Value:    "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// docker org
	dockerOrg := resolvers.Secret{
		Key:           "DOCKER_ORG",
		Type:          "env",
		Scope:         resolvers.GetSecretScope("extension"),
		EnvironmentID: productionEnv.Model.ID,
	}
	db.FirstOrInit(&dockerOrg, dockerOrg)
	db.Save(&dockerOrg)
	eValue = resolvers.SecretValue{
		SecretID: dockerOrg.Model.ID,
		UserID:   user.Model.ID,
		Value:    "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	dockerOrgDev := resolvers.Secret{
		Key:           "DOCKER_ORG",
		Type:          "env",
		Scope:         resolvers.GetSecretScope("extension"),
		EnvironmentID: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&dockerOrgDev, dockerOrgDev)
	db.Save(&dockerOrgDev)
	eValue = resolvers.SecretValue{
		SecretID: dockerOrgDev.Model.ID,
		UserID:   user.Model.ID,
		Value:    "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// docker host
	dockerHost := resolvers.Secret{
		Key:           "DOCKER_HOST",
		Type:          "env",
		Scope:         resolvers.GetSecretScope("extension"),
		EnvironmentID: productionEnv.Model.ID,
	}
	db.FirstOrInit(&dockerHost, dockerHost)
	db.Save(&dockerHost)
	eValue = resolvers.SecretValue{
		SecretID: dockerHost.Model.ID,
		UserID:   user.Model.ID,
		Value:    "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	dockerHostDev := resolvers.Secret{
		Key:           "DOCKER_HOST",
		Type:          "env",
		Scope:         resolvers.GetSecretScope("extension"),
		EnvironmentID: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&dockerHostDev, dockerHostDev)
	db.Save(&dockerHostDev)
	eValue = resolvers.SecretValue{
		SecretID: dockerHostDev.Model.ID,
		UserID:   user.Model.ID,
		Value:    "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// docker user
	dockerUser := resolvers.Secret{
		Key:           "DOCKER_USER",
		Type:          "env",
		Scope:         resolvers.GetSecretScope("extension"),
		EnvironmentID: productionEnv.Model.ID,
	}
	db.FirstOrInit(&dockerUser, dockerUser)
	db.Save(&dockerUser)
	eValue = resolvers.SecretValue{
		SecretID: dockerUser.Model.ID,
		UserID:   user.Model.ID,
		Value:    "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// docker user
	dockerUserDev := resolvers.Secret{
		Key:           "DOCKER_USER",
		Type:          "env",
		Scope:         resolvers.GetSecretScope("extension"),
		EnvironmentID: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&dockerUserDev, dockerUserDev)
	db.Save(&dockerUserDev)
	eValue = resolvers.SecretValue{
		SecretID: dockerUserDev.Model.ID,
		UserID:   user.Model.ID,
		Value:    "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// docker email
	dockerEmail := resolvers.Secret{
		Key:           "DOCKER_EMAIL",
		Type:          "env",
		Scope:         resolvers.GetSecretScope("extension"),
		EnvironmentID: productionEnv.Model.ID,
	}
	db.FirstOrInit(&dockerEmail, dockerEmail)
	db.Save(&dockerEmail)
	eValue = resolvers.SecretValue{
		SecretID: dockerEmail.Model.ID,
		UserID:   user.Model.ID,
		Value:    "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// docker email
	dockerEmailDev := resolvers.Secret{
		Key:           "DOCKER_EMAIL",
		Type:          "env",
		Scope:         resolvers.GetSecretScope("extension"),
		EnvironmentID: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&dockerEmailDev, dockerEmailDev)
	db.Save(&dockerEmailDev)
	eValue = resolvers.SecretValue{
		SecretID: dockerEmailDev.Model.ID,
		UserID:   user.Model.ID,
		Value:    "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// docker pass
	dockerPass := resolvers.Secret{
		Key:           "DOCKER_PASS",
		Type:          "env",
		Scope:         resolvers.GetSecretScope("extension"),
		EnvironmentID: productionEnv.Model.ID,
		IsSecret:      true,
	}
	db.FirstOrInit(&dockerPass, dockerPass)
	db.Save(&dockerPass)
	eValue = resolvers.SecretValue{
		SecretID: dockerPass.Model.ID,
		UserID:   user.Model.ID,
		Value:    "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// docker pass
	dockerPassDev := resolvers.Secret{
		Key:           "DOCKER_PASS",
		Type:          "env",
		Scope:         resolvers.GetSecretScope("extension"),
		EnvironmentID: developmentEnv.Model.ID,
		IsSecret:      true,
	}
	db.FirstOrInit(&dockerPassDev, dockerPassDev)
	db.Save(&dockerPassDev)
	eValue = resolvers.SecretValue{
		SecretID: dockerPassDev.Model.ID,
		UserID:   user.Model.ID,
		Value:    "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// s3 bucket
	s3Bucket := resolvers.Secret{
		Key:           "S3_BUCKET",
		Type:          "env",
		Scope:         resolvers.GetSecretScope("extension"),
		EnvironmentID: productionEnv.Model.ID,
	}
	db.FirstOrInit(&s3Bucket, s3Bucket)
	db.Save(&s3Bucket)
	eValue = resolvers.SecretValue{
		SecretID: s3Bucket.Model.ID,
		UserID:   user.Model.ID,
		Value:    "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	s3BucketDev := resolvers.Secret{
		Key:           "S3_BUCKET",
		Type:          "env",
		Scope:         resolvers.GetSecretScope("extension"),
		EnvironmentID: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&s3BucketDev, s3BucketDev)
	db.Save(&s3BucketDev)
	eValue = resolvers.SecretValue{
		SecretID: s3BucketDev.Model.ID,
		UserID:   user.Model.ID,
		Value:    "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// ssl arn
	sslArn := resolvers.Secret{
		Key:           "SSL_ARN",
		Type:          "env",
		Scope:         resolvers.GetSecretScope("extension"),
		EnvironmentID: productionEnv.Model.ID,
	}
	db.FirstOrInit(&sslArn, sslArn)
	db.Save(&sslArn)
	eValue = resolvers.SecretValue{
		SecretID: sslArn.Model.ID,
		UserID:   user.Model.ID,
		Value:    "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	sslArnDev := resolvers.Secret{
		Key:           "SSL_ARN",
		Type:          "env",
		Scope:         resolvers.GetSecretScope("extension"),
		EnvironmentID: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&sslArnDev, sslArnDev)
	db.Save(&sslArnDev)
	eValue = resolvers.SecretValue{
		SecretID: sslArnDev.Model.ID,
		UserID:   user.Model.ID,
		Value:    "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	certificateAuthorityDev := resolvers.Secret{
		Key:           "CERTIFICATE_AUTHORITY",
		Type:          "file",
		Scope:         resolvers.GetSecretScope("extension"),
		EnvironmentID: developmentEnv.Model.ID,
		IsSecret:      true,
	}
	db.FirstOrInit(&certificateAuthorityDev, certificateAuthorityDev)
	db.Save(&certificateAuthorityDev)
	eValue = resolvers.SecretValue{
		SecretID: certificateAuthorityDev.Model.ID,
		UserID:   user.Model.ID,
		Value:    "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	certificateAuthorityProd := resolvers.Secret{
		Key:           "CERTIFICATE_AUTHORITY",
		Type:          "file",
		Scope:         resolvers.GetSecretScope("extension"),
		EnvironmentID: productionEnv.Model.ID,
		IsSecret:      true,
	}
	db.FirstOrInit(&certificateAuthorityProd, certificateAuthorityProd)
	db.Save(&certificateAuthorityProd)
	eValue = resolvers.SecretValue{
		SecretID: certificateAuthorityProd.Model.ID,
		UserID:   user.Model.ID,
		Value:    "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	clientKeyDev := resolvers.Secret{
		Key:           "CLIENT_KEY",
		Type:          "file",
		Scope:         resolvers.GetSecretScope("extension"),
		EnvironmentID: developmentEnv.Model.ID,
		IsSecret:      true,
	}

	db.FirstOrInit(&clientKeyDev, clientKeyDev)
	db.Save(&clientKeyDev)
	eValue = resolvers.SecretValue{
		SecretID: clientKeyDev.Model.ID,
		UserID:   user.Model.ID,
		Value:    "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	clientKeyProd := resolvers.Secret{
		Key:           "CLIENT_KEY",
		Type:          "file",
		Scope:         resolvers.GetSecretScope("extension"),
		EnvironmentID: productionEnv.Model.ID,
		IsSecret:      true,
	}
	db.FirstOrInit(&clientKeyProd, clientKeyProd)
	db.Save(&clientKeyProd)
	eValue = resolvers.SecretValue{
		SecretID: clientKeyProd.Model.ID,
		UserID:   user.Model.ID,
		Value:    "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	clientCertDev := resolvers.Secret{
		Key:           "CLIENT_CERTIFICATE",
		Type:          "file",
		Scope:         resolvers.GetSecretScope("extension"),
		EnvironmentID: developmentEnv.Model.ID,
		IsSecret:      true,
	}

	db.FirstOrInit(&clientCertDev, clientCertDev)
	db.Save(&clientCertDev)
	eValue = resolvers.SecretValue{
		SecretID: clientCertDev.Model.ID,
		UserID:   user.Model.ID,
		Value:    "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	clientCertProd := resolvers.Secret{
		Key:           "CLIENT_CERTIFICATE",
		Type:          "file",
		Scope:         resolvers.GetSecretScope("extension"),
		EnvironmentID: productionEnv.Model.ID,
		IsSecret:      true,
	}

	db.FirstOrInit(&clientCertProd, clientCertProd)
	db.Save(&clientCertProd)
	eValue = resolvers.SecretValue{
		SecretID: clientCertProd.Model.ID,
		UserID:   user.Model.ID,
		Value:    "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// kubeconfig
	kubeConfig := resolvers.Secret{
		Key:           "KUBECONFIG",
		Type:          "file",
		Scope:         resolvers.GetSecretScope("extension"),
		EnvironmentID: productionEnv.Model.ID,
	}
	db.FirstOrInit(&kubeConfig, kubeConfig)
	db.Save(&kubeConfig)
	eValue = resolvers.SecretValue{
		SecretID: kubeConfig.Model.ID,
		UserID:   user.Model.ID,
		Value:    "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	kubeConfigDev := resolvers.Secret{
		Key:           "KUBECONFIG",
		Type:          "file",
		Scope:         resolvers.GetSecretScope("extension"),
		EnvironmentID: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&kubeConfigDev, kubeConfigDev)
	db.Save(&kubeConfigDev)
	eValue = resolvers.SecretValue{
		SecretID: kubeConfigDev.Model.ID,
		UserID:   user.Model.ID,
		Value:    "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	serviceSpec := resolvers.ServiceSpec{
		Name:                   "default",
		CpuRequest:             "500",
		CpuLimit:               "500",
		MemoryRequest:          "500",
		MemoryLimit:            "500",
		TerminationGracePeriod: "300",
	}
	db.FirstOrInit(&serviceSpec, resolvers.ServiceSpec{
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
	ext := resolvers.Extension{
		Type:          plugins.GetType("workflow"),
		Key:           "dockerbuilder",
		Name:          "Docker Builder",
		Component:     "",
		EnvironmentID: productionEnv.Model.ID,
		Config:        postgres.Jsonb{marshalledDbConfig},
	}
	db.FirstOrInit(&ext, ext)
	db.Save(&ext)

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
	ext = resolvers.Extension{
		Type:          plugins.GetType("workflow"),
		Key:           "dockerbuilder",
		Name:          "Docker Builder",
		Component:     "",
		EnvironmentID: developmentEnv.Model.ID,
		Config:        postgres.Jsonb{marshalledDbConfig},
	}

	db.FirstOrInit(&ext, ext)
	db.Save(&ext)

	// load balancer
	lbConfig := []map[string]interface{}{
		map[string]interface{}{"key": "KUBECONFIG", "value": kubeConfig.Model.ID.String()},
		map[string]interface{}{"key": "SSL_CERT_ARN", "value": sslArn.Model.ID.String()},
		map[string]interface{}{"key": "ACCESS_LOG_S3_BUCKET", "value": s3Bucket.Model.ID.String()},
		map[string]interface{}{"key": "HOSTED_ZONE_ID", "value": hostedZoneID.Model.ID.String()},
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
	ext = resolvers.Extension{
		Type:          plugins.GetType("once"),
		Key:           "kubernetesloadbalancers",
		Name:          "Load Balancer",
		Component:     "LoadBalancer",
		EnvironmentID: productionEnv.Model.ID,
		Config:        postgres.Jsonb{marshalledLbConfig},
	}

	db.FirstOrInit(&ext, ext)
	db.Save(&ext)

	lbConfig = []map[string]interface{}{
		map[string]interface{}{"key": "KUBECONFIG", "value": kubeConfigDev.Model.ID.String()},
		map[string]interface{}{"key": "SSL_CERT_ARN", "value": sslArnDev.Model.ID.String()},
		map[string]interface{}{"key": "ACCESS_LOG_S3_BUCKET", "value": s3BucketDev.Model.ID.String()},
		map[string]interface{}{"key": "HOSTED_ZONE_ID", "value": hostedZoneIDDev.Model.ID.String()},
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
	ext = resolvers.Extension{
		Type:          plugins.GetType("once"),
		Key:           "kubernetesloadbalancers",
		Name:          "Load Balancer",
		Component:     "LoadBalancer",
		EnvironmentID: developmentEnv.Model.ID,
		Config:        postgres.Jsonb{marshalledLbConfig},
	}

	db.FirstOrInit(&ext, ext)
	db.Save(&ext)

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
	ext = resolvers.Extension{
		Type:          plugins.GetType("deployment"),
		Key:           "kubernetesdeployments",
		Name:          "Kubernetes",
		Component:     "",
		EnvironmentID: productionEnv.Model.ID,
		Config:        postgres.Jsonb{marshalledKubeConfig},
	}

	db.FirstOrInit(&ext, ext)
	db.Save(&ext)

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
	ext = resolvers.Extension{
		Type:          plugins.GetType("deployment"),
		Key:           "kubernetesdeployments",
		Name:          "Kubernetes",
		Component:     "",
		EnvironmentID: developmentEnv.Model.ID,
		Config:        postgres.Jsonb{marshalledKubeConfig},
	}

	db.FirstOrInit(&ext, ext)
	db.Save(&ext)

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

	r := &resolvers.Resolver{DB: x.DB, Events: x.Events, Redis: x.Redis}
	http.Handle("/query", resolvers.CorsMiddleware(r.AuthMiddleware(&relay.Handler{Schema: x.Schema})))

	log.Info(fmt.Sprintf("running GraphQL server on %v", x.ServiceAddress))
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s", x.ServiceAddress), handlers.LoggingHandler(os.Stdout, http.DefaultServeMux)))
}

func (x *CodeAmp) Start(events chan transistor.Event) error {
	var err error

	db, err := gorm.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s password=%s",
		viper.GetString("plugins.codeamp.postgres.host"),
		viper.GetString("plugins.codeamp.postgres.port"),
		viper.GetString("plugins.codeamp.postgres.user"),
		viper.GetString("plugins.codeamp.postgres.dbname"),
		viper.GetString("plugins.codeamp.postgres.sslmode"),
		viper.GetString("plugins.codeamp.postgres.password"),
	))
	//defer x.DB.Close()

	schema, err := assets.Asset("plugins/codeamp/schema.graphql")
	if err != nil {
		log.Panic(err)
	}

	parsedSchema, err := graphql.ParseSchema(string(schema), &resolvers.Resolver{DB: db, Events: events})
	if err != nil {
		log.Panic(err)
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

	x.Events = events
	x.SocketIO = sio
	x.Schema = parsedSchema
	x.Redis = redisClient

	// DEBUG
	db.LogMode(false)

	x.DB = db

	go x.Listen()

	return nil
}

func (x *CodeAmp) Stop() {
	log.Info("stopping CodeAmp service")
}

func (x *CodeAmp) Subscribe() []string {
	return []string{
		"plugins.GitCommit",
		"plugins.HeartBeat",
		"plugins.WebsocketMsg",
		"plugins.ProjectExtension:status",
		"plugins.ProjectExtension:update",
		"plugins.ProjectExtension:complete",
		"plugins.ReleaseExtension:status",
		"plugins.Release:status",
		"plugins.Release:complete",
	}
}

func (x *CodeAmp) Process(e transistor.Event) error {
	log.InfoWithFields("process CodeAmp event", log.Fields{
		"event_name": e.Name,
	})

	methodName := fmt.Sprintf("%sEventHandler", strings.Split(e.PayloadModel, ".")[1])

	if _, ok := reflect.TypeOf(x).MethodByName(methodName); ok {
		fmt.Println(reflect.ValueOf(x).MethodByName(methodName).Call([]reflect.Value{reflect.ValueOf(e)}))
	} else {
		log.InfoWithFields("*EventHandler not implemented", log.Fields{
			"event_name":    e.Name,
			"payload_model": e.PayloadModel,
			"method_name":   methodName,
		})
	}

	return nil
}

func (x *CodeAmp) HeartBeatEventHandler(e transistor.Event) {
	payload := e.Payload.(plugins.HeartBeat)

	var projects []resolvers.Project

	x.DB.Find(&projects)
	for _, project := range projects {
		switch payload.Tick {
		case "minute":
			x.GitSync(&project)
		}
	}
}

func (x *CodeAmp) GitCommitEventHandler(e transistor.Event) error {
	payload := e.Payload.(plugins.GitCommit)

	var project resolvers.Project
	var feature resolvers.Feature

	if x.DB.Where("repository = ?", payload.Repository).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"repository": payload.Repository,
		})
		return nil
	}

	if x.DB.Where("project_id = ? AND hash = ?", project.ID, payload.Hash).First(&feature).RecordNotFound() {
		feature = resolvers.Feature{
			ProjectID:  project.ID,
			Message:    payload.Message,
			User:       payload.User,
			Hash:       payload.Hash,
			ParentHash: payload.ParentHash,
			Ref:        payload.Ref,
			Created:    payload.Created,
		}
		x.DB.Save(&feature)
	} else {
		log.InfoWithFields("feature already exists", log.Fields{
			"repository": payload.Repository,
			"hash":       payload.Hash,
		})
	}

	x.Events <- transistor.NewEvent(plugins.WebsocketMsg{
		Event: fmt.Sprintf("projects/%s/features", project.Slug),
	}, nil)

	return nil
}

func (x *CodeAmp) WebsocketMsgEventHandler(e transistor.Event) error {
	payload := e.Payload.(plugins.WebsocketMsg)

	if payload.Channel == "" {
		payload.Channel = "general"
	}

	x.SocketIO.BroadcastTo(payload.Channel, payload.Event, payload.Payload, nil)

	return nil
}

func (x *CodeAmp) ProjectExtensionEventHandler(e transistor.Event) error {
	payload := e.Payload.(plugins.ProjectExtension)
	var extension resolvers.ProjectExtension
	var project resolvers.Project

	if e.Matches("plugins.ProjectExtension:status") {
		if x.DB.Where("id = ?", payload.ID).Find(&extension).RecordNotFound() {
			log.InfoWithFields("extension not found", log.Fields{
				"id": payload.ID,
			})
			return fmt.Errorf(fmt.Sprintf("Could not handle ProjectExtension status event because ProjectExtension not found given payload id: %s.", payload.ID))
		}

		if x.DB.Where("id = ?", extension.ProjectID).Find(&project).RecordNotFound() {
			log.InfoWithFields("project not found", log.Fields{
				"id": extension.ProjectID,
			})
			return fmt.Errorf(fmt.Sprintf("Could not handle ProjectExtension status event because Project not found given payload id: %s.", extension.ProjectID))
		}

		extension.State = payload.State
		extension.StateMessage = payload.StateMessage
		//extension.Artifacts = postgres.Jsonb{marshalledArtifacts}
		x.DB.Save(&extension)

		x.Events <- transistor.NewEvent(plugins.WebsocketMsg{
			Event:   fmt.Sprintf("projects/%s/extensions", project.Slug),
			Payload: extension,
		}, nil)

		if payload.State == plugins.GetState("complete") {
			// TODO: handle extension init complete
		}
	}

	return nil
}

func (x *CodeAmp) ReleaseExtensionEventHandler(e transistor.Event) error {
	payload := e.Payload.(plugins.ReleaseExtension)

	var releaseExtension resolvers.ReleaseExtension
	var release resolvers.Release

	if e.Matches("plugins.ReleaseExtension:status") {
		if x.DB.Where("id = ?", payload.Release.ID).Find(&release).RecordNotFound() {
			log.InfoWithFields("release", log.Fields{
				"id": payload.Release.ID,
			})
			return nil
		}

		if x.DB.Where("id = ?", payload.ID).Find(&releaseExtension).RecordNotFound() {
			log.InfoWithFields("release extension not found", log.Fields{
				"id": payload.ID,
			})
			return nil
		}

		releaseExtension.State = payload.State
		releaseExtension.StateMessage = payload.StateMessage
		//releaseExtension.Artifacts = postgres.Jsonb{marshalledReArtifacts}
		x.DB.Save(&releaseExtension)
		x.DB.Save(&release)

		if payload.State == plugins.GetState("complete") {
			x.ReleaseExtensionCompleted(&releaseExtension)
		}

		if payload.State == plugins.GetState("failed") {
			release.State = plugins.GetState("failed")
			release.StateMessage = payload.StateMessage
			x.DB.Save(&release)
		}
	}

	return nil
}

func (x *CodeAmp) GitSync(project *resolvers.Project) error {
	var feature resolvers.Feature
	var release resolvers.Release
	var headFeature resolvers.Feature
	hash := ""

	// Get latest release and deployed feature hash
	if x.DB.Where("project_id = ?", project.ID).Order("created_at DESC").First(&release).RecordNotFound() {
		// get latest feature if there is no releases
		x.DB.Where("project_id = ?", project.ID).Order("created_at DESC").First(&feature)
		hash = feature.Hash
	} else {
		if x.DB.Where("id = ?", release.HeadFeatureID).Find(&headFeature).RecordNotFound() {
			log.InfoWithFields("can not find head feature", log.Fields{
				"id": release.HeadFeatureID,
			})
		}
		hash = headFeature.Hash
	}

	// get branches of entire environments
	projectSettingsCollection := []resolvers.ProjectSettings{}
	if x.DB.Where("project_id = ?", project.Model.ID.String()).Find(&projectSettingsCollection).RecordNotFound() {
		gitSync := plugins.GitSync{
			Action: plugins.GetAction("update"),
			State:  plugins.GetState("waiting"),
			Project: plugins.Project{
				ID:         project.Model.ID.String(),
				Repository: project.Repository,
			},
			Git: plugins.Git{
				Url:           project.GitUrl,
				Protocol:      project.GitProtocol,
				Branch:        "master",
				RsaPrivateKey: project.RsaPrivateKey,
				RsaPublicKey:  project.RsaPublicKey,
			},
			From: hash,
		}

		x.Events <- transistor.NewEvent(gitSync, nil)
	} else {
		for _, projectSettings := range projectSettingsCollection {
			gitSync := plugins.GitSync{
				Action: plugins.GetAction("update"),
				State:  plugins.GetState("waiting"),
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

			x.Events <- transistor.NewEvent(gitSync, nil)
		}
	}

	return nil
}

func (x *CodeAmp) ReleaseExtensionCompleted(re *resolvers.ReleaseExtension) {
	project := resolvers.Project{}
	release := resolvers.Release{}
	releaseExtensions := []resolvers.ReleaseExtension{}

	if x.DB.Where("id = ?", re.ReleaseID).First(&release).RecordNotFound() {
		log.InfoWithFields("release not found", log.Fields{
			"releaseExtension": re,
		})
		return
	}

	if x.DB.Where("id = ?", release.ProjectID).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"release": release,
		})
		return
	}

	if x.DB.Where("release_id = ?", re.ReleaseID).Find(&releaseExtensions).RecordNotFound() {
		log.InfoWithFields("release extensions not found", log.Fields{
			"releaseExtension": re,
		})
		return
	}

	x.Events <- transistor.NewEvent(plugins.WebsocketMsg{
		Event:   fmt.Sprintf("projects/%s/releases/reCompleted", project.Slug),
		Payload: release,
	}, nil)

	// loop through and check if all same-type elease extensions are completed
	done := true
	for _, fre := range releaseExtensions {
		if fre.Type == re.Type && fre.State != plugins.GetState("complete") {
			done = false
		}
	}

	if done {
		switch re.Type {
		case plugins.GetType("workflow"):
			x.WorkflowProjectExtensionsCompleted(&release)
		case plugins.GetType("deployment"):
			x.DeploymentProjectExtensionsCompleted(&release)
		}
	}
}

func (x *CodeAmp) WorkflowProjectExtensionsCompleted(release *resolvers.Release) {
	// find all related deployment extensions
	depProjectExtensions := []resolvers.ProjectExtension{}
	aggregateReleaseExtensionArtifacts := make(map[string]interface{})
	found := false

	if x.DB.Where("project_id = ? and environment_id = ?", release.ProjectID, release.EnvironmentID).Find(&depProjectExtensions).RecordNotFound() {
		log.InfoWithFields("deployment extensions not found", log.Fields{
			"release": release,
		})
		return
	}

	for _, de := range depProjectExtensions {
		var ext resolvers.Extension
		if x.DB.Where("id = ?", de.ExtensionID).First(&ext).RecordNotFound() {
			log.InfoWithFields("extension spec not found", log.Fields{
				"extension spec": de,
			})
		}
		if plugins.Type(ext.Type) == plugins.GetType("workflow") {
			releaseExtension := resolvers.ReleaseExtension{}

			if x.DB.Where("release_id = ? AND extension_id = ? AND state = ?", release.Model.ID, de.Model.ID, string(plugins.GetState("complete"))).Find(&releaseExtension).RecordNotFound() {
				log.InfoWithFields("release extension not found", log.Fields{
					"release_id":   release.Model.ID,
					"extension_id": de.Model.ID,
					"state":        plugins.GetState("complete"),
				})
			}

			// put all releaseextension artifacts inside release artifacts
			unmarshalledArtifacts := make(map[string]interface{})
			err := json.Unmarshal(releaseExtension.Artifacts.RawMessage, &unmarshalledArtifacts)
			if err != nil {
				log.InfoWithFields(err.Error(), log.Fields{})
				return
			}

			for k, v := range unmarshalledArtifacts {
				key := fmt.Sprintf("%s_%s", strings.ToUpper(ext.Key), strings.ToUpper(k))
				aggregateReleaseExtensionArtifacts[key] = v
			}
		}

		if plugins.Type(ext.Type) == plugins.GetType("deployment") {
			found = true
		}
	}

	// persist workflow artifacts
	// release.Artifacts = plugins.MapStringStringToHstore(releaseExtensionArtifacts)
	x.DB.Save(release)

	// if there are no deployment workflows, then release is complete
	if !found {
		x.ReleaseCompleted(release)
	}

	project := resolvers.Project{}

	if x.DB.Where("id = ?", release.ProjectID).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"release": release,
		})
	}

	services := []resolvers.Service{}
	if x.DB.Where("project_id = ?", release.ProjectID).Find(&services).RecordNotFound() {
		log.InfoWithFields("no services found for this project", log.Fields{
			"release": release,
		})
	}

	headFeature := resolvers.Feature{}
	if x.DB.Where("id = ?", release.HeadFeatureID).First(&headFeature).RecordNotFound() {
		log.InfoWithFields("head feature not found", log.Fields{
			"id": release.HeadFeatureID,
		})
		return
	}

	tailFeature := resolvers.Feature{}
	if x.DB.Where("id = ?", release.TailFeatureID).First(&tailFeature).RecordNotFound() {
		log.InfoWithFields("tail feature not found", log.Fields{
			"id": release.TailFeatureID,
		})
		return
	}

	environment := resolvers.Environment{}
	if x.DB.Where("id = ?", release.EnvironmentID).First(&environment).RecordNotFound() {
		log.InfoWithFields("environment not found", log.Fields{
			"id": release.EnvironmentID,
		})
		return
	}

	// get all branches relevant for the projec
	branch := "master"
	projectSettings := resolvers.ProjectSettings{}
	if x.DB.Where("environment_id = ? and project_id = ?", environment.Model.ID.String(),
		project.Model.ID.String()).First(&projectSettings).RecordNotFound() {
		log.InfoWithFields("no env project branch found", log.Fields{})
	} else {
		branch = projectSettings.GitBranch
	}

	releaseEvent := plugins.Release{
		Action:       plugins.GetAction("create"),
		State:        plugins.GetState("waiting"),
		Environment:  environment.Name,
		StateMessage: "create release event",
		ID:           release.Model.ID.String(),
		HeadFeature: plugins.Feature{
			ID:         headFeature.Model.ID.String(),
			Hash:       headFeature.Hash,
			ParentHash: headFeature.ParentHash,
			User:       headFeature.User,
			Message:    headFeature.Message,
			Created:    headFeature.Created,
		},
		TailFeature: plugins.Feature{
			ID:         tailFeature.Model.ID.String(),
			Hash:       tailFeature.Hash,
			ParentHash: tailFeature.ParentHash,
			User:       tailFeature.User,
			Message:    tailFeature.Message,
			Created:    tailFeature.Created,
		},
		User: "",
		Project: plugins.Project{
			ID:         project.Model.ID.String(),
			Slug:       project.Slug,
			Repository: project.Repository,
		},
		Git: plugins.Git{
			Url:           project.GitUrl,
			Branch:        branch,
			RsaPrivateKey: project.RsaPrivateKey,
		},
	}
	releaseExtensionEvents := []plugins.ReleaseExtension{}

	for _, extension := range depProjectExtensions {
		ext := resolvers.Extension{}
		if x.DB.Where("id= ?", extension.ExtensionID).Find(&ext).RecordNotFound() {
			log.InfoWithFields("extension spec not found", log.Fields{
				"extension": extension,
			})
		}

		if plugins.Type(ext.Type) == plugins.GetType("workflow") {
			releaseExtension := resolvers.ReleaseExtension{}

			if x.DB.Where("release_id = ? AND extension_id = ? AND state = ?", release.Model.ID, extension.Model.ID, plugins.GetState("complete")).Find(&releaseExtension).RecordNotFound() {
				log.InfoWithFields("release extension not found", log.Fields{
					"release_id":   release.Model.ID,
					"extension_id": extension.Model.ID,
					"state":        plugins.GetState("complete"),
				})
			}
		}

		if plugins.Type(ext.Type) == plugins.GetType("deployment") {

			// create ReleaseExtension
			releaseExtension := resolvers.ReleaseExtension{
				ReleaseID:          release.Model.ID,
				FeatureHash:        "",
				ServicesSignature:  "",
				SecretsSignature:   "",
				ProjectExtensionID: extension.Model.ID,
				State:              plugins.GetState("waiting"),
				Type:               plugins.GetType("deployment"),
				StateMessage:       "initialized",
			}

			x.DB.Save(&releaseExtension)
			unmarshalledConfig := make(map[string]interface{})

			err := json.Unmarshal(extension.Config.RawMessage, &unmarshalledConfig)
			if err != nil {
				log.Info(err.Error())
			}

			config, err := x.ExtractConfig(unmarshalledConfig, ext.Key, x.DB)
			if err != nil {
				log.Info(err.Error())
			}

			releaseExtensionEvents = append(releaseExtensionEvents, plugins.ReleaseExtension{
				ID:           releaseExtension.Model.ID.String(),
				Action:       plugins.GetAction("create"),
				Slug:         ext.Key,
				State:        releaseExtension.State,
				Config:       config,
				Artifacts:    map[string]interface{}{},
				Release:      releaseEvent,
				StateMessage: releaseExtension.StateMessage,
			})

		}
	}

	// send out release extension event for each re
	for _, re := range releaseExtensionEvents {
		//re.Release.Artifacts = aggregateReleaseExtensionArtifacts
		x.Events <- transistor.NewEvent(re, nil)
	}
}

func (x *CodeAmp) DeploymentProjectExtensionsCompleted(release *resolvers.Release) {
	// find all related deployment extensions
	depProjectExtensions := []resolvers.ProjectExtension{}
	// releaseExtensionArtifacts := map[string]string{}

	if x.DB.Where("project_id = ?", release.ProjectID).Find(&depProjectExtensions).RecordNotFound() {
		log.InfoWithFields("deployment extensions not found", log.Fields{
			"release": release,
		})
		return
	}

	for _, de := range depProjectExtensions {
		var ext resolvers.Extension
		if x.DB.Where("id = ?", de.ExtensionID).First(&ext).RecordNotFound() {
			log.InfoWithFields("extension spec not found", log.Fields{
				"id": de.ExtensionID,
			})
		}

		if plugins.Type(ext.Type) == plugins.GetType("deployment") {
			releaseExtension := resolvers.ReleaseExtension{}

			if x.DB.Where("release_id = ? AND extension_id = ? AND state = ?", release.Model.ID, de.Model.ID, plugins.GetState("complete")).Find(&releaseExtension).RecordNotFound() {
				log.InfoWithFields("release extension not found", log.Fields{
					"release_id":   release.Model.ID,
					"extension_id": de.Model.ID,
					"state":        plugins.GetState("complete"),
				})
			}
		}
	}

	x.ReleaseCompleted(release)
}

func (x *CodeAmp) ReleaseCompleted(release *resolvers.Release) {
	project := resolvers.Project{}
	if x.DB.Where("id = ?", release.ProjectID).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"release": release,
		})
	}

	// mark release as complete
	release.State = plugins.GetState("complete")
	release.StateMessage = "Release completed"

	x.DB.Save(release)
}

func (x *CodeAmp) ReleaseCreated(release *resolvers.Release) {
	project := resolvers.Project{}

	if x.DB.Where("id = ?", release.ProjectID).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"release": release,
		})
		return
	}

	// loop through extensions and send ReleaseWorkflow events
	projectExtensions := []resolvers.ProjectExtension{}
	if x.DB.Where("project_id = ? and environment_id = ?", release.ProjectID, release.EnvironmentID).Find(&projectExtensions).RecordNotFound() {
		log.InfoWithFields("project has no extensions", log.Fields{
			"project_id":     release.ProjectID,
			"environment_id": release.EnvironmentID,
		})
	}

	services := []resolvers.Service{}
	if x.DB.Where("project_id = ? and environment_id = ?", release.ProjectID, release.EnvironmentID).Find(&services).RecordNotFound() {
		log.InfoWithFields("project has no services", log.Fields{
			"project_id":     release.ProjectID,
			"environment_id": release.EnvironmentID,
		})
	}

	headFeature := resolvers.Feature{}
	if x.DB.Where("id = ?", release.HeadFeatureID).First(&headFeature).RecordNotFound() {
		log.InfoWithFields("head feature not found", log.Fields{
			"id": release.HeadFeatureID,
		})
		return
	}

	tailFeature := resolvers.Feature{}
	if x.DB.Where("id = ?", release.TailFeatureID).First(&tailFeature).RecordNotFound() {
		log.InfoWithFields("tail feature not found", log.Fields{
			"id": release.TailFeatureID,
		})
		return
	}

	environment := resolvers.Environment{}
	if x.DB.Where("id = ?", release.EnvironmentID).First(&environment).RecordNotFound() {
		log.InfoWithFields("environment not found", log.Fields{
			"id": release.EnvironmentID,
		})
		return
	}

	// get all branches relevant for the projec
	branch := "master"
	projectSettings := resolvers.ProjectSettings{}
	if x.DB.Where("environment_id = ? and project_id = ?", environment.Model.ID.String(),
		project.Model.ID.String()).First(&projectSettings).RecordNotFound() {
		log.InfoWithFields("no env project branch found", log.Fields{})
	} else {
		branch = projectSettings.GitBranch
	}

	releaseEvent := plugins.Release{
		ID:          release.Model.ID.String(),
		Action:      plugins.GetAction("create"),
		State:       plugins.GetState("waiting"),
		Environment: environment.Name,
		HeadFeature: plugins.Feature{
			ID:         headFeature.Model.ID.String(),
			Hash:       headFeature.Hash,
			ParentHash: headFeature.ParentHash,
			User:       headFeature.User,
			Message:    headFeature.Message,
			Created:    headFeature.Created,
		},
		TailFeature: plugins.Feature{
			ID:         tailFeature.Model.ID.String(),
			Hash:       tailFeature.Hash,
			ParentHash: tailFeature.ParentHash,
			User:       tailFeature.User,
			Message:    tailFeature.Message,
			Created:    tailFeature.Created,
		},
		User: release.User.Email,
		Project: plugins.Project{
			ID:         project.Model.ID.String(),
			Repository: project.Repository,
		},
		Git: plugins.Git{
			Url:           project.GitUrl,
			Branch:        branch,
			RsaPrivateKey: project.RsaPrivateKey,
		},
	}
	for _, extension := range projectExtensions {
		ext := resolvers.Extension{}
		if x.DB.Where("id= ?", extension.ExtensionID).Find(&ext).RecordNotFound() {
			log.InfoWithFields("extension spec not found", log.Fields{
				"id": extension.ExtensionID,
			})
		}

		// ONLY SEND WORKFLOW TYPE, EVENTs
		if plugins.Type(ext.Type) == plugins.GetType("workflow") {
			// create ReleaseExtension
			releaseExtension := resolvers.ReleaseExtension{
				ReleaseID:          release.Model.ID,
				FeatureHash:        "",
				ServicesSignature:  "",
				SecretsSignature:   "",
				ProjectExtensionID: extension.Model.ID,
				State:              plugins.GetState("waiting"),
				Type:               plugins.GetType("workflow"),
			}

			x.DB.Save(&releaseExtension)

			unmarshalledConfig := make(map[string]interface{})

			err := json.Unmarshal(extension.Config.RawMessage, &unmarshalledConfig)
			if err != nil {
				log.Info(err.Error())
			}

			config, err := x.ExtractConfig(unmarshalledConfig, ext.Key, x.DB)
			if err != nil {
				log.Info(err.Error())
			}

			x.Events <- transistor.NewEvent(plugins.ReleaseExtension{
				ID:        releaseExtension.Model.ID.String(),
				Action:    plugins.GetAction("create"),
				Slug:      ext.Key,
				State:     releaseExtension.State,
				Release:   releaseEvent,
				Config:    config,
				Artifacts: map[string]interface{}{},
			}, nil)
		}
	}

	// send web socket message notifying release has been created
	x.Events <- transistor.NewEvent(plugins.WebsocketMsg{
		Event:   fmt.Sprintf("projects/%s/releases", project.Slug),
		Payload: release,
	}, nil)
}

/* fills in Config by querying config ids and getting the actual value */
func (x *CodeAmp) ExtractConfig(config map[string]interface{}, extKey string, db *gorm.DB) (map[string]interface{}, error) {
	c := make(map[string]interface{})

	for _, val := range config["config"].([]interface{}) {
		val := val.(map[string]interface{})
		// check if val is UUID. If so, query in environment variables for id
		secretID := uuid.FromStringOrNil(val["value"].(string))
		if secretID != uuid.Nil {
			secret := resolvers.SecretValue{}
			if db.Where("secret_id = ?", secretID).Order("created_at desc").First(&secret).RecordNotFound() {
				log.InfoWithFields("secret not found", log.Fields{
					"secret_id": secretID,
				})
			}
			c[fmt.Sprintf("%s_%s", strings.ToUpper(extKey), strings.ToUpper(val["key"].(string)))] = secret.Value
		} else {
			c[fmt.Sprintf("%s_%s", strings.ToUpper(extKey), strings.ToUpper(val["key"].(string)))] = val["value"].(string)
		}
	}

	for key, val := range config["custom"].(map[string]interface{}) {
		c[fmt.Sprintf("%s_%s", strings.ToUpper(extKey), strings.ToUpper(key))] = val
	}

	return c, nil
}

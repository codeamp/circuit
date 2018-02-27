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
		&resolvers.ExtensionSpec{},
		&resolvers.Extension{},
		&resolvers.EnvironmentVariable{},
		&resolvers.EnvironmentVariableValue{},
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
		UserId: user.Model.ID,
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
	hostedZoneId := resolvers.EnvironmentVariable{
		Key:           "HOSTED_ZONE_ID",
		Type:          "env",
		Scope:         resolvers.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
	}
	db.FirstOrInit(&hostedZoneId, hostedZoneId)
	db.Save(&hostedZoneId)

	eValue := resolvers.EnvironmentVariableValue{
		EnvironmentVariableId: hostedZoneId.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	hostedZoneIdDev := resolvers.EnvironmentVariable{
		Key:           "HOSTED_ZONE_ID",
		Type:          "env",
		Scope:         resolvers.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&hostedZoneIdDev, hostedZoneIdDev)
	db.Save(&hostedZoneIdDev)
	eValue = resolvers.EnvironmentVariableValue{
		EnvironmentVariableId: hostedZoneIdDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// hosted zone name
	hostedZone := resolvers.EnvironmentVariable{
		Key:           "HOSTED_ZONE_NAME",
		Type:          "env",
		Scope:         resolvers.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
	}
	db.FirstOrInit(&hostedZone, hostedZone)
	db.Save(&hostedZone)
	eValue = resolvers.EnvironmentVariableValue{
		EnvironmentVariableId: hostedZone.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	hostedZoneDev := resolvers.EnvironmentVariable{
		Key:           "HOSTED_ZONE_NAME",
		Type:          "env",
		Scope:         resolvers.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&hostedZoneDev, hostedZoneDev)
	db.Save(&hostedZoneDev)
	eValue = resolvers.EnvironmentVariableValue{
		EnvironmentVariableId: hostedZoneDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// aws secret key
	awsSecretKey := resolvers.EnvironmentVariable{
		Key:           "AWS_SECRET_KEY",
		Type:          "env",
		Scope:         resolvers.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
		IsSecret:      true,
	}
	db.FirstOrInit(&awsSecretKey, awsSecretKey)
	db.Save(&awsSecretKey)
	eValue = resolvers.EnvironmentVariableValue{
		EnvironmentVariableId: awsSecretKey.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	awsSecretKeyDev := resolvers.EnvironmentVariable{
		Key:           "AWS_SECRET_KEY",
		Type:          "env",
		Scope:         resolvers.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
		IsSecret:      true,
	}
	db.FirstOrInit(&awsSecretKeyDev, awsSecretKeyDev)
	db.Save(&awsSecretKeyDev)
	eValue = resolvers.EnvironmentVariableValue{
		EnvironmentVariableId: awsSecretKeyDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// aws access key
	awsAccessKey := resolvers.EnvironmentVariable{
		Key:           "AWS_ACCESS_KEY_ID",
		Type:          "env",
		Scope:         resolvers.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
	}
	db.FirstOrInit(&awsAccessKey, awsAccessKey)
	db.Save(&awsAccessKey)
	eValue = resolvers.EnvironmentVariableValue{
		EnvironmentVariableId: awsAccessKey.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	awsAccessKeyDev := resolvers.EnvironmentVariable{
		Key:           "AWS_ACCESS_KEY_ID",
		Type:          "env",
		Scope:         resolvers.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&awsAccessKeyDev, awsAccessKeyDev)
	db.Save(&awsAccessKeyDev)
	eValue = resolvers.EnvironmentVariableValue{
		EnvironmentVariableId: awsAccessKeyDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// docker org
	dockerOrg := resolvers.EnvironmentVariable{
		Key:           "DOCKER_ORG",
		Type:          "env",
		Scope:         resolvers.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
	}
	db.FirstOrInit(&dockerOrg, dockerOrg)
	db.Save(&dockerOrg)
	eValue = resolvers.EnvironmentVariableValue{
		EnvironmentVariableId: dockerOrg.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	dockerOrgDev := resolvers.EnvironmentVariable{
		Key:           "DOCKER_ORG",
		Type:          "env",
		Scope:         resolvers.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&dockerOrgDev, dockerOrgDev)
	db.Save(&dockerOrgDev)
	eValue = resolvers.EnvironmentVariableValue{
		EnvironmentVariableId: dockerOrgDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// docker host
	dockerHost := resolvers.EnvironmentVariable{
		Key:           "DOCKER_HOST",
		Type:          "env",
		Scope:         resolvers.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
	}
	db.FirstOrInit(&dockerHost, dockerHost)
	db.Save(&dockerHost)
	eValue = resolvers.EnvironmentVariableValue{
		EnvironmentVariableId: dockerHost.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	dockerHostDev := resolvers.EnvironmentVariable{
		Key:           "DOCKER_HOST",
		Type:          "env",
		Scope:         resolvers.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&dockerHostDev, dockerHostDev)
	db.Save(&dockerHostDev)
	eValue = resolvers.EnvironmentVariableValue{
		EnvironmentVariableId: dockerHostDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// docker user
	dockerUser := resolvers.EnvironmentVariable{
		Key:           "DOCKER_USER",
		Type:          "env",
		Scope:         resolvers.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
	}
	db.FirstOrInit(&dockerUser, dockerUser)
	db.Save(&dockerUser)
	eValue = resolvers.EnvironmentVariableValue{
		EnvironmentVariableId: dockerUser.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// docker user
	dockerUserDev := resolvers.EnvironmentVariable{
		Key:           "DOCKER_USER",
		Type:          "env",
		Scope:         resolvers.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&dockerUserDev, dockerUserDev)
	db.Save(&dockerUserDev)
	eValue = resolvers.EnvironmentVariableValue{
		EnvironmentVariableId: dockerUserDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// docker email
	dockerEmail := resolvers.EnvironmentVariable{
		Key:           "DOCKER_EMAIL",
		Type:          "env",
		Scope:         resolvers.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
	}
	db.FirstOrInit(&dockerEmail, dockerEmail)
	db.Save(&dockerEmail)
	eValue = resolvers.EnvironmentVariableValue{
		EnvironmentVariableId: dockerEmail.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// docker email
	dockerEmailDev := resolvers.EnvironmentVariable{
		Key:           "DOCKER_EMAIL",
		Type:          "env",
		Scope:         resolvers.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&dockerEmailDev, dockerEmailDev)
	db.Save(&dockerEmailDev)
	eValue = resolvers.EnvironmentVariableValue{
		EnvironmentVariableId: dockerEmailDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// docker pass
	dockerPass := resolvers.EnvironmentVariable{
		Key:           "DOCKER_PASS",
		Type:          "env",
		Scope:         resolvers.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
		IsSecret:      true,
	}
	db.FirstOrInit(&dockerPass, dockerPass)
	db.Save(&dockerPass)
	eValue = resolvers.EnvironmentVariableValue{
		EnvironmentVariableId: dockerPass.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// docker pass
	dockerPassDev := resolvers.EnvironmentVariable{
		Key:           "DOCKER_PASS",
		Type:          "env",
		Scope:         resolvers.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
		IsSecret:      true,
	}
	db.FirstOrInit(&dockerPassDev, dockerPassDev)
	db.Save(&dockerPassDev)
	eValue = resolvers.EnvironmentVariableValue{
		EnvironmentVariableId: dockerPassDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// s3 bucket
	s3Bucket := resolvers.EnvironmentVariable{
		Key:           "S3_BUCKET",
		Type:          "env",
		Scope:         resolvers.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
	}
	db.FirstOrInit(&s3Bucket, s3Bucket)
	db.Save(&s3Bucket)
	eValue = resolvers.EnvironmentVariableValue{
		EnvironmentVariableId: s3Bucket.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	s3BucketDev := resolvers.EnvironmentVariable{
		Key:           "S3_BUCKET",
		Type:          "env",
		Scope:         resolvers.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&s3BucketDev, s3BucketDev)
	db.Save(&s3BucketDev)
	eValue = resolvers.EnvironmentVariableValue{
		EnvironmentVariableId: s3BucketDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// ssl arn
	sslArn := resolvers.EnvironmentVariable{
		Key:           "SSL_ARN",
		Type:          "env",
		Scope:         resolvers.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
	}
	db.FirstOrInit(&sslArn, sslArn)
	db.Save(&sslArn)
	eValue = resolvers.EnvironmentVariableValue{
		EnvironmentVariableId: sslArn.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	sslArnDev := resolvers.EnvironmentVariable{
		Key:           "SSL_ARN",
		Type:          "env",
		Scope:         resolvers.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&sslArnDev, sslArnDev)
	db.Save(&sslArnDev)
	eValue = resolvers.EnvironmentVariableValue{
		EnvironmentVariableId: sslArnDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	certificateAuthorityDev := resolvers.EnvironmentVariable{
		Key:           "CERTIFICATE_AUTHORITY",
		Type:          "file",
		Scope:         resolvers.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
		IsSecret:      true,
	}
	db.FirstOrInit(&certificateAuthorityDev, certificateAuthorityDev)
	db.Save(&certificateAuthorityDev)
	eValue = resolvers.EnvironmentVariableValue{
		EnvironmentVariableId: certificateAuthorityDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	certificateAuthorityProd := resolvers.EnvironmentVariable{
		Key:           "CERTIFICATE_AUTHORITY",
		Type:          "file",
		Scope:         resolvers.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
		IsSecret:      true,
	}
	db.FirstOrInit(&certificateAuthorityProd, certificateAuthorityProd)
	db.Save(&certificateAuthorityProd)
	eValue = resolvers.EnvironmentVariableValue{
		EnvironmentVariableId: certificateAuthorityProd.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	clientKeyDev := resolvers.EnvironmentVariable{
		Key:           "CLIENT_KEY",
		Type:          "file",
		Scope:         resolvers.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
		IsSecret:      true,
	}

	db.FirstOrInit(&clientKeyDev, clientKeyDev)
	db.Save(&clientKeyDev)
	eValue = resolvers.EnvironmentVariableValue{
		EnvironmentVariableId: clientKeyDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	clientKeyProd := resolvers.EnvironmentVariable{
		Key:           "CLIENT_KEY",
		Type:          "file",
		Scope:         resolvers.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
		IsSecret:      true,
	}
	db.FirstOrInit(&clientKeyProd, clientKeyProd)
	db.Save(&clientKeyProd)
	eValue = resolvers.EnvironmentVariableValue{
		EnvironmentVariableId: clientKeyProd.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	clientCertDev := resolvers.EnvironmentVariable{
		Key:           "CLIENT_CERTIFICATE",
		Type:          "file",
		Scope:         resolvers.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
		IsSecret:      true,
	}

	db.FirstOrInit(&clientCertDev, clientCertDev)
	db.Save(&clientCertDev)
	eValue = resolvers.EnvironmentVariableValue{
		EnvironmentVariableId: clientCertDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	clientCertProd := resolvers.EnvironmentVariable{
		Key:           "CLIENT_CERTIFICATE",
		Type:          "file",
		Scope:         resolvers.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
		IsSecret:      true,
	}

	db.FirstOrInit(&clientCertProd, clientCertProd)
	db.Save(&clientCertProd)
	eValue = resolvers.EnvironmentVariableValue{
		EnvironmentVariableId: clientCertProd.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	// kubeconfig
	kubeConfig := resolvers.EnvironmentVariable{
		Key:           "KUBECONFIG",
		Type:          "file",
		Scope:         resolvers.GetEnvironmentVariableScope("extension"),
		EnvironmentId: productionEnv.Model.ID,
	}
	db.FirstOrInit(&kubeConfig, kubeConfig)
	db.Save(&kubeConfig)
	eValue = resolvers.EnvironmentVariableValue{
		EnvironmentVariableId: kubeConfig.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
	}
	db.FirstOrInit(&eValue, eValue)
	db.Save(&eValue)

	kubeConfigDev := resolvers.EnvironmentVariable{
		Key:           "KUBECONFIG",
		Type:          "file",
		Scope:         resolvers.GetEnvironmentVariableScope("extension"),
		EnvironmentId: developmentEnv.Model.ID,
	}
	db.FirstOrInit(&kubeConfigDev, kubeConfigDev)
	db.Save(&kubeConfigDev)
	eValue = resolvers.EnvironmentVariableValue{
		EnvironmentVariableId: kubeConfigDev.Model.ID,
		UserId:                user.Model.ID,
		Value:                 "",
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
	extensionSpec := resolvers.ExtensionSpec{
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
	extensionSpec = resolvers.ExtensionSpec{
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
	extensionSpec = resolvers.ExtensionSpec{
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
	extensionSpec = resolvers.ExtensionSpec{
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
	extensionSpec = resolvers.ExtensionSpec{
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
	extensionSpec = resolvers.ExtensionSpec{
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
			ProjectId:  project.ID,
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

func (x *CodeAmp) ExtensionEventHandler(e transistor.Event) error {
	payload := e.Payload.(plugins.Extension)
	var extension resolvers.Extension
	var project resolvers.Project

	if e.Matches("plugins.Extension:status") {
		if x.DB.Where("id = ?", payload.Id).Find(&extension).RecordNotFound() {
			log.InfoWithFields("extension not found", log.Fields{
				"id": payload.Id,
			})
			return fmt.Errorf(fmt.Sprintf("Could not handle Extension status event because Extension not found given payload id: %s.", payload.Id))
		}

		if x.DB.Where("id = ?", extension.ProjectId).Find(&project).RecordNotFound() {
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
		if x.DB.Where("id = ?", payload.Release.Id).Find(&release).RecordNotFound() {
			log.InfoWithFields("release", log.Fields{
				"id": payload.Release.Id,
			})
			return nil
		}

		if x.DB.Where("id = ?", payload.Id).Find(&releaseExtension).RecordNotFound() {
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
		x.DB.Save(&releaseExtension)

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
				Id:         project.Model.ID.String(),
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
					Id:         project.Model.ID.String(),
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

	if x.DB.Where("id = ?", re.ReleaseId).First(&release).RecordNotFound() {
		log.InfoWithFields("release not found", log.Fields{
			"releaseExtension": re,
		})
		return
	}

	if x.DB.Where("id = ?", release.ProjectId).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"release": release,
		})
		return
	}

	if x.DB.Where("release_id = ?", re.ReleaseId).Find(&releaseExtensions).RecordNotFound() {
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
			x.WorkflowExtensionsCompleted(&release)
		case plugins.GetType("deployment"):
			x.DeploymentExtensionsCompleted(&release)
		}
	}
}

func (x *CodeAmp) WorkflowExtensionsCompleted(release *resolvers.Release) {
	// find all related deployment extensions
	depExtensions := []resolvers.Extension{}
	aggregateReleaseExtensionArtifacts := make(map[string]interface{})
	found := false

	if x.DB.Where("project_id = ? and environment_id = ?", release.ProjectId, release.EnvironmentId).Find(&depExtensions).RecordNotFound() {
		log.InfoWithFields("deployment extensions not found", log.Fields{
			"release": release,
		})
		return
	}

	for _, de := range depExtensions {
		var extensionSpec resolvers.ExtensionSpec
		if x.DB.Where("id = ?", de.ExtensionSpecId).First(&extensionSpec).RecordNotFound() {
			log.InfoWithFields("extension spec not found", log.Fields{
				"extension spec": de,
			})
		}
		if plugins.Type(extensionSpec.Type) == plugins.GetType("workflow") {
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
				key := fmt.Sprintf("%s_%s", strings.ToUpper(extensionSpec.Key), strings.ToUpper(k))
				aggregateReleaseExtensionArtifacts[key] = v
			}
		}

		if plugins.Type(extensionSpec.Type) == plugins.GetType("deployment") {
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

	if x.DB.Where("id = ?", release.ProjectId).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"release": release,
		})
	}

	services := []resolvers.Service{}
	if x.DB.Where("project_id = ?", release.ProjectId).Find(&services).RecordNotFound() {
		log.InfoWithFields("no services found for this project", log.Fields{
			"release": release,
		})
	}

	// get secrets from release snapshot env vars
	secrets, pluginServices, err := x.GetSecretsAndServicesFromSnapshot(release)
	if err != nil {
		log.Info(err.Error())
		return
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
	if x.DB.Where("id = ?", release.EnvironmentId).First(&environment).RecordNotFound() {
		log.InfoWithFields("environment not found", log.Fields{
			"id": release.EnvironmentId,
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
		Id:           release.Model.ID.String(),
		HeadFeature: plugins.Feature{
			Id:         headFeature.Model.ID.String(),
			Hash:       headFeature.Hash,
			ParentHash: headFeature.ParentHash,
			User:       headFeature.User,
			Message:    headFeature.Message,
			Created:    headFeature.Created,
		},
		TailFeature: plugins.Feature{
			Id:         tailFeature.Model.ID.String(),
			Hash:       tailFeature.Hash,
			ParentHash: tailFeature.ParentHash,
			User:       tailFeature.User,
			Message:    tailFeature.Message,
			Created:    tailFeature.Created,
		},
		User: "",
		Project: plugins.Project{
			Id:             project.Model.ID.String(),
			Action:         plugins.GetAction("update"),
			Repository:     project.Repository,
			NotifyChannels: []string{}, // not sure what channels can be notified with this
			Services:       pluginServices,
		},
		Git: plugins.Git{
			Url:           project.GitUrl,
			Branch:        branch,
			RsaPrivateKey: project.RsaPrivateKey,
		},
		Secrets: secrets,
	}
	releaseExtensionEvents := []plugins.ReleaseExtension{}

	for _, extension := range depExtensions {
		extensionSpec := resolvers.ExtensionSpec{}
		if x.DB.Where("id= ?", extension.ExtensionSpecId).Find(&extensionSpec).RecordNotFound() {
			log.InfoWithFields("extension spec not found", log.Fields{
				"extension": extension,
			})
		}

		if plugins.Type(extensionSpec.Type) == plugins.GetType("workflow") {
			releaseExtension := resolvers.ReleaseExtension{}

			if x.DB.Where("release_id = ? AND extension_id = ? AND state = ?", release.Model.ID, extension.Model.ID, plugins.GetState("complete")).Find(&releaseExtension).RecordNotFound() {
				log.InfoWithFields("release extension not found", log.Fields{
					"release_id":   release.Model.ID,
					"extension_id": extension.Model.ID,
					"state":        plugins.GetState("complete"),
				})
			}
		}

		if plugins.Type(extensionSpec.Type) == plugins.GetType("deployment") {

			// create ReleaseExtension
			releaseExtension := resolvers.ReleaseExtension{
				ReleaseId:         release.Model.ID,
				FeatureHash:       "",
				ServicesSignature: "",
				SecretsSignature:  "",
				ExtensionId:       extension.Model.ID,
				State:             plugins.GetState("waiting"),
				Type:              plugins.GetType("deployment"),
				StateMessage:      "initialized",
			}

			x.DB.Save(&releaseExtension)
			unmarshalledConfig := make(map[string]interface{})

			err := json.Unmarshal(extension.Config.RawMessage, &unmarshalledConfig)
			if err != nil {
				log.Info(err.Error())
			}

			formValues, err := x.GetFilledFormValues(unmarshalledConfig, extensionSpec.Key, x.DB)
			if err != nil {
				log.Info(err.Error())
			}

			extensionEvent := plugins.Extension{
				Id:     extension.Model.ID.String(),
				Config: formValues,
				// Artifacts: plugins.HstoreToMapStringString(extension.Artifacts),
			}

			releaseExtensionEvents = append(releaseExtensionEvents, plugins.ReleaseExtension{
				Id:           releaseExtension.Model.ID.String(),
				Action:       plugins.GetAction("create"),
				Slug:         extensionSpec.Key,
				State:        releaseExtension.State,
				Artifacts:    map[string]string{},
				Release:      releaseEvent,
				Extension:    extensionEvent,
				StateMessage: releaseExtension.StateMessage,
			})

		}
	}

	// send out release extension event for each re
	for _, re := range releaseExtensionEvents {
		re.Release.Artifacts = aggregateReleaseExtensionArtifacts
		x.Events <- transistor.NewEvent(re, nil)
	}
}

func (x *CodeAmp) DeploymentExtensionsCompleted(release *resolvers.Release) {
	// find all related deployment extensions
	depExtensions := []resolvers.Extension{}
	// releaseExtensionArtifacts := map[string]string{}

	if x.DB.Where("project_id = ?", release.ProjectId).Find(&depExtensions).RecordNotFound() {
		log.InfoWithFields("deployment extensions not found", log.Fields{
			"release": release,
		})
		return
	}

	for _, de := range depExtensions {
		var extensionSpec resolvers.ExtensionSpec
		if x.DB.Where("id = ?", de.ExtensionSpecId).First(&extensionSpec).RecordNotFound() {
			log.InfoWithFields("extension spec not found", log.Fields{
				"id": de.ExtensionSpecId,
			})
		}

		if plugins.Type(extensionSpec.Type) == plugins.GetType("deployment") {
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
	if x.DB.Where("id = ?", release.ProjectId).First(&project).RecordNotFound() {
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

	if x.DB.Where("id = ?", release.ProjectId).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"release": release,
		})
		return
	}

	// loop through extensions and send ReleaseWorkflow events
	projectExtensions := []resolvers.Extension{}
	if x.DB.Where("project_id = ? and environment_id = ?", release.ProjectId, release.EnvironmentId).Find(&projectExtensions).RecordNotFound() {
		log.InfoWithFields("project has no extensions", log.Fields{
			"project_id":     release.ProjectId,
			"environment_id": release.EnvironmentId,
		})
	}

	services := []resolvers.Service{}
	if x.DB.Where("project_id = ? and environment_id = ?", release.ProjectId, release.EnvironmentId).Find(&services).RecordNotFound() {
		log.InfoWithFields("project has no services", log.Fields{
			"project_id":     release.ProjectId,
			"environment_id": release.EnvironmentId,
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
	if x.DB.Where("id = ?", release.EnvironmentId).First(&environment).RecordNotFound() {
		log.InfoWithFields("environment not found", log.Fields{
			"id": release.EnvironmentId,
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

	secrets, pluginServices, err := x.GetSecretsAndServicesFromSnapshot(release)
	if err != nil {
		log.Info(err.Error())
		return
	}

	releaseEvent := plugins.Release{
		Id:          release.Model.ID.String(),
		Action:      plugins.GetAction("create"),
		State:       plugins.GetState("waiting"),
		Environment: environment.Name,
		HeadFeature: plugins.Feature{
			Id:         headFeature.Model.ID.String(),
			Hash:       headFeature.Hash,
			ParentHash: headFeature.ParentHash,
			User:       headFeature.User,
			Message:    headFeature.Message,
			Created:    headFeature.Created,
		},
		TailFeature: plugins.Feature{
			Id:         tailFeature.Model.ID.String(),
			Hash:       tailFeature.Hash,
			ParentHash: tailFeature.ParentHash,
			User:       tailFeature.User,
			Message:    tailFeature.Message,
			Created:    tailFeature.Created,
		},
		User: release.User.Email,
		Project: plugins.Project{
			Id:         project.Model.ID.String(),
			Repository: project.Repository,
			Services:   pluginServices,
		},
		Git: plugins.Git{
			Url:           project.GitUrl,
			Branch:        branch,
			RsaPrivateKey: project.RsaPrivateKey,
		},
		Secrets: secrets,
	}
	for _, extension := range projectExtensions {
		extensionSpec := resolvers.ExtensionSpec{}
		if x.DB.Where("id= ?", extension.ExtensionSpecId).Find(&extensionSpec).RecordNotFound() {
			log.InfoWithFields("extension spec not found", log.Fields{
				"id": extension.ExtensionSpecId,
			})
		}

		// ONLY SEND WORKFLOW TYPE, EVENTs
		if plugins.Type(extensionSpec.Type) == plugins.GetType("workflow") {
			// create ReleaseExtension
			releaseExtension := resolvers.ReleaseExtension{
				ReleaseId:         release.Model.ID,
				FeatureHash:       "",
				ServicesSignature: "",
				SecretsSignature:  "",
				ExtensionId:       extension.Model.ID,
				State:             plugins.GetState("waiting"),
				Type:              plugins.GetType("workflow"),
			}

			x.DB.Save(&releaseExtension)

			unmarshalledConfig := make(map[string]interface{})

			err := json.Unmarshal(extension.Config.RawMessage, &unmarshalledConfig)
			if err != nil {
				log.Info(err.Error())
			}

			formValues, err := x.GetFilledFormValues(unmarshalledConfig, extensionSpec.Key, x.DB)
			if err != nil {
				log.Info(err.Error())
			}

			extensionEvent := plugins.Extension{
				Id:     extension.Model.ID.String(),
				Config: formValues,
				// Artifacts:  plugins.HstoreToMapStringString(extension.Artifacts),
			}
			x.Events <- transistor.NewEvent(plugins.ReleaseExtension{
				Id:        releaseExtension.Model.ID.String(),
				Action:    plugins.GetAction("create"),
				Slug:      extensionSpec.Key,
				State:     releaseExtension.State,
				Release:   releaseEvent,
				Extension: extensionEvent,
				Artifacts: map[string]string{},
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
func (x *CodeAmp) GetFilledFormValues(configWithEnvVarIds map[string]interface{}, extensionSpecKey string, db *gorm.DB) (map[string]interface{}, error) {
	formValues := make(map[string]interface{})
	// iter through custom + config and add to formvalues interface
	for _, val := range configWithEnvVarIds["config"].([]interface{}) {
		val := val.(map[string]interface{})
		// check if val is UUID. If so, query in environment variables for id
		valId := uuid.FromStringOrNil(val["value"].(string))
		if valId != uuid.Nil {
			envVar := resolvers.EnvironmentVariableValue{}

			if db.Where("environment_variable_id = ?", valId).Order("created_at desc").First(&envVar).RecordNotFound() {
				log.InfoWithFields("envvarvalue not found", log.Fields{
					"environment_variable_id": valId,
				})
			}
			formValues[fmt.Sprintf("%s_%s", strings.ToUpper(extensionSpecKey), strings.ToUpper(val["key"].(string)))] = envVar.Value
		} else {
			formValues[fmt.Sprintf("%s_%s", strings.ToUpper(extensionSpecKey), strings.ToUpper(val["key"].(string)))] = val["value"].(string)
		}
	}

	for key, val := range configWithEnvVarIds["custom"].(map[string]interface{}) {
		// check if val is UUID. If so, query in environment variables for id
		formValues[fmt.Sprintf("%s_%s", strings.ToUpper(extensionSpecKey), strings.ToUpper(key))] = val
	}
	return formValues, nil
}

func (x *CodeAmp) GetSecretsAndServicesFromSnapshot(release *resolvers.Release) ([]plugins.Secret, []plugins.Service, error) {
	secrets := []plugins.Secret{}
	unmarshalledSnapshot := map[string]interface{}{}
	err := json.Unmarshal(release.Snapshot.RawMessage, &unmarshalledSnapshot)
	if err != nil {
		log.Info(err.Error())
		return nil, nil, err
	}

	for _, envvar := range unmarshalledSnapshot["environmentVariables"].([]interface{}) {
		key := envvar.(map[string]interface{})["key"].(string)
		val := envvar.(map[string]interface{})["value"].(string)
		evType := plugins.GetType(envvar.(map[string]interface{})["type"].(string))

		secrets = append(secrets, plugins.Secret{
			Key:   key,
			Value: val,
			Type:  evType,
		})
	}

	pluginServices := []plugins.Service{}
	for _, service := range unmarshalledSnapshot["services"].([]interface{}) {
		pluginListeners := []plugins.Listener{}
		for _, listener := range service.(map[string]interface{})["container_ports"].([]interface{}) {
			intPort, _ := strconv.Atoi(listener.(map[string]interface{})["port"].(string))
			pluginListeners = append(pluginListeners, plugins.Listener{
				Port:     int32(intPort),
				Protocol: listener.(map[string]interface{})["protocol"].(string),
			})
		}

		intTerminationGracePeriod, _ := strconv.Atoi(service.(map[string]interface{})["service_spec"].(map[string]interface{})["termination_grace_period"].(string))
		intReplicas, _ := strconv.Atoi(service.(map[string]interface{})["count"].(string))
		pluginServices = append(pluginServices, plugins.Service{
			Id:        service.(map[string]interface{})["id"].(string),
			Command:   service.(map[string]interface{})["command"].(string),
			Name:      service.(map[string]interface{})["name"].(string),
			Listeners: pluginListeners,
			State:     plugins.GetState("waiting"),
			Spec: plugins.ServiceSpec{
				Id:                            service.(map[string]interface{})["service_spec"].(map[string]interface{})["id"].(string),
				CpuRequest:                    fmt.Sprintf("%sm", service.(map[string]interface{})["service_spec"].(map[string]interface{})["cpu_request"].(string)),
				CpuLimit:                      fmt.Sprintf("%sm", service.(map[string]interface{})["service_spec"].(map[string]interface{})["cpu_limit"].(string)),
				MemoryRequest:                 fmt.Sprintf("%sMi", service.(map[string]interface{})["service_spec"].(map[string]interface{})["memory_request"].(string)),
				MemoryLimit:                   fmt.Sprintf("%sMi", service.(map[string]interface{})["service_spec"].(map[string]interface{})["memory_limit"].(string)),
				TerminationGracePeriodSeconds: int64(intTerminationGracePeriod),
			},
			Type:     string(service.(map[string]interface{})["type"].(string)),
			Replicas: int64(intReplicas),
		})
	}
	return secrets, pluginServices, nil
}

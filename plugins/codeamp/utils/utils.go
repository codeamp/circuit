package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/satori/go.uuid"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	oidc "github.com/coreos/go-oidc"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	UserId      string   `json:"userId"`
	Email       string   `json:"email"`
	Verified    bool     `json:"email_verified"`
	Groups      []string `json:"groups"`
	Permissions []string `json:"permissions"`
	TokenError  string   `json:"tokenError"`
}

type Cache struct {
	UserId      uuid.UUID `json:"userId"`
	Permissions []string  `json:"permissions"`
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func GetTempDir() (string, error) {
	for {
		filePath := fmt.Sprintf("/tmp/%s", uuid.NewV1().String())
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			log.Info("directory does not exist")
			// create the file
			err = os.MkdirAll(filePath, os.ModeDir)
			if err != nil {
				log.Info(err.Error())
				return "", err
			}
			return filePath, nil
		}
	}
}

func SetupKubeConfig(config map[string]interface{}, key string) (string, error) {
	randomDirectory, err := GetTempDir()
	if err != nil {
		log.Info(err.Error())
		return "", err
	}

	err = ioutil.WriteFile(fmt.Sprintf("%s/kubeconfig", randomDirectory), []byte(config[fmt.Sprintf("%sKUBECONFIG", key)].(string)), 0644)
	if err != nil {
		log.Info(err.Error())
		return "", err
	}

	if err != nil {
		log.Info("ERROR: %s", err.Error())
		return "", err
	}
	log.Info("Using kubeconfig file: %s", fmt.Sprintf("%s/kubeconfig", randomDirectory))

	// generate client cert, client key
	// certificate authority
	err = ioutil.WriteFile(fmt.Sprintf("%s/admin.pem", randomDirectory),
		[]byte(config[fmt.Sprintf("%sCLIENT_CERTIFICATE", key)].(string)), 0644)
	if err != nil {
		log.Info("ERROR: %s", err.Error())
		return "", err
	}

	err = ioutil.WriteFile(fmt.Sprintf("%s/admin-key.pem", randomDirectory),
		[]byte(config[fmt.Sprintf("%sCLIENT_KEY", key)].(string)), 0644)
	if err != nil {
		log.Info("ERROR: %s", err.Error())
		return "", err
	}

	err = ioutil.WriteFile(fmt.Sprintf("%s/ca.pem", randomDirectory),
		[]byte(config[fmt.Sprintf("%sCERTIFICATE_AUTHORITY", key)].(string)), 0644)
	if err != nil {
		log.Info("ERROR: %s", err.Error())
		return "", err
	}

	return fmt.Sprintf("%s/kubeconfig", randomDirectory), nil
}

/* fills in Config by querying config ids and getting the actual value */
func GetFilledFormValues(configWithEnvVarIds map[string]interface{}, extensionSpecKey string, db *gorm.DB) (map[string]interface{}, error) {
	formValues := make(map[string]interface{})
	// iter through custom + config and add to formvalues interface
	for _, val := range configWithEnvVarIds["config"].([]interface{}) {
		val := val.(map[string]interface{})
		// check if val is UUID. If so, query in environment variables for id
		valId := uuid.FromStringOrNil(val["value"].(string))
		if valId != uuid.Nil {
			envVar := models.EnvironmentVariableValue{}

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

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func AuthMiddleware(next http.Handler, db *gorm.DB, redisClient *redis.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := Claims{}
		authString := r.Header.Get("Authorization")
		ctx := context.Context(context.Background())

		if len(authString) < 8 {
			claims.TokenError = "invalid access token"
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "jwt", claims)))
			return
		}

		bearerToken := authString[7:len(authString)]

		// Initialize a provider by specifying dex's issuer URL.
		provider, err := oidc.NewProvider(ctx, viper.GetString("plugins.codeamp.oidc_uri"))
		if err != nil {
			claims.TokenError = err.Error()
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "jwt", claims)))
			return
		}

		// Create an ID token parser, but only trust ID tokens issued to "example-app"
		idTokenVerifier := provider.Verifier(&oidc.Config{ClientID: viper.GetString("plugins.codeamp.oidc_client_id")})

		idToken, err := idTokenVerifier.Verify(ctx, bearerToken)
		if err != nil {
			claims.TokenError = fmt.Sprintf("could not verify bearer token: %v", err.Error())
			w.Header().Set("Www-Authenticate", "Bearer token_type=\"JWT\"")
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "jwt", claims)))
			return
		}

		if err := idToken.Claims(&claims); err != nil {
			claims.TokenError = fmt.Sprintf("failed to parse claims: %v", err.Error())
			w.Header().Set("Www-Authenticate", "Bearer token_type=\"JWT\"")
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "jwt", claims)))
			return
		}

		if !claims.Verified {
			claims.TokenError = fmt.Sprintf("email (%q) in returned claims was not verified", claims.Email)
			w.Header().Set("Www-Authenticate", "Bearer token_type=\"JWT\"")
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "jwt", claims)))
			return
		}

		c, err := redisClient.Get(fmt.Sprintf("%s_%s", idToken.Nonce, claims.Email)).Result()
		if err == redis.Nil {
			user := models.User{}
			if db.Where("email = ?", claims.Email).Find(&user).RecordNotFound() {
				user.Email = claims.Email
				db.Create(&user)
			}

			db.Model(&user).Association("Permissions").Find(&user.Permissions)

			var permissions []string
			for _, permission := range user.Permissions {
				permissions = append(permissions, permission.Value)
			}

			// Add user scope
			permissions = append(permissions, fmt.Sprintf("user/%s", user.ID.String()))
			claims.UserId = user.ID.String()
			claims.Permissions = permissions

			serializedClaims, err := json.Marshal(claims)
			if err != nil {
				log.Panic(err)
			}

			err = redisClient.Set(fmt.Sprintf("%s_%s", idToken.Nonce, claims.Email), serializedClaims, 24*time.Hour).Err()
			if err != nil {
				log.Panic(err)
			}
		} else if err != nil {
			log.Panic(err)
		} else {
			err := json.Unmarshal([]byte(c), &claims)
			if err != nil {
				log.Panic(err)
			}
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "jwt", claims)))
	})
}

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// allow cross domain AJAX requests
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Cache-Control, Content-Language, Content-Type, Expires, Last-Modified, Pragma, WWW-Authenticate")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
		if r.Method == "OPTIONS" {
			//handle preflight in here
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func CheckAuth(ctx context.Context, scopes []string) (string, error) {
	claims := ctx.Value("jwt").(Claims)

	if claims.UserId == "" {
		return "", errors.New(claims.TokenError)
	}

	if transistor.SliceContains("admin1", claims.Permissions) {
		return claims.UserId, nil
	}

	if len(scopes) == 0 {
		return claims.UserId, nil
	} else {
		for _, scope := range scopes {
			if transistor.SliceContains(scope, claims.Permissions) {
				return claims.UserId, nil
			}
		}
		return "", errors.New("you dont have permission to access this resource")
	}

	return claims.UserId, nil
}

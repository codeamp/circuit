package graphql_resolver_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codeamp/transistor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	log "github.com/codeamp/logger"

	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/test"
)

type MiddlewareTestSuite struct {
	suite.Suite
	Resolver *graphql_resolver.Resolver

	Middleware *graphql_resolver.Middleware
}

func (ts *MiddlewareTestSuite) SetupTest() {
	migrators := []interface{}{
		&model.Project{},
		&model.ProjectBookmark{},
		&model.ProjectEnvironment{},
		&model.ProjectExtension{},
		&model.ProjectSettings{},
		&model.UserPermission{},
		&model.Environment{},
		&model.Extension{},
		&model.Service{},
		&model.ServiceSpec{},
		&model.Secret{},
		&model.Feature{},
	}

	db, err := test.SetupResolverTest(migrators)
	if err != nil {
		log.Fatal(err.Error())
	}

	ts.Resolver = &graphql_resolver.Resolver{DB: db, Events: make(chan transistor.Event, 10)}
	ts.Middleware = &graphql_resolver.Middleware{ts.Resolver}
}

func (ts *MiddlewareTestSuite) GetTestHandler() http.HandlerFunc {
	fn := func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")

		// In the future we could report back on the status of our DB, or our cache
		// (e.g. Redis) by performing a simple PING, and include them in the response.
		io.WriteString(w, `{"alive": true}`)
	}
	return http.HandlerFunc(fn)
}

func (ts *MiddlewareTestSuite) TestAuthFailure() {
	testserver := httptest.NewServer(ts.Middleware.Auth(ts.GetTestHandler()))
	defer testserver.Close()

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	rr := httptest.NewRecorder()
	handler := ts.Middleware.Auth(ts.GetTestHandler())
	handler.ServeHTTP(rr, req)

	assert.Equal(ts.T(), http.StatusForbidden, rr.Code)
}

func (ts *MiddlewareTestSuite) TestAuthSuccess() {
	testserver := httptest.NewServer(ts.Middleware.Auth(ts.GetTestHandler()))
	defer testserver.Close()

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	req = req.WithContext(test.ResolverAuthContext())
	req.Header.Set("Authorization", "Bearer abcd12345")

	rr := httptest.NewRecorder()
	handler := ts.Middleware.Auth(ts.GetTestHandler())
	handler.ServeHTTP(rr, req)

	assert.Equal(ts.T(), http.StatusOK, rr.Code)
}

func (ts *MiddlewareTestSuite) TestCors() {
	testserver := httptest.NewServer(ts.Middleware.Auth(ts.GetTestHandler()))
	defer testserver.Close()

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	handler := ts.Middleware.Cors(ts.GetTestHandler())
	handler.ServeHTTP(httptest.NewRecorder(), req)
}

func (ts *MiddlewareTestSuite) TestCorsOptions() {
	testserver := httptest.NewServer(ts.Middleware.Auth(ts.GetTestHandler()))
	defer testserver.Close()

	req, err := http.NewRequest("OPTIONS", "/", nil)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	handler := ts.Middleware.Cors(ts.GetTestHandler())
	handler.ServeHTTP(httptest.NewRecorder(), req)
}

func (ts *MiddlewareTestSuite) TearDownTest() {
}

func TestMiddleware(t *testing.T) {
	suite.Run(t, new(MiddlewareTestSuite))
}

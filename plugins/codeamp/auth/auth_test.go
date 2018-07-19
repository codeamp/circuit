package auth_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/codeamp/circuit/test"

	"github.com/codeamp/circuit/plugins/codeamp/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AuthTestSuite struct {
	suite.Suite
}

var ValidUUID = [...]string{"123e4567-e89b-12d3-a456-426655440000", "11075553-5309-494B-9085-2D79A6ED1EB3"}

func (ts *AuthTestSuite) SetupTest() {
}

func (ts *AuthTestSuite) TestAuthFailureBadID() {
	ctx := test.BuildAuthContext("", "test@example.com", []string{})

	userID, err := auth.CheckAuth(ctx, []string{})
	assert.NotNil(ts.T(), err)
	assert.Equal(ts.T(), "", userID)
}

func (ts *AuthTestSuite) TestAuthFailureNoJWT() {
	var ctx context.Context
	userID, err := auth.CheckAuth(ctx, []string{})
	assert.NotNil(ts.T(), err)
	assert.Equal(ts.T(), "", userID)
}

func (ts *AuthTestSuite) TestAuthEmptyPermissions() {
	ctx := test.BuildAuthContext(ValidUUID[0], "test@example.com", []string{})

	_, err := auth.CheckAuth(ctx, []string{})
	assert.Nil(ts.T(), err)
}

func (ts *AuthTestSuite) TestAuthSelfSuccess() {
	userID := ValidUUID[0]
	ctx := test.BuildAuthContext(userID, "test@example.com", []string{fmt.Sprintf("user/%s", userID)})

	_userID, err := auth.CheckAuth(ctx, []string{fmt.Sprintf("user/%s", userID)})
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), userID, _userID)
}

func (ts *AuthTestSuite) TestAuthOtherUserSuccess() {
	userID := ValidUUID[0]
	ctx := test.BuildAuthContext(userID, "test@example.com", []string{fmt.Sprintf("user/%s", ValidUUID[1])})

	_userID, err := auth.CheckAuth(ctx, []string{fmt.Sprintf("user/%s", ValidUUID[1])})
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), userID, _userID)
}

func (ts *AuthTestSuite) TestAuthSelfWithSelfPermissionSuccess() {
	userID := ValidUUID[0]
	ctx := test.BuildAuthContext(userID, "test@example.com", []string{fmt.Sprintf("user/%s", userID)})

	_userID, err := auth.CheckAuth(ctx, []string{fmt.Sprintf("user/%s", userID)})
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), userID, _userID)
}

func (ts *AuthTestSuite) TestAuthSelfFailure() {
	userID := ValidUUID[0]
	ctx := test.BuildAuthContext(userID, "test@example.com", []string{})

	_userID, err := auth.CheckAuth(ctx, []string{fmt.Sprintf("user/%s", userID)})
	assert.NotNil(ts.T(), err)
	assert.Equal(ts.T(), userID, _userID)
}

func (ts *AuthTestSuite) TestAuthProjectPermissionHierarchySuccess() {
	userID := ValidUUID[0]
	ctx := test.BuildAuthContext(userID, "test@example.com", []string{"projects"})

	_userID, err := auth.CheckAuth(ctx, []string{"projects/checkr/judy"})
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), userID, _userID)

	_userID, err = auth.CheckAuth(ctx, []string{"projects/checkr/checkr"})
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), userID, _userID)
}

func (ts *AuthTestSuite) TestAuthProjectPermissionHierarchyFailureMoreDifferently() {
	userID := ValidUUID[0]
	ctx := test.BuildAuthContext(userID, "test@example.com", []string{"projects/not/checkr"})

	_userID, err := auth.CheckAuth(ctx, []string{"projects"})
	assert.NotNil(ts.T(), err)
	assert.Equal(ts.T(), userID, _userID)
}

func (ts *AuthTestSuite) TestAuthProjectPermissionHierarchyFailure() {
	userID := ValidUUID[0]
	ctx := test.BuildAuthContext(userID, "test@example.com", []string{"projects/not/checkr"})

	_userID, err := auth.CheckAuth(ctx, []string{"projects/checkr/judy"})
	assert.NotNil(ts.T(), err)
	assert.Equal(ts.T(), userID, _userID)
}

func (ts *AuthTestSuite) TestAuthProjectPermissionSuccess() {
	userID := ValidUUID[0]
	ctx := test.BuildAuthContext(userID, "test@example.com", []string{"projects/checkr/judy"})

	_userID, err := auth.CheckAuth(ctx, []string{"projects/checkr/judy"})
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), userID, _userID)
}

func (ts *AuthTestSuite) TestAuthProjectPermissionsEmpty() {
	userID := ValidUUID[0]
	ctx := test.BuildAuthContext(userID, "test@example.com", []string{"projects/checkr/judy"})

	_userID, err := auth.CheckAuth(ctx, []string{})
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), userID, _userID)
}

func (ts *AuthTestSuite) TestAuthProjectPermissionDenied() {
	userID := ValidUUID[0]
	ctx := test.BuildAuthContext(userID, "test@example.com", []string{"projects/checkr/judy"})

	_userID, err := auth.CheckAuth(ctx, []string{"projects/codeamp/judy"})
	assert.NotNil(ts.T(), err)
	assert.Equal(ts.T(), userID, _userID)
}

func (ts *AuthTestSuite) TestAuthProjectPermissionFailure() {
	userID := ValidUUID[0]
	ctx := test.BuildAuthContext(userID, "test@example.com", []string{"projects/checkr/reginald"})

	_userID, err := auth.CheckAuth(ctx, []string{"projects/checkr/judy"})
	assert.NotNil(ts.T(), err)
	assert.Equal(ts.T(), userID, _userID)
}

func (ts *AuthTestSuite) TestAuthIsAdminUserSuccess() {
	userID := ValidUUID[0]
	ctx := test.BuildAuthContext(userID, "test@example.com", []string{"admin"})

	_userID, err := auth.CheckAuth(ctx, []string{"admin"})
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), userID, _userID)
}

func (ts *AuthTestSuite) TestAuthIsAdminUserFailure() {
	userID := ValidUUID[0]
	ctx := test.BuildAuthContext(userID, "test@example.com", []string{"projects/checkr/judy"})
	_userID, err := auth.CheckAuth(ctx, []string{"admin"})
	assert.NotNil(ts.T(), err)
	assert.Equal(ts.T(), userID, _userID)
}

func (ts *AuthTestSuite) TearDownTest() {
}

func TestSuiteAuth(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}

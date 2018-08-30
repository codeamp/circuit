package codeamp_resolvers

import (
	"encoding/json"

	"github.com/jinzhu/gorm"
	graphql "github.com/graph-gophers/graphql-go"
	uuid "github.com/satori/go.uuid"
)

// User
type User struct {
	Model `json:",inline"`
	// Email
	Email string `json:"email"`
	// Password
	Password string `json:"password" gorm:"type:varchar(255)"`
	// Permissions
	Permissions []UserPermission `json:"permissions"`
}

// User permission
type UserPermission struct {
	Model `json:",inline"`
	// UserID
	UserID uuid.UUID `json:"userID" gorm:"type:uuid"`
	// Value
	Value string `json:"value"`
}

// User resolver
type UserResolver struct {
	User
	DB *gorm.DB
}

// ID
func (r *UserResolver) ID() graphql.ID {
	return graphql.ID(r.User.Model.ID.String())
}

// Email
func (r *UserResolver) Email() string {
	return r.User.Email
}

// Permissions
func (r *UserResolver) Permissions() []string {
	var permissions []string

	r.DB.Model(r.User).Association("Permissions").Find(&r.User.Permissions)

	for _, permission := range r.User.Permissions {
		permissions = append(permissions, permission.Value)
	}

	return permissions
}

// Created
func (r *UserResolver) Created() graphql.Time {
	return graphql.Time{Time: r.User.Model.CreatedAt}
}

func (r *UserResolver) MarshalJSON() ([]byte, error) {
	return json.Marshal(&r.User)
}

func (r *UserResolver) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.User)
}

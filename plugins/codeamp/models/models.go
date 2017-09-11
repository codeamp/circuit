package codeamp_models

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type Model struct {
	ID        uuid.UUID  `sql:"type:uuid;default:uuid_generate_v4()" json:"id" gorm:"primary_key"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt" sql:"index"`
}

type User struct {
	Model       `json:",inline"`
	Email       string `json:"email" gorm:"type:varchar(100);unique_index"`
	Password    string `json:"password" gorm:"type:varchar(255)"`
	Permissions []UserPermission
}

type UserPermission struct {
	Model  `json:",inline"`
	UserId uuid.UUID `json:"userId" gorm:"type:uuid"`
	Value  string    `json:"value"`
}

type Project struct {
	Model         `json:",inline"`
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	Repository    string `json:"repository"`
	Secret        string `json:"secret"`
	GitUrl        string `json:"GitUrl"`
	GitProtocol   string `json:"GitProtocol"`
	RsaPrivateKey string `json:"rsaPrivateKey"`
	RsaPublicKey  string `json:"rsaPublicKey"`
}

type Release struct {
	Model     `json:",inline"`
	ProjectId uuid.UUID `bson:"projectId" json:"projectId" gorm:"type:uuid"`
	UserId    uuid.UUID `bson:"userId" json:"-" gorm:"type:uuid"`
}

type Bookmark struct {
	Model     `json:",inline"`
	UserId    uuid.UUID `json:"userId" gorm:"type:uuid"`
	ProjectId uuid.UUID `json:"projectId" gorm:"type:uuid"`
}

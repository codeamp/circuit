package resolvers

import (
	"github.com/codeamp/circuit/plugins"
	"github.com/jinzhu/gorm"
)

type KeyValueResolver struct {
	db       *gorm.DB
	KeyValue plugins.KeyValue
}

func (r *KeyValueResolver) Key() string {
	return r.KeyValue.Key
}

func (r *KeyValueResolver) Value() string {
	return r.KeyValue.Value
}

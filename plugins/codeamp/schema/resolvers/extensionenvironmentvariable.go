package resolvers

import (
	"github.com/codeamp/circuit/plugins"
	"github.com/jinzhu/gorm"
)

type ExtensionEnvironmentVariableResolver struct {
	db                           *gorm.DB
	ExtensionEnvironmentVariable plugins.ExtensionEnvironmentVariable
}

func (r *ExtensionEnvironmentVariableResolver) ProjectId() *string {
	return &r.ExtensionEnvironmentVariable.ProjectId
}

func (r *ExtensionEnvironmentVariableResolver) Key() string {
	return r.ExtensionEnvironmentVariable.Key
}

func (r *ExtensionEnvironmentVariableResolver) Type() string {
	return r.ExtensionEnvironmentVariable.Type
}

func (r *ExtensionEnvironmentVariableResolver) EnvironmentVariableId() *string {
	return &r.ExtensionEnvironmentVariable.EnvironmentVariableId
}

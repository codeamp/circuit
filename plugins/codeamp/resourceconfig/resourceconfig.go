package resourceconfig

type ResourceConfig interface {
	Import() error
	Export() (interface{}, error)
	GetConfig() *string
}

type BaseResourceConfig struct {
	ResourceConfig
	config *string
}

func (b *BaseResourceConfig) GetConfig() *string {
	return b.config
}

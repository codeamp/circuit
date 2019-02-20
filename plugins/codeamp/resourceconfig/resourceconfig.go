package resourceconfig

type ResourceConfig interface {
	Import() error
	ExportYAML() (string, error)
	GetConfig() string
	GetChildResourceConfigs() ([]ResourceConfig, error)
}

type BaseResourceConfig struct {
	ResourceConfig
	config string
}

func (b *BaseResourceConfig) Import() error {
	childConfigs, err := b.GetChildResourceConfigs()
	if err != nil {
		return err
	}

	for _, config := range childConfigs {
		err := config.Import()
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *BaseResourceConfig) GetChildResourceConfigs() ([]ResourceConfig, error) {
	return []ResourceConfig{}, nil
}

func (b *BaseResourceConfig) GetConfig() string {
	return b.config
}

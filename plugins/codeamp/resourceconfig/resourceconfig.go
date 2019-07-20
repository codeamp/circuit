package resourceconfig

// A ResourceConfig is an object that is used for importing and exporting
// resources in CodeAmp. Its purpose is to decouple implementation behaviors
// of Export and Import from objects when the respective operation is occurring
// and provide a general-purpose interface for any resource that wants to implement
// import/export behaviors.
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

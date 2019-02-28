package resourceconfig

// A ResourceConfig is an object that is used for importing and exporting
// resources in CodeAmp. Its purpose is to decouple implementation behaviors
// of Export and Import from objects when the respective operation is occurring
// and provide a general-purpose interface for any resource that wants to implement
// import/export behaviors.
type ResourceConfig interface {
	Import(interface{}) error
	Export(interface{}) (interface{}, error)
}

// BaseResourceConfig is used for
// incorporating into any struct that wants
// to implement ResourceConfig
type BaseResourceConfig struct{}

// Import takes in an input in its exported format
// and persists it into the system's database
func (b *BaseResourceConfig) Import(obj interface{}) error {
	return nil
}

// Export takes in an input that can be transformed into
// its corresponding export format e.g. transforming a model.Service
// into a resourceconfig.Service
func (b *BaseResourceConfig) Export(obj interface{}) (interface{}, error) {
	return obj, nil
}

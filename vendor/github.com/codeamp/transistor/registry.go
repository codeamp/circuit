package transistor

type Creator func() Plugin

var PluginRegistry = map[string]Creator{}

func RegisterPlugin(name string, creator Creator) {
	PluginRegistry[name] = creator
}

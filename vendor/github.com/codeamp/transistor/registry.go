package transistor

import (
	"reflect"
)

type Creator func() Plugin

var PluginRegistry = map[string]Creator{}
var EventRegistry = make(map[string]interface{})

func RegisterPlugin(name string, creator Creator) {
	PluginRegistry[name] = creator
}

func RegisterEvent(i interface{}) {
	EventRegistry[reflect.TypeOf(i).String()] = i
}

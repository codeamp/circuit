package transistor

import "reflect"

type Creator func() Plugin

var PluginRegistry = map[string]Creator{}
var EventRegistry = make(map[string]interface{})

func RegisterPlugin(name string, creator Creator, events ...interface{}) {
	PluginRegistry[name] = creator
	for _, i := range events {
		EventRegistry[reflect.TypeOf(i).String()] = i
	}
}

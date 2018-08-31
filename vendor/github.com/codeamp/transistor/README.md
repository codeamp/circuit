# Transistor [![CircleCI](https://circleci.com/gh/codeamp/transistor.svg?style=svg)](https://circleci.com/gh/codeamp/transistor) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) [![Coverage Status](https://coveralls.io/repos/github/codeamp/transistor/badge.svg?branch=master)](https://coveralls.io/github/codeamp/transistor?branch=master) [![Go Report Card](https://goreportcard.com/badge/codeamp/transistor)](https://goreportcard.com/report/codeamp/transistor) [![codebeat badge](https://codebeat.co/badges/b977a7e7-1e94-43e1-9e58-463cff99add3)](https://codebeat.co/projects/github-com-codeamp-transistor-master)

Transistor allows you to run distributed workload on one or accross multiple hosts. It's a plugin based system that allows plugins to subscribe to multiple events and internal scheduler takes care of delivery. This allows multiple plugins to recieve same message do some work and respond with updated or different message.

We use a central file `api.go` that keeps all available  messages in one place.

```go
package plugins

import "github.com/codeamp/transistor"

func init() {
}

type Hello struct {
    Action  string
      Message string
}
```

you also need to register all api events with transistor so that we are able to transform it to correct type when json event payload is recieved.

All plugins need to implement `Start`, `Stop`, `Subscribe` and `Process` methods.

```go
type Plugin interface {
	// Start starts the Plugin service, whatever that may be
	Start(chan Event) error

	// Stop stops the services and closes any necessary channels and connections
	Stop()

	// Subscribe takes in an event message and validates it for Process
	Subscribe() []string

	// Process takes in an event message and tries to process it
	Process(Event) error
}
```

You can create new events

```go
func (x *ExamplePlugin1) Start(e chan transistor.Event) error {
	log.Info("starting ExamplePlugin")

	event := Hello{
		Action:  "examplePlugin1",
		Message: "Hello World from ExamplePlugin1",
	}

	e <- transistor.NewEvent(transistor.EventName("exampleplugin1"), transistor.Action("create"), nil)

	return nil
}
```

or respond to existing one and keep track of parent event

```go
func (x *ExamplePlugin2) Process(e transistor.Event) error {
	if e.Name == "exampleplugin2:create" {
		hello := e.Payload().(Hello)
		log.Info("ExamplePlugin2 received a message:", hello)
	}
	return nil
}
```

Transistor can run on a multiple or single host. To run on multiple hosts you will need a Redis connection. This is a minimal example to set up 2 plugins:

```go
func main() {
	config := transistor.Config{
		Server:   "0.0.0.0:16379",
		Database: "0",
		Pool:     "30",
		Process:  "1",
		Queueing: true,
		Plugins: map[string]interface{}{
			"exampleplugin1": map[string]interface{}{
				"hello":   "world1",
				"workers": 1,
			},
			"exampleplugin2": map[string]interface{}{
				"hello":   "world2",
				"workers": 1,
			},
		},
		EnabledPlugins: []string{"exampleplugin1", "exampleplugin2"},
	}

	t, err := transistor.NewTransistor(config)
	if err != nil {
		log.Fatal(err)
	}

	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-signals
		if sig == os.Interrupt || sig == syscall.SIGTERM {
			log.Info("Shutting down circuit. SIGTERM recieved!\n")
			// If Queueing is ON then workers are responsible for closing Shutdown chan
			if !t.Config.Queueing {
				t.Stop()
			}
		}
	}()

	log.InfoWithFields("plugins loaded", log.Fields{
		"plugins": strings.Join(t.PluginNames(), ","),
	})

	t.Run()
}
```

if you want to run on a single host and without redis you need to set `Queueing: false` in config. You can see and run a minimal example that uses Redis in `example/` folder.

Transistor was build to power Checkr's deployment pipeline and it's used to build and deploy over 100 microservices to kubernetes.

[GoDoc](https://godoc.org/github.com/codeamp/transistor)
